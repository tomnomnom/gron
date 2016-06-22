package main

import "unicode"

// The javascript reserved words cannot be used as unquoted keys
var reservedWords = []string{
	"break",
	"case",
	"catch",
	"class",
	"const",
	"continue",
	"debugger",
	"default",
	"delete",
	"do",
	"else",
	"export",
	"extends",
	"false",
	"finally",
	"for",
	"function",
	"if",
	"import",
	"in",
	"instanceof",
	"new",
	"null",
	"return",
	"super",
	"switch",
	"this",
	"throw",
	"true",
	"try",
	"typeof",
	"var",
	"void",
	"while",
	"with",
	"yield",
}

// validIdentifier checks to see if a string is a valid
// JavaScript identifier
// E.g:
//     justLettersAndNumbers1 -> true
//     a key with spaces      -> false
//     1startsWithANumber	  -> false
func validIdentifier(s string) bool {
	for i, r := range s {
		if i == 0 && !validFirstRune(r) {
			return false
		}
		if i != 0 && !validSecondaryRune(r) {
			return false
		}
	}

	// Check the list of reserved words
	for _, i := range reservedWords {
		if s == i {
			return false
		}
	}

	return true
}

// validFirstRune returns true for runes that are valid
// as the first rune in a key.
// E.g:
//     'r' -> true
//     '7' -> false
func validFirstRune(r rune) bool {
	return unicode.In(r,
		unicode.Lu,
		unicode.Ll,
		unicode.Lm,
		unicode.Lo,
		unicode.Nl,
	) || r == '$' || r == '_'
}

// validSecondaryRune returns true for runes that are valid
// as anything other than the first rune in a key.
func validSecondaryRune(r rune) bool {
	return validFirstRune(r) ||
		unicode.In(r, unicode.Mn, unicode.Mc, unicode.Nd, unicode.Pc)
}
