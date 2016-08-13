package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"unicode"
	"unicode/utf8"
)

// formatter is an interchangeable statement formatter. It's
// interchangeable so that it can be swapped out with one
// that, for example, uses colors
var formatter monoFormatter

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

	// Find where the two strings start to differ, keeping track
	// of where any numbers start so that we can compare them properly
	numStart := -1
	diffStart := -1
	for i, ra := range ss[a] {
		rb, _ := utf8.DecodeRuneInString(ss[b][i:])

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
		return ss[a][diffStart:] < ss[b][diffStart:]
	}

	// Read and compare the numbers from each string
	return readNum(ss[a][numStart:]) < readNum(ss[b][numStart:])
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

// Contains seaches the statements for a given statement
// Mostly to make testing things easier
func (ss statements) Contains(search string) bool {
	for _, i := range ss {
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

// statementFormatters handle the formatting of the output text
type statementFormatter interface {
	prefix(string, interface{}) (string, error)
	value(interface{}) string
	assignment(string, interface{}) string
}

// monoFormatter formats statements in monochrome
type monoFormatter struct{}

// value uses json.Marshal to format scalars
// E.g:
//     a string -> "a string"
//     7.0000   -> 7
func (f monoFormatter) value(s interface{}) string {
	out, err := json.Marshal(s)
	if err != nil {
		// It shouldn't be possible to be given a value we can't marshal
		panic(fmt.Sprintf("failed to marshal value: %#v", s))
	}
	return string(out)
}

// prefix takes the previous prefix and the next identifier and
// returns a new prefix or an error on failure
func (f monoFormatter) prefix(prev string, next interface{}) (string, error) {
	switch v := next.(type) {
	case int:
		// Next identifier is an array key
		return fmt.Sprintf("%s[%d]", prev, v), nil
	case string:
		// Next identifier is an object key, either bare or quoted
		if validIdentifier(v) {
			// This is a fairly hot code path, and concatination has
			// proven to be faster than fmt.Sprintf, despite the allocations
			return prev + "." + v, nil
		}
		return prev + "[" + f.value(v) + "]", nil
	default:
		return "", fmt.Errorf("could not form prefix for %#v", next)
	}
}

// assignment formats an assignment
func (f monoFormatter) assignment(key string, value interface{}) string {
	var valStr string

	switch vv := value.(type) {

	case map[string]interface{}:
		valStr = "{}"
	case []interface{}:
		valStr = "[]"
	case json.Number:
		valStr = vv.String()
	case float64, string:
		valStr = f.value(vv)
	case bool:
		valStr = fmt.Sprintf("%t", vv)
	case nil:
		valStr = "null"
	}
	return fmt.Sprintf("%s = %s;", key, valStr)
}
