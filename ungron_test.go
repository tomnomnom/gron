package main

import "testing"

func TestLex(t *testing.T) {
	cases := []struct {
		in   string
		want []token
	}{
		{`json.foo = 1;`, []token{
			{`json`, typBare},
			{`foo`, typBare},
			{`1`, typValue},
		}},

		{`json.foo = "bar";`, []token{
			{`json`, typBare},
			{`foo`, typBare},
			{`"bar"`, typValue},
		}},

		{`json[0] = "bar";`, []token{
			{`json`, typBare},
			{`0`, typNumeric},
			{`"bar"`, typValue},
		}},

		{`json["foo"] = "bar";`, []token{
			{`json`, typBare},
			{`"foo"`, typQuoted},
			{`"bar"`, typValue},
		}},

		{`json.foo["bar"][0] = "bar";`, []token{
			{`json`, typBare},
			{`foo`, typBare},
			{`"bar"`, typQuoted},
			{`0`, typNumeric},
			{`"bar"`, typValue},
		}},

		{`not an identifier at all`, []token{
			{`not`, typBare},
		}},

		{`alsonotanidentifier`, []token{
			{`alsonotanidentifier`, typBare},
		}},

		{`wat!`, []token{
			{`wat`, typBare},
		}},
	}

	for _, c := range cases {
		l := newLexer(c.in)
		have := l.lex()

		if len(have) != len(c.want) {
			t.Logf("Want: %#v", c.want)
			t.Logf("Have: %#v", have)
			t.Fatalf("want %d tokens, have %d", len(c.want), len(have))
		}

		for i := range have {
			if have[i] != c.want[i] {
				t.Logf("Want: %#v", c.want)
				t.Logf("Have: %#v", have)
				t.Errorf("Want `%#v` in position %d, have `%#v`", c.want[i], i, have[i])
			}
		}
	}
}
