package jsoninline_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/hydrz/jsoninline"
)

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

	tests := []struct {
		substr   string
		expected bool
	}{
		{`"China":`, false},
		{`"USA":`, false},
		{`"NestedFoo":`, false},
		{`"NestedBar":`, false},
		{`"province": "Guangdong"`, true},
		{`"city": "Shenzhen"`, true},
		{`"state": "California"`, true},
		{`"city": "Los Angeles"`, true},
		{`"foo_field": "FooValue"`, true},
		{`"bar_field": "BarValue"`, true},
	}

	for _, tt := range tests {
		contains := strings.Contains(output, tt.substr)
		if contains != tt.expected {
			if tt.expected {
				t.Errorf("Expected output to contain %q, but it did not.\nOutput: %s", tt.substr, output)
			} else {
				t.Errorf("Did not expect output to contain %q.\nOutput: %s", tt.substr, output)
			}
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
