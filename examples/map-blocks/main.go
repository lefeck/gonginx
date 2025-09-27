package main

import (
	"fmt"
	"log"

	"github.com/lefeck/gonginx/dumper"
	"github.com/lefeck/gonginx/parser"
)

func main() {
	// Example nginx configuration with map blocks
	configString := `
http {
	# Backend selection based on host
	map $http_host $backend {
		default backend_default;
		example.com backend_main;
		api.example.com backend_api;
		~^www\. backend_www;
		~^mobile\. backend_mobile;
	}
	
	# Rate limiting based on request method
	map $request_method $limit_key {
		default "";
		POST $binary_remote_addr;
		PUT $binary_remote_addr;
		DELETE $binary_remote_addr;
	}
	
	# SSL redirect based on scheme
	map $scheme $redirect_https {
		default 0;
		http 1;
	}
	
	# Content type mapping
	map $uri $content_type {
		default "text/html";
		~\.js$ "application/javascript";
		~\.css$ "text/css";
		~\.json$ "application/json";
		~\.(jpg|jpeg|png|gif)$ "image/*";
	}

	upstream backend_main {
		server 192.168.1.10:8080;
		server 192.168.1.11:8080;
	}
	
	upstream backend_api {
		server 192.168.1.20:3000;
		server 192.168.1.21:3000;
	}
	
	upstream backend_www {
		server 192.168.1.30:8080;
	}
	
	upstream backend_mobile {
		server 192.168.1.40:8080;
	}
	
	upstream backend_default {
		server 192.168.1.50:8080;
	}

	server {
		listen 80;
		server_name _;
		
		# Use map variables for dynamic configuration
		set $backend_to_use $backend;
		
		# Conditional SSL redirect
		if ($redirect_https) {
			return 301 https://$server_name$request_uri;
		}
		
		location / {
			proxy_pass http://$backend_to_use;
			proxy_set_header Host $host;
			proxy_set_header X-Real-IP $remote_addr;
		}
		
		location /api {
			# Rate limiting using map variable
			limit_req zone=api_limit key=$limit_key burst=10 nodelay;
			proxy_pass http://backend_api;
		}
	}
	
	server {
		listen 443 ssl;
		server_name _;
		
		ssl_certificate /etc/ssl/certs/server.crt;
		ssl_certificate_key /etc/ssl/private/server.key;
		
		location / {
			proxy_pass http://$backend;
			add_header Content-Type $content_type;
		}
	}
}
`

	fmt.Println("=== Map Blocks Example ===\n")

	// Parse the configuration
	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	if err != nil {
		log.Fatal("Failed to parse config:", err)
	}

	// 1. Find and display all map blocks
	fmt.Println("1. Found Map Blocks:")
	maps := conf.FindMaps()
	fmt.Printf("   Total map blocks: %d\n\n", len(maps))

	for i, mapBlock := range maps {
		fmt.Printf("   Map %d: %s -> %s\n", i+1, mapBlock.Variable, mapBlock.MappedVariable)
		fmt.Printf("   Mappings (%d):\n", len(mapBlock.Mappings))

		for _, mapping := range mapBlock.Mappings {
			fmt.Printf("     %s -> %s\n", mapping.Pattern, mapping.Value)
		}

		defaultValue := mapBlock.GetDefaultValue()
		if defaultValue != "" {
			fmt.Printf("   Default value: %s\n", defaultValue)
		}
		fmt.Println()
	}

	// 2. Find specific map by variables
	fmt.Println("2. Finding Specific Maps:")
	backendMap := conf.FindMapByVariables("$http_host", "$backend")
	if backendMap != nil {
		fmt.Printf("   Found backend selection map with %d mappings\n", len(backendMap.Mappings))
		fmt.Printf("   Default backend: %s\n", backendMap.GetDefaultValue())
	}

	limitMap := conf.FindMapByVariables("$request_method", "$limit_key")
	if limitMap != nil {
		fmt.Printf("   Found rate limiting map with %d mappings\n", len(limitMap.Mappings))
	}
	fmt.Println()

	// 3. Manipulate map blocks
	fmt.Println("3. Manipulating Map Blocks:")
	if backendMap != nil {
		// Add a new mapping
		originalCount := len(backendMap.Mappings)
		backendMap.AddMapping("test.example.com", "backend_test")
		fmt.Printf("   Added new mapping: test.example.com -> backend_test\n")
		fmt.Printf("   Mappings count: %d -> %d\n", originalCount, len(backendMap.Mappings))

		// Update default value
		backendMap.SetDefaultValue("backend_updated_default")
		fmt.Printf("   Updated default value to: %s\n", backendMap.GetDefaultValue())
	}
	fmt.Println()

	// 4. Display the modified configuration
	fmt.Println("4. Modified Configuration:")
	fmt.Println("   (showing only the first map block with changes)")

	if len(maps) > 0 {
		mapOutput := dumper.DumpDirective(maps[0], dumper.IndentedStyle)
		fmt.Println(mapOutput)
	}
	fmt.Println()

	// 5. Demonstrate map usage patterns
	fmt.Println("5. Common Map Usage Patterns:")
	fmt.Println("   - Backend Selection: Route requests to different upstreams based on hostname")
	fmt.Println("   - Rate Limiting: Apply different rate limits based on request method")
	fmt.Println("   - SSL Redirect: Conditionally redirect HTTP to HTTPS")
	fmt.Println("   - Content Type: Set appropriate content types based on file extensions")
	fmt.Println("   - A/B Testing: Route users to different backends based on cookies or headers")
	fmt.Println("   - Geographic Routing: Route based on client location (with geo module)")
	fmt.Println("   - Security: Block or allow requests based on various criteria")

	fmt.Println("\n=== Map Blocks Demo Complete ===")
}
