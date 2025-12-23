package jsoninline_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/hydrz/jsoninline"
)

// TestMarshalBothInlines ensures when both inline structs are present,
// fields are inlined and later inline fields win on key conflicts.
func TestMarshalBothInlines(t *testing.T) {
	u := User{
		ID:    10,
		Name:  "Both",
		Email: "both@example.com",
		China: &China{
			Province: "Guangdong",
			City:     "Shenzhen",
			NestedFoo: &NestedFoo{
				FooField: "FooA",
			},
		},
		USA: &USA{
			State: "California",
			City:  "San Francisco",
			NestedBar: &NestedBar{
				BarField: "BarB",
			},
		},
	}

	b, err := json.Marshal(jsoninline.V(u))
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("failed to parse marshaled JSON: %v -- %s", err, string(b))
	}

	// Both province and state should be present
	if got, _ := m["province"].(string); got != "Guangdong" {
		t.Fatalf("expected province=Guangdong, got %v, raw: %s", m["province"], string(b))
	}
	if got, _ := m["state"].(string); got != "California" {
		t.Fatalf("expected state=California, got %v, raw: %s", m["state"], string(b))
	}

	// city is present; because USA is later in struct, its city should win
	if got, _ := m["city"].(string); got != "San Francisco" {
		t.Fatalf("expected city to be USA value (San Francisco), got %v", m["city"])
	}

	// nested fields from both should appear
	if got, _ := m["foo_field"].(string); got != "FooA" {
		t.Fatalf("expected foo_field present, got %v", m["foo_field"])
	}
	if got, _ := m["bar_field"].(string); got != "BarB" {
		t.Fatalf("expected bar_field present, got %v", m["bar_field"])
	}
}

// TestMarshalOmitNilInlines ensures nil inline pointer fields are omitted from output.
func TestMarshalOmitNilInlines(t *testing.T) {
	u := User{
		ID:    11,
		Name:  "NilBoth",
		Email: "nil@example.com",
		China: nil,
		USA:   nil,
	}

	b, err := json.Marshal(jsoninline.V(u))
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	out := string(b)

	// Should not contain inline fields
	unexpected := []string{"province", "state", "city", "foo_field", "bar_field"}
	for _, s := range unexpected {
		if strings.Contains(out, s) {
			t.Fatalf("did not expect %q in output: %s", s, out)
		}
	}
}

// TestUnmarshalBothInlineFields verifies unmarshalling into a struct with multiple
// inline pointers will populate each inline struct (both receive the 'city' value).
func TestUnmarshalBothInlineFields(t *testing.T) {
	data := `{
        "id": 20,
        "name": "Amb",
        "email": "amb@example.com",
        "province": "Guangdong",
        "state": "California",
        "city": "SharedCity",
        "foo_field": "FooX",
        "bar_field": "BarY"
    }`

	var u User
	if err := json.Unmarshal([]byte(data), jsoninline.V(&u)); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if u.China == nil {
		t.Fatalf("expected China set, got nil")
	}
	if u.USA == nil {
		t.Fatalf("expected USA set, got nil")
	}

	// Both inline structs receive the single city value from JSON
	if u.China.City != "SharedCity" {
		t.Fatalf("expected China.City=SharedCity, got %q", u.China.City)
	}
	if u.USA.City != "SharedCity" {
		t.Fatalf("expected USA.City=SharedCity, got %q", u.USA.City)
	}

	if u.China.Province != "Guangdong" {
		t.Fatalf("expected China.Province set, got %q", u.China.Province)
	}
	if u.USA.State != "California" {
		t.Fatalf("expected USA.State set, got %q", u.USA.State)
	}

	if u.China.NestedFoo == nil || u.China.NestedFoo.FooField != "FooX" {
		t.Fatalf("expected China.NestedFoo populated, got %+v", u.China.NestedFoo)
	}
	if u.USA.NestedBar == nil || u.USA.NestedBar.BarField != "BarY" {
		t.Fatalf("expected USA.NestedBar populated")
	}
}

// TestSliceElementTypes ensures both slices of structs and slices of pointers
// are handled correctly by the Unmarshal implementation.
func TestSliceElementTypes(t *testing.T) {
	data := `[
        {"id":1,"name":"A","email":"a@x","province":"P1","city":"C1"},
        {"id":2,"name":"B","email":"b@x","state":"S2","city":"C2"}
    ]`

	var usersPtr []*User
	if err := json.Unmarshal([]byte(data), jsoninline.V(&usersPtr)); err != nil {
		t.Fatalf("unmarshal ptr-slice failed: %v", err)
	}
	if len(usersPtr) != 2 {
		t.Fatalf("expected 2 users, got %d", len(usersPtr))
	}
	if usersPtr[0].China == nil || usersPtr[0].China.City != "C1" {
		t.Fatalf("unexpected usersPtr[0].China: %+v", usersPtr[0].China)
	}

	var usersVal []User
	if err := json.Unmarshal([]byte(data), jsoninline.V(&usersVal)); err != nil {
		t.Fatalf("unmarshal value-slice failed: %v", err)
	}
	if len(usersVal) != 2 {
		t.Fatalf("expected 2 users, got %d", len(usersVal))
	}
	if usersVal[1].USA == nil || usersVal[1].USA.City != "C2" {
		t.Fatalf("unexpected usersVal[1].USA: %+v", usersVal[1].USA)
	}
}

// TestSchemaForSlice ensures schema generation works for slices with inline types.
func TestSchemaForSlice(t *testing.T) {
	schema, err := jsoninline.For[User](&jsonschema.ForOptions{})
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}
	// Basic sanity: ensure schema properties contain inlined fields
	if _, ok := schema.Properties["city"]; !ok {
		t.Fatalf("expected schema to include 'city' property; got properties: %v", schema.Properties)
	}
}
