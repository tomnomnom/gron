package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
)

func main() {
	filename := "test-input.json"
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		todo(err)
	}

	// The 'JSON' object could actually be an object, array or scalar,
	// so the best we can do for now is an empty interface type
	var top interface{}
	err = json.Unmarshal(raw, &top)
	if err != nil {
		todo(err)
	}

	err = printStatements("json", top)
	if err != nil {
		todo(err)
	}
}

func keyMustBeQuoted(s string) bool {
	r := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*$`)
	return !r.MatchString(s)
}

func makePrefix(prev string, next interface{}) (string, error) {
	switch v := next.(type) {
	case int:
		return fmt.Sprintf("%s[%d]", prev, v), nil
	case string:
		if keyMustBeQuoted(v) {
			return fmt.Sprintf("%s[%s]", prev, escape(v)), nil
		} else {
			return fmt.Sprintf("%s.%s", prev, v), nil
		}
	default:
		return "", fmt.Errorf("could not form prefix for %#v", next)
	}
}

func printStatements(prefix string, v interface{}) error {
	switch vv := v.(type) {

	case map[string]interface{}:
		// It's an object
		fmt.Printf("%s = {};\n", prefix)

		for k, sub := range vv {
			newPrefix, err := makePrefix(prefix, k)
			if err != nil {
				return err
			}
			printStatements(newPrefix, sub)
		}

	case []interface{}:
		// It's an array
		fmt.Printf("%s = [];\n", prefix)

		for k, sub := range vv {
			newPrefix, err := makePrefix(prefix, k)
			if err != nil {
				return err
			}
			printStatements(newPrefix, sub)
		}

	case float64:
		fmt.Printf("%s = %s;\n", prefix, escape(vv))

	case string:
		fmt.Printf("%s = %s;\n", prefix, escape(vv))

	case bool:
		fmt.Printf("%s = %t;\n", prefix, vv)

	case nil:
		fmt.Printf("%s = null;\n", prefix)
	}

	return nil
}

func escape(s interface{}) string {
	// I'm pretty sure it's safe to ignore this error
	// ...maybe. I'll work something into this I promise
	out, _ := json.Marshal(s)
	return string(out)
}

func todo(err error) {
	log.Fatalf("TODO: %s", err)
}
