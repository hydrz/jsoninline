# jsoninline

jsoninline is a small Go utility that helps inline nested struct fields into the parent object when marshaling to JSON. It provides `InlineMarshaler` and a helper `V()` to wrap values that should be inlined.

Features

- Inline fields tagged with `,inline` into their parent JSON object.
- Support for pointers, structs, slices/arrays, and basic types.

Installation

```bash
go get github.com/hydrz/jsoninline@latest
```

Quick start

1. Use `jsoninline.InlineMarshaler` as the field type and add `,inline` to the JSON tag.
2. Wrap a value with `jsoninline.V(value)` to mark it for inlining.
3. Marshal using the standard `encoding/json` package.

Example

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/hydrz/jsoninline"
)

type DNSServerOption struct {
	Type  string               `json:"type"`
	Tag   string               `json:"tag"`
	Local LocalDNSServerOption `json:",inline"`
	UDP   UDPDNSServerOption   `json:",inline"`
	TLS   TLSDNSServerOption   `json:",inline"`
}

type LocalDNSServerOption struct {
	PreferGO bool `json:"prefer_go,omitempty"`
}

type ServerOptions struct {
	Server     string `json:"server,omitempty"`
	ServerPort int    `json:"server_port,omitempty"`
}

type UDPDNSServerOption struct {
	ServerOptions
}

type TLSDNSServerOption struct {
	ServerOptions
	TLS any `json:"tls,omitempty"`
}

func main() {
	options := []DNSServerOption{
		{
			Type: "local",
			Tag:  "local-dns",
			Local: LocalDNSServerOption{
				PreferGO: true,
			},
		},
		{
			Type: "udp",
			Tag:  "udp-dns",
			UDP: UDPDNSServerOption{
				ServerOptions: ServerOptions{
					Server:     "1.1.1.1",
					ServerPort: 53,
				},
			},
		},
		{
			Type: "tls",
			Tag:  "tls-dns",
			TLS: TLSDNSServerOption{
				ServerOptions: ServerOptions{
					Server:     "dns.google",
					ServerPort: 853,
				},
				TLS: map[string]interface{}{
					"sni": "dns.google",
				},
			},
		},
	}
	bytes, err := json.MarshalIndent(jsoninline.V(options), "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(bytes))

	// Output:
	// [
	// 	{
	// 	  "prefer_go": true,
	// 	  "tag": "local-dns",
	// 	  "type": "local"
	// 	},
	// 	{
	// 	  "server": "1.1.1.1",
	// 	  "server_port": 53,
	// 	  "tag": "udp-dns",
	// 	  "type": "udp"
	// 	},
	// 	{
	// 	  "type": "tls"
	// 	  "server": "dns.google",
	// 	  "server_port": 853,
	// 	  "tag": "tls-dns",
	// 	  "tls": {
	// 	     "sni": "dns.google"
	// 	  },
	// 	}
	// ]
}
```

Unmarshal example

You can also use `jsoninline.V` when unmarshaling to populate structs that expect inlined fields. Example:

```go
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
```

License

See the `LICENSE` file in the repository.
