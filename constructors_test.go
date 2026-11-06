package main

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestUngronEmptyConstructors(t *testing.T) {
	for _, tc := range []struct{ value, want string }{
		{"new Array", "[]"}, {"new Array()", "[]"}, {"new Array ( )", "[]"},
		{"new Object", "{}"}, {"new Object()", "{}"}, {"new\tObject\t(\t)", "{}"},
		{`"new Array()"`, `"new Array()"`}, {"null", "null"},
	} {
		t.Run(tc.value, func(t *testing.T) {
			var out bytes.Buffer
			code, err := ungron(strings.NewReader("json = "+tc.value+";\n"), &out, optMonochrome)
			if code != exitOK || err != nil {
				t.Fatalf("code = %d, error = %v", code, err)
			}
			var got, want interface{}
			if err := json.Unmarshal(out.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(tc.want), &want); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("got %#v, want %#v", got, want)
			}
		})
	}
}

func TestUngronRejectsNonemptyConstructors(t *testing.T) {
	for _, value := range []string{"new Array(2)", "new Array(1, 2)", "new Object(1)", "new Date()", "new ArrayExtra", "newArray()", "new Ar ray()", "new Array(}", "new Object()junk"} {
		t.Run(value, func(t *testing.T) {
			var out bytes.Buffer
			code, err := ungron(strings.NewReader("json = "+value+";\n"), &out, optMonochrome)
			if code == exitOK || err == nil {
				t.Fatalf("accepted unsupported constructor %q: %q", value, out.String())
			}
		})
	}
}
