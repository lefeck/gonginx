package main

import (
	"fmt"
	"log"

	"github.com/lefeck/gonginx/dumper"
	"github.com/lefeck/gonginx/parser"
)

func main() {
	// Example nginx configuration with limit_req_zone directives
	configString := `
http {
	# Rate limiting zones
	limit_req_zone $binary_remote_addr zone=global:10m rate=10r/s;
	limit_req_zone $server_name zone=perserver:5m rate=50r/s;
	limit_req_zone $request_uri zone=perpath:20m rate=5r/s;
	limit_req_zone $http_user_agent zone=peragent:1g rate=2r/m sync;
	limit_req_zone $cookie_user_id zone=peruser:100m rate=20r/s;
	
	# Advanced rate limiting with different variables
	limit_req_zone $binary_remote_addr$request_method zone=per_ip_method:50m rate=30r/s;
	limit_req_zone $geoip_country_code zone=percountry:10m rate=100r/s;
	limit_req_zone $http_x_real_ip zone=trusted:5m rate=1000r/s;

	upstream backend {
		server backend1.example.com:8080;
		server backend2.example.com:8080;
		server backend3.example.com:8080;
	}

	# API server with strict rate limiting
	server {
		listen 80;
		server_name api.example.com;
		
		access_log /var/log/nginx/api_access.log combined;
		error_log /var/log/nginx/api_error.log;
		
		# Global rate limiting
		limit_req zone=global burst=20 nodelay;
		
		location /api/v1/ {
			# Strict rate limiting for API endpoints
			limit_req zone=perpath burst=5 nodelay;
			limit_req zone=peruser burst=10 delay=5;
			
			proxy_pass http://backend;
			proxy_set_header Host $host;
			proxy_set_header X-Real-IP $remote_addr;
			
			# Rate limit status headers
			add_header X-Rate-Limit-Zone "perpath,peruser";
			add_header X-Rate-Limit-Burst "5,10";
		}
		
		location /api/v2/ {
			# Different rate limiting for v2 API
			limit_req zone=per_ip_method burst=15 nodelay;
			
			proxy_pass http://backend;
			proxy_set_header Host $host;
			proxy_set_header X-Real-IP $remote_addr;
		}
		
		location /login {
			# Very strict rate limiting for login
			limit_req zone=global burst=3 nodelay;
			limit_req zone=perpath burst=1 delay=2;
			
			proxy_pass http://backend;
		}
		
		location /health {
			# No rate limiting for health checks
			proxy_pass http://backend;
		}
	}
	
	# Web server with moderate rate limiting
	server {
		listen 80;
		server_name web.example.com;
		
		# Moderate global rate limiting
		limit_req zone=global burst=50 delay=10;
		
		location / {
			# Per-path rate limiting for web content
			limit_req zone=perpath burst=20 nodelay;
			
			proxy_pass http://backend;
			proxy_set_header Host $host;
		}
		
		location /static/ {
			# Higher limits for static content
			limit_req zone=perserver burst=100 nodelay;
			
			expires 1y;
			add_header Cache-Control "public, immutable";
			proxy_pass http://backend;
		}
		
		location /upload {
			# Strict limits for upload endpoint
			limit_req zone=peruser burst=2 delay=5;
			
			client_max_body_size 10m;
			proxy_pass http://backend;
		}
	}
	
	# Admin server with user-based rate limiting
	server {
		listen 80;
		server_name admin.example.com;
		
		# Admin-specific rate limiting
		limit_req zone=peruser burst=30 nodelay;
		
		location /admin/ {
			# Additional path-based limiting
			limit_req zone=perpath burst=10 delay=3;
			
			auth_basic "Admin Area";
			auth_basic_user_file /etc/nginx/.htpasswd;
			
			proxy_pass http://backend;
		}
		
		location /admin/analytics {
			# Higher limits for analytics
			limit_req zone=peruser burst=100 nodelay;
			
			proxy_pass http://backend;
		}
	}
	
	# Country-specific rate limiting
	server {
		listen 80;
		server_name global.example.com;
		
		# Different limits based on country
		limit_req zone=percountry burst=50 nodelay;
		
		location / {
			proxy_pass http://backend;
			
			# Add country info to headers
			add_header X-Country-Code $geoip_country_code;
			add_header X-Rate-Limit-Zone "percountry";
		}
	}
	
	# Monitoring and status server
	server {
		listen 8080;
		server_name status.example.com;
		
		location /nginx_status {
			stub_status on;
			access_log off;
			
			# No rate limiting for monitoring
			allow 127.0.0.1;
			allow 10.0.0.0/8;
			deny all;
		}
		
		location /rate_limit_status {
			return 200 "Rate Limiting Status:
Global Zone: 10r/s (burst=20,50)
Per-Server Zone: 50r/s (burst=100)
Per-Path Zone: 5r/s (burst=5,10,20)
Per-User Zone: 20r/s (burst=2,10,30,100)
Per-Agent Zone: 2r/m (sync enabled)
Per-IP-Method Zone: 30r/s (burst=15)
Per-Country Zone: 100r/s (burst=50)
Trusted IPs Zone: 1000r/s
";
			add_header Content-Type text/plain;
		}
	}
}
`

	fmt.Println("=== Limit Req Zone Example ===\n")

	// Parse the configuration
	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	if err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	// Demonstrate limit_req_zone operations
	fmt.Println("1. Finding all limit_req_zone directives:")
	limitReqZones := conf.FindLimitReqZones()
	fmt.Printf("Found %d limit_req_zone directives:\n", len(limitReqZones))

	for i, zone := range limitReqZones {
		fmt.Printf("  %d. Zone: %s\n", i+1, zone.ZoneName)
		fmt.Printf("     Key: %s\n", zone.Key)
		fmt.Printf("     Size: %s", zone.ZoneSize)

		// Show size in bytes
		sizeBytes, err := zone.GetZoneSizeBytes()
		if err == nil {
			fmt.Printf(" (%d bytes)", sizeBytes)
		}
		fmt.Println()

		fmt.Printf("     Rate: %s", zone.Rate)

		// Show rate details
		rateNum, err1 := zone.GetRateNumber()
		rateUnit, err2 := zone.GetRateUnit()
		if err1 == nil && err2 == nil {
			fmt.Printf(" (%.1f requests per %s)", rateNum, rateUnit)
		}
		fmt.Println()

		if zone.Sync {
			fmt.Printf("     Sync: enabled\n")
		}
		fmt.Println()
	}

	fmt.Println("2. Finding specific zones:")

	// Find zone by name
	globalZone := conf.FindLimitReqZoneByName("global")
	if globalZone != nil {
		fmt.Printf("Found global zone: %s -> %s (%s)\n",
			globalZone.Key, globalZone.ZoneName, globalZone.Rate)

		rateNum, _ := globalZone.GetRateNumber()
		fmt.Printf("  Rate: %.0f requests per second\n", rateNum)

		sizeBytes, _ := globalZone.GetZoneSizeBytes()
		fmt.Printf("  Memory: %d bytes (%.1f MB)\n", sizeBytes, float64(sizeBytes)/(1024*1024))
	}

	// Find zones by key
	ipBasedZones := conf.FindLimitReqZonesByKey("$binary_remote_addr")
	fmt.Printf("\nFound %d zones using $binary_remote_addr:\n", len(ipBasedZones))
	for _, zone := range ipBasedZones {
		fmt.Printf("  - %s: %s\n", zone.ZoneName, zone.Rate)
	}

	fmt.Println("\n3. Manipulating zones:")

	if globalZone != nil {
		fmt.Println("Modifying global zone...")

		// Update rate
		originalRate := globalZone.Rate
		err := globalZone.SetRate("15r/s")
		if err != nil {
			fmt.Printf("Error updating rate: %v\n", err)
		} else {
			fmt.Printf("Updated rate from %s to %s\n", originalRate, globalZone.Rate)
		}

		// Update zone size
		originalSize := globalZone.ZoneSize
		err = globalZone.SetZoneSize("20m")
		if err != nil {
			fmt.Printf("Error updating size: %v\n", err)
		} else {
			fmt.Printf("Updated size from %s to %s\n", originalSize, globalZone.ZoneSize)
		}
	}

	fmt.Println("\n4. Testing validation:")

	if globalZone != nil {
		// Test invalid rate
		err := globalZone.SetRate("invalid")
		if err != nil {
			fmt.Printf("✓ Validation works - rejected invalid rate: %v\n", err)
		}

		// Test invalid size
		err = globalZone.SetZoneSize("invalid")
		if err != nil {
			fmt.Printf("✓ Validation works - rejected invalid size: %v\n", err)
		}

		// Test valid updates
		err = globalZone.SetRate("25.5r/m")
		if err == nil {
			fmt.Printf("✓ Successfully set rate to 25.5r/m\n")
		}
	}

	fmt.Println("\n5. Zone analysis:")

	// Analyze zones by type
	var totalMemory int64
	ratesByUnit := make(map[string][]float64)

	for _, zone := range limitReqZones {
		// Sum up memory usage
		if size, err := zone.GetZoneSizeBytes(); err == nil {
			totalMemory += size
		}

		// Group rates by unit
		if rateNum, err1 := zone.GetRateNumber(); err1 == nil {
			if rateUnit, err2 := zone.GetRateUnit(); err2 == nil {
				ratesByUnit[rateUnit] = append(ratesByUnit[rateUnit], rateNum)
			}
		}
	}

	fmt.Printf("Total memory allocated: %.1f MB\n", float64(totalMemory)/(1024*1024))

	for unit, rates := range ratesByUnit {
		var sum float64
		for _, rate := range rates {
			sum += rate
		}
		avg := sum / float64(len(rates))
		fmt.Printf("Average rate for %s: %.1f requests per %s (%d zones)\n",
			unit, avg, unit, len(rates))
	}

	fmt.Println("\n6. Dumping modified configuration:")

	// Dump the modified configuration
	style := &dumper.Style{
		SortDirectives:    false,
		StartIndent:       0,
		Indent:            4,
		SpaceBeforeBlocks: true,
	}

	output := dumper.DumpConfig(conf, style)
	fmt.Println("\n" + output)

	fmt.Print("\n=== Rate Limiting Features Demonstrated ===\n")
	fmt.Println("✓ Multiple rate limiting zones with different keys")
	fmt.Println("✓ Various memory sizes (k, m, g)")
	fmt.Println("✓ Different rate units (r/s, r/m)")
	fmt.Println("✓ Sync parameter for distributed setups")
	fmt.Println("✓ Rate and size validation")
	fmt.Println("✓ Zone finding by name and key")
	fmt.Println("✓ Memory usage analysis")
	fmt.Println("✓ Integration with limit_req directives")
	fmt.Println("✓ Multi-layer rate limiting strategies")
	fmt.Println("✓ Different policies for different endpoints")
	fmt.Println("✓ Country-based and user-based limiting")
	fmt.Println("✓ API protection and abuse prevention")
}
