package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestGron(t *testing.T) {
	cases := []struct {
		inFile  string
		outFile string
	}{
		{"testdata/one.json", "testdata/one.gron"},
		{"testdata/two.json", "testdata/two.gron"},
		{"testdata/three.json", "testdata/three.gron"},
		{"testdata/github.json", "testdata/github.gron"},
	}

	for _, c := range cases {
		in, err := os.Open(c.inFile)
		if err != nil {
			t.Fatalf("failed to open input file: %s", err)
		}

		want, err := ioutil.ReadFile(c.outFile)
		if err != nil {
			t.Fatalf("failed to open want file: %s", err)
		}

		out := &bytes.Buffer{}
		code, err := gron(in, out, optMonochrome)

		if code != exitOK {
			t.Errorf("want exitOK; have %d", code)
		}
		if err != nil {
			t.Errorf("want nil error; have %s", err)
		}

		if !reflect.DeepEqual(want, out.Bytes()) {
			t.Logf("want: %s", want)
			t.Logf("have: %s", out.Bytes())
			t.Errorf("gronned %s does not match %s", c.inFile, c.outFile)
		}
	}

}

func TestPrefixedFilenameWriter(t *testing.T) {
	prefixes := []string{
		"testdata/one.json: ",
		"testdata/two.json: ",
		"testdata/three.json: ",
		"testdata/github.json: ",
	}
	lines := []string{
		"Line 1",
		"Another line 2",
		"a 3rd Line",
	}

	for _, p := range prefixes {
		wrappedWriter := &bytes.Buffer{}
		writer := NewPrefixedWriter(wrappedWriter, p)
		for _, l := range lines {
			fmt.Fprintln(writer, l)
		}
		s := bufio.NewScanner(wrappedWriter)
		var have []string
		for s.Scan() {
			have = append(have, s.Text())
		}
		if len(have) != len(lines) {
			t.Errorf("want %d lines, have %d lines", len(lines), len(have))
		}
		for i, l := range lines {
			want := p + l
			if have[i] != want {
				t.Errorf("line %d\nwant: %s\nhave: %s", i, want, have[i])
			}
		}
	}
}

func TestGronStream(t *testing.T) {
	cases := []struct {
		inFile  string
		outFile string
	}{
		{"testdata/stream.json", "testdata/stream.gron"},
		{"testdata/scalar-stream.json", "testdata/scalar-stream.gron"},
	}

	for _, c := range cases {
		in, err := os.Open(c.inFile)
		if err != nil {
			t.Fatalf("failed to open input file: %s", err)
		}

		want, err := ioutil.ReadFile(c.outFile)
		if err != nil {
			t.Fatalf("failed to open want file: %s", err)
		}

		out := &bytes.Buffer{}
		code, err := gronStream(in, out, optMonochrome)

		if code != exitOK {
			t.Errorf("want exitOK; have %d", code)
		}
		if err != nil {
			t.Errorf("want nil error; have %s", err)
		}

		if !reflect.DeepEqual(want, out.Bytes()) {
			t.Logf("want: %s", want)
			t.Logf("have: %s", out.Bytes())
			t.Errorf("gronned %s does not match %s", c.inFile, c.outFile)
		}
	}

}

func TestLargeGronStream(t *testing.T) {
	cases := []struct {
		inFile  string
		outFile string
	}{
		{"testdata/long-stream.json", "testdata/long-stream.gron"},
	}

	for _, c := range cases {
		in, err := os.Open(c.inFile)
		if err != nil {
			t.Fatalf("failed to open input file: %s", err)
		}

		want, err := ioutil.ReadFile(c.outFile)
		if err != nil {
			t.Fatalf("failed to open want file: %s", err)
		}

		out := &bytes.Buffer{}
		code, err := gronStream(in, out, optMonochrome)

		if code != exitOK {
			t.Errorf("want exitOK; have %d", code)
		}
		if err != nil {
			t.Errorf("want nil error; have %s", err)
		}

		if !reflect.DeepEqual(want, out.Bytes()) {
			t.Logf("want: %s", want)
			t.Logf("have: %s", out.Bytes())
			t.Errorf("gronned %s does not match %s", c.inFile, c.outFile)
		}
	}

}

func TestUngron(t *testing.T) {
	cases := []struct {
		inFile  string
		outFile string
	}{
		{"testdata/one.gron", "testdata/one.json"},
		{"testdata/two.gron", "testdata/two.json"},
		{"testdata/three.gron", "testdata/three.json"},
		{"testdata/grep-separators.gron", "testdata/grep-separators.json"},
		{"testdata/github.gron", "testdata/github.json"},
		{"testdata/large-line.gron", "testdata/large-line.json"},
	}

	for _, c := range cases {
		wantF, err := ioutil.ReadFile(c.outFile)
		if err != nil {
			t.Fatalf("failed to open want file: %s", err)
		}

		var want interface{}
		err = json.Unmarshal(wantF, &want)
		if err != nil {
			t.Fatalf("failed to unmarshal JSON from want file: %s", err)
		}

		in, err := os.Open(c.inFile)
		if err != nil {
			t.Fatalf("failed to open input file: %s", err)
		}

		out := &bytes.Buffer{}
		code, err := ungron(in, out, optMonochrome)

		if code != exitOK {
			t.Errorf("want exitOK; have %d", code)
		}
		if err != nil {
			t.Errorf("want nil error; have %s", err)
		}

		var have interface{}
		err = json.Unmarshal(out.Bytes(), &have)
		if err != nil {
			t.Fatalf("failed to unmarshal JSON from ungron output: %s", err)
		}

		if !reflect.DeepEqual(want, have) {
			t.Logf("want: %#v", want)
			t.Logf("have: %#v", have)
			t.Errorf("ungronned %s does not match %s", c.inFile, c.outFile)
		}

	}
}

