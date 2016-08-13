package main

import (
	"encoding/json"
	"fmt"
)

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
