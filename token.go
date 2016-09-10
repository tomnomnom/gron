package main

import (
	"encoding/json"
	"fmt"
)

// A token is a chunk of text from a statement with a type
type token struct {
	text string
	typ  tokenTyp
}

// A tokenTyp identifies what kind of token something is
type tokenTyp int

const (
	// A bare word is a unquoted key; like 'foo' in json.foo = 1;
	typBare tokenTyp = iota

	// Numeric key; like '2' in json[2] = "foo";
	typNumericKey

	// A quoted key; like 'foo bar' in json["foo bar"] = 2;
	typQuotedKey

	// Punctuation types
	typDot    // .
	typLBrace // [
	typRBrace // ]
	typEquals // =
	typSemi   // ;

	// Value types
	typString      // "foo"
	typNumber      // 4
	typTrue        // true
	typFalse       // false
	typNull        // null
	typEmptyArray  // []
	typEmptyObject // {}

	// Ignored token
	typIgnored

	// Error token
	typError
)

// a sprintFn adds color to its input
type sprintFn func(...interface{}) string

// mapping of token types to the appropriate color sprintFn
var sprintFns = map[tokenTyp]sprintFn{
	typBare:        bareColor.SprintFunc(),
	typNumericKey:  numColor.SprintFunc(),
	typQuotedKey:   strColor.SprintFunc(),
	typLBrace:      braceColor.SprintFunc(),
	typRBrace:      braceColor.SprintFunc(),
	typString:      strColor.SprintFunc(),
	typNumber:      numColor.SprintFunc(),
	typTrue:        boolColor.SprintFunc(),
	typFalse:       boolColor.SprintFunc(),
	typNull:        boolColor.SprintFunc(),
	typEmptyArray:  braceColor.SprintFunc(),
	typEmptyObject: braceColor.SprintFunc(),
}

// isValue returns true if the token is a valid value type
func (t token) isValue() bool {
	switch t.typ {
	case typString, typNumber, typTrue, typFalse, typNull, typEmptyArray, typEmptyObject:
		return true
	default:
		return false
	}
}

// isPunct returns true is the token is a punctuation type
func (t token) isPunct() bool {
	switch t.typ {
	case typDot, typLBrace, typRBrace, typEquals, typSemi:
		return true
	default:
		return false
	}
}

// format returns the formatted version of the token text
func (t token) format() string {
	if t.typ == typEquals {
		return " " + t.text + " "
	}
	return t.text
}

// formatColor returns the colored formatted version of the token text
func (t token) formatColor() string {
	text := t.text
	if t.typ == typEquals {
		text = " " + text + " "
	}
	fn, ok := sprintFns[t.typ]
	if ok {
		return fn(text)
	}
	return text

}

// valueTokenFromInterface takes any valid value and
// returns a value token to represent it
func valueTokenFromInterface(v interface{}) token {
	switch vv := v.(type) {

	case map[string]interface{}:
		return token{"{}", typEmptyObject}
	case []interface{}:
		return token{"[]", typEmptyArray}
	case json.Number:
		return token{vv.String(), typNumber}
	case string:
		return token{quoteString(vv), typString}
	case bool:
		if vv {
			return token{"true", typTrue}
		}
		return token{"false", typFalse}
	case nil:
		return token{"null", typNull}
	default:
		return token{"", typError}
	}
}

// quoteString takes a string and returns a quoted and
// escaped string valid for use in gron output
func quoteString(s string) string {
	out, err := json.Marshal(s)
	if err != nil {
		// It shouldn't be possible to be given a string we can't marshal
		// so just bomb out in spectacular style
		panic(fmt.Sprintf("failed to marshal string: %s", s))
	}
	return string(out)
}
