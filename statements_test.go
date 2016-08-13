package main

import (
	"bytes"
	"encoding/json"
	"reflect"
	"sort"
	"testing"
)

func TestStatementsSimple(t *testing.T) {

	j := []byte(`{
		"dotted": "A dotted value",
		"a quoted": "value",
		"bool1": true,
		"bool2": false,
		"anull": null,
		"anarr": [1, 1.5],
		"anob": {
			"foo": "bar"
		},
		"else": 1,
		"id": 66912849
	}`)

	ss, err := makeStatementsFromJSON(bytes.NewReader(j))

	if err != nil {
		t.Errorf("Want nil error from makeStatements() but got %s", err)
	}

	wants := []string{
		`json = {};`,
		`json.dotted = "A dotted value";`,
		`json["a quoted"] = "value";`,
		`json.bool1 = true;`,
		`json.bool2 = false;`,
		`json.anull = null;`,
		`json.anarr = [];`,
		`json.anarr[0] = 1;`,
		`json.anarr[1] = 1.5;`,
		`json.anob = {};`,
		`json.anob.foo = "bar";`,
		`json["else"] = 1;`,
		`json.id = 66912849;`,
	}

	t.Logf("Have: %#v", ss)
	for _, want := range wants {
		if !ss.Contains(want) {
			t.Errorf("Statement group should contain `%s` but doesn't", want)
		}
	}

}

func TestStatementsSorting(t *testing.T) {
	want := statements{
		`json.a = true;`,
		`json.b = true;`,
		`json.c[0] = true;`,
		`json.c[2] = true;`,
		`json.c[10] = true;`,
		`json.c[11] = true;`,
		`json.c[21][2] = true;`,
		`json.c[21][11] = true;`,
	}

	have := statements{
		`json.c[11] = true;`,
		`json.c[21][2] = true;`,
		`json.c[0] = true;`,
		`json.c[2] = true;`,
		`json.b = true;`,
		`json.c[10] = true;`,
		`json.c[21][11] = true;`,
		`json.a = true;`,
	}

	sort.Sort(have)

	for i := range want {
		if have[i] != want[i] {
			t.Errorf("Statements sorted incorrectly; want `%s` at index %d, have `%s`", want[i], i, have[i])
		}
	}
}

// A cheeky check of statements sorting using some of Dave Koelle's Alphanum test data
// See here: http://www.davekoelle.com/alphanum.html
func TestStatementsSortAlphanum(t *testing.T) {
	have := statements{
		"1000X Radonius Maximus",
		"10X Radonius",
		"200X Radonius",
		"20X Radonius",
		"20X Radonius Prime",
		"30X Radonius",
		"40X Radonius",
		"Allegia 50 Clasteron",
		"Allegia 500 Clasteron",
		"Allegia 50B Clasteron",
		"Allegia 51 Clasteron",
		"Allegia 6R Clasteron",
		"Alpha 100",
		"Alpha 2",
		"Alpha 200",
		"Alpha 2A",
		"Alpha 2A-8000",
		"Alpha 2A-900",
		"Callisto Morphamax",
		"Callisto Morphamax 500",
		"Callisto Morphamax 5000",
		"Callisto Morphamax 600",
		"Callisto Morphamax 6000 SE",
		"Callisto Morphamax 6000 SE2",
		"Callisto Morphamax 700",
		"Callisto Morphamax 7000",
		"Xiph Xlater 10000",
		"Xiph Xlater 2000",
		"Xiph Xlater 300",
		"Xiph Xlater 40",
		"Xiph Xlater 5",
		"Xiph Xlater 50",
		"Xiph Xlater 500",
		"Xiph Xlater 5000",
		"Xiph Xlater 58",
	}
	want := statements{
		"10X Radonius",
		"20X Radonius",
		"20X Radonius Prime",
		"30X Radonius",
		"40X Radonius",
		"200X Radonius",
		"1000X Radonius Maximus",
		"Allegia 6R Clasteron",
		"Allegia 50 Clasteron",
		"Allegia 50B Clasteron",
		"Allegia 51 Clasteron",
		"Allegia 500 Clasteron",
		"Alpha 2",
		"Alpha 2A",
		"Alpha 2A-900",
		"Alpha 2A-8000",
		"Alpha 100",
		"Alpha 200",
		"Callisto Morphamax",
		"Callisto Morphamax 500",
		"Callisto Morphamax 600",
		"Callisto Morphamax 700",
		"Callisto Morphamax 5000",
		"Callisto Morphamax 6000 SE",
		"Callisto Morphamax 6000 SE2",
		"Callisto Morphamax 7000",
		"Xiph Xlater 5",
		"Xiph Xlater 40",
		"Xiph Xlater 50",
		"Xiph Xlater 58",
		"Xiph Xlater 300",
		"Xiph Xlater 500",
		"Xiph Xlater 2000",
		"Xiph Xlater 5000",
		"Xiph Xlater 10000",
	}

	sort.Sort(have)

	for i := range want {
		if have[i] != want[i] {
			t.Errorf("Statements sorted incorrectly; want `%s` at index %d, have `%s`", want[i], i, have[i])
		}
	}
}

