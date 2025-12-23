package jsoninline_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/hydrz/jsoninline"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	China *China `json:",inline"`
	USA   *USA   `json:",inline"`
}

type China struct {
	City      string     `json:"city,omitempty"`
	Province  string     `json:"province,omitempty"`
	NestedFoo *NestedFoo `json:",inline"`
	NestedBar *NestedBar `json:",inline"`
}

type USA struct {
	City      string     `json:"city,omitempty"`
	State     string     `json:"state,omitempty"`
	NestedFoo *NestedFoo `json:",inline"`
	NestedBar *NestedBar `json:",inline"`
}

type NestedFoo struct {
	FooField string `json:"foo_field,omitempty"`
}

type NestedBar struct {
	BarField string `json:"bar_field,omitempty"`
}

func TestMarshaler(t *testing.T) {
	users := []User{
		{
			ID:    1,
			Name:  "Alice",
			Email: "alice@example.com",
			China: &China{
				Province: "Guangdong",
				City:     "Shenzhen",
				NestedFoo: &NestedFoo{
					FooField: "FooValue",
				},
			},
		},
		{
			ID:    2,
			Name:  "Bob",
			Email: "bob@example.com",
			USA: &USA{
				State: "California",
				City:  "Los Angeles",
				NestedBar: &NestedBar{
					BarField: "BarValue",
				},
			},
		},
	}

	nestedBytes, err := json.MarshalIndent(jsoninline.V(users), "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal users: %v", err)
	}

	// check output contains inlined fields
	output := string(nestedBytes)

	unexpectedSubstrings := []string{
		`"China":`,
		`"USA":`,
		`"NestedFoo":`,
		`"NestedBar":`,
	}

	for _, substr := range unexpectedSubstrings {
		if strings.Contains(output, substr) {
			t.Errorf("Did not expect output to contain %q.\nOutput: %s", substr, output)
		}
	}

	expectedSubstrings := []string{
		`"province": "Guangdong"`,
		`"city": "Shenzhen"`,
		`"state": "California"`,
		`"city": "Los Angeles"`,
		`"foo_field": "FooValue"`,
		`"bar_field": "BarValue"`,
	}

	for _, substr := range expectedSubstrings {
		if !strings.Contains(output, substr) {
			t.Errorf("Expected output to contain %q, but it did not.\nOutput: %s", substr, output)
		}
	}
}

func TestUnmarshaler(t *testing.T) {
	jsonData := `[
  {
	"id": 1,
	"name": "Alice",
	"email": "alice@example.com",
	"province": "Guangdong",
	"city": "Shenzhen"
  },
  {
	"id": 2,
	"name": "Bob",
	"email": "bob@example.com",
	"state": "California",
	"city": "Los Angeles"
  }
]`

	var users []*User
	if err := json.Unmarshal([]byte(jsonData), jsoninline.V(&users)); err != nil {
		t.Fatalf("Failed to unmarshal users: %v", err)
	}

	if len(users) != 2 {
		t.Fatalf("Expected 2 users, got %d", len(users))
	}

	alice := users[0]
	if alice.China == nil || alice.China.Province != "Guangdong" || alice.China.City != "Shenzhen" {
		t.Errorf("Alice's China info not unmarshaled correctly: %+v", alice.China)
	}

	bob := users[1]
	if bob.USA == nil || bob.USA.State != "California" || bob.USA.City != "Los Angeles" {
		t.Errorf("Bob's USA info not unmarshaled correctly: %+v", bob.USA)
	}

	schema, err := jsoninline.For[[]User](&jsonschema.ForOptions{})
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	resolved, err := schema.Resolve(&jsonschema.ResolveOptions{})
	if err != nil {
		t.Fatalf("Failed to resolve schema: %v", err)
	}

	var parsed any
	if err := json.Unmarshal([]byte(jsonData), &parsed); err != nil {
		t.Fatalf("Failed to parse jsonData for schema validation: %v", err)
	}
	if err := resolved.Validate(parsed); err != nil {
		t.Errorf("Schema validation failed: %v", err)
	}

}
