//go:build example

package main

import (
	"encoding/json"
	"fmt"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/hydrz/jsoninline"
)

// Example types mirroring test fixtures; used to demonstrate schema generation.
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
}

type USA struct {
	City      string     `json:"city,omitempty"`
	State     string     `json:"state,omitempty"`
	NestedBar *NestedBar `json:",inline"`
}

type NestedFoo struct {
	FooField string `json:"foo_field,omitempty"`
}

type NestedBar struct {
	BarField string `json:"bar_field,omitempty"`
}

func main() {
	schema, err := jsoninline.For[User](&jsonschema.ForOptions{})
	if err != nil {
		panic(err)
	}

	// Marshal the schema to pretty JSON and print it.
	b, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}
