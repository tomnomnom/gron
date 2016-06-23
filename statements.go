package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"unicode"
	"unicode/utf8"
)

// statements is a list of assignment statements.
// E.g statement: json.foo = "bar";
type statements []string

// Add adds a new statement to the list given the prefix and value
func (ss *statements) Add(prefix, value string) {
	*ss = append(*ss, fmt.Sprintf("%s = %s;", prefix, value))
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

// Less compares two statements for sort.Sort
// A limited kind of natural sort to keep array indexes in order
func (ss statements) Less(a, b int) bool {

	// Two statements should never be identical, but I can't bring
	// myself not to guard against the possibility
	if ss[a] == ss[b] {
		return true
	}

	// Run through the strings until we find a difference, tracking
	// where numbers start in both strings
	var ra, rb rune

	numStart := -1
	for i, w := 0, 0; i < len(ss[a]); i += w {
		ra, w = utf8.DecodeRuneInString(ss[a][i:])

		// Check there's actually enough bytes left in
		// string B to get another rune
		if i > len(ss[b]) {
			return true
		}

		rb, _ = utf8.DecodeRuneInString(ss[b][i:])

		// If we've hit the start of a number in both strings we
		// need to keep track of where the numbers start
		if numStart == -1 && unicode.IsNumber(ra) && unicode.IsNumber(rb) {
			numStart = i
		}

		// Found a difference
		if ra != rb {
			break
		}

		// If there's no difference, and the runes aren't numbers then
		// reset numStart so we can spot the start of the next number
		if !unicode.IsNumber(ra) {
			numStart = -1
		}
	}

	// If both the runes aren't numbers just compare the two runes
	if numStart == -1 {
		return ra < rb
	}

	// Read and compare the numbers from each string
	return readNum(ss[a][numStart:]) < readNum(ss[b][numStart:])
}

// readNum reads digits from a string until it hits a non-digit,
// returning the digits as an integer
func readNum(str string) int {
	buf := bytes.Buffer{}
	for _, r := range str {
		if !unicode.IsNumber(r) {
			break
		}

		// WriteRune's error is always nil
		_, _ = buf.WriteRune(r)
	}

	// If we've failed to parse a number then zero is
	// just fine; it's being used for sorting only
	num, _ := strconv.Atoi(buf.String())
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

// makeStatements takes a prefix and interface value and returns
// a statements list or an error on failure
func makeStatements(prefix string, v interface{}) (statements, error) {
	ss := make(statements, 0)

	switch vv := v.(type) {

	case map[string]interface{}:
		// It's an object
		ss.Add(prefix, "{}")

		for k, sub := range vv {
			newPrefix, err := makePrefix(prefix, k)
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
		ss.Add(prefix, "[]")

		for k, sub := range vv {
			newPrefix, err := makePrefix(prefix, k)
			if err != nil {
				return ss, err
			}
			extra, err := makeStatements(newPrefix, sub)
			if err != nil {
				return ss, err
			}
			ss.AddMulti(extra)
		}

	case float64:
		ss.Add(prefix, formatValue(vv))

	case string:
		ss.Add(prefix, formatValue(vv))

	case bool:
		ss.Add(prefix, fmt.Sprintf("%t", vv))

	case nil:
		ss.Add(prefix, "null")
	}

	return ss, nil
}

// formatValue uses json.Marshal to format scalars
// E.g:
//     a string -> "a string"
//     7.0000   -> 7
func formatValue(s interface{}) string {
	// I'm pretty sure it's safe to ignore this error
	// ...maybe. I'll work something into this I promise
	out, _ := json.Marshal(s)
	return string(out)
}

// makePrefix takes the previous prefix and the next key and
// returns a new prefix or an error on failure
func makePrefix(prev string, next interface{}) (string, error) {
	switch v := next.(type) {
	case int:
		return fmt.Sprintf("%s[%d]", prev, v), nil
	case string:
		if validIdentifier(v) {
			// This is a fairly hot code path, and concatination has
			// proven to be faster than fmt.Sprintf, despite the allocations
			return prev + "." + v, nil
		}
		return prev + "[" + formatValue(v) + "]", nil
	default:
		return "", fmt.Errorf("could not form prefix for %#v", next)
	}
}
