package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestGronValues(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name: "scalar values",
			input: "json = {};\n" +
				"json.name = \"Ada\";\n" +
				"json.count = 42;\n" +
				"json.enabled = true;\n" +
				"json.disabled = false;\n" +
				"json.missing = null;\n" +
				"json.items = [];\n",
			want: "Ada\n42\ntrue\nfalse\nnull\n",
		},
		{
			name:  "decoded string",
			input: "json.text = \"héllo\\nworld\";\n",
			want:  "héllo\nworld\n",
		},
		{
			name:  "root scalar",
			input: "json = \"root\";\n",
			want:  "root\n",
		},
		{name: "empty input"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var out bytes.Buffer
			code, err := gronValues(strings.NewReader(c.input), &out, optMonochrome)
			if err != nil || code != exitOK {
				t.Fatalf("gronValues returned exit code %d and error %v", code, err)
			}
			if got := out.String(); got != c.want {
				t.Errorf("values = %q, want %q", got, c.want)
			}
		})
	}
}
