//go:build example

package main

import (
	"encoding/json"
	"fmt"

	"github.com/hydrz/jsoninline"
)

// Demonstrates unmarshaling JSON with inlined fields into Go structs
// using jsoninline.V to wrap the destination.
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	China *China `json:",inline"`
	USA   *USA   `json:",inline"`
}

func (u *User) String() string {
	data, _ := json.Marshal(jsoninline.V(u))
	return string(data)
}

type China struct {
	City     string `json:"city,omitempty"`
	Province string `json:"province,omitempty"`
}

type USA struct {
	City  string `json:"city,omitempty"`
	State string `json:"state,omitempty"`
}

var jsonData = `[
  {
    "id": 1,
    "name": "Alice",
    "email": "alice@example.com",
    "province": "Guangdong",
    "city": "Shenzhen"
  },
  {
    "id": 2,
    "name": "Bob",
    "email": "bob@example.com",
    "state": "California",
    "city": "Los Angeles"
  }
]`

func main() {

	var users []*User
	if err := json.Unmarshal([]byte(jsonData), jsoninline.V(&users)); err != nil {
		panic(err)
	}

	for i, u := range users {
		fmt.Printf("user %d: %+v\n", i, u)
	}
	// Output:
	// user 0: {"city":"Shenzhen","email":"alice@example.com","id":1,"name":"Alice","province":"Guangdong"}
	// user 1: {"city":"Los Angeles","email":"bob@example.com","id":2,"name":"Bob","state":"California"}
}
