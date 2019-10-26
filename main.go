package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"

	"github.com/tomnomnom/gron/internal/gron"
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
	var a actionFn = gron.Gron
	if ungronFlag {
		a = gron.Ungron
	} else if streamFlag {
		a = gron.GronStream
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

func fatal(code int, err error) {
	fmt.Fprintf(os.Stderr, "%s\n", err)
	os.Exit(code)
}
