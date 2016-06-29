package main

import "testing"

func TestNextIdentifier(t *testing.T) {

	cases := []struct {
		in   string
		want string
		err  error
	}{
		// Generic error conditions
		{"", "", errEmptyStatement},
		{" = 1;", "", errIdentEnd},
		{"= 1;", "", errIdentEnd},
		{"!= 1;", "", errBadIdentStart},

		// Bare identifiers
		{`json`, ``, errBadBareIdent},
		{`json.foo = 1;`, `json`, nil},
		{`json["foo"] = 1;`, `json`, nil},
		{`json[0] = 1;`, `json`, nil},

		// Array keys
		{`[1foo] = 1;`, ``, errNonNumArrayKey},
		{`[0`, ``, errBadArrayKey},
		{`[0] = 1;`, `0`, nil},

		// Quoted keys
		{`["foo] = 1;`, ``, errBadQuotedKey},
		{`['foo] = 1;`, ``, errBadQuotedKey},
		{`[foo] = 1;`, ``, errBadQuotedKey},
		{`[foo"] = 1;`, ``, errBadQuotedKey},
		{`[foo'] = 1;`, ``, errBadQuotedKey},
		{`['foo'] = 1;`, ``, errBadQuotedKey},
		{`["foo"] = 1;`, `foo`, nil},
		{`["f\"oo"] = 1;`, `f\"oo`, nil},
		{`["f\"oo"] = 1;`, `f\"oo`, nil},
	}

	for _, c := range cases {
		have, err := nextIdentifier(c.in)

		if have != c.want {
			t.Errorf("Want `%s` but have `%s` for nextIdentifier(%s)", c.want, have, c.in)
		}

		if err != c.err {
			t.Errorf("Want error `%s` but have `%s` for nextIdentifier(%s)", c.err, err, c.in)
		}
	}
}
