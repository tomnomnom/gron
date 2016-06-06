package main

import (
	"encoding/json"
	"fmt"
	"regexp"
)

// The javascript reserved words cannot be used as unquoted keys
var reservedWords = []string{
	"break",
	"case",
	"catch",
	"class",
	"const",
	"continue",
	"debugger",
	"default",
	"delete",
	"do",
	"else",
	"export",
	"extends",
	"false",
	"finally",
	"for",
	"function",
	"if",
	"import",
	"in",
	"instanceof",
	"new",
	"null",
	"return",
	"super",
	"switch",
	"this",
	"throw",
	"true",
	"try",
	"typeof",
	"var",
	"void",
	"while",
	"with",
	"yield",
}

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
func (ss statements) Less(i, j int) bool {
	return ss[i] < ss[j]
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

// keyMustBeQuoted checks to see if a key in a JSON object
// must be quoted or not.
// E.g:
//     justLettersAndNumbers1 -> false
//     a key with spaces      -> true
//     1startsWithANumber	  -> true
func keyMustBeQuoted(s string) bool {
	r := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	if !r.MatchString(s) {
		return true
	}

	// Check the list of reserved words
	for _, i := range reservedWords {
		if s == i {
			return true
		}
	}
	return false
}

// makePrefix takes the previous prefix and the next key and
// returns a new prefix or an error on failure
func makePrefix(prev string, next interface{}) (string, error) {
	switch v := next.(type) {
	case int:
		return fmt.Sprintf("%s[%d]", prev, v), nil
	case string:
		if keyMustBeQuoted(v) {
			return fmt.Sprintf("%s[%s]", prev, formatValue(v)), nil
		}
		return fmt.Sprintf("%s.%s", prev, v), nil
	default:
		return "", fmt.Errorf("could not form prefix for %#v", next)
	}
}
