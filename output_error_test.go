package main

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

type outputTestWriter struct {
	calls  int
	failAt int
	err    error
	short  bool
}

func (w *outputTestWriter) Write(p []byte) (int, error) {
	w.calls++
	if w.calls == w.failAt {
		if w.short {
			return len(p) - 1, nil
		}
		return 0, w.err
	}
	return len(p), nil
}

func TestOutputFailures(t *testing.T) {
	sentinel := errors.New("output unavailable")
	actions := []struct {
		name   string
		action actionFn
		input  string
	}{
		{"gron", gron, `{"a":1,"b":2}`},
		{"stream", gronStream, "{\"a\":1}\n{\"b\":2}\n"},
		{"ungron", ungron, "json.a = 1;\njson.b = 2;\n"},
	}
	for _, a := range actions {
		for _, opts := range []int{0, optMonochrome, optMonochrome | optNoSort, optMonochrome | optJSON} {
			if a.name == "ungron" && opts&optJSON != 0 {
				continue
			}
			for _, short := range []bool{false, true} {
				for _, failAt := range []int{1, 2, 3} {
					if a.name == "ungron" && failAt > 1 {
						continue
					}
					w := &outputTestWriter{failAt: failAt, err: sentinel, short: short}
					code, err := a.action(strings.NewReader(a.input), w, opts)
					want := sentinel
					if short {
						want = io.ErrShortWrite
					}
					if code != 7 || !errors.Is(err, want) {
						t.Errorf("%s opts=%d short=%v failAt=%d: code=%d err=%v, want 7 wrapping %v", a.name, opts, short, failAt, code, err, want)
					}
					if w.calls != failAt {
						t.Errorf("%s: wrote %d times after failure at %d", a.name, w.calls, failAt)
					}
				}
			}
		}
	}
}

type noReadOutputInput struct{ t *testing.T }

func (r noReadOutputInput) Read([]byte) (int, error) {
	r.t.Error("read input after initial output failed")
	return 0, io.EOF
}

func TestStreamStopsBeforeReadingWhenHeaderWriteFails(t *testing.T) {
	sentinel := errors.New("header failed")
	w := &outputTestWriter{failAt: 1, err: sentinel}
	code, err := gronStream(noReadOutputInput{t}, w, optMonochrome)
	if code != 7 || !errors.Is(err, sentinel) {
		t.Fatalf("code=%d err=%v", code, err)
	}
}

func TestOutputSuccessAndInputErrors(t *testing.T) {
	for _, a := range []struct {
		name        string
		action      actionFn
		input, want string
	}{
		{"gron", gron, `{"a":1}`, "json = {};\njson.a = 1;\n"},
		{"stream", gronStream, "1\n2\n", "json = [];\njson[0] = 1;\njson[1] = 2;\n"},
		{"ungron", ungron, "json.a = 1;\n", "{\n  \"a\": 1\n}\n"},
	} {
		var out bytes.Buffer
		code, err := a.action(strings.NewReader(a.input), &out, optMonochrome)
		if code != exitOK || err != nil || out.String() != a.want {
			t.Errorf("%s: code=%d err=%v out=%q", a.name, code, err, out.String())
		}
	}
	var out bytes.Buffer
	code, err := gron(strings.NewReader("{"), &out, optMonochrome)
	if code != exitFormStatements || err == nil || out.Len() != 0 {
		t.Errorf("input error changed: %d %v %q", code, err, out.String())
	}
}
