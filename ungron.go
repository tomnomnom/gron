// Ungronning is the reverse of gronning: turn statements
// back into JSON. The expected input grammar is:
//
//   Input ::= '--'* Statement (Statement | '--')*
//   Statement ::= Path Space* "=" Space* Value ";" "\n"
//   Path ::= (BareWord) ("." BareWord | ("[" Key "]"))*
//   Value ::= String | Number | "true" | "false" | "null" | "[]" | "{}"
//   BareWord ::= (UnicodeLu | UnicodeLl | UnicodeLm | UnicodeLo | UnicodeNl | '$' | '_') (UnicodeLu | UnicodeLl | UnicodeLm | UnicodeLo | UnicodeNl | UnicodeMn | UnicodeMc | UnicodeNd | UnicodePc | '$' | '_')*
//   Key ::= [0-9]+ | String
//   String ::= '"' (UnescapedRune | ("\" (["\/bfnrt] | ('u' Hex))))* '"'
//   UnescapedRune ::= [^#x0-#x1f"\]

package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/pkg/errors"
)

// errRecoverable is an error type to represent errors that
// can be recovered from; e.g. an empty line in the input
type errRecoverable struct {
	msg string
}

func (e errRecoverable) Error() string {
	return e.msg
}

// A lexer holds the state for lexing statements
type lexer struct {
	text       string  // The raw input text
	pos        int     // The current byte offset in the text
	width      int     // The width of the current rune in bytes
	cur        rune    // The rune at the current position
	prev       rune    // The rune at the previous position
	tokens     []token // The tokens that have been emitted
	tokenStart int     // The starting position of the current token
}

// newLexer returns a new lexer for the provided input string
func newLexer(text string) *lexer {
	return &lexer{
		text:       text,
		pos:        0,
		tokenStart: 0,
		tokens:     make([]token, 0),
	}
}

