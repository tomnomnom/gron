package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
)

const (
	exitOK = iota
	exitOpenFile
	exitReadInput
	exitJSONDecode
	exitFormStatements
	exitFetchURL
	exitUnknown
)

func init() {
	flag.Usage = func() {
		h := "Transform JSON (from a file, URL, or stdin) into discrete assignments to make it greppable\n\n"

		h += "Usage:\n"
		h += "  gron [file|url]\n\n"

		h += "Exit Codes:\n"
		h += fmt.Sprintf("  %d\t%s\n", exitOK, "OK")
		h += fmt.Sprintf("  %d\t%s\n", exitOpenFile, "Failed to open file")
		h += fmt.Sprintf("  %d\t%s\n", exitReadInput, "Failed to read input")
		h += fmt.Sprintf("  %d\t%s\n", exitJSONDecode, "Failed to decode JSON")
		h += fmt.Sprintf("  %d\t%s\n", exitFormStatements, "Failed to from statements")
		h += fmt.Sprintf("  %d\t%s\n", exitFetchURL, "Failed to fetch URL")
		h += "\n"

		h += "Examples:\n"
		h += "  gron /tmp/apiresponse.json\n"
		h += "  gron http://headers.jsontest.com/ \n"
		h += "  curl -s http://headers.jsontest.com/ | gron\n"

		fmt.Fprintf(os.Stderr, h)
	}
}

func main() {
	ungronFlag := flag.Bool("ungron", false, "Turn statements into JSON instead")
	flag.Parse()

	var raw io.Reader

	filename := flag.Arg(0)
	if filename == "" || filename == "-" {
		raw = os.Stdin
	} else {
		if !validURL(filename) {
			r, err := os.Open(filename)
			if err != nil {
				fatal(exitOpenFile, "failed to open file", err)
			}
			raw = r
		} else {
			r, err := getURL(filename)
			if err != nil {
				fatal(exitFetchURL, "failed to fetch URL", err)
			}
			raw = r
		}
	}

	if *ungronFlag {
		ungron(raw)
	} else {
		gron(raw)
	}

}

func gron(r io.Reader) {

	b, err := ioutil.ReadAll(r)
	if err != nil {
		fatal(exitReadInput, "failed to read input", err)
	}

	// The 'JSON' might be an object, array or scalar, so the
	// best we can do for now is an empty interface type
	var top interface{}

	err = json.Unmarshal(b, &top)
	if err != nil {
		fatal(exitJSONDecode, "failed to decode JSON", err)
	}

	ss, err := makeStatements("json", top)
	if err != nil {
		fatal(exitFormStatements, "failed to form statements", err)
	}

	// Go's maps do not have well-defined ordering, but we want a consistent
	// output for a given input, so we must sort the statements
	sort.Sort(ss)

	for _, s := range ss {
		fmt.Println(s)
	}

	os.Exit(exitOK)
}

func ungron(r io.Reader) {
	scanner := bufio.NewScanner(r)

	// Get all the idividually parsed statements
	var parsed []interface{}
	for scanner.Scan() {
		l := newLexer(scanner.Text())
		u, err := ungronTokens(l.lex())

		if err != nil {
			fatal(exitUnknown, "TODO: Add proper error for ungron tokens", err)
		}

		parsed = append(parsed, u)
	}
	// TODO: Handle any scanner errors

	if len(parsed) == 0 {
		fatal(exitUnknown, "TODO: Add proper error for no parsed statements", nil)
	}

	merged := parsed[0]
	for _, p := range parsed[1:] {
		m, err := recursiveMerge(merged, p)
		if err != nil {
			fatal(exitUnknown, "TODO: Add proper error for merging statements", err)
		}
		merged = m
	}

	// If there's only one top level key and it's "json", make that the top level thing
	mergedMap, ok := merged.(map[string]interface{})
	if ok {
		if len(mergedMap) == 1 {
			if _, exists := mergedMap["json"]; exists {
				merged = mergedMap["json"]
			}
		}
	}

	j, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		fatal(exitUnknown, "TODO: Add proper error for JSON marshal failure", err)
	}

	fmt.Printf("%s\n", j)

	os.Exit(exitOK)

}

func fatal(code int, msg string, err error) {
	fmt.Fprintf(os.Stderr, "%s (%s)\n", msg, err)
	os.Exit(code)
}
