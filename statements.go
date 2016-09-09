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

// A statement is a complete assignment
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

func (s statement) colorString() string {
	out := make([]string, 0, len(s)+2)
	for _, t := range s {
		out = append(out, t.formatColor())
	}
	return strings.Join(out, "")
}

func (s statement) withBare(k string) statement {
	return append(
		s,
		token{".", typDot},
		token{k, typBare},
	)
}

func (s statement) withQuotedKey(k string) statement {
	return append(
		s,
		token{"[", typLBrace},
		token{quoteString(k), typQuotedKey},
		token{"]", typRBrace},
	)
}

func (s statement) withNumericKey(k int) statement {
	return append(s,
		token{"[", typLBrace},
		token{strconv.Itoa(k), typNumericKey},
		token{"]", typRBrace},
	)
}

// statements is a list of assignment statements.
// E.g statement: json.foo = "bar";
type statements []statement

// Add makes a copy of a statement and appends it to the list of statements
func (ss *statements) Add(s statement, new token) {
	add := append(s, token{"=", typEquals}, new, token{";", typSemi})
	*ss = append(*ss, add)
}

// AddFull adds a new statement to the list given the entire statement
func (ss *statements) AddFull(s statement) {
	*ss = append(*ss, s)
}

// AddMulti adds a whole other list of statements
func (ss *statements) AddMulti(l statements) {
	*ss = append(*ss, l...)
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
func (ss statements) ungron() (interface{}, error) {

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

// makeStatementsFromJSON takes an io.Reader containing JSON
// and returns statements or an error on failure
func makeStatementsFromJSON(r io.Reader) (statements, error) {
	var top interface{}
	d := json.NewDecoder(r)
	d.UseNumber()
	err := d.Decode(&top)
	if err != nil {
		return nil, err
	}
	return makeStatements(statement{{"json", typBare}}, top)
}

// makeStatements takes a prefix and interface value and returns
// a statements list or an error on failure
func makeStatements(prefix statement, v interface{}) (statements, error) {
	ss := make(statements, 0)

	// Add a statement for the current prefix and value
	ss.Add(prefix, valueTokenFromInterface(v))

	// Recurse into objects and arrays
	switch vv := v.(type) {

	case map[string]interface{}:
		// It's an object
		for k, sub := range vv {
			var newPrefix statement
			if validIdentifier(k) {
				newPrefix = prefix.withBare(k)
			} else {
				newPrefix = prefix.withQuotedKey(k)
			}
			extra, err := makeStatements(newPrefix, sub)
			if err != nil {
				return ss, err
			}
			ss.AddMulti(extra)
		}

	case []interface{}:
		// It's an array
		for k, sub := range vv {
			extra, err := makeStatements(prefix.withNumericKey(k), sub)
			if err != nil {
				return ss, err
			}
			ss.AddMulti(extra)
		}
	}

	return ss, nil
}
