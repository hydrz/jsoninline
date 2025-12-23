package jsoninline

import (
	"fmt"
	"reflect"

	_ "unsafe"

	"github.com/google/jsonschema-go/jsonschema"
	_ "github.com/google/jsonschema-go/jsonschema"
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
	if opts == nil {
		opts = &jsonschema.ForOptions{}
	}
	if opts.TypeSchemas == nil {
		opts.TypeSchemas = make(map[reflect.Type]*jsonschema.Schema)
	}
	newT, schemaMap, err := schemaType(t, map[reflect.Type]bool{})
	if err != nil {
		return nil, err
	}
	for k, v := range schemaMap {
		opts.TypeSchemas[k] = v
	}
	// Debug: print resulting struct fields to ensure inline worked.
	fmt.Println("DEBUG: ForType resulting struct fields:")
	if newT.Kind() == reflect.Struct {
		for _, f := range reflect.VisibleFields(newT) {
			fmt.Printf(" - %s (Anonymous=%v, Type=%v, Tag=%q)\n", f.Name, f.Anonymous, f.Type, f.Tag)
		}
	} else if newT.Kind() == reflect.Slice || newT.Kind() == reflect.Array {
		elem := newT.Elem()
		if elem.Kind() == reflect.Struct {
			for _, f := range reflect.VisibleFields(elem) {
				fmt.Printf(" - %s (Anonymous=%v, Type=%v, Tag=%q)\n", f.Name, f.Anonymous, f.Type, f.Tag)
			}
		}
	}

	return jsonschema.ForType(newT, opts)
}
func schemaType(t reflect.Type, seen map[reflect.Type]bool) (reflect.Type, map[reflect.Type]*jsonschema.Schema, error) {
	// Handle slices/arrays by processing their element type recursively.
	if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		elem := t.Elem()
		newElem, innerMap, err := schemaType(elem, seen)
		if err != nil {
			return nil, nil, err
		}
		// merge discovered schemas
		schemaMap := make(map[reflect.Type]*jsonschema.Schema)
		for k, v := range innerMap {
			schemaMap[k] = v
		}

		if newElem == elem {
			return t, schemaMap, nil
		}
		if t.Kind() == reflect.Slice {
			return reflect.SliceOf(newElem), schemaMap, nil
		}
		// array
		return reflect.ArrayOf(t.Len(), newElem), schemaMap, nil
	}

	if t.Kind() != reflect.Struct {
		return t, nil, nil
	}

	// Prevent infinite recursion on self-referential types.
	if seen[t] {
		return t, nil, nil
	}
	seen[t] = true

	fields := reflect.VisibleFields(t)
	out := make([]reflect.StructField, 0, len(fields))
	schemaMap := make(map[reflect.Type]*jsonschema.Schema)
	changed := false
	// track inline inner types and their processed (possibly new) types

	for _, f := range fields {
		info := fieldJSONInfo(f)

		// Copy the field and preserve exported/unexported status via PkgPath.
		sf := reflect.StructField{
			Name:      f.Name,
			Type:      f.Type,
			Tag:       f.Tag,
			PkgPath:   f.PkgPath,
			Anonymous: f.Anonymous,
		}

		if info.settings["inline"] {
			// For inline-tagged fields we expand the inner struct's visible fields
			// into the parent struct so JSON Schema generation sees the properties
			// directly (avoids promotion conflicts when multiple inlined structs
			// contain the same field name).
			changed = true

			// Determine inner element type (dereference pointers)
			inner := f.Type
			ptrWrapped := false
			if inner.Kind() == reflect.Ptr {
				inner = inner.Elem()
				ptrWrapped = true
			}

			// Recursively process the inner type to handle nested inline fields.
			newInner, innerMap, err := schemaType(inner, seen)
			if err != nil {
				return nil, nil, err
			}
			// merge any discovered schemas
			for k, v := range innerMap {
				schemaMap[k] = v
			}

			// Append each visible exported field from the inner type as a new
			// field on the parent struct. Use a unique Go field name to avoid
			// collisions while preserving the original JSON tag.
			for _, cf := range reflect.VisibleFields(newInner) {
				if cf.PkgPath != "" { // unexported
					continue
				}
				cinfo := fieldJSONInfo(cf)
				if cinfo.omit {
					continue
				}

				// Unique Go field name to avoid duplicate names in the generated struct.
				uniqueName := fmt.Sprintf("%s_%s", f.Name, cf.Name)

				childType := cf.Type
				if ptrWrapped && childType.Kind() == reflect.Struct {
					// If original inline field was a pointer to struct, keep the
					// child field types as pointer where appropriate to reflect
					// nullable behavior. Wrap struct child types in pointer.
					childType = reflect.PtrTo(childType)
				}

				childSF := reflect.StructField{
					Name:      uniqueName,
					Type:      childType,
					Tag:       cf.Tag,
					PkgPath:   "",
					Anonymous: false,
				}
				out = append(out, childSF)
			}
			continue
		}

		out = append(out, sf)
	}

	if !changed {
		return t, nil, nil
	}

	// Build a new struct type with the modified Anonymous flags.
	// Note: reflect.Type is immutable, so we create a new one instead of mutating.
	newT := reflect.StructOf(out)

	return newT, schemaMap, nil
}
