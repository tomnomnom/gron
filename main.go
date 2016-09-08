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

// Exit codes
const (
	exitOK = iota
	exitOpenFile
	exitReadInput
	exitFormStatements
	exitFetchURL
	exitParseStatements
	exitJSONEncode
)

// Option bitfields
const (
	optMonochrome = iota + 1
	optNoSort
)

// Output colors
var (
	strColor   = color.New(color.FgYellow)
	braceColor = color.New(color.FgMagenta)
	bareColor  = color.New(color.FgBlue, color.Bold)
	numColor   = color.New(color.FgRed)
	boolColor  = color.New(color.FgCyan)
)

// gronVersion stores the current gron version, set at build
// time with the ldflags -X option
var gronVersion = "dev"

func init() {
	flag.Usage = func() {
		h := "Transform JSON (from a file, URL, or stdin) into discrete assignments to make it greppable\n\n"

		h += "Usage:\n"
		h += "  gron [OPTIONS] [FILE|URL|-]\n\n"

		h += "Options:\n"
		h += "  -u, --ungron     Reverse the operation (turn assignments back into JSON)\n"
		h += "  -m, --monochrome Monochrome (don't colorize output)\n"
		h += "      --no-sort    Don't sort output (faster)\n"
		h += "      --version    Print version information\n\n"

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
		noSortFlag     bool
		versionFlag    bool
	)

	flag.BoolVar(&ungronFlag, "ungron", false, "")
	flag.BoolVar(&ungronFlag, "u", false, "")
	flag.BoolVar(&monochromeFlag, "monochrome", false, "")
	flag.BoolVar(&monochromeFlag, "m", false, "")
	flag.BoolVar(&noSortFlag, "no-sort", false, "")
	flag.BoolVar(&versionFlag, "version", false, "")

	flag.Parse()

	// Print version information
	if versionFlag {
		fmt.Printf("gron version %s\n", gronVersion)
		os.Exit(exitOK)
	}

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

	var opts int
	if monochromeFlag {
		opts = opts | optMonochrome
	}
	if noSortFlag {
		opts = opts | optNoSort
	}

	var a actionFn = gron
	if ungronFlag {
		a = ungron
	}
	exitCode, err := a(raw, os.Stdout, opts)

	if exitCode != exitOK {
		fatal(exitCode, err)
	}

	os.Exit(exitOK)
}

type actionFn func(io.Reader, io.Writer, int) (int, error)

func gron(r io.Reader, w io.Writer, opts int) (int, error) {

	ss, err := makeStatementsFromJSON(r)
	if err != nil {
		return exitFormStatements, fmt.Errorf("failed to form statements: %s", err)
	}

	// Go's maps do not have well-defined ordering, but we want a consistent
	// output for a given input, so we must sort the statements
	if opts&optNoSort == 0 {
		sort.Sort(ss)
	}

	if opts&optMonochrome > 0 {
		for _, s := range ss {
			fmt.Fprintln(w, s.String())
		}
	} else {
		for _, s := range ss {
			fmt.Fprintln(w, s.colorString())
		}
	}

	return exitOK, nil
}

func ungron(r io.Reader, w io.Writer, opts int) (int, error) {
	scanner := bufio.NewScanner(r)

	// Make a list of statements from the input
	var ss statements
	for scanner.Scan() {
		s := statementFromString(scanner.Text())
		ss.AddFull(s)
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
	if opts&optMonochrome == 0 {
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
