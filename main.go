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
	"github.com/mattn/go-colorable"
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
	optMonochrome = 1 << iota
	optNoSort
	optJSON
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
		h += "  -c, --colorize   Colorize output (default on tty)\n"
		h += "  -m, --monochrome Monochrome (don't colorize output)\n"
		h += "  -s, --stream     Treat each line of input as a separate JSON object\n"
		h += "  -k, --insecure   Disable certificate validation\n"
		h += "  -j, --json       Represent gron data as JSON stream\n"
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
		colorizeFlag   bool
		monochromeFlag bool
		streamFlag     bool
		noSortFlag     bool
		versionFlag    bool
		insecureFlag   bool
		jsonFlag       bool
	)

	flag.BoolVar(&ungronFlag, "ungron", false, "")
	flag.BoolVar(&ungronFlag, "u", false, "")
	flag.BoolVar(&colorizeFlag, "colorize", false, "")
	flag.BoolVar(&colorizeFlag, "c", false, "")
	flag.BoolVar(&monochromeFlag, "monochrome", false, "")
	flag.BoolVar(&monochromeFlag, "m", false, "")
	flag.BoolVar(&streamFlag, "s", false, "")
	flag.BoolVar(&streamFlag, "stream", false, "")
	flag.BoolVar(&noSortFlag, "no-sort", false, "")
	flag.BoolVar(&versionFlag, "version", false, "")
	flag.BoolVar(&insecureFlag, "k", false, "")
	flag.BoolVar(&insecureFlag, "insecure", false, "")
	flag.BoolVar(&jsonFlag, "j", false, "")
	flag.BoolVar(&jsonFlag, "json", false, "")

	flag.Parse()

	// Print version information
	if versionFlag {
		fmt.Printf("gron version %s\n", gronVersion)
		os.Exit(exitOK)
	}

	// Determine what the program's input should be:
	// file, HTTP URL or stdin
	var rawInput io.Reader
	filename := flag.Arg(0)
	if filename == "" || filename == "-" {
		rawInput = os.Stdin
	} else if validURL(filename) {
		r, err := getURL(filename, insecureFlag)
		if err != nil {
			fatal(exitFetchURL, err)
		}
		rawInput = r
	} else {
		r, err := os.Open(filename)
		if err != nil {
			fatal(exitOpenFile, err)
		}
		rawInput = r
	}

	var opts int
	// The monochrome option should be forced if the output isn't a terminal
	// to avoid doing unnecessary work calling the color functions
	switch {
	case colorizeFlag:
		color.NoColor = false
	case monochromeFlag || color.NoColor:
		opts = opts | optMonochrome
	}
	if noSortFlag {
		opts = opts | optNoSort
	}
	if jsonFlag {
		opts = opts | optJSON
	}

	// Pick the appropriate action: gron, ungron or gronStream
	var a actionFn = gron
	if ungronFlag {
		a = ungron
	} else if streamFlag {
		a = gronStream
	}
	exitCode, err := a(rawInput, colorable.NewColorableStdout(), opts)

	if exitCode != exitOK {
		fatal(exitCode, err)
	}

	os.Exit(exitOK)
}

// an actionFn represents a main action of the program, it accepts
// an input, output and a bitfield of options; returning an exit
// code and any error that occurred
type actionFn func(io.Reader, io.Writer, int) (int, error)

// gron is the default action. Given JSON as the input it returns a list
// of assignment statements. Possible options are optNoSort and optMonochrome
func gron(r io.Reader, w io.Writer, opts int) (int, error) {
	var err error

	var conv statementconv
	if opts&optMonochrome > 0 {
		conv = statementToString
	} else {
		conv = statementToColorString
	}

	ss, err := statementsFromJSON(r, statement{{"json", typBare}})
	if err != nil {
		goto out
	}

	// Go's maps do not have well-defined ordering, but we want a consistent
	// output for a given input, so we must sort the statements
	if opts&optNoSort == 0 {
		sort.Sort(ss)
	}

	for _, s := range ss {
		if opts&optJSON > 0 {
			s, err = s.jsonify()
			if err != nil {
				goto out
			}
		}
		fmt.Fprintln(w, conv(s))
	}

out:
	if err != nil {
		return exitFormStatements, fmt.Errorf("failed to form statements: %s", err)
	}
	return exitOK, nil
}

// gronStream is like the gron action, but it treats the input as one
// JSON object per line. There's a bit of code duplication from the
// gron action, but it'd be fairly messy to combine the two actions
func gronStream(r io.Reader, w io.Writer, opts int) (int, error) {
	var err error
	errstr := "failed to form statements"
	var i int
	var sc *bufio.Scanner
	var buf []byte

	var conv func(s statement) string
	if opts&optMonochrome > 0 {
		conv = statementToString
	} else {
		conv = statementToColorString
	}

	// Helper function to make the prefix statements for each line
	makePrefix := func(index int) statement {
		return statement{
			{"json", typBare},
			{"[", typLBrace},
			{fmt.Sprintf("%d", index), typNumericKey},
			{"]", typRBrace},
		}
	}

	// The first line of output needs to establish that the top-level
	// thing is actually an array...
	top := statement{
		{"json", typBare},
		{"=", typEquals},
		{"[]", typEmptyArray},
		{";", typSemi},
	}

	if opts&optJSON > 0 {
		top, err = top.jsonify()
		if err != nil {
			goto out
		}
	}

	fmt.Fprintln(w, conv(top))

	// Read the input line by line
	sc = bufio.NewScanner(r)
	buf = make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	i = 0
	for sc.Scan() {

		line := bytes.NewBuffer(sc.Bytes())

		var ss statements
		ss, err = statementsFromJSON(line, makePrefix(i))
		i++
		if err != nil {
			goto out
		}

		// Go's maps do not have well-defined ordering, but we want a consistent
		// output for a given input, so we must sort the statements
		if opts&optNoSort == 0 {
			sort.Sort(ss)
		}

		for _, s := range ss {
			if opts&optJSON > 0 {
				s, err = s.jsonify()
				if err != nil {
					goto out
				}

			}
			fmt.Fprintln(w, conv(s))
		}
	}
	if err = sc.Err(); err != nil {
		errstr = "error reading multiline input: %s"
	}

out:
	if err != nil {
		return exitFormStatements, fmt.Errorf(errstr+": %s", err)
	}
	return exitOK, nil

}

// ungron is the reverse of gron. Given assignment statements as input,
// it returns JSON. The only option is optMonochrome
func ungron(r io.Reader, w io.Writer, opts int) (int, error) {
	scanner := bufio.NewScanner(r)
	var maker statementmaker

	if opts&optJSON > 0 {
		maker = statementFromJSONSpec
	} else {
		maker = statementFromStringMaker
	}

	// Make a list of statements from the input
	var ss statements
	for scanner.Scan() {
		s, err := maker(scanner.Text())
		if err != nil {
			return exitParseStatements, err
		}
		ss.add(s)
	}
	if err := scanner.Err(); err != nil {
		return exitReadInput, fmt.Errorf("failed to read input statements")
	}

	// turn the statements into a single merged interface{} type
	merged, err := ss.toInterface()
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
	out := &bytes.Buffer{}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	err = enc.Encode(merged)
	if err != nil {
		return exitJSONEncode, errors.Wrap(err, "failed to convert statements to JSON")
	}
	j := out.Bytes()

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

	// For whatever reason, the monochrome version of the JSON
	// has a trailing newline character, but the colorized version
	// does not. Strip the whitespace so that neither has the newline
	// character on the end, and then we'll add a newline in the
	// Fprintf below
	j = bytes.TrimSpace(j)

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
