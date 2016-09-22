package main

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// A statement is a slice of tokens representing an assignment statement.
// An assignment statement is something like:
//
//   json.city = "Leeds";
//
// Where 'json', '.', 'city', '=', '"Leeds"' and ';' are discrete tokens.
// Statements are stored as tokens to make sorting more efficient, and so
// that the same type can easily be used when gronning and ungronning.
type statement []token

// String returns the string form of a statement rather than the
// underlying slice of tokens
func (s statement) String() string {
	out := make([]string, 0, len(s)+2)
	for _, t := range s {
		out = append(out, t.format())
	}
	return strings.Join(out, "")
}

// colorString returns the string form of a statement with ASCII color codes
func (s statement) colorString() string {
	out := make([]string, 0, len(s)+2)
	for _, t := range s {
		out = append(out, t.formatColor())
	}
	return strings.Join(out, "")
}

// withBare returns a copy of a statement with a new bare
// word token appended to it
func (s statement) withBare(k string) statement {
	new := make(statement, len(s), len(s)+2)
	copy(new, s)
	return append(
		new,
		token{".", typDot},
		token{k, typBare},
	)
}

// withQuotedKey returns a copy of a statement with a new
// quoted key token appended to it
func (s statement) withQuotedKey(k string) statement {
	new := make(statement, len(s), len(s)+3)
	copy(new, s)
	return append(
		new,
		token{"[", typLBrace},
		token{quoteString(k), typQuotedKey},
		token{"]", typRBrace},
	)
}

// withNumericKey returns a copy of a statement with a new
// numeric key token appended to it
func (s statement) withNumericKey(k int) statement {
	new := make(statement, len(s), len(s)+3)
	copy(new, s)
	return append(
		new,
		token{"[", typLBrace},
		token{strconv.Itoa(k), typNumericKey},
		token{"]", typRBrace},
	)
}

// statements is a list of assignment statements.
// E.g statement: json.foo = "bar";
type statements []statement

// addWithValue takes a statement representing a path, copies it,
// adds a value token to the end of the statement and appends
// the new statement to the list of statements
func (ss *statements) addWithValue(path statement, value token) {
	s := make(statement, len(path), len(path)+3)
	copy(s, path)
	s = append(s, token{"=", typEquals}, value, token{";", typSemi})
	*ss = append(*ss, s)
}

// add appends a new complete statement to list of statements
func (ss *statements) add(s statement) {
	*ss = append(*ss, s)
}

// Len returns the number of statements for sort.Sort
func (ss statements) Len() int {
	return len(ss)
}

// Swap swaps two statements for sort.Sort
func (ss statements) Swap(i, j int) {
	ss[i], ss[j] = ss[j], ss[i]
}

// statementFromString takes statement string, lexes it and returns
// the corresponding statement
func statementFromString(str string) statement {
	l := newLexer(str)
	s := l.lex()
	return s
}

// ungron turns statements into a proper datastructure
func (ss statements) toInterface() (interface{}, error) {

	// Get all the individually parsed statements
	var parsed []interface{}
	for _, s := range ss {
		u, err := ungronTokens(s)

		switch err.(type) {
		case nil:
			// no problem :)
		case errRecoverable:
			continue
		default:
			return nil, errors.Wrapf(err, "ungron failed for `%s`", s)
		}

		parsed = append(parsed, u)
	}

	if len(parsed) == 0 {
		return nil, fmt.Errorf("no statements were parsed")
	}

	merged := parsed[0]
	for _, p := range parsed[1:] {
		m, err := recursiveMerge(merged, p)
		if err != nil {
			return nil, errors.Wrap(err, "failed to merge statements")
		}
		merged = m
	}
	return merged, nil

}

// Less compares two statements for sort.Sort
// Implements a natural sort to keep array indexes in order
func (ss statements) Less(a, b int) bool {

	diffStart := -1
	for i := range ss[a] {

		if len(ss[b]) < i+1 {
			// b must be shorter than a, so it
			// should come first
			return false
		}

		// The tokens match, so just carry on
		if ss[a][i] == ss[b][i] {
			continue
		}

		// We've found a difference
		diffStart = i
		break
	}

	// If diffStart is still -1 then the only difference must be
	// that string B is longer than A, so A should come first
	if diffStart == -1 {
		return true
	}

	// Get the tokens that differ
	ta := ss[a][diffStart]
	tb := ss[b][diffStart]

	// An equals always comes first
	if ta.typ == typEquals {
		return true
	}
	if tb.typ == typEquals {
		return false
	}

	// If both tokens are numeric keys do an integer comparison
	if ta.typ == typNumericKey && tb.typ == typNumericKey {
		ia, _ := strconv.Atoi(ta.text)
		ib, _ := strconv.Atoi(tb.text)
		return ia < ib
	}

	// If neither token is a number, just do a string comparison
	if ta.typ != typNumber || tb.typ != typNumber {
		return ta.text < tb.text
	}

	// We have two numbers to compare so turn them into json.Number
	// for comparison
	na, _ := json.Number(ta.text).Float64()
	nb, _ := json.Number(tb.text).Float64()
	return na < nb

}

// Contains searches the statements for a given statement
// Mostly to make testing things easier
func (ss statements) Contains(search statement) bool {
	for _, i := range ss {
		if reflect.DeepEqual(i, search) {
			return true
		}
	}
	return false
}

// statementsFromJSON takes an io.Reader containing JSON
// and returns statements or an error on failure
func statementsFromJSON(r io.Reader) (statements, error) {
	//var top interface{}
	d := json.NewDecoder(r)
	d.UseNumber()
	//err := d.Decode(&top)
	//if err != nil {
	//return nil, err
	//}
	ss := make(statements, 0, 32)
	ss.fill(statement{{"json", typBare}}, d)
	return ss, nil
}

// fill takes a prefix statement and some value and recursively fills
// the statement list using that value
func (ss *statements) fill(prefix statement, dec *json.Decoder) {

	t, err := dec.Token()
	if err == io.EOF {
		ss.addWithValue(prefix, token{"", typError})
		return
	}
	if err != nil {
		ss.addWithValue(prefix, token{"", typError})
		return
	}

	switch vv := t.(type) {
	case json.Delim:
		// It's a map or an array
		if vv == json.Delim('{') {
			ss.addWithValue(prefix, token{"{}", typEmptyObject})
			for {
				t, err := dec.Token()
				if err != nil {
					ss.addWithValue(prefix, token{"", typError})
					return
				}

				key, ok := t.(string)
				if !ok {
					ss.addWithValue(prefix, token{"", typError})
					return
				}

				if validIdentifier(key) {
					ss.fill(prefix.withBare(key), dec)
				} else {
					ss.fill(prefix.withQuotedKey(key), dec)
				}

				if !dec.More() {
					break
				}
			}

		} else if vv == json.Delim('[') {
			ss.addWithValue(prefix, token{"[]", typEmptyArray})
			i := 0
			for {
				ss.fill(prefix.withNumericKey(i), dec)
				i++
				if !dec.More() {
					break
				}
			}

		}

		// Consume the closing delimiter
		_, _ = dec.Token()

	case json.Number:
		ss.addWithValue(prefix, token{vv.String(), typNumber})
	case string:
		ss.addWithValue(prefix, token{quoteString(vv), typString})
	case bool:
		if vv {
			ss.addWithValue(prefix, token{"true", typTrue})
		} else {
			ss.addWithValue(prefix, token{"false", typFalse})
		}
	case nil:
		ss.addWithValue(prefix, token{"null", typNull})
	default:
		ss.addWithValue(prefix, token{"", typError})
	}

}
