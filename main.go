package main

import (
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
)

func init() {
	flag.Usage = func() {
		h := "Transform JSON from a file (or stdin) into discrete assignments to make it greppable\n\n"

		h += "Usage:\n"
		h += "  gron [jsonfile]\n\n"

		h += "Exit Codes:\n"
		h += fmt.Sprintf("  %d\t%s\n", exitOK, "OK")
		h += fmt.Sprintf("  %d\t%s\n", exitOpenFile, "Failed to open file")
		h += fmt.Sprintf("  %d\t%s\n", exitReadInput, "Failed to read input")
		h += fmt.Sprintf("  %d\t%s\n", exitJSONDecode, "Failed to decode JSON")
		h += fmt.Sprintf("  %d\t%s\n", exitFormStatements, "Failed to from statements")
		h += "\n"

		h += "Examples:\n"
		h += "  gron /tmp/apiresponse.json\n"
		h += "  curl -s http://headers.jsontest.com/ | gron\n"

		fmt.Fprintf(os.Stderr, h)
	}
}

func main() {
	flag.Parse()

	var raw io.Reader

	filename := flag.Arg(0)
	if filename == "" {
		raw = os.Stdin
	} else {
		r, err := os.Open(filename)
		if err != nil {
			fatal(exitOpenFile, "failed to open file", err)
		}
		raw = r
	}

	b, err := ioutil.ReadAll(raw)
	if err != nil {
		fatal(exitReadInput, "failed to read input", err)
	}

	// The 'JSON' be an object, array or scalar,
	// so the best we can do for now is an empty interface type
	var top interface{}

	err = json.Unmarshal(b, &top)
	if err != nil {
		fatal(exitJSONDecode, "failed to decode JSON", err)
	}

	ss, err := makeStatements("json", top)
	if err != nil {
		fatal(exitFormStatements, "failed to form statements", err)
	}

	sort.Sort(ss)

	for _, s := range ss {
		fmt.Println(s)
	}

	os.Exit(exitOK)
}

func fatal(code int, msg string, err error) {
	fmt.Fprintf(os.Stderr, "%s (%s)\n", msg, err)
	os.Exit(code)
}
