package main

import (
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
