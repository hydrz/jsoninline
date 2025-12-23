package jsoninline

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
)

func V(v any) *InlineMarshaler {
	return &InlineMarshaler{V: v}
}

type InlineMarshaler struct {
	V any
}

func (im InlineMarshaler) MarshalJSON() ([]byte, error) {
	return marshal(im.V)
}

// UnmarshalJSON implements json.Unmarshaler for InlineMarshaler.
// im.V must be a non-nil pointer to the destination value (struct, slice, etc.).
func (im *InlineMarshaler) UnmarshalJSON(data []byte) error {
	if im == nil || im.V == nil {
		return errors.New("jsoninline: nil target for UnmarshalJSON")
	}
	return unmarshal(data, im.V)
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

func unmarshal(data []byte, p any) error {
	v := reflect.ValueOf(p)
	if v.Kind() != reflect.Ptr {
		return errors.New("jsoninline: V must be a pointer")
	}

	ve := v.Elem()

	// handle slices/arrays
	if ve.Kind() == reflect.Slice || ve.Kind() == reflect.Array {
		var raws []json.RawMessage
		if err := json.Unmarshal(data, &raws); err != nil {
			return err
		}

		elemType := ve.Type().Elem()

		if ve.Kind() == reflect.Slice {
			slice := reflect.MakeSlice(ve.Type(), 0, len(raws))
			for _, raw := range raws {
				var elemVal reflect.Value
				if elemType.Kind() == reflect.Ptr && elemType.Elem().Kind() == reflect.Struct {
					// element is pointer to struct (*T)
					newElem := reflect.New(elemType.Elem()) // *T
					im := &InlineMarshaler{V: newElem.Interface()}
					if err := im.UnmarshalJSON(raw); err != nil {
						return err
					}
					elemVal = newElem
				} else if elemType.Kind() == reflect.Struct {
					// element is struct (T)
					newElemPtr := reflect.New(elemType) // *T
					im := &InlineMarshaler{V: newElemPtr.Interface()}
					if err := im.UnmarshalJSON(raw); err != nil {
						return err
					}
					elemVal = newElemPtr.Elem()
				} else {
					// other types
					newElemPtr := reflect.New(elemType)
					if err := unmarshalRaw(raw, newElemPtr.Interface()); err != nil {
						return err
					}
					elemVal = newElemPtr.Elem()
				}
				slice = reflect.Append(slice, elemVal)
			}
			ve.Set(slice)
			return nil
		}

		// array
		if len(raws) != ve.Len() {
			return errors.New("jsoninline: array length mismatch")
		}
		for i, raw := range raws {
			if elemType.Kind() == reflect.Ptr && elemType.Elem().Kind() == reflect.Struct {
				newElem := reflect.New(elemType.Elem()) // *T
				im := &InlineMarshaler{V: newElem.Interface()}
				if err := im.UnmarshalJSON(raw); err != nil {
					return err
				}
				ve.Index(i).Set(newElem)
			} else if elemType.Kind() == reflect.Struct {
				newElemPtr := reflect.New(elemType)
				im := &InlineMarshaler{V: newElemPtr.Interface()}
				if err := im.UnmarshalJSON(raw); err != nil {
					return err
				}
				ve.Index(i).Set(newElemPtr.Elem())
			} else {
				newElemPtr := reflect.New(elemType)
				if err := unmarshalRaw(raw, newElemPtr.Interface()); err != nil {
					return err
				}
				ve.Index(i).Set(newElemPtr.Elem())
			}
		}
		return nil
	}

	// non-struct fallback to default unmarshal
	if ve.Kind() != reflect.Struct {
		return json.Unmarshal(data, p)
	}

	// struct: parse top-level map and populate fields, handling ",inline" tags
	top := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &top); err != nil {
		return err
	}

	t := ve.Type()
	out := reflect.New(t).Elem()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" { // unexported
			continue
		}
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

		fv := out.Field(i)

		if strings.Contains(tag, ",inline") {
			// For inline fields, unmarshal the whole object into the inline struct.
			if fv.Kind() == reflect.Ptr {
				fv.Set(reflect.New(fv.Type().Elem()))
				if err := json.Unmarshal(data, fv.Interface()); err != nil {
					return err
				}
			} else {
				if err := json.Unmarshal(data, fv.Addr().Interface()); err != nil {
					return err
				}
			}
			continue
		}

		raw, ok := top[name]
		if !ok {
			// not present in JSON; leave zero value (or nil pointer)
			continue
		}

		if fv.Kind() == reflect.Ptr {
			fv.Set(reflect.New(fv.Type().Elem()))
			if err := json.Unmarshal(raw, fv.Interface()); err != nil {
				return err
			}
		} else {
			if err := json.Unmarshal(raw, fv.Addr().Interface()); err != nil {
				return err
			}
		}
	}

	ve.Set(out)
	return nil
}

// unmarshalRaw raw JSON into target, handling structs with inline tags recursively.
func unmarshalRaw(raw json.RawMessage, v interface{}) error {
	// If target implements json.Unmarshaler, let json.Unmarshal handle it.
	if um, ok := v.(json.Unmarshaler); ok {
		return um.UnmarshalJSON(raw)
	}
	return json.Unmarshal(raw, v)
}
