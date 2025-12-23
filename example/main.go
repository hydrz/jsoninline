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
