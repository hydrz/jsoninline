package jsoninline_test

import (
	"encoding/json"
	"strings"
	"testing"

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
	City     string `json:"city,omitempty"`
	Province string `json:"province,omitempty"`
}

type USA struct {
	City  string `json:"city,omitempty"`
	State string `json:"state,omitempty"`
}

func Test(t *testing.T) {
	users := []User{
		{
			ID:    1,
			Name:  "Alice",
			Email: "alice@example.com",
			China: &China{
				Province: "Guangdong",
				City:     "Shenzhen",
			},
		},
		{
			ID:    2,
			Name:  "Bob",
			Email: "bob@example.com",
			USA: &USA{
				State: "California",
				City:  "Los Angeles",
			},
		},
	}

	nestedBytes, err := json.MarshalIndent(jsoninline.V(users), "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal users: %v", err)
	}

	// check output contains inlined fields
	output := string(nestedBytes)

	expectedSubstrings := []string{
		`"province": "Guangdong"`,
		`"city": "Shenzhen"`,
		`"state": "California"`,
		`"city": "Los Angeles"`,
	}

	for _, substr := range expectedSubstrings {
		if !strings.Contains(output, substr) {
			t.Errorf("Expected output to contain %q, but it did not.\nOutput: %s", substr, output)
		}
	}
}
