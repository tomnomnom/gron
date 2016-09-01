package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestGron(t *testing.T) {
	cases := []struct {
		inFile  string
		outFile string
	}{
		{"testdata/one.json", "testdata/one.gron"},
		{"testdata/two.json", "testdata/two.gron"},
		{"testdata/three.json", "testdata/three.gron"},
	}

	for _, c := range cases {
		in, err := os.Open(c.inFile)
		if err != nil {
			t.Fatalf("failed to open input file: %s", err)
		}

		want, err := ioutil.ReadFile(c.outFile)
		if err != nil {
			t.Fatalf("failed to open want file: %s", err)
		}

		out := &bytes.Buffer{}
		code, err := gron(in, out, true)

		if code != exitOK {
			t.Errorf("want exitOK; have %d", code)
		}
		if err != nil {
			t.Errorf("want nil error; have %s", err)
		}

		if !reflect.DeepEqual(want, out.Bytes()) {
			t.Logf("want: %s", want)
			t.Logf("have: %s", out.Bytes())
			t.Errorf("gronned %s does not match %s", c.inFile, c.outFile)
		}
	}

}
