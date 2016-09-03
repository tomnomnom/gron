package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/fatih/color"
	"github.com/nwidger/jsoncolor"
	"github.com/pkg/errors"
)

const (
	exitOK = iota
	exitOpenFile
	exitReadInput
	exitFormStatements
	exitFetchURL
	exitParseStatements
	exitJSONEncode
)

var (
	strColor   = color.New(color.FgYellow)
	braceColor = color.New(color.FgMagenta)
	bareColor  = color.New(color.FgBlue, color.Bold)
	numColor   = color.New(color.FgRed)
	boolColor  = color.New(color.FgCyan)
)

func init() {
	flag.Usage = func() {
		h := "Transform JSON (from a file, URL, or stdin) into discrete assignments to make it greppable\n\n"

		h += "Usage:\n"
		h += "  gron [OPTIONS] [FILE|URL|-]\n\n"

		h += "Options:\n"
		h += "  -u, --ungron     Reverse the operation (turn assignments back into JSON)\n"
		h += "  -m, --monochrome Monochrome (don't colorize output)\n\n"

		h += "Exit Codes:\n"
		h += fmt.Sprintf("  %d\t%s\n", exitOK, "OK")
		h += fmt.Sprintf("  %d\t%s\n", exitOpenFile, "Failed to open file")
		h += fmt.Sprintf("  %d\t%s\n", exitReadInput, "Failed to read input")
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

func main() {
	var (
		ungronFlag     bool
		monochromeFlag bool
	)

	flag.BoolVar(&ungronFlag, "ungron", false, "Turn statements into JSON instead")
	flag.BoolVar(&ungronFlag, "u", false, "Turn statements into JSON instead")
	flag.BoolVar(&monochromeFlag, "monochrome", false, "Monochrome (don't colorize output)")
	flag.BoolVar(&monochromeFlag, "m", false, "Monochrome (don't colorize output)")

	flag.Parse()

	var raw io.Reader

	filename := flag.Arg(0)
	if filename == "" || filename == "-" {
		raw = os.Stdin
	} else {
		if !validURL(filename) {
			r, err := os.Open(filename)
			if err != nil {
				fatal(exitOpenFile, err)
			}
			raw = r
		} else {
			r, err := getURL(filename)
			if err != nil {
				fatal(exitFetchURL, err)
			}
			raw = r
		}
	}

	var a actionFn = gron
	if ungronFlag {
		a = ungron
	}
	exitCode, err := a(raw, os.Stdout, monochromeFlag)

	if exitCode != exitOK {
		fatal(exitCode, err)
	}

	os.Exit(exitOK)
}

type actionFn func(io.Reader, io.Writer, bool) (int, error)

func gron(r io.Reader, w io.Writer, monochrome bool) (int, error) {

	formatter = colorFormatter{}
	if monochrome {
		formatter = monoFormatter{}
	}

	ss, err := makeStatementsFromJSON(r)
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

func ungron(r io.Reader, w io.Writer, monochrome bool) (int, error) {
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
		return exitParseStatements, err
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

	// Marshal the output into JSON to display to the user
	j, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		return exitJSONEncode, errors.Wrap(err, "failed to convert statements to JSON")
	}

	// If the output isn't monochrome, add color to the JSON
	if !monochrome {
		c, err := colorizeJSON(j)

		// If we failed to colorize the JSON for whatever reason,
		// we'll just fall back to monochrome output, otherwise
		// replace the monochrome JSON with glorious technicolor
		if err == nil {
			j = c
		}
	}

	fmt.Fprintf(w, "%s\n", j)

	return exitOK, nil
}

func colorizeJSON(src []byte) ([]byte, error) {
	out := &bytes.Buffer{}
	f := jsoncolor.NewFormatter()

	f.StringColor = strColor
	f.ObjectColor = braceColor
	f.ArrayColor = braceColor
	f.FieldColor = bareColor
	f.NumberColor = numColor
	f.TrueColor = boolColor
	f.FalseColor = boolColor
	f.NullColor = boolColor

	err := f.Format(out, src)
	if err != nil {
		return out.Bytes(), err
	}
	return out.Bytes(), nil
}

func fatal(code int, err error) {
	fmt.Fprintf(os.Stderr, "%s\n", err)
	os.Exit(code)
}
