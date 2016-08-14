package main

import (
	"encoding/json"
	"fmt"

	"github.com/fatih/color"
)

var (
	strColor   = color.New(color.FgYellow).SprintFunc()
	braceColor = color.New(color.FgMagenta).SprintFunc()
	bareColor  = color.New(color.FgBlue, color.Bold).SprintFunc()
	numColor   = color.New(color.FgRed).SprintFunc()
	boolColor  = color.New(color.FgCyan).SprintFunc()
)

// statementFormatters handle the formatting of statements
// They're responsible for creating prefixes/keys, escaping
// values and formatting assignments
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

// colorFormatter formats statements in color
type colorFormatter struct{}

// value uses json.Marshal to format scalars
// E.g:
//     a string -> "a string"
//     7.0000   -> 7
func (f colorFormatter) value(s interface{}) string {
	out, err := json.Marshal(s)
	if err != nil {
		// It shouldn't be possible to be given a value we can't marshal
		panic(fmt.Sprintf("failed to marshal value: %#v", s))
	}
	return string(out)
}

// prefix takes the previous prefix and the next identifier and
// returns a new prefix or an error on failure
func (f colorFormatter) prefix(prev string, next interface{}) (string, error) {
	if prev == "json" {
		prev = bareColor(prev)
	}

	switch v := next.(type) {
	case int:
		// Next identifier is an array key
		return prev + braceColor("[") + numColor(v) + braceColor("]"), nil
	case string:
		// Next identifier is an object key, either bare or quoted
		if validIdentifier(v) {
			return prev + "." + bareColor(v), nil
		}
		return prev + braceColor("[") + strColor(f.value(v)) + braceColor("]"), nil
	default:
		return "", fmt.Errorf("could not form prefix for %#v", next)
	}
}

// assignment formats an assignment
func (f colorFormatter) assignment(key string, value interface{}) string {
	if key == "json" {
		key = bareColor(key)
	}

	var valStr string

	switch vv := value.(type) {

	case map[string]interface{}:
		valStr = braceColor("{}")
	case []interface{}:
		valStr = braceColor("[]")
	case json.Number:
		valStr = numColor(vv.String())
	case float64:
		valStr = numColor(f.value(vv))
	case string:
		valStr = strColor(f.value(vv))
	case bool:
		valStr = boolColor(fmt.Sprintf("%t", vv))
	case nil:
		valStr = boolColor("null")
	}
	return fmt.Sprintf("%s = %s;", key, valStr)
}
