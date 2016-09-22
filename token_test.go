package main

import "testing"

var cases = []struct {
	in   string
	want string
}{
	{"foo", `"foo"`},
	{"<3", `"<3"`},
	{"&", `"&"`},
	{"\b", `"\b"`},
	{"\f", `"\f"`},
	{"\n", `"\n"`},
	{"\r", `"\r"`},
	{"\t", `"\t"`},
	{"wat \u001e", `"wat \u001E"`},
	{"Hello, 世界", `"Hello, 世界"`},
	{`Quotes are fun """`, `"Quotes are fun \"\"\""`},
	{`Slashes \\\`, `"Slashes \\\\\\"`},
}

func TestQuoteString(t *testing.T) {

	for _, c := range cases {
		have := quoteString(c.in)

		if have != c.want {
			t.Errorf("have `%s` for quoteString(%s); want %s", have, c.in, c.want)
		}
	}
}

func BenchmarkQuoteString(b *testing.B) {

	for i := 0; i < b.N; i++ {
		for _, c := range cases {
			_ = quoteString(c.in)
		}
	}
}
