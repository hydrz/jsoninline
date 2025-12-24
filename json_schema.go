package jsoninline

import (
	"reflect"
	"slices"

	_ "unsafe"

	"github.com/google/jsonschema-go/jsonschema"
)

type jsonInfo struct {
	omit     bool            // unexported or first tag element is "-"
	name     string          // Go field name or first tag element. Empty if omit is true.
	settings map[string]bool // "omitempty", "omitzero", etc.
}

//go:linkname fieldJSONInfo github.com/google/jsonschema-go/jsonschema.fieldJSONInfo
func fieldJSONInfo(f reflect.StructField) jsonInfo

func For[T any](opts *jsonschema.ForOptions) (*jsonschema.Schema, error) {
	t := reflect.TypeFor[T]()
	return ForType(t, opts)
}

func ForType(t reflect.Type, opts *jsonschema.ForOptions) (*jsonschema.Schema, error) {

	schema, err := jsonschema.ForType(t, opts)
	if err != nil {
		return nil, err
	}

	if err := handleInline(t, schema); err != nil {
		return nil, err
	}

	return schema, nil
}

func handleInline(t reflect.Type, schema *jsonschema.Schema) error {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.Array, reflect.Slice:
		elemType := t.Elem()
		return handleInline(elemType, schema.Items)
	case reflect.Struct:
		for _, field := range reflect.VisibleFields(t) {
			info := fieldJSONInfo(field)
			if info.omit {
				continue
			}

			propSchema, ok := schema.Properties[info.name]
			if !ok {
				continue
			}

			if info.settings["inline"] {
				if schema.OneOf == nil {
					schema.OneOf = make([]*jsonschema.Schema, 0)
				}

				schema.OneOf = append(schema.OneOf, propSchema)
				delete(schema.Properties, info.name)
				schema.Required = slices.DeleteFunc(schema.Required, func(s string) bool {
					return s == info.name
				})
			}

			if err := handleInline(field.Type, propSchema); err != nil {
				return err
			}
		}
	}
	return nil
}
