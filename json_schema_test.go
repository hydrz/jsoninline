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

	if slices.Contains(keys, "China") {
		t.Errorf("Expected 'China' not to be required in schema")
	}

	if slices.Contains(keys, "USA") {
		t.Errorf("Expected 'USA' not to be required in schema")
	}

	if !slices.Contains(keys, "city") {
		t.Errorf("Expected 'city' to be present in schema")
	}

	if !slices.Contains(keys, "province") {
		t.Errorf("Expected 'province' to be present in schema")
	}

	if !slices.Contains(keys, "state") {
		t.Errorf("Expected 'state' to be present in schema")
	}

	if slices.Contains(keys, "NestedFoo") {
		t.Errorf("Expected 'NestedFoo' not to be required in schema")
	}

	if slices.Contains(keys, "NestedBar") {
		t.Errorf("Expected 'NestedBar' not to be required in schema")
	}

	if !slices.Contains(keys, "foo_field") {
		t.Errorf("Expected 'foo_field' to be present in schema")
	}

	if !slices.Contains(keys, "bar_field") {
		t.Errorf("Expected 'bar_field' to be present in schema")
	}
}
