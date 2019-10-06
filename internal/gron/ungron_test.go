package gron

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
			{`.`, typDot},
			{`foo`, typBare},
			{`=`, typEquals},
			{`1`, typNumber},
			{`;`, typSemi},
		}},

		{`json.foo = "bar";`, []token{
			{`json`, typBare},
			{`.`, typDot},
			{`foo`, typBare},
			{`=`, typEquals},
			{`"bar"`, typString},
			{`;`, typSemi},
		}},

		{`json.foo = "ba;r";`, []token{
			{`json`, typBare},
			{`.`, typDot},
			{`foo`, typBare},
			{`=`, typEquals},
			{`"ba;r"`, typString},
			{`;`, typSemi},
		}},

		{`json.foo = "ba\"r ;";`, []token{
			{`json`, typBare},
			{`.`, typDot},
			{`foo`, typBare},
			{`=`, typEquals},
			{`"ba\"r ;"`, typString},
			{`;`, typSemi},
		}},

		{`json = "\\";`, []token{
			{`json`, typBare},
			{`=`, typEquals},
			{`"\\"`, typString},
			{`;`, typSemi},
		}},

		{`json = "\\\\";`, []token{
			{`json`, typBare},
			{`=`, typEquals},
			{`"\\\\"`, typString},
			{`;`, typSemi},
		}},

		{`json = "f\oo\\";`, []token{
			{`json`, typBare},
			{`=`, typEquals},
			{`"f\oo\\"`, typString},
			{`;`, typSemi},
		}},

		{`json.value = "\u003c ;";`, []token{
			{`json`, typBare},
			{`.`, typDot},
			{`value`, typBare},
			{`=`, typEquals},
			{`"\u003c ;"`, typString},
			{`;`, typSemi},
		}},

		{`json[0] = "bar";`, []token{
			{`json`, typBare},
			{`[`, typLBrace},
			{`0`, typNumericKey},
			{`]`, typRBrace},
			{`=`, typEquals},
			{`"bar"`, typString},
			{`;`, typSemi},
		}},

		{`json["foo"] = "bar";`, []token{
			{`json`, typBare},
			{`[`, typLBrace},
			{`"foo"`, typQuotedKey},
			{`]`, typRBrace},
			{`=`, typEquals},
			{`"bar"`, typString},
			{`;`, typSemi},
		}},

		{`json.foo["bar"][0] = "bar";`, []token{
			{`json`, typBare},
			{`.`, typDot},
			{`foo`, typBare},
			{`[`, typLBrace},
			{`"bar"`, typQuotedKey},
			{`]`, typRBrace},
			{`[`, typLBrace},
			{`0`, typNumericKey},
			{`]`, typRBrace},
			{`=`, typEquals},
			{`"bar"`, typString},
			{`;`, typSemi},
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
			{`[`, typLBrace},
			{``, typError},
		}},

		{`json.[2] = 1;`, []token{
			{`json`, typBare},
			{`.`, typDot},
			{``, typError},
		}},

		{`json[1 = 1;`, []token{
			{`json`, typBare},
			{`[`, typLBrace},
			{`1`, typNumericKey},
			{``, typError},
		}},

		{`json["foo] = 1;`, []token{
			{`json`, typBare},
			{`[`, typLBrace},
			{`"foo] = 1;`, typQuotedKey},
			{``, typError},
		}},

		{`--`, []token{
			{`--`, typIgnored},
		}},

		{`json  =  1;`, []token{
			{`json`, typBare},
			{`=`, typEquals},
			{`1`, typNumber},
			{`;`, typSemi},
		}},

		{`json=1;`, []token{
			{`json`, typBare},
			{`=`, typEquals},
			{`1`, typNumber},
			{`;`, typSemi},
		}},
	}

	for _, c := range cases {
		l := newLexer(c.in)
		have := l.lex()

		if len(have) != len(c.want) {
			t.Logf("Input: %#v", c.in)
			t.Logf("Want: %#v", c.want)
			t.Logf("Have: %#v", have)
			t.Fatalf("want %d tokens, have %d", len(c.want), len(have))
		}

		for i := range have {
			if have[i] != c.want[i] {
				t.Logf("Input: %#v", c.in)
				t.Logf("Want: %#v", c.want)
				t.Logf("Have: %#v", have)
				t.Errorf("Want `%#v` in position %d, have `%#v`", c.want[i], i, have[i])
			}
		}
	}
}

func TestUngronTokensSimple(t *testing.T) {
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

func TestUngronTokensInvalid(t *testing.T) {
	cases := []struct {
		in []token
	}{
		{[]token{{``, typError}}},                           // Error token
		{[]token{{`foo`, typString}}},                       // Invalid value
		{[]token{{`"foo`, typQuotedKey}, {"1", typNumber}}}, // Invalid quoted key
		{[]token{{`foo`, typNumericKey}, {"1", typNumber}}}, // Invalid numeric key
		{[]token{{``, -255}, {"1", typNumber}}},             // Invalid token type
	}

	for _, c := range cases {
		_, err := ungronTokens(c.in)
		if err == nil {
			t.Errorf("want non-nil error for %#v; have nil", c.in)
		}
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
