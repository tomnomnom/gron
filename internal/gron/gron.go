package gron

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"sort"

	"github.com/fatih/color"
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

// Gron is the default action. Given JSON as the input it returns a list
// of assignment statements. Possible options are optNoSort and optMonochrome
func Gron(r io.Reader, w io.Writer, opts int) (int, error) {
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

// GronStream is like the gron action, but it treats the input as one
// JSON object per line. There's a bit of code duplication from the
// gron action, but it'd be fairly messy to combine the two actions
func GronStream(r io.Reader, w io.Writer, opts int) (int, error) {
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
