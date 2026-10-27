package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
	"testing/iotest"
)

func TestStatementsFromJSONLeadingBOM(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"object", "\xEF\xBB\xBF{\"value\":1}", []string{`json = {};`, `json.value = 1;`}},
		{"array", "\xEF\xBB\xBF[true]", []string{`json = [];`, `json[0] = true;`}},
		{"number", "\xEF\xBB\xBF9007199254740993", []string{`json = 9007199254740993;`}},
		{"null", "\xEF\xBB\xBFnull", []string{`json = null;`}},
		{"whitespace after BOM", "\xEF\xBB\xBF \r\n{}", []string{`json = {};`}},
		{"BOM in string", "\xEF\xBB\xBF\"a\uFEFFb\"", []string{"json = \"a\uFEFFb\";"}},
		{"escaped BOM in string", "\xEF\xBB\xBF\"\\uFEFF\"", []string{"json = \"\uFEFF\";"}},
		{"interior BOM without prefix", "{\"\uFEFF\":\"\uFEFF\"}", []string{`json = {};`, "json[\"\uFEFF\"] = \"\uFEFF\";"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, reader := range []struct {
				name string
				r    io.Reader
			}{
				{"reader", strings.NewReader(test.input)},
				{"one byte", iotest.OneByteReader(strings.NewReader(test.input))},
			} {
				t.Run(reader.name, func(t *testing.T) {
					ss, err := statementsFromJSON(reader.r, statement{{"json", typBare}})
					if err != nil {
						t.Fatal(err)
					}
					if len(ss) != len(test.want) {
						t.Fatalf("want %d statements, got %d", len(test.want), len(ss))
					}
					for _, want := range statementsFromStringSlice(test.want) {
						if !ss.Contains(want) {
							t.Errorf("missing %s", want)
						}
					}
				})
			}
		})
	}
}

func TestStatementsFromJSONInvalidBOM(t *testing.T) {
	for _, test := range []struct{ name, input string }{
		{"empty", ""},
		{"first byte only", "\xEF"},
		{"first two bytes only", "\xEF\xBB"},
		{"wrong second byte", "\xEF\x00\xBF{}"},
		{"wrong third byte", "\xEF\xBB\x00{}"},
		{"UTF16 BOM", "\xFF\xFE{}"},
		{"BOM only", "\xEF\xBB\xBF"},
		{"invalid JSON after BOM", "\xEF\xBB\xBF{"},
		{"repeated BOM", "\xEF\xBB\xBF\xEF\xBB\xBF{}"},
		{"whitespace before BOM", " \xEF\xBB\xBF{}"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := statementsFromJSON(iotest.OneByteReader(strings.NewReader(test.input)), statement{{"json", typBare}})
			if err == nil {
				t.Errorf("invalid input %q was accepted", test.input)
			}
		})
	}
}

func TestStatementsFromJSONNonBOMPrefixUnchanged(t *testing.T) {
	for _, input := range []string{"", "\xEF", "\xEF\xBB", "\xEF\x00\xBF{}", "\xEF\xBB\x00{}", "\xFF\xFE{}", " \xEF\xBB\xBF{}"} {
		var value interface{}
		want := json.NewDecoder(strings.NewReader(input)).Decode(&value)
		_, got := statementsFromJSON(iotest.OneByteReader(strings.NewReader(input)), statement{{"json", typBare}})
		if want == nil || got == nil || got.Error() != want.Error() {
			t.Errorf("input %q: want unchanged decoder error %v, got %v", input, want, got)
		}
	}
}

func TestStatementsFromJSONBOMReaderError(t *testing.T) {
	want := errors.New("input failed")
	for _, prefix := range []string{"", "\xEF\xBB\xBF"} {
		r := io.MultiReader(strings.NewReader(prefix), iotest.ErrReader(want))
		_, err := statementsFromJSON(r, statement{{"json", typBare}})
		if err != want {
			t.Errorf("prefix %q: want reader error %v, got %v", prefix, want, err)
		}
	}
}

type bomDocumentReader struct {
	input       string
	readPastEnd bool
}

func (r *bomDocumentReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if len(r.input) == 0 {
		r.readPastEnd = true
		return 0, io.EOF
	}
	p[0] = r.input[0]
	r.input = r.input[1:]
	return 1, nil
}

func TestStatementsFromJSONBOMDoesNotReadPastDocument(t *testing.T) {
	for _, input := range []string{"{}", "\xEF\xBB\xBF{}"} {
		r := &bomDocumentReader{input: input}
		if _, err := statementsFromJSON(r, statement{{"json", typBare}}); err != nil {
			t.Fatal(err)
		}
		if r.readPastEnd {
			t.Errorf("input %q: read beyond the complete document", input)
		}
	}
}

func TestGronStreamLeadingBOM(t *testing.T) {
	input := "\xEF\xBB\xBF{\"value\":1}\n\xEF\xBB\xBF{\"value\":2}\n"
	var out bytes.Buffer
	code, err := gronStream(strings.NewReader(input), &out, optMonochrome)
	if code != exitOK || err != nil {
		t.Fatalf("want success, got exit %d, error %v", code, err)
	}
	want := "json = [];\njson[0] = {};\njson[0].value = 1;\njson[1] = {};\njson[1].value = 2;\n"
	if out.String() != want {
		t.Errorf("want %q, got %q", want, out.String())
	}
}
