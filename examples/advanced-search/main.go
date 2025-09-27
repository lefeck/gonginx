package main

import (
	"fmt"
	"log"

	"github.com/lefeck/gonginx/config"
	"github.com/lefeck/gonginx/parser"
)

func main() {
	// Example nginx configuration with multiple servers, upstreams, and SSL
	configString := `
http {
	upstream api_backend {
		server 192.168.1.10:8080 weight=3;
		server 192.168.1.11:8080 weight=2;
		server 192.168.1.12:8080 backup;
	}
	
	upstream web_backend {
		server 10.0.0.1:3000;
		server 10.0.0.2:3000;
	}

	server {
		listen 80;
		server_name example.com www.example.com;
		return 301 https://$server_name$request_uri;
	}

	server {
		listen 443 ssl http2;
		server_name example.com www.example.com;
		
		ssl_certificate /etc/ssl/certs/example.com.crt;
		ssl_certificate_key /etc/ssl/private/example.com.key;
		ssl_protocols TLSv1.2 TLSv1.3;
		
		location / {
			proxy_pass http://web_backend;
			proxy_set_header Host $host;
		}
		
		location /api {
			proxy_pass http://api_backend;
			proxy_set_header X-Real-IP $remote_addr;
		}
		
		location ~ \.(jpg|jpeg|png|gif|css|js)$ {
			expires 30d;
			root /var/www/static;
		}
	}

	server {
		listen 443 ssl;
		server_name api.example.com;
		
		ssl_certificate /etc/ssl/certs/api.example.com.crt;
		ssl_certificate_key /etc/ssl/private/api.example.com.key;
		
		location / {
			proxy_pass http://api_backend;
		}
	}
}
`

	// Parse the configuration
	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	if err != nil {
		log.Fatal("Failed to parse config:", err)
	}

	fmt.Println("=== Advanced Search Examples ===\n")

	// 1. Find servers by server_name
	fmt.Println("1. Finding servers by server_name:")
	servers := conf.FindServersByName("example.com")
	fmt.Printf("   Found %d servers with server_name 'example.com'\n", len(servers))
	for i, server := range servers {
		listens := server.FindDirectives("listen")
		if len(listens) > 0 {
			fmt.Printf("   Server %d: listens on %s\n", i+1, listens[0].GetParameters()[0].GetValue())
		}
	}

	// 2. Find upstream by name
	fmt.Println("2. Finding upstream by name:")
	upstream := conf.FindUpstreamByName("api_backend")
	if upstream != nil {
		fmt.Printf("   Found upstream '%s' with %d servers:\n", upstream.UpstreamName, len(upstream.UpstreamServers))
		for i, server := range upstream.UpstreamServers {
			fmt.Printf("   Server %d: %s", i+1, server.Address)
			if len(server.Parameters) > 0 {
				fmt.Printf(" (weight=%s)", server.Parameters["weight"])
			}
			if len(server.Flags) > 0 {
				fmt.Printf(" [%s]", server.Flags[0])
			}
			fmt.Println()
		}
	} else {
		fmt.Println("   Upstream not found")
	}
	fmt.Println()

	// 3. Find locations by pattern
	fmt.Println("3. Finding locations by pattern:")
	locations := conf.FindLocationsByPattern("/api")
	fmt.Printf("   Found %d locations with pattern '/api'\n", len(locations))

	// Also search for regex patterns
	regexLocations := conf.FindLocationsByPattern("~ \\.(jpg|jpeg|png|gif|css|js)$")
	fmt.Printf("   Found %d locations with regex pattern for static files\n", len(regexLocations))
	fmt.Println()

	// 4. Get all SSL certificates
	fmt.Println("4. Getting all SSL certificates:")
	certificates := conf.GetAllSSLCertificates()
	fmt.Printf("   Found %d SSL certificates:\n", len(certificates))
	for i, cert := range certificates {
		fmt.Printf("   Certificate %d: %s\n", i+1, cert)
	}
	fmt.Println()

	// 5. Get all upstream servers
	fmt.Println("5. Getting all upstream servers:")
	allServers := conf.GetAllUpstreamServers()
	fmt.Printf("   Found %d total upstream servers across all upstreams:\n", len(allServers))
	for i, server := range allServers {
		fmt.Printf("   Server %d: %s", i+1, server.Address)
		if len(server.Parameters) > 0 {
			for key, value := range server.Parameters {
				fmt.Printf(" (%s=%s)", key, value)
			}
		}
		if len(server.Flags) > 0 {
			fmt.Printf(" [%s]", server.Flags[0])
		}
		fmt.Println()
	}
	fmt.Println()

	// 6. Advanced search combinations
	fmt.Println("6. Advanced search combinations:")

	// Find all servers that have SSL certificates
	fmt.Println("   Servers with SSL certificates:")
	allServerDirectives := conf.FindDirectives("server")
	sslServerCount := 0
	for _, directive := range allServerDirectives {
		if server, ok := directive.(*config.Server); ok {
			sslCerts := server.FindDirectives("ssl_certificate")
			if len(sslCerts) > 0 {
				sslServerCount++
				serverNames := server.FindDirectives("server_name")
				if len(serverNames) > 0 {
					fmt.Printf("   - %s\n", serverNames[0].GetParameters()[0].GetValue())
				}
			}
		}
	}
	fmt.Printf("   Total SSL-enabled servers: %d\n", sslServerCount)

	fmt.Println("\n=== Search Complete ===")
}
