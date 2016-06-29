package main

import (
	"errors"
	"unicode"
	"unicode/utf8"
)

var (
	errEmptyStatement = errors.New("statement is empty")
	errIdentEnd       = errors.New("end of identifiers")
	errBadIdentStart  = errors.New("invalid first character for identifier")
	errBadBareIdent   = errors.New("failed to parse bare identifier")
	errBadArrayKey    = errors.New("failed to parse array key")
	errNonNumArrayKey = errors.New("non-number found in array key")
	errBadQuotedKey   = errors.New("failed to parse quoted key")
)

func nextIdentifier(s string) (string, error) {

	if s == "" {
		return "", errEmptyStatement
	}

	first, _ := utf8.DecodeRuneInString(s)

	// If we've got a space or equals there's no
	// more identifiers left
	if first == ' ' || first == '=' {
		return "", errIdentEnd
	}

	// The next identifier might be 'bare'
	// e.g. the 'foo' in 'foo.bar' is bare
	if validFirstRune(first) || first == '.' {
		return nextBareIdentifier(s)
	}

	// If the identifier isn't bare, it's an error for it
	// to not start with an opening square brace
	if first != '[' {
		return "", errBadIdentStart
	}

	// The rune after the opening square brace
	// determines if it's a map or array key
	second, _ := utf8.DecodeRuneInString(s[1:])

	if unicode.IsNumber(second) {
		return nextArrayKey(s)
	}

	return nextQuotedKey(s)
}

func nextBareIdentifier(s string) (string, error) {
	for i, r := range s {
		if r == '.' || r == '[' || r == ' ' {
			return s[:i], nil
		}
	}
	return "", errBadBareIdent
}

func nextArrayKey(s string) (string, error) {
	// The first char is an opening brace so ignore it
	for i, r := range s[1:] {
		if r == ']' {
			return s[1 : i+1], nil
		}
		if !unicode.IsNumber(r) {
			return "", errNonNumArrayKey
		}
	}
	return "", errBadArrayKey
}

func nextQuotedKey(s string) (string, error) {
	inEscape := false
	if s[1] != '"' {
		return "", errBadQuotedKey
	}

	for i, r := range s[2:] {
		if !inEscape && r == '"' {
			return s[2 : i+2], nil
		}

		inEscape = false
		if r == '\\' {
			inEscape = true
		}
	}
	return "", errBadQuotedKey
}
