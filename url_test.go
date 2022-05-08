package main

import (
	"os"
	"testing"
)

func TestValidURL(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{"http://test.com", true},
		{"https://test.com", true},
		{"HttPs://test.com", true},
		{"/test/test.com", false},
		{"", false},
	}

	for _, test := range tests {
		have := validURL(test.url)
		if have != test.want {
			t.Errorf("Want %t for validURL(%s); have %t", test.want, test.url, have)
		}
	}
}

func TestConfigureProxyHttp(t *testing.T) {
	tests := []struct {
		url          string
		httpProxy    string
		envHttpProxy string
		noProxy      string
		envNoProxy   string
		hasProxy     bool
	}{
		// http proxy via env variables
		{"http://test1.com", undefinedProxy, "http://localhost:1234", undefinedProxy, "", true},
		{"https://test1.com", undefinedProxy, "http://localhost:1234", undefinedProxy, "", false},
		{"schema://test1.com", undefinedProxy, "http://localhost:1234", undefinedProxy, "", false},

		// http proxy with env variables, overwritten by argument
		{"http://test2.com", "", "http://localhost:1234", undefinedProxy, "", false},
		{"https://test2.com", "", "http://localhost:1234", undefinedProxy, "", false},
		{"schema://test2.com", "", "http://localhost:1234", undefinedProxy, "", false},

		// http proxy with env variables, domain excluded by no_proxy
		{"http://test3.com", undefinedProxy, "http://localhost:1234", undefinedProxy, "test3.com,foobar3.com", false},
		{"http://foobar3.com", undefinedProxy, "http://localhost:1234", undefinedProxy, "test3.com,foobar3.com", false},
		{"http://test.foobar3.com", undefinedProxy, "http://localhost:1234", undefinedProxy, ".foobar3.com", false},
		{"https://test3.com", undefinedProxy, "http://localhost:1234", undefinedProxy, "test3.com,foobar3.com", false},
		{"schema://test3.com", undefinedProxy, "http://localhost:1234", undefinedProxy, "test3.com,foobar3.com", false},

		// http proxy with env variables, domain excluded by no_proxy, overwritten by argument
		{"http://test4.com", undefinedProxy, "http://localhost:1234", "", "test4.com,foobar4.com", true},
		{"http://foobar4.com", undefinedProxy, "http://localhost:1234", "", "test4.com,foobar4.com", true},
		{"http://test.foobar4.com", undefinedProxy, "http://localhost:1234", "", ".foobar4.com", true},
		{"https://test4.com", undefinedProxy, "http://localhost:1234", "", "test4.com,foobar4.com", false},
		{"schema://test4.com", undefinedProxy, "http://localhost:1234", "", "test4.com,foobar4.com", false},
	}

	for _, test := range tests {
		os.Setenv("http_proxy", test.envHttpProxy)
		os.Setenv("no_proxy", test.envNoProxy)
		proxy := configureProxy(test.url, test.httpProxy, test.noProxy)
		hasProxy := proxy != nil
		if hasProxy != test.hasProxy {
			t.Errorf("Want %t for configureProxy; have %t; %v", test.hasProxy, hasProxy, test)
		}
		os.Unsetenv("http_proxy")
		os.Unsetenv("no_proxy")
	}
}

func TestConfigureProxyHttps(t *testing.T) {
	tests := []struct {
		url           string
		httpsProxy    string
		envHttpsProxy string
		noProxy       string
		envNoProxy    string
		hasProxy      bool
	}{
		// https proxy via env variables
		{"http://test1.com", undefinedProxy, "http://localhost:1234", undefinedProxy, "", false},
		{"https://test1.com", undefinedProxy, "http://localhost:1234", undefinedProxy, "", true},
		{"schema://test1.com", undefinedProxy, "http://localhost:1234", undefinedProxy, "", false},

		// https proxy with env variables, overwritten by argument
		{"http://test2.com", "", "http://localhost:1234", undefinedProxy, "", false},
		{"https://test2.com", "", "http://localhost:1234", undefinedProxy, "", false},
		{"schema://test2.com", "", "http://localhost:1234", undefinedProxy, "", false},

		// https proxy with env variables, domain excluded by no_proxy
		{"http://test3.com", undefinedProxy, "http://localhost:1234", undefinedProxy, "test3.com,foobar3.com", false},
		{"http://foobar3.com", undefinedProxy, "http://localhost:1234", undefinedProxy, "test3.com,foobar3.com", false},
		{"http://test.foobar3.com", undefinedProxy, "http://localhost:1234", undefinedProxy, ".foobar3.com", false},
		{"https://test3.com", undefinedProxy, "http://localhost:1234", undefinedProxy, "test3.com,foobar3.com", false},
		{"schema://test3.com", undefinedProxy, "http://localhost:1234", undefinedProxy, "test3.com,foobar3.com", false},

		// https proxy with env variables, domain excluded by no_proxy, overwritten by argument
		{"http://test4.com", undefinedProxy, "http://localhost:1234", "", "test4.com,foobar4.com", false},
		{"http://foobar4.com", undefinedProxy, "http://localhost:1234", "", "test4.com,foobar4.com", false},
		{"http://test.foobar4.com", undefinedProxy, "http://localhost:1234", "", ".foobar4.com", false},
		{"https://test4.com", undefinedProxy, "http://localhost:1234", "", "test4.com,foobar4.com", true},
		{"schema://test4.com", undefinedProxy, "http://localhost:1234", "", "test4.com,foobar4.com", false},
	}

	for _, test := range tests {
		os.Setenv("https_proxy", test.envHttpsProxy)
		os.Setenv("no_proxy", test.envNoProxy)
		proxy := configureProxy(test.url, test.httpsProxy, test.noProxy)
		hasProxy := proxy != nil
		if hasProxy != test.hasProxy {
			t.Errorf("Want %t for configureProxy; have %t; %v", test.hasProxy, hasProxy, test)
		}
		os.Unsetenv("https_proxy")
		os.Unsetenv("no_proxy")
	}
}
