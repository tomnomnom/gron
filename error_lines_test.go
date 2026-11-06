package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestUngronErrorLineNumbers(t *testing.T) {
	for _, tc := range []struct {
		name, input, want string
		opts              int
	}{
		{"first line", `"test";`, "line 1", optMonochrome},
		{"after blank and separator", "json = {};\n\n--\njson.bad = nope;\n", "line 4", optMonochrome},
		{"JSON layout", "[[],{}]\nnot JSON\n", "line 2", optMonochrome | optJSON},
		{"merge mismatch", "json = {};\n\njson = [];\n", "line 3", optMonochrome},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := &bytes.Buffer{}
			code, err := ungron(strings.NewReader(tc.input), out, tc.opts)
			if code != exitParseStatements || err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("code = %d, error = %v; want parse error containing %q", code, err, tc.want)
			}
			if out.Len() != 0 {
				t.Errorf("error wrote partial output: %q", out.String())
			}
		})
	}
}

func TestUngronLineNumbersPreserveRecoverableInput(t *testing.T) {
	var out bytes.Buffer
	code, err := ungron(strings.NewReader("\n--\njson = {};\njson.ok = true;\n"), &out, optMonochrome)
	if code != exitOK || err != nil || !strings.Contains(out.String(), `"ok": true`) {
		t.Fatalf("code = %d, error = %v, output = %q", code, err, out.String())
	}
}
