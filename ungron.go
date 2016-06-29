package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// A lexer holds the state for lexing statements
type lexer struct {
	text       string
	pos        int
	width      int
	cur        rune
	prev       rune
	tokens     []token
	tokenStart int
}

// A tokenTyp identifies what kind of token something is
type tokenTyp int

// Token types
const (
	typBare tokenTyp = iota
	typNumeric
	typQuoted
	typValue
)

// A token is a chunk of text from a statement with a type
type token struct {
	text string
	typ  tokenTyp
}

// next gets the next rune in the input and updates the lexer state
func (l *lexer) next() rune {
	r, w := utf8.DecodeRuneInString(l.text[l.pos:])

	l.pos += w
	l.width = w

	l.prev = l.cur
	l.cur = r

	return r
}

// backup moves the lexer back one rune
// can only be used once per call of next()
func (l *lexer) backup() {
	l.pos -= l.width
}

// peek returns the next rune in the input
// without moving the internal pointer
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// ignore skips the current token
func (l *lexer) ignore() {
	l.tokenStart = l.pos
}

// emit adds the current token to the token slice and
// moves the tokenStart pointer to the current position
func (l *lexer) emit(typ tokenTyp) {
	t := token{
		text: l.text[l.tokenStart:l.pos],
		typ:  typ,
	}
	l.tokenStart = l.pos

	l.tokens = append(l.tokens, t)
}

// lex runs the lexer and returns the lexed tokens
func (l *lexer) lex() []token {

	for next := lexStatement; next != nil; {
		next = next(l)
	}
	return l.tokens
}

// accept moves the pointer if the next rune is in
// the set of valid runes
func (l *lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

// acceptRun continually accepts runes from the
// set of valid runes
func (l *lexer) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
	}
	l.backup()
}

// a runeCheck is a function that determines if a rune is valid
// or not so that we can do complex checks against runes
type runeCheck func(rune) bool

// acceptFunc accepts a rune if the provided runeCheck
// function returns true
func (l *lexer) acceptFunc(fn runeCheck) {
	if fn(l.next()) {
		return
	}
	l.backup()
}

// acceptRunFunc continually accepts runes for as long
// as the runeCheck function returns true
func (l *lexer) acceptRunFunc(fn runeCheck) {
	for fn(l.next()) {
	}
	l.backup()
}

// acceptUntil accepts runes until it hits a delimiter
// rune contained in the provided string
func (l *lexer) acceptUntil(delims string) {
	for !strings.ContainsRune(delims, l.next()) {
		if l.cur == utf8.RuneError {
			return
		}
	}
	l.backup()
}

// acceptUntilUnescaped accepts runes until it hits a delimiter
// rune contained in the provided string, unless that rune was
// escaped with a backslash
func (l *lexer) acceptUntilUnescaped(delims string) {
	for !strings.ContainsRune(delims, l.next()) && l.prev != '\\' {
		if l.cur == utf8.RuneError {
			return
		}
	}
	l.backup()
}

// newLexer returns a new lexer for the provided string
func newLexer(text string) *lexer {
	return &lexer{
		text:       text,
		pos:        0,
		tokenStart: 0,
		tokens:     make([]token, 0),
	}
}

// a lexFn accepts a lexer, performs some action on it and
// then returns an appropriate lexFn for the next stage
type lexFn func(*lexer) lexFn

// lexStatement is the highest level lexFn. Its only job
// is to determine which more specific lexFn to use
func lexStatement(l *lexer) lexFn {
	r := l.peek()

	switch {
	case r == '.' || validFirstRune(r):
		return lexBareWord
	case r == '[':
		return lexBraces
	case r == ' ':
		return lexValue
	default:
		return nil
	}

}

// lexBareWord lexes for bare identifiers.
// E.g: the 'foo' in 'foo.bar' or 'foo[0]' is a bare identifier
func lexBareWord(l *lexer) lexFn {
	// Skip over a starting dot
	l.accept(".")
	l.ignore()

	l.acceptFunc(validFirstRune)
	l.acceptRunFunc(validSecondaryRune)
	l.emit(typBare)

	return lexStatement
}

// lexBraces lexes keys contained within square braces
func lexBraces(l *lexer) lexFn {
	l.accept("[")
	switch {
	case unicode.IsNumber(l.peek()):
		return lexNumericKey
	case l.peek() == '"':
		return lexQuotedKey
	default:
		return nil
	}
}

// lexNumericKey lexes numeric keys between square braces
func lexNumericKey(l *lexer) lexFn {
	l.accept("[")
	l.ignore()

	l.acceptRunFunc(unicode.IsNumber)
	l.emit(typNumeric)

	l.accept("]")
	l.ignore()
	return lexStatement
}

