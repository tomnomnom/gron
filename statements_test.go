package main

import (
	"encoding/json"
	"testing"
)

func TestStatementsSimple(t *testing.T) {

	j := []byte(`{
		"dotted": "A dotted value",
		"a quoted": "value",
		"bool1": true,
		"bool2": false,
		"anull": null,
		"anarr": [1, 1.5],
		"anob": {
			"foo": "bar"
		},
		"else": 1
	}`)

	var top interface{}
	err := json.Unmarshal(j, &top)
	if err != nil {
		t.Errorf("Failed to unmarshal test file: %s", err)
	}

	ss, err := makeStatements("json", top)

	if err != nil {
		t.Errorf("Want nil error from makeStatements() but got %s", err)
	}

	wants := []string{
		`json = {};`,
		`json.dotted = "A dotted value";`,
		`json["a quoted"] = "value";`,
		`json.bool1 = true;`,
		`json.bool2 = false;`,
		`json.anull = null;`,
		`json.anarr = [];`,
		`json.anarr[0] = 1;`,
		`json.anarr[1] = 1.5;`,
		`json.anob = {};`,
		`json.anob.foo = "bar";`,
		`json["else"] = 1;`,
	}

	for _, want := range wants {
		if !ss.Contains(want) {
			t.Errorf("Statement group should contain `%s` but doesn't", want)
		}
	}

}

func TestPrefixHappy(t *testing.T) {
	tests := []struct {
		prev string
		next interface{}
		want string
	}{
		{"j", 123, "j[123]"},
		{"j", 1, "j[1]"},
		{"j", "dotted", "j.dotted"},
		{"j", "un-dotted", "j[\"un-dotted\"]"},
	}

	for _, test := range tests {
		r, err := makePrefix(test.prev, test.next)
		if err != nil {
			t.Errorf("Want nil error from makePrefix(%s, %#v); have: %s", test.prev, test.next, err)
		}
		if r != test.want {
			t.Errorf("Want %s from makePrefix(%s, %#v); have: %s", test.want, test.prev, test.next, r)
		}
	}
}

func TestKeyMustBeQuoted(t *testing.T) {
	tests := []struct {
		key  string
		want bool
	}{
		// Fine to be dotted
		{"dotted", false},
		{"dotted123", false},
		{"_under_scores", false},
		{"ಠ_ಠ", false},

		// Invalid chars
		{"is-quoted", true},
		{"Definitely quoted!", true},

		// Reserved words
		{"true", true},
		{"else", true},
		{"null", true},
	}

	for _, test := range tests {
		have := keyMustBeQuoted(test.key)
		if have != test.want {
			t.Errorf("Want %t for keyMustBeQuoted(%s); have %t", test.want, test.key, have)
		}
	}
}

func TestValidFirstRune(t *testing.T) {
	tests := []struct {
		in   rune
		want bool
	}{
		{'r', true},
		{'ಠ', true},
		{'4', false},
		{'-', false},
	}

	for _, test := range tests {
		have := validFirstRune(test.in)
		if have != test.want {
			t.Errorf("Want %t for validFirstRune(%#U); have %t", test.want, test.in, have)
		}
	}
}

func TestValidSecondaryRune(t *testing.T) {
	tests := []struct {
		in   rune
		want bool
	}{
		{'r', true},
		{'ಠ', true},
		{'4', true},
		{'-', false},
	}

	for _, test := range tests {
		have := validSecondaryRune(test.in)
		if have != test.want {
			t.Errorf("Want %t for validSecondaryRune(%#U); have %t", test.want, test.in, have)
		}
	}
}

func BenchmarkKeyMustBeQuoted(b *testing.B) {
	for i := 0; i < b.N; i++ {
		keyMustBeQuoted("must-be-quoted")
	}
}

func BenchmarkKeyMustBeQuotedUnquoted(b *testing.B) {
	for i := 0; i < b.N; i++ {
		keyMustBeQuoted("canbeunquoted")
	}
}

func BenchmarkKeyMustBeQuotedReserved(b *testing.B) {
	for i := 0; i < b.N; i++ {
		keyMustBeQuoted("function")
	}
}

func BenchmarkMakeStatements(b *testing.B) {
	j := []byte(`{
		"dotted": "A dotted value",
		"a quoted": "value",
		"bool1": true,
		"bool2": false,
		"anull": null,
		"anarr": [1, 1.5],
		"anob": {
			"foo": "bar"
		},
		"else": 1
	}`)

	var top interface{}
	err := json.Unmarshal(j, &top)
	if err != nil {
		b.Fatalf("Failed to unmarshal test file: %s", err)
	}

	for i := 0; i < b.N; i++ {
		_, _ = makeStatements("json", top)
	}
}

func BenchmarkMakePrefixUnquoted(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = makePrefix("json", "isunquoted")
	}
}

func BenchmarkMakePrefixQuoted(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = makePrefix("json", "this-is-quoted")
	}
}

func BenchmarkMakePrefixInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = makePrefix("json", 212)
	}
}
