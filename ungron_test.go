package main

import (
	"reflect"
	"testing"
)

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

		{`json.foo = "ba;r";`, []token{
			{`json`, typBare},
			{`foo`, typBare},
			{`"ba;r"`, typValue},
		}},

		{`json.foo = "ba\"r ;";`, []token{
			{`json`, typBare},
			{`foo`, typBare},
			{`"ba\"r ;"`, typValue},
		}},

		{`json.value = "\u003c ;"`, []token{
			{`json`, typBare},
			{`value`, typBare},
			{`"\u003c ;"`, typValue},
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
			{``, typError},
		}},

		{`json[ = 1;`, []token{
			{`json`, typBare},
			{`[`, typError},
		}},

		{`json.[2] = 1;`, []token{
			{`json`, typBare},
			{``, typError},
		}},

		{`json[1 = 1;`, []token{
			{`json`, typBare},
			{`1`, typNumeric},
			{``, typError},
		}},

		{`json["foo] = 1;`, []token{
			{`json`, typBare},
			{`"foo] = 1;`, typQuoted},
			{``, typError},
		}},

		{`--`, []token{
			{`--`, typIgnored},
		}},

		{`json  =  1;`, []token{
			{`json`, typBare},
			{`1`, typValue},
		}},

		{`json=1;`, []token{
			{`json`, typBare},
			{`1`, typValue},
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

func TestUngronTokens(t *testing.T) {
	in := `json.contact["e-mail"][0] = "mail@tomnomnom.com";`
	want := map[string]interface{}{
		"json": map[string]interface{}{
			"contact": map[string]interface{}{
				"e-mail": []interface{}{
					"mail@tomnomnom.com",
				},
			},
		},
	}

	l := newLexer(in)
	tokens := l.lex()
	have, err := ungronTokens(tokens)

	if err != nil {
		t.Fatalf("failed to ungron statement: %s", err)
	}

	t.Logf("Have: %#v", have)
	t.Logf("Want: %#v", want)

	eq := reflect.DeepEqual(have, want)
	if !eq {
		t.Errorf("Have and want datastructures are unequal")
	}
}

func TestMerge(t *testing.T) {
	a := map[string]interface{}{
		"json": map[string]interface{}{
			"contact": map[string]interface{}{
				"e-mail": []interface{}{
					0: "mail@tomnomnom.com",
				},
			},
		},
	}

	b := map[string]interface{}{
		"json": map[string]interface{}{
			"contact": map[string]interface{}{
				"e-mail": []interface{}{
					1: "test@tomnomnom.com",
					3: "foo@tomnomnom.com",
				},
				"twitter": "@TomNomNom",
			},
		},
	}

	want := map[string]interface{}{
		"json": map[string]interface{}{
			"contact": map[string]interface{}{
				"e-mail": []interface{}{
					0: "mail@tomnomnom.com",
					1: "test@tomnomnom.com",
					3: "foo@tomnomnom.com",
				},
				"twitter": "@TomNomNom",
			},
		},
	}

	t.Logf("A: %#v", a)
	t.Logf("B: %#v", b)
	have, err := recursiveMerge(a, b)
	if err != nil {
		t.Fatalf("failed to merge datastructures: %s", err)
	}

	t.Logf("Have: %#v", have)
	t.Logf("Want: %#v", want)
	eq := reflect.DeepEqual(have, want)
	if !eq {
		t.Errorf("Have and want datastructures are unequal")
	}

}
