package jsoninline_test

import (
	"maps"
	"slices"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/hydrz/jsoninline"
)

func TestSchema(t *testing.T) {
	schema, err := jsoninline.For[User](&jsonschema.ForOptions{})
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	keys := slices.Collect(maps.Keys(schema.Properties))

	tests := []struct {
		key      string
		expected bool
	}{
		{"China", false},
		{"USA", false},
		{"city", true},
		{"province", true},
		{"state", true},
		{"NestedFoo", false},
		{"NestedBar", false},
		{"foo_field", true},
		{"bar_field", true},
	}

	for _, tt := range tests {
		if got := slices.Contains(keys, tt.key); got != tt.expected {
			t.Errorf("For key '%s', expected presence: %v, got: %v", tt.key, tt.expected, got)
		}
	}
}

type SchemaWrapper struct {
	Schema string `json:"$schema"`
	Users  []User `json:"users"`
}

func TestSchemaWithWrapper(t *testing.T) {
	schema, err := jsoninline.For[SchemaWrapper](&jsonschema.ForOptions{})
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	usersSchema, ok := schema.Properties["users"]
	if !ok {
		t.Fatalf("Expected 'users' property in schema")
	}

	keys := slices.Collect(maps.Keys(usersSchema.Items.Properties))

	tests := []struct {
		key      string
		expected bool
	}{
		{"China", false},
		{"USA", false},
		{"city", true},
		{"province", true},
		{"state", true},
		{"NestedFoo", false},
		{"NestedBar", false},
		{"foo_field", true},
		{"bar_field", true},
	}

	for _, tt := range tests {
		if got := slices.Contains(keys, tt.key); got != tt.expected {
			t.Errorf("For key '%s', expected presence: %v, got: %v", tt.key, tt.expected, got)
		}
	}
}
