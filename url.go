package main

import (
	"bufio"
	"io"
	"net/http"
	"strings"
)

func validURL(url string) bool {
	return strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://")
}

func getURL(url string) (io.Reader, error) {

	resp, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	return bufio.NewReader(resp.Body), err
}
