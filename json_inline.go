package jsoninline

import (
	"encoding/json"
	"reflect"
	"strings"
)

func V(v any) InlineMarshaler {
	return InlineMarshaler{V: v}
}

type InlineMarshaler struct {
	V any
}

func (im InlineMarshaler) MarshalJSON() ([]byte, error) {
	return marshal(im.V)
}

func marshal(p any) ([]byte, error) {
	v := reflect.ValueOf(p)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// handle slices/arrays by marshaling each element individually
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		out := make([]interface{}, 0, v.Len())
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i).Interface()
			b, err := marshal(elem)
			if err != nil {
				return nil, err
			}
			var val interface{}
			if err := json.Unmarshal(b, &val); err != nil {
				return nil, err
			}
			out = append(out, val)
		}
		return json.Marshal(out)
	}

	// for non-structs fallback to default marshal
	if v.Kind() != reflect.Struct {
		return json.Marshal(p)
	}

	m := make(map[string]any)
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" { // unexported
			continue
		}

		// skip the helper InlineStruct field itself so it doesn't appear in output
		if field.Type == reflect.TypeOf(InlineMarshaler{}) {
			continue
		}

		tag := field.Tag.Get("json")
		if tag == "-" {
			continue
		}

		parts := strings.Split(tag, ",")
		name := parts[0]
		if name == "" {
			name = field.Name
		}

		fv := v.Field(i)

		// handle inline fields
		if strings.Contains(tag, ",inline") {
			if fv.Kind() == reflect.Ptr {
				if fv.IsNil() {
					continue
				}
			}
			inlineBytes, err := json.Marshal(fv.Interface())
			if err != nil {
				return nil, err
			}
			var inlineMap map[string]interface{}
			if err := json.Unmarshal(inlineBytes, &inlineMap); err != nil {
				return nil, err
			}
			for k, v := range inlineMap {
				m[k] = v
			}
			delete(m, name)
			continue
		}

		if fv.Kind() == reflect.Ptr && fv.IsNil() {
			if strings.Contains(tag, "omitempty") {
				continue
			}
			m[name] = nil
			continue
		}

		valBytes, err := json.Marshal(fv.Interface())
		if err != nil {
			return nil, err
		}
		var val interface{}
		if err := json.Unmarshal(valBytes, &val); err != nil {
			return nil, err
		}
		m[name] = val
	}

	return json.Marshal(m)
}
