package main

import (
	"encoding/json"
	"testing"
)

func TestValueTokenFromInterface(t *testing.T) {
	cases := []struct {
		in   interface{}
		want token
	}{
		{make(map[string]interface{}), token{"{}", typEmptyObject}},
		{make([]interface{}, 0), token{"[]", typEmptyArray}},
		{json.Number("1.2"), token{"1.2", typNumber}},
		{"foo", token{`"foo"`, typString}},
		{"<3", token{`"<3"`, typString}},
		{true, token{"true", typTrue}},
		{false, token{"false", typFalse}},
		{nil, token{"null", typNull}},
		{struct{}{}, token{"", typError}},
	}

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
