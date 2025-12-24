package jsoninline_test

import (
	"maps"
	"slices"
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

type SchemaWrapper struct {
	Schema string `json:"$schema"`
	Users  []User `json:"users"`
}

func TestSchema(t *testing.T) {
	schema, err := jsoninline.For[SchemaWrapper](&jsonschema.ForOptions{})
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	tests := []struct {
		schema *jsonschema.Schema
		props  []string
	}{
		{
			schema: schema,
			props:  []string{"$schema", "users"},
		},
		{
			schema: schema.Properties["users"].Items,
			props:  []string{"id", "name", "email"},
		},
		{
			schema: schema.Properties["users"].Items.OneOf[0],
			props:  []string{"province", "city"},
		},
		{
			schema: schema.Properties["users"].Items.OneOf[1],
			props:  []string{"state", "city"},
		},
	}

	for _, tt := range tests {
		actualProps := slices.Collect(maps.Keys(tt.schema.Properties))
		slices.Sort(actualProps)
		slices.Sort(tt.props)
		if !slices.Equal(actualProps, tt.props) {
			t.Errorf("Expected properties %v, got %v", tt.props, actualProps)
		}
	}
}
