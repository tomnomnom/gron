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
	exitParseStatements
	exitJSONEncode
)

func init() {
	flag.Usage = func() {
		h := "Transform JSON (from a file, URL, or stdin) into discrete assignments to make it greppable\n\n"

		h += "Usage:\n"
		h += "  gron [OPTIONS] [FILE|URL|-]\n\n"

		h += "Options:\n"
		h += "  -u, --ungron\tReverse the operation (turn assignments back into JSON)\n\n"

		h += "Exit Codes:\n"
		h += fmt.Sprintf("  %d\t%s\n", exitOK, "OK")
		h += fmt.Sprintf("  %d\t%s\n", exitOpenFile, "Failed to open file")
		h += fmt.Sprintf("  %d\t%s\n", exitReadInput, "Failed to read input")
		h += fmt.Sprintf("  %d\t%s\n", exitJSONDecode, "Failed to decode JSON")
		h += fmt.Sprintf("  %d\t%s\n", exitFormStatements, "Failed to form statements")
		h += fmt.Sprintf("  %d\t%s\n", exitFetchURL, "Failed to fetch URL")
		h += fmt.Sprintf("  %d\t%s\n", exitParseStatements, "Failed to parse statements")
		h += fmt.Sprintf("  %d\t%s\n", exitJSONEncode, "Failed to encode JSON")
		h += "\n"

		h += "Examples:\n"
		h += "  gron /tmp/apiresponse.json\n"
		h += "  gron http://jsonplaceholder.typicode.com/users/1 \n"
		h += "  curl -s http://jsonplaceholder.typicode.com/users/1 | gron\n"
		h += "  gron http://jsonplaceholder.typicode.com/users/1 | grep company | gron --ungron\n"

		fmt.Fprintf(os.Stderr, h)
	}
}

var ungronFlag bool

func main() {
	flag.BoolVar(&ungronFlag, "ungron", false, "Turn statements into JSON instead")
	flag.BoolVar(&ungronFlag, "u", false, "Turn statements into JSON instead")

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

	var exitCode int
	var err error
	if ungronFlag {
		exitCode, err = ungron(raw, os.Stdout)
	} else {
		exitCode, err = gron(raw, os.Stdout)
	}

	if exitCode != exitOK {
		fatal(exitCode, "Fatal", err)
	}

	os.Exit(exitOK)
}

func gron(r io.Reader, w io.Writer) (int, error) {

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return exitReadInput, fmt.Errorf("failed to read input: %s", err)
	}

	// The 'JSON' might be an object, array or scalar, so the
	// best we can do for now is an empty interface type
	var top interface{}

	err = json.Unmarshal(b, &top)
	if err != nil {
		return exitJSONDecode, fmt.Errorf("failed to decode JSON: %s", err)
	}

	ss, err := makeStatements("json", top)
	if err != nil {
		return exitFormStatements, fmt.Errorf("failed to form statements: %s", err)
	}

	// Go's maps do not have well-defined ordering, but we want a consistent
	// output for a given input, so we must sort the statements
	sort.Sort(ss)

	for _, s := range ss {
		fmt.Fprintln(w, s)
	}

	return exitOK, nil
}

func ungron(r io.Reader, w io.Writer) (int, error) {
	scanner := bufio.NewScanner(r)

	// Make a list of statements from the input
	var ss statements
	for scanner.Scan() {
		ss.AddFull(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return exitReadInput, fmt.Errorf("failed to read input statements")
	}

	// ungron the statements
	merged, err := ss.ungron()
	if err != nil {
		return exitParseStatements, fmt.Errorf("failed to parse input statements")
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
		return exitJSONEncode, fmt.Errorf("failed to convert statements to JSON: %s", err)
	}

	fmt.Fprintf(w, "%s\n", j)

	return exitOK, nil
}

func fatal(code int, msg string, err error) {
	fmt.Fprintf(os.Stderr, "%s (%s)\n", msg, err)
	os.Exit(code)
}
