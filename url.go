package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"
)

func validURL(url string) bool {
	r := regexp.MustCompile("(?i)^http(?:s)?://")
	return r.MatchString(url)
}

func getURL(url string, insecure bool) (io.Reader, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
	}
	client := http.Client{
		Transport: tr,
		Timeout:   20 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", fmt.Sprintf("gron/%s", gronVersion))
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	return bufio.NewReader(resp.Body), err
}
