package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

func validURL(url string) bool {
	r := regexp.MustCompile("(?i)^http(?:s)?://")
	return r.MatchString(url)
}

func configureProxy(url string, proxy string, noProxy string) func(*http.Request) (*neturl.URL, error) {
	cURL, err := neturl.Parse(url)
	if err != nil {
		return nil
	}

	// Direct arguments are superior to environment variables.
	if proxy == undefinedProxy {
		proxy = os.Getenv(fmt.Sprintf("%s_proxy", cURL.Scheme))
	}
	if noProxy == undefinedProxy {
		noProxy = os.Getenv("no_proxy")
	}

	// Skip setting a proxy if no proxy has been set through env variable or
	// argument.
	if proxy == "" {
		return nil
	}

	// Test if any of the hosts mentioned in the noProxy variable or the
	// no_proxy env variable. Skip setting up the proxy if a match is found.
	noProxyHosts := strings.Split(noProxy, ",")
	if len(noProxyHosts) > 0 {
		for _, noProxyHost := range noProxyHosts {
			if len(noProxyHost) == 0 {
				continue
			}
			// Test for direct matches of the hostname.
			if cURL.Host == noProxyHost {
				return nil
			}
			// Match through wildcard-like pattern, e.g. ".foobar.com" should
			// match all subdomains of foobar.com.
			if strings.HasPrefix(noProxyHost, ".") && strings.HasSuffix(cURL.Host, noProxyHost) {
				return nil
			}
		}
	}

	proxyURL, err := neturl.Parse(proxy)
	if err != nil {
		return nil
	}

	return http.ProxyURL(proxyURL)
}

func getURL(url string, insecure bool, proxyURL string, noProxy string) (io.Reader, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
	}
	// Set proxy if defined.
	proxy := configureProxy(url, proxyURL, noProxy)
	if proxy != nil {
		tr.Proxy = proxy
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
