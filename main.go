package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
)

const (
	exitOK = iota
	exitInvalidArgs
)

func main() {
	flag.Parse()

	filename := flag.Arg(0)
	if filename == "" {
		os.Exit(exitInvalidArgs)
	}

	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		todo(err)
	}

	// The 'JSON' be an object, array or scalar,
	// so the best we can do for now is an empty interface type
	var top interface{}
	err = json.Unmarshal(raw, &top)
	if err != nil {
		todo(err)
	}

	ss, err := makeStatements("json", top)
	if err != nil {
		todo(err)
	}

	sort.Sort(ss)

	for _, s := range ss {
		fmt.Println(s)
	}

	os.Exit(exitOK)
}

func todo(err error) {
	log.Fatalf("TODO: %s", err)
}