func TestGronJ(t *testing.T) {
	cases := []struct {
		inFile  string
		outFile string
	}{
		{"testdata/one.json", "testdata/one.jgron"},
		{"testdata/two.json", "testdata/two.jgron"},
		{"testdata/three.json", "testdata/three.jgron"},
		{"testdata/github.json", "testdata/github.jgron"},
	}

	for _, c := range cases {
		in, err := os.Open(c.inFile)
		if err != nil {
			t.Fatalf("failed to open input file: %s", err)
		}

		want, err := ioutil.ReadFile(c.outFile)
		if err != nil {
			t.Fatalf("failed to open want file: %s", err)
		}

		out := &bytes.Buffer{}
		code, err := gron(in, out, optMonochrome|optJSON)

		if code != exitOK {
			t.Errorf("want exitOK; have %d", code)
		}
		if err != nil {
			t.Errorf("want nil error; have %s", err)
		}

		if !reflect.DeepEqual(want, out.Bytes()) {
			t.Logf("want: %s", want)
			t.Logf("have: %s", out.Bytes())
			t.Errorf("gronned %s does not match %s", c.inFile, c.outFile)
		}
	}

}

func TestGronStreamJ(t *testing.T) {
	cases := []struct {
		inFile  string
		outFile string
	}{
		{"testdata/stream.json", "testdata/stream.jgron"},
		{"testdata/scalar-stream.json", "testdata/scalar-stream.jgron"},
	}

	for _, c := range cases {
		in, err := os.Open(c.inFile)
		if err != nil {
			t.Fatalf("failed to open input file: %s", err)
		}

		want, err := ioutil.ReadFile(c.outFile)
		if err != nil {
			t.Fatalf("failed to open want file: %s", err)
		}

		out := &bytes.Buffer{}
		code, err := gronStream(in, out, optMonochrome|optJSON)

		if code != exitOK {
			t.Errorf("want exitOK; have %d", code)
		}
		if err != nil {
			t.Errorf("want nil error; have %s", err)
		}

		if !reflect.DeepEqual(want, out.Bytes()) {
			t.Logf("want: %s", want)
			t.Logf("have: %s", out.Bytes())
			t.Errorf("gronned %s does not match %s", c.inFile, c.outFile)
		}
	}

}

func TestUngronJ(t *testing.T) {
	cases := []struct {
		inFile  string
		outFile string
	}{
		{"testdata/one.jgron", "testdata/one.json"},
		{"testdata/two.jgron", "testdata/two.json"},
		{"testdata/three.jgron", "testdata/three.json"},
		{"testdata/github.jgron", "testdata/github.json"},
	}

	for _, c := range cases {
		wantF, err := ioutil.ReadFile(c.outFile)
		if err != nil {
			t.Fatalf("failed to open want file: %s", err)
		}

		var want interface{}
		err = json.Unmarshal(wantF, &want)
		if err != nil {
			t.Fatalf("failed to unmarshal JSON from want file: %s", err)
		}

		in, err := os.Open(c.inFile)
		if err != nil {
			t.Fatalf("failed to open input file: %s", err)
		}

		out := &bytes.Buffer{}
		code, err := ungron(in, out, optMonochrome|optJSON)

		if code != exitOK {
			t.Errorf("want exitOK; have %d", code)
		}
		if err != nil {
			t.Errorf("want nil error; have %s", err)
		}

		var have interface{}
		err = json.Unmarshal(out.Bytes(), &have)
		if err != nil {
			t.Fatalf("failed to unmarshal JSON from ungron output: %s", err)
		}

		if !reflect.DeepEqual(want, have) {
			t.Logf("want: %#v", want)
			t.Logf("have: %#v", have)
			t.Errorf("ungronned %s does not match %s", c.inFile, c.outFile)
		}

	}
}

func BenchmarkBigJSON(b *testing.B) {
	in, err := os.Open("testdata/big.json")
	if err != nil {
		b.Fatalf("failed to open test data file: %s", err)
	}

	for i := 0; i < b.N; i++ {
		out := &bytes.Buffer{}
		_, err = in.Seek(0, 0)
		if err != nil {
			b.Fatalf("failed to rewind input: %s", err)
		}

		_, err := gron(in, out, optMonochrome|optNoSort)
		if err != nil {
			b.Fatalf("failed to gron: %s", err)
		}
	}
}
