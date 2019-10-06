package gron

import "testing"

func TestValidIdentifier(t *testing.T) {
	tests := []struct {
		key  string
		want bool
	}{
		// Valid Identifiers
		{"dotted", true},
		{"dotted123", true},
		{"_under_scores", true},
		{"ಠ_ಠ", true},

		// Invalid chars
		{"is-quoted", false},
		{"Definitely quoted!", false},

		// Reserved words
		{"true", false},
		{"else", false},
		{"null", false},

		// Empty string
		{"", false},
	}

	for _, test := range tests {
		have := validIdentifier(test.key)
		if have != test.want {
			t.Errorf("Want %t for validIdentifier(%s); have %t", test.want, test.key, have)
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

func BenchmarkValidIdentifier(b *testing.B) {
	for i := 0; i < b.N; i++ {
		validIdentifier("must-be-quoted")
	}
}

func BenchmarkValidIdentifierUnquoted(b *testing.B) {
	for i := 0; i < b.N; i++ {
		validIdentifier("canbeunquoted")
	}
}

func BenchmarkValidIdentifierReserved(b *testing.B) {
	for i := 0; i < b.N; i++ {
		validIdentifier("function")
	}
}
