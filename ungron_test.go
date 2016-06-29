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

func TestUngronTokens(t *testing.T) {
	in := `json.contact["e-mail"][0] = "mail@tomnomnom.com";`
	l := newLexer(in)
	tokens := l.lex()

	have, err := ungronTokens(tokens)
	if err != nil {
		t.Fatalf("failed to ungron statement: %s", err)
	}

	t.Logf("Have: %#v", have)

	top, ok := have.(map[string]interface{})
	if !ok {
		t.Fatalf("failed to convert top level to map[string]interface{}")
	}

	rawJ, ok := top["json"]
	if !ok {
		t.Fatalf("top level should have key 'json' but doesn't")
	}

	j, ok := rawJ.(map[string]interface{})
	if !ok {
		t.Fatalf("failed to convert json level to map[string]interface{}")
	}

	rawContact, ok := j["contact"]
	if !ok {
		t.Fatalf("json should have key 'contact' but doesn't")
	}

	contact, ok := rawContact.(map[string]interface{})
	if !ok {
		t.Fatalf("failed to convert contact to map[string]interface{}")
	}

	rawEmail, ok := contact["e-mail"]
	if !ok {
		t.Fatalf("contact should have key 'e-mail' but doesn't")
	}

	email, ok := rawEmail.([]interface{})
	if !ok {
		t.Fatalf("failed to convert email to []interface{}")
	}

	if len(email) != 1 {
		t.Fatalf("want length 1 for email but have %d", len(email))
	}

	addr, ok := email[0].(string)
	if !ok {
		t.Fatalf("failed to convert email address to string")
	}

	if addr != "mail@tomnomnom.com" {
		t.Fatalf("Want `mail@tomnomnom.com`; have `%s`", addr)
	}
}
