package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"sort"
)

const (
	exitOK = iota
	exitInvalidArgs
)

func main() {
	flag.Parse()

	filename := flag.Arg(0)
	if filename == "" {
		os.Exit(exitInvalidArgs)
	}

	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		todo(err)
	}

	// The 'JSON' be an object, array or scalar,
	// so the best we can do for now is an empty interface type
	var top interface{}
	err = json.Unmarshal(raw, &top)
	if err != nil {
		todo(err)
	}

	ss, err := makeStatements("json", top)
	if err != nil {
		todo(err)
	}

	sort.Sort(ss)

	for _, s := range ss {
		fmt.Println(s)
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

type statements []string

func (ss *statements) Add(prefix, value string) {
	*ss = append(*ss, fmt.Sprintf("%s = %s;", prefix, value))
}

func (ss *statements) AddGroup(g statements) {
	*ss = append(*ss, g...)
}

func (ss statements) Len() int {
	return len(ss)
}

func (ss statements) Swap(i, j int) {
	ss[i], ss[j] = ss[j], ss[i]
}

func (ss statements) Less(i, j int) bool {
	return ss[i] < ss[j]
}

func (ss statements) Contains(search string) bool {
	for _, i := range ss {
		if i == search {
			return true
		}
	}
	return false
}

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
			ss.AddGroup(extra)
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
			ss.AddGroup(extra)
		}

	case float64:
		ss.Add(prefix, escape(vv))

	case string:
		ss.Add(prefix, escape(vv))

	case bool:
		ss.Add(prefix, fmt.Sprintf("%t", vv))

	case nil:
		ss.Add(prefix, "null")
	}

	return ss, nil
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