func TestStatementLess(t *testing.T) {
	ss := statements{
		"1",
		"2",
		"20",
		"Alpha 2",
		"Alpha 2AA",
		"Xiph Xlater 50",
		"Xiph Xlater 58",
		"200X Radonius",
		"20X Radonius",
	}

	cases := []struct {
		a    int
		b    int
		want bool
	}{
		{0, 1, true},
		{3, 4, true},
		{4, 3, false},
		{2, 1, false},
		{5, 6, true},
		{6, 5, false},
		{0, 5, true},
		{5, 0, false},
		{7, 8, false},
		{8, 7, true},
	}

	for _, c := range cases {
		have := ss.Less(c.a, c.b)

		if have != c.want {
			t.Errorf("`%s` < `%s` should be %t but isn't", ss[c.a], ss[c.b], c.want)
		}
	}

}

func BenchmarkStatementsLess(b *testing.B) {
	ss := statements{
		`json.c[21][2] = true;`,
		`json.c[21][11] = true;`,
	}

	for i := 0; i < b.N; i++ {
		_ = ss.Less(0, 1)
	}
}

func BenchmarkMakeStatements(b *testing.B) {
	j := []byte(`{
		"dotted": "A dotted value",
		"a quoted": "value",
		"bool1": true,
		"bool2": false,
		"anull": null,
		"anarr": [1, 1.5],
		"anob": {
			"foo": "bar"
		},
		"else": 1
	}`)

	var top interface{}
	err := json.Unmarshal(j, &top)
	if err != nil {
		b.Fatalf("Failed to unmarshal test file: %s", err)
	}

	for i := 0; i < b.N; i++ {
		_, _ = makeStatements("json", top)
	}
}

func TestUngronStatements(t *testing.T) {
	in := statements{
		`json.contact = {};`,
		`json.contact["e-mail"][0] = "mail@tomnomnom.com";`,
		`json.contact["e-mail"][1] = "test@tomnomnom.com";`,
		`json.contact["e-mail"][3] = "foo@tomnomnom.com";`,
		`json.contact.twitter = "@TomNomNom";`,
	}

	want := map[string]interface{}{
		"json": map[string]interface{}{
			"contact": map[string]interface{}{
				"e-mail": []interface{}{
					0: "mail@tomnomnom.com",
					1: "test@tomnomnom.com",
					3: "foo@tomnomnom.com",
				},
				"twitter": "@TomNomNom",
			},
		},
	}

	have, err := in.ungron()

	if err != nil {
		t.Fatalf("want nil error but have: %s", err)
	}

	t.Logf("Have: %#v", have)
	t.Logf("Want: %#v", want)

	eq := reflect.DeepEqual(have, want)
	if !eq {
		t.Errorf("have and want are not equal")
	}
}
