package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"unicode"
	"unicode/utf8"
)

// formatter is an interchangeable statement formatter. It's
// interchangeable so that it can be swapped between the
// monochrome and color formatters
var formatter statementFormatter = colorFormatter{}

// statements is a list of assignment statements.
// E.g statement: json.foo = "bar";
type statements []string

// Add adds a new statement to the list given the key and a value
func (ss *statements) Add(key string, value interface{}) {
	*ss = append(*ss, formatter.assignment(key, value))
}

// AddFull adds a new statement to the list given the entire statement
func (ss *statements) AddFull(s string) {
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

// ungron turns statements into a proper datastructur
func (ss statements) ungron() (interface{}, error) {

	// Get all the idividually parsed statements
	var parsed []interface{}
	for _, s := range ss {
		l := newLexer(s)
		u, err := ungronTokens(l.lex())

		if err != nil {
			return nil, fmt.Errorf("failed to translate tokens into datastructure: %s", err)
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
			return nil, fmt.Errorf("failed to merge statements: %s", err)
		}
		merged = m
	}
	return merged, nil

}

// Less compares two statements for sort.Sort
// Implements a natural sort to keep array indexes in order
func (ss statements) Less(a, b int) bool {

	// Two statements should never be identical, but I can't bring
	// myself not to guard against the possibility
	if ss[a] == ss[b] {
		return true
	}

	// The statements may contain ANSI color codes. We don't
	// want to sort based on the colors so we need to strip
	// them out first
	aStr := stripColors(ss[a])
	bStr := stripColors(ss[b])

	// Find where the two strings start to differ, keeping track
	// of where any numbers start so that we can compare them properly
	numStart := -1
	diffStart := -1
	for i, ra := range aStr {
		rb, _ := utf8.DecodeRuneInString(bStr[i:])

		// Are we looking at a number?
		isNum := unicode.IsNumber(ra) && unicode.IsNumber(rb)

		// If we are looking at a number but don't have a start
		// position then this is the start of a new number
		if isNum && numStart == -1 {
			numStart = i
		}

		// Found a difference
		if ra != rb {
			diffStart = i
			break
		}

		// There was no difference yet, so if we're not looking at a
		// number: reset numStart so we start looking again
		if !isNum {
			numStart = -1
		}
	}

	// If diffStart is still -1 then the only difference must be
	// that string B is longer than A, so A should come first
	if diffStart == -1 {
		return true
	}

	// If we don't have a start position for a number, that means the
	// difference we found wasn't numeric, so do a regular comparison
	// on the remainder of the strings
	if numStart == -1 {
		return aStr[diffStart:] < bStr[diffStart:]
	}

	// Read and compare the numbers from each string
	return readNum(aStr[numStart:]) < readNum(bStr[numStart:])
}

// stripColors removes ANSI colors from a string
func stripColors(in string) string {
	var buf bytes.Buffer

	inColor := false
	for i := 0; i < len(in); {
		r, l := utf8.DecodeRuneInString(in[i:])
		i += l

		if inColor && r == 'm' {
			inColor = false
			continue
		}

		// Escape char
		if r == 0x001B {
			inColor = true
		}

		if inColor {
			continue
		}

		// buf.WriteRune doesn't actually ever return
		// an error, so it's safe to ignore it
		_, _ = buf.WriteRune(r)
	}

	return buf.String()
}

// readNum reads digits from a string until it hits a non-digit,
// returning the digits as an integer
func readNum(str string) int {
	numEnd := len(str)
	for i, r := range str {
		// If we hit a non-number then we've found the end
		if !unicode.IsNumber(r) {
			numEnd = i
			break
		}

	}
	// If we've failed to parse a number then zero is
	// just fine; it's being used for sorting only
	num, _ := strconv.Atoi(str[:numEnd])
	return num
}

// Contains searches the statements for a given statement
// Mostly to make testing things easier
func (ss statements) Contains(search string) bool {
	for _, i := range ss {
		// The Contains method is used exclusively for testing
		// so while stripping the colors out of every statement
		// every time Contains is called isn't very efficient,
		// it won't affect users' performance.
		i = stripColors(i)
		if i == search {
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
	return makeStatements("json", top)
}

// makeStatements takes a prefix and interface value and returns
// a statements list or an error on failure
func makeStatements(prefix string, v interface{}) (statements, error) {
	ss := make(statements, 0)

	// Add a statement for the current prefix and value
	ss.Add(prefix, v)

	// Recurse into objects and arrays
	switch vv := v.(type) {

	case map[string]interface{}:
		// It's an object
		for k, sub := range vv {
			newPrefix, err := formatter.prefix(prefix, k)
			if err != nil {
				return ss, err
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
			newPrefix, err := formatter.prefix(prefix, k)
			if err != nil {
				return ss, err
			}
			extra, err := makeStatements(newPrefix, sub)
			if err != nil {
				return ss, err
			}
			ss.AddMulti(extra)
		}
	}

	return ss, nil
}