// lex runs the lexer and returns the lexed statement
func (l *lexer) lex() statement {

	for lexfn := lexStatement; lexfn != nil; {
		lexfn = lexfn(l)
	}
	return l.tokens
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
func (l *lexer) acceptFunc(fn runeCheck) bool {
	if fn(l.next()) {
		return true
	}
	l.backup()
	return false
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

	// Read until we hit an unescaped rune or the end of the input
	inEscape := false
	for {
		r := l.next()
		if r == '\\' && !inEscape {
			inEscape = true
			continue
		}
		if strings.ContainsRune(delims, r) && !inEscape {
			l.backup()
			return
		}
		if l.cur == utf8.RuneError {
			return
		}
		inEscape = false
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
	case r == ' ', r == '=':
		return lexValue
	case r == '-':
		// grep -A etc can add '--' lines to output
		// we'll save the text but not actually do
		// anything with them
		return lexIgnore
	case r == utf8.RuneError:
		return nil
	default:
		l.emit(typError)
		return nil
	}

}

// lexBareWord lexes for bare identifiers.
// E.g: the 'foo' in 'foo.bar' or 'foo[0]' is a bare identifier
func lexBareWord(l *lexer) lexFn {
	if l.accept(".") {
		l.emit(typDot)
	}

	if !l.acceptFunc(validFirstRune) {
		l.emit(typError)
		return nil
	}
	l.acceptRunFunc(validSecondaryRune)
	l.emit(typBare)

	return lexStatement
}

// lexBraces lexes keys contained within square braces
func lexBraces(l *lexer) lexFn {
	l.accept("[")
	l.emit(typLBrace)

	switch {
	case unicode.IsNumber(l.peek()):
		return lexNumericKey
	case l.peek() == '"':
		return lexQuotedKey
	default:
		l.emit(typError)
		return nil
	}
}

// lexNumericKey lexes numeric keys between square braces
func lexNumericKey(l *lexer) lexFn {
	l.accept("[")
	l.ignore()

	l.acceptRunFunc(unicode.IsNumber)
	l.emit(typNumericKey)

	if l.accept("]") {
		l.emit(typRBrace)
	} else {
		l.emit(typError)
		return nil
	}
	l.ignore()
	return lexStatement
}

// lexQuotedKey lexes quoted keys between square braces
func lexQuotedKey(l *lexer) lexFn {
	l.accept("[")
	l.ignore()

	l.accept(`"`)

	l.acceptUntilUnescaped(`"`)
	l.accept(`"`)
	l.emit(typQuotedKey)

	if l.accept("]") {
		l.emit(typRBrace)
	} else {
		l.emit(typError)
		return nil
	}
	l.ignore()
	return lexStatement
}

// lexValue lexes a value at the end of a statement
func lexValue(l *lexer) lexFn {
	l.acceptRun(" ")
	l.ignore()

	if l.accept("=") {
		l.emit(typEquals)
	} else {
		return nil
	}
	l.acceptRun(" ")
	l.ignore()

	switch {

	case l.accept(`"`):
		l.acceptUntilUnescaped(`"`)
		l.accept(`"`)
		l.emit(typString)

	case l.accept("t"):
		l.acceptRun("rue")
		l.emit(typTrue)

	case l.accept("f"):
		l.acceptRun("alse")
		l.emit(typFalse)

	case l.accept("n"):
		l.acceptRun("ul")
		l.emit(typNull)

	case l.accept("["):
		l.accept("]")
		l.emit(typEmptyArray)

	case l.accept("{"):
		l.accept("}")
		l.emit(typEmptyObject)

	default:
		// Assume number
		l.acceptUntil(";")
		l.emit(typNumber)
	}

	l.acceptRun(" ")
	l.ignore()

	if l.accept(";") {
		l.emit(typSemi)
	}

	// The value should always be the last thing
	// in the statement
	return nil
}

// lexIgnore accepts runes until the end of the input
// and emits them as a typIgnored token
func lexIgnore(l *lexer) lexFn {
	l.acceptRunFunc(func(r rune) bool {
		return r != utf8.RuneError
	})
	l.emit(typIgnored)
	return nil
}

// ungronTokens turns a slice of tokens into an actual datastructure
func ungronTokens(ts []token) (interface{}, error) {
	if len(ts) == 0 {
		return nil, errRecoverable{"empty input"}
	}

	if ts[0].typ == typIgnored {
		return nil, errRecoverable{"ignored token"}
	}

	if ts[len(ts)-1].typ == typError {
		return nil, errors.New("invalid statement")
	}

	// The last token should be typSemi so we need to check
	// the second to last token is a value rather than the
	// last one
	if len(ts) > 1 && !ts[len(ts)-2].isValue() {
		return nil, errors.New("statement has no value")
	}

	t := ts[0]
	switch {
	case t.isPunct():
		// Skip the token
		val, err := ungronTokens(ts[1:])
		if err != nil {
			return nil, err
		}
		return val, nil

	case t.isValue():
		var val interface{}
		d := json.NewDecoder(strings.NewReader(t.text))
		d.UseNumber()
		err := d.Decode(&val)
		if err != nil {
			return nil, fmt.Errorf("invalid value `%s`", t.text)
		}
		return val, nil

	case t.typ == typBare:
		val, err := ungronTokens(ts[1:])
		if err != nil {
			return nil, err
		}
		out := make(map[string]interface{})
		out[t.text] = val
		return out, nil

	case t.typ == typQuotedKey:
		val, err := ungronTokens(ts[1:])
		if err != nil {
			return nil, err
		}
		key := ""
		err = json.Unmarshal([]byte(t.text), &key)
		if err != nil {
			return nil, fmt.Errorf("invalid quoted key `%s`", t.text)
		}

		out := make(map[string]interface{})
		out[key] = val
		return out, nil

	case t.typ == typNumericKey:
		key, err := strconv.Atoi(t.text)
		if err != nil {
			return nil, fmt.Errorf("invalid integer key `%s`", t.text)
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
		return nil, fmt.Errorf("unexpected token `%s`", t.text)
	}
}

// recursiveMerge merges maps and slices, or returns b for scalars
func recursiveMerge(a, b interface{}) (interface{}, error) {
	switch a.(type) {

	case map[string]interface{}:
		bMap, ok := b.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("cannot merge object with non-object")
		}
		return recursiveMapMerge(a.(map[string]interface{}), bMap)

	case []interface{}:
		bSlice, ok := b.([]interface{})
		if !ok {
			return nil, fmt.Errorf("cannot merge array with non-array")
		}
		return recursiveSliceMerge(a.([]interface{}), bSlice)

	case string, int, float64, bool, nil:
		// Can't merge them, second one wins
		return b, nil

	default:
		return nil, fmt.Errorf("unexpected data type for merge")
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
				return nil, err
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
				return nil, err
			}
			out[k] = merged
		}
	}
	return out, nil
}
