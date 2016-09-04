package main

import (
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
