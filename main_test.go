package main

import (
	"bytes"
	"encoding/json"
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

func TestUngron(t *testing.T) {
	cases := []struct {
		inFile  string
		outFile string
	}{
		{"testdata/one.gron", "testdata/one.json"},
		{"testdata/two.gron", "testdata/two.json"},
		{"testdata/three.gron", "testdata/three.json"},
	}

	for _, c := range cases {
		wantF, err := ioutil.ReadFile(c.outFile)
		if err != nil {
			t.Fatalf("failed to open want file: %s", err)
		}

		var want interface{}
		err = json.Unmarshal(wantF, &want)
		if err != nil {
			t.Fatalf("failed to unmarshal JSON from want file: %s", err)
		}

		in, err := os.Open(c.inFile)
		if err != nil {
			t.Fatalf("failed to open input file: %s", err)
		}

		out := &bytes.Buffer{}
		code, err := ungron(in, out, true)

		if code != exitOK {
			t.Errorf("want exitOK; have %d", code)
		}
		if err != nil {
			t.Errorf("want nil error; have %s", err)
		}

		var have interface{}
		err = json.Unmarshal(out.Bytes(), &have)
		if err != nil {
			t.Fatalf("failed to unmarshal JSON from ungron output: %s", err)
		}

		if !reflect.DeepEqual(want, have) {
			t.Logf("want: %#v", want)
			t.Logf("have: %#v", have)
			t.Errorf("ungronned %s does not match %s", c.inFile, c.outFile)
		}

	}
}
