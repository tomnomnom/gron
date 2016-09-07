package main

import "testing"

func TestPrefixHappy(t *testing.T) {
	var f monoFormatter

	tests := []struct {
		prev string
		next interface{}
		want string
	}{
		{"j", 123, "j[123]"},
		{"j", 1, "j[1]"},
		{"j", "dotted", "j.dotted"},
		{"j", "un-dotted", "j[\"un-dotted\"]"},
	}

	for _, test := range tests {
		r, err := f.prefix(test.prev, test.next)
		if err != nil {
			t.Errorf("Want nil error from f.prefix(%s, %#v); have: %s", test.prev, test.next, err)
		}
		if r != test.want {
			t.Errorf("Want %s from f.prefix(%s, %#v); have: %s", test.want, test.prev, test.next, r)
		}
	}
}

func BenchmarkFormatterMakePrefixUnquoted(b *testing.B) {
	var f monoFormatter
	for i := 0; i < b.N; i++ {
		_, _ = f.prefix("json", "isunquoted")
	}
}

func BenchmarkFormatterMakePrefixQuoted(b *testing.B) {
	var f monoFormatter
	for i := 0; i < b.N; i++ {
		_, _ = f.prefix("json", "this-is-quoted")
	}
}

func BenchmarkFormatterMakePrefixInt(b *testing.B) {
	var f monoFormatter
	for i := 0; i < b.N; i++ {
		_, _ = f.prefix("json", 212)
	}
}

func BenchmarkFormatterAssignmentString(b *testing.B) {
	var f monoFormatter
	for i := 0; i < b.N; i++ {
		_ = f.assignment("json.foo", "bar")
	}
}

func BenchmarkFormatterAssignmentMap(b *testing.B) {
	var f monoFormatter
	val := make(map[string]interface{})
	for i := 0; i < b.N; i++ {
		_ = f.assignment("json.foo", val)
	}
}

func BenchmarkFormatterValueString(b *testing.B) {
	var f monoFormatter
	for i := 0; i < b.N; i++ {
		_ = f.value("a string")
	}
}
