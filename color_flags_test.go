package main

import (
	"flag"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestColorFlagAliases(t *testing.T) {
	for _, tc := range []struct {
		alias    string
		original string
	}{
		{"--color", "--colorize"},
		{"--no-color", "--monochrome"},
	} {
		t.Run(tc.alias, func(t *testing.T) {
			run := func(option string) []byte {
				cmd := exec.Command(os.Args[0], "-test.run=^TestColorFlagHelper$", "--", option)
				cmd.Env = append(os.Environ(), "GRON_COLOR_FLAG_HELPER=1")
				cmd.Stdin = strings.NewReader(`{"hello":"world"}`)
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("CLI failed: %v\n%s", err, out)
				}
				if !strings.Contains(string(out), "world") {
					t.Errorf("missing input value in output: %q", out)
				}
				return out
			}
			alias, original := run(tc.alias), run(tc.original)
			if string(alias) != string(original) {
				t.Errorf("%s output %q differs from %s output %q", tc.alias, alias, tc.original, original)
			}
			if tc.alias == "--no-color" && strings.Contains(string(alias), "\x1b[") {
				t.Errorf("--no-color emitted ANSI escapes: %q", alias)
			}
		})
	}
}

func TestColorFlagHelper(t *testing.T) {
	if os.Getenv("GRON_COLOR_FLAG_HELPER") != "1" {
		return
	}
	os.Args = append([]string{"gron"}, os.Args[3:]...)
	flag.CommandLine = flag.NewFlagSet("gron", flag.ExitOnError)
	main()
}
