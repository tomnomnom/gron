package gron

import (
	"encoding/json"
	"testing"
)

var cases = []struct {
	in   interface{}
	want token
}{
	{make(map[string]interface{}), token{"{}", typEmptyObject}},
	{make([]interface{}, 0), token{"[]", typEmptyArray}},
	{json.Number("1.2"), token{"1.2", typNumber}},
	{"foo", token{`"foo"`, typString}},
	{"<3", token{`"<3"`, typString}},
	{"&", token{`"&"`, typString}},
	{"\b", token{`"\b"`, typString}},
	{"\f", token{`"\f"`, typString}},
	{"\n", token{`"\n"`, typString}},
	{"\r", token{`"\r"`, typString}},
	{"\t", token{`"\t"`, typString}},
	{"wat \u001e", token{`"wat \u001E"`, typString}},
	{"Hello, 世界", token{`"Hello, 世界"`, typString}},
	{true, token{"true", typTrue}},
	{false, token{"false", typFalse}},
	{nil, token{"null", typNull}},
	{struct{}{}, token{"", typError}},
}

func TestValueTokenFromInterface(t *testing.T) {

	for _, c := range cases {
		have := valueTokenFromInterface(c.in)

		if have != c.want {
			t.Logf("input: %#v", have)
			t.Logf("have: %#v", have)
			t.Logf("want: %#v", c.want)
			t.Errorf("have != want")
		}
	}
}

func BenchmarkValueTokenFromInterface(b *testing.B) {

	for i := 0; i < b.N; i++ {
		for _, c := range cases {
			_ = valueTokenFromInterface(c.in)
		}
	}
}
