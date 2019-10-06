package gron

import (
	"bytes"
	"encoding/json"
	"reflect"
	"sort"
	"testing"
)

func statementsFromStringSlice(strs []string) statements {
	ss := make(statements, len(strs))
	for i, str := range strs {
		ss[i] = statementFromString(str)
	}
	return ss
}

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
		"id": 66912849,
		"": 2
	}`)

	ss, err := statementsFromJSON(bytes.NewReader(j), statement{{"json", typBare}})

	if err != nil {
		t.Errorf("Want nil error from makeStatementsFromJSON() but got %s", err)
	}

	wants := statementsFromStringSlice([]string{
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
		`json[""] = 2;`,
	})

	t.Logf("Have: %#v", ss)
	for _, want := range wants {
		if !ss.Contains(want) {
			t.Errorf("Statement group should contain `%s` but doesn't", want)
		}
	}

}

func TestStatementsSorting(t *testing.T) {
	want := statementsFromStringSlice([]string{
		`json.a = true;`,
		`json.b = true;`,
		`json.c[0] = true;`,
		`json.c[2] = true;`,
		`json.c[10] = true;`,
		`json.c[11] = true;`,
		`json.c[21][2] = true;`,
		`json.c[21][11] = true;`,
	})

	have := statementsFromStringSlice([]string{
		`json.c[11] = true;`,
		`json.c[21][2] = true;`,
		`json.c[0] = true;`,
		`json.c[2] = true;`,
		`json.b = true;`,
		`json.c[10] = true;`,
		`json.c[21][11] = true;`,
		`json.a = true;`,
	})

	sort.Sort(have)

	for i := range want {
		if !reflect.DeepEqual(have[i], want[i]) {
			t.Errorf("Statements sorted incorrectly; want `%s` at index %d, have `%s`", want[i], i, have[i])
		}
	}
}

func BenchmarkStatementsLess(b *testing.B) {
	ss := statementsFromStringSlice([]string{
		`json.c[21][2] = true;`,
		`json.c[21][11] = true;`,
	})

	for i := 0; i < b.N; i++ {
		_ = ss.Less(0, 1)
	}
}

func BenchmarkFill(b *testing.B) {
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
		ss := make(statements, 0)
		ss.fill(statement{{"json", typBare}}, top)
	}
}

func TestUngronStatementsSimple(t *testing.T) {
	in := statementsFromStringSlice([]string{
		`json.contact = {};`,
		`json.contact["e-mail"][0] = "mail@tomnomnom.com";`,
		`json.contact["e-mail"][1] = "test@tomnomnom.com";`,
		`json.contact["e-mail"][3] = "foo@tomnomnom.com";`,
		`json.contact.twitter = "@TomNomNom";`,
	})

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

	have, err := in.toInterface()

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

func TestUngronStatementsInvalid(t *testing.T) {
	cases := []statements{
		statementsFromStringSlice([]string{``}),
		statementsFromStringSlice([]string{`this isn't a statement at all`}),
		statementsFromStringSlice([]string{`json[0] = 1;`, `json.bar = 1;`}),
	}

	for _, c := range cases {
		_, err := c.toInterface()
		if err == nil {
			t.Errorf("want non-nil error; have nil")
		}
	}
}

func TestStatement(t *testing.T) {
	s := statement{
		token{"json", typBare},
		token{".", typDot},
		token{"foo", typBare},
		token{"=", typEquals},
		token{"2", typNumber},
		token{";", typSemi},
	}

	have := s.String()
	want := "json.foo = 2;"
	if have != want {
		t.Errorf("have: `%s` want: `%s`", have, want)
	}
}