// lexQuotedKey lexes quoted keys between square braces
func lexQuotedKey(l *lexer) lexFn {
	l.accept("[")
	l.ignore()

	l.next()
	l.acceptUntilUnescaped("\"")
	l.next()
	l.emit(typQuoted)

	l.accept("]")
	l.ignore()
	return lexStatement
}

// lexValue lexes a value at the end of a statement
func lexValue(l *lexer) lexFn {
	l.accept(" ")
	if !l.accept("=") {
		return nil
	}
	l.accept(" ")
	l.ignore()
	l.acceptUntil(";")
	l.emit(typValue)
	return nil
}

// ungronTokens turns a slice of tokens into an actual datastructure
func ungronTokens(ts []token) (interface{}, error) {
	if len(ts) == 0 {
		return nil, fmt.Errorf("zero tokens provided to ungronTokens")
	}
	if ts[len(ts)-1].typ != typValue {
		return nil, fmt.Errorf("last token in slice is not a value")
	}

	t := ts[0]
	switch t.typ {

	case typValue:
		var val interface{}
		err := json.Unmarshal([]byte(t.text), &val)
		if err != nil {
			return nil, fmt.Errorf("failed to handle quoted key `%s`", t.text)
		}
		return val, nil

	case typBare:
		val, err := ungronTokens(ts[1:])
		if err != nil {
			return nil, err
		}
		out := make(map[string]interface{})
		out[t.text] = val
		return out, nil

	case typQuoted:
		val, err := ungronTokens(ts[1:])
		if err != nil {
			return nil, err
		}
		key := ""
		err = json.Unmarshal([]byte(t.text), &key)
		if err != nil {
			return nil, fmt.Errorf("failed to handle quoted key `%s`", t.text)
		}

		out := make(map[string]interface{})
		out[key] = val
		return out, nil

	case typNumeric:
		key, err := strconv.Atoi(t.text)
		if err != nil {
			return nil, fmt.Errorf("failed to convert array key to int: %s", err)
		}

		val, err := ungronTokens(ts[1:])
		if err != nil {
			return nil, err
		}

		// There needs to be at least key + 1 space in the array
		out := make([]interface{}, key+1)
		out[key] = val
		return out, nil

	default:
		return nil, fmt.Errorf("failed to ungron tokens")
	}
}

// recursiveMerge merges maps and slices, or returns b for scalars
func recursiveMerge(a, b interface{}) (interface{}, error) {
	switch a.(type) {

	case map[string]interface{}:
		bMap, ok := b.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("cannot merge map[string]interface{} with non-map")
		}
		return recursiveMapMerge(a.(map[string]interface{}), bMap)

	case []interface{}:
		bSlice, ok := b.([]interface{})
		if !ok {
			return nil, fmt.Errorf("cannot merge []interface{} with non-slice")
		}
		return recursiveSliceMerge(a.([]interface{}), bSlice)

	case string, int, float64, bool, nil:
		// Can't merge them, second one wins
		return b, nil

	default:
		return nil, fmt.Errorf("cannot merge datastructures that are not []interface{} or map[string]interface{}")
	}
}

// recursiveMapMerge recursively merges map[string]interface{} values
func recursiveMapMerge(a, b map[string]interface{}) (map[string]interface{}, error) {
	// Merge keys from b into a
	for k, v := range b {
		_, exists := a[k]
		if !exists {
			// Doesn't exist in a, just add it in
			a[k] = v
		} else {
			// Does exist, merge the values
			merged, err := recursiveMerge(a[k], b[k])
			if err != nil {
				return nil, fmt.Errorf("error merging map values: %s", err)
			}

			a[k] = merged
		}
	}
	return a, nil
}

// recursiveSliceMerge recursively merged []interface{} values
func recursiveSliceMerge(a, b []interface{}) ([]interface{}, error) {
	// We need a new slice with the capacity of whichever
	// slive is biggest
	outLen := len(a)
	if len(b) > outLen {
		outLen = len(b)
	}
	out := make([]interface{}, outLen)

	// Copy the values from 'a' into the output slice
	copy(out, a)

	// Add the values from 'b'; merging existing keys
	for k, v := range b {
		if out[k] == nil {
			out[k] = v
		} else if v != nil {
			merged, err := recursiveMerge(out[k], b[k])
			if err != nil {
				return nil, fmt.Errorf("error merging slice values: %s", err)
			}
			out[k] = merged
		}
	}
	return out, nil
}
