package main

import (
	"fmt"
	"log"

	"github.com/lefeck/gonginx/dumper"
	"github.com/lefeck/gonginx/parser"
)

func main() {
	// Example nginx configuration with limit_conn_zone directives
	configString := `
http {
	# Connection limiting zones
	limit_conn_zone $binary_remote_addr zone=addr:10m;
	limit_conn_zone $server_name zone=perserver:5m;
	limit_conn_zone $request_uri zone=perpath:20m;
	limit_conn_zone $cookie_user_id zone=peruser:100m sync;
	limit_conn_zone $http_user_agent zone=peragent:1g;
	
	# Advanced connection limiting with different variables
	limit_conn_zone $binary_remote_addr$request_method zone=per_ip_method:50m;
	limit_conn_zone $geoip_country_code zone=percountry:30m sync;
	limit_conn_zone $http_x_real_ip zone=trusted:15m;

	upstream backend {
		server backend1.example.com:8080;
		server backend2.example.com:8080;
		server backend3.example.com:8080;
	}

	# API server with strict connection limiting
	server {
		listen 80;
		server_name api.example.com;
		
		access_log /var/log/nginx/api_access.log combined;
		error_log /var/log/nginx/api_error.log;
		
		# Global connection limiting per IP
		limit_conn addr 50;
		
		location /api/v1/ {
			# Strict connection limiting for API endpoints
			limit_conn addr 10;
			limit_conn peruser 20;
			
			proxy_pass http://backend;
			proxy_set_header Host $host;
			proxy_set_header X-Real-IP $remote_addr;
			
			# Connection limit status headers
			add_header X-Conn-Limit-Zone "addr,peruser";
			add_header X-Conn-Limit-Max "10,20";
		}
		
		location /api/v2/ {
			# Different connection limiting for v2 API
			limit_conn per_ip_method 15;
			
			proxy_pass http://backend;
			proxy_set_header Host $host;
			proxy_set_header X-Real-IP $remote_addr;
		}
		
		location /streaming/ {
			# Higher connection limits for streaming
			limit_conn addr 100;
			limit_conn peruser 5;
			
			proxy_pass http://backend;
			proxy_http_version 1.1;
			proxy_set_header Upgrade $http_upgrade;
			proxy_set_header Connection "upgrade";
		}
		
		location /health {
			# No connection limiting for health checks
			proxy_pass http://backend;
		}
	}
	
	# Web server with moderate connection limiting
	server {
		listen 80;
		server_name web.example.com;
		
		# Moderate global connection limiting
		limit_conn addr 200;
		limit_conn perserver 1000;
		
		location / {
			# Per-path connection limiting for web content
			limit_conn perpath 30;
			
			proxy_pass http://backend;
			proxy_set_header Host $host;
		}
		
		location /static/ {
			# Higher limits for static content
			limit_conn addr 500;
			limit_conn perserver 2000;
			
			expires 1y;
			add_header Cache-Control "public, immutable";
			proxy_pass http://backend;
		}
		
		location /download/ {
			# Limited concurrent downloads per user
			limit_conn peruser 3;
			
			proxy_pass http://backend;
			proxy_buffering off;
			proxy_read_timeout 300s;
		}
	}
	
	# Admin server with user-based connection limiting
	server {
		listen 80;
		server_name admin.example.com;
		
		# Admin-specific connection limiting
		limit_conn peruser 10;
		
		location /admin/ {
			# Additional IP-based limiting
			limit_conn addr 5;
			
			auth_basic "Admin Area";
			auth_basic_user_file /etc/nginx/.htpasswd;
			
			proxy_pass http://backend;
		}
		
		location /admin/dashboard {
			# Higher limits for dashboard
			limit_conn addr 20;
			limit_conn peruser 5;
			
			proxy_pass http://backend;
		}
		
		location /admin/bulk-operations {
			# Very strict limits for bulk operations
			limit_conn addr 1;
			limit_conn peruser 1;
			
			proxy_pass http://backend;
			proxy_read_timeout 600s;
		}
	}
	
	# Country-specific connection limiting
	server {
		listen 80;
		server_name global.example.com;
		
		# Different limits based on country
		limit_conn percountry 1000;
		limit_conn addr 50;
		
		location / {
			proxy_pass http://backend;
			
			# Add country info to headers
			add_header X-Country-Code $geoip_country_code;
			add_header X-Conn-Limit-Zone "percountry,addr";
		}
	}
	
	# WebSocket server with special connection handling
	server {
		listen 80;
		server_name ws.example.com;
		
		# WebSocket-specific connection limiting
		limit_conn addr 20;
		limit_conn peruser 5;
		
		location /ws {
			# WebSocket upgrade
			proxy_pass http://backend;
			proxy_http_version 1.1;
			proxy_set_header Upgrade $http_upgrade;
			proxy_set_header Connection "upgrade";
			proxy_set_header Host $host;
			
			# Long-lived connections
			proxy_read_timeout 3600s;
			proxy_send_timeout 3600s;
		}
	}
	
	# Monitoring and status server
	server {
		listen 8080;
		server_name status.example.com;
		
		location /nginx_status {
			stub_status on;
			access_log off;
			
			# No connection limiting for monitoring
			allow 127.0.0.1;
			allow 10.0.0.0/8;
			deny all;
		}
		
		location /conn_limit_status {
			return 200 "Connection Limiting Status:
IP Address Zone: 50-500 connections (10-100MB memory)
Per-Server Zone: 1000-2000 connections (5MB memory)
Per-Path Zone: 30 connections (20MB memory)
Per-User Zone: 3-20 connections (100MB memory, sync enabled)
Per-Agent Zone: Variable (1GB memory)
Per-IP-Method Zone: 15 connections (50MB memory)
Per-Country Zone: 1000 connections (30MB memory, sync enabled)
Trusted IPs Zone: Variable (15MB memory)
";
			add_header Content-Type text/plain;
		}
		
		location /conn_recommendations {
			return 200 "Connection Limit Recommendations:
- Use $binary_remote_addr for memory efficiency
- Set sync=true for distributed deployments
- Monitor connection usage patterns
- Adjust limits based on server capacity
- Consider different limits for different content types
- Use multiple zones for fine-grained control
";
			add_header Content-Type text/plain;
		}
	}
}
`

	fmt.Println("=== Limit Conn Zone Example ===\n")

	// Parse the configuration
	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	if err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	// Demonstrate limit_conn_zone operations
	fmt.Println("1. Finding all limit_conn_zone directives:")
	limitConnZones := conf.FindLimitConnZones()
	fmt.Printf("Found %d limit_conn_zone directives:\n", len(limitConnZones))

	for i, zone := range limitConnZones {
		fmt.Printf("  %d. Zone: %s\n", i+1, zone.ZoneName)
		fmt.Printf("     Key: %s\n", zone.Key)
		fmt.Printf("     Size: %s", zone.ZoneSize)

		// Show size in bytes
		sizeBytes, err := zone.GetZoneSizeBytes()
		if err == nil {
			fmt.Printf(" (%d bytes)", sizeBytes)
		}
		fmt.Println()

		// Show estimated max connections
		maxConnections, err := zone.EstimateMaxConnections()
		if err == nil {
			fmt.Printf("     Max Connections: ~%d\n", maxConnections)
		}

		if zone.Sync {
			fmt.Printf("     Sync: enabled\n")
		}

		// Show memory usage estimate
		fmt.Printf("     Memory Usage: %s\n", zone.GetMemoryUsageEstimate())
		fmt.Println()
	}

	fmt.Println("2. Finding specific zones:")

	// Find zone by name
	addrZone := conf.FindLimitConnZoneByName("addr")
	if addrZone != nil {
		fmt.Printf("Found addr zone: %s -> %s (%s)\n",
			addrZone.Key, addrZone.ZoneName, addrZone.ZoneSize)

		maxConnections, _ := addrZone.EstimateMaxConnections()
		fmt.Printf("  Max connections: %d\n", maxConnections)

		sizeBytes, _ := addrZone.GetZoneSizeBytes()
		fmt.Printf("  Memory: %d bytes (%.1f MB)\n", sizeBytes, float64(sizeBytes)/(1024*1024))

		// Get recommendations
		recommendations, err := addrZone.GetRecommendedLimits()
		if err == nil {
			fmt.Printf("  Recommended limits:\n")
			fmt.Printf("    Conservative: %d connections\n", recommendations["conservative"])
			fmt.Printf("    Moderate: %d connections\n", recommendations["moderate"])
			fmt.Printf("    Aggressive: %d connections\n", recommendations["aggressive"])
		}
	}

	// Find zones by key
	ipBasedZones := conf.FindLimitConnZonesByKey("$binary_remote_addr")
	fmt.Printf("\nFound %d zones using $binary_remote_addr:\n", len(ipBasedZones))
	for _, zone := range ipBasedZones {
		maxConn, _ := zone.EstimateMaxConnections()
		fmt.Printf("  - %s: %s (~%d connections)\n", zone.ZoneName, zone.ZoneSize, maxConn)
	}

	fmt.Println("\n3. Manipulating zones:")

	if addrZone != nil {
		fmt.Println("Modifying addr zone...")

		// Update zone size
		originalSize := addrZone.ZoneSize
		err := addrZone.SetZoneSize("20m")
		if err != nil {
			fmt.Printf("Error updating size: %v\n", err)
		} else {
			fmt.Printf("Updated size from %s to %s\n", originalSize, addrZone.ZoneSize)

			// Show new capacity
			newMaxConnections, _ := addrZone.EstimateMaxConnections()
			fmt.Printf("New capacity: ~%d connections\n", newMaxConnections)
		}

		// Enable sync
		addrZone.SetSync(true)
		fmt.Printf("Enabled sync for distributed setup\n")
	}

	fmt.Println("\n4. Testing validation:")

	if addrZone != nil {
		// Test invalid size
		err := addrZone.SetZoneSize("invalid")
		if err != nil {
			fmt.Printf("✓ Validation works - rejected invalid size: %v\n", err)
		}

		// Test valid updates
		err = addrZone.SetZoneSize("50m")
		if err == nil {
			fmt.Printf("✓ Successfully set size to 50m\n")
		}
	}

	fmt.Println("\n5. Zone compatibility analysis:")

	// Test zone compatibility
	ipZones := conf.FindLimitConnZonesByKey("$binary_remote_addr")
	if len(ipZones) >= 2 {
		zone1 := ipZones[0]
		zone2 := ipZones[1]

		if zone1.IsCompatibleWith(zone2) {
			fmt.Printf("✓ Zones '%s' and '%s' are compatible (same key)\n",
				zone1.ZoneName, zone2.ZoneName)
		}
	}

	fmt.Println("\n6. Memory and capacity analysis:")

	// Analyze zones by memory usage
	var totalMemory int64
	totalConnections := int64(0)

	for _, zone := range limitConnZones {
		if size, err := zone.GetZoneSizeBytes(); err == nil {
			totalMemory += size
		}
		if connections, err := zone.EstimateMaxConnections(); err == nil {
			totalConnections += connections
		}
	}

	fmt.Printf("Total memory allocated: %.1f MB\n", float64(totalMemory)/(1024*1024))
	fmt.Printf("Total connection capacity: ~%d connections\n", totalConnections)

	// Group zones by sync setting
	syncZones := 0
	for _, zone := range limitConnZones {
		if zone.Sync {
			syncZones++
		}
	}
	fmt.Printf("Zones with sync enabled: %d/%d\n", syncZones, len(limitConnZones))

	fmt.Println("\n7. Dumping modified configuration:")

	// Dump the modified configuration
	style := &dumper.Style{
		SortDirectives:    false,
		StartIndent:       0,
		Indent:            4,
		SpaceBeforeBlocks: true,
	}

	output := dumper.DumpConfig(conf, style)
	fmt.Println("\n" + output)

	fmt.Print("\n=== Connection Limiting Features Demonstrated ===\n")
	fmt.Println("✓ Multiple connection limiting zones with different keys")
	fmt.Println("✓ Various memory sizes (k, m, g)")
	fmt.Println("✓ Connection capacity estimation")
	fmt.Println("✓ Sync parameter for distributed setups")
	fmt.Println("✓ Size validation and error handling")
	fmt.Println("✓ Zone finding by name and key")
	fmt.Println("✓ Memory usage analysis")
	fmt.Println("✓ Zone compatibility checking")
	fmt.Println("✓ Recommended limit calculations")
	fmt.Println("✓ Integration with limit_conn directives")
	fmt.Println("✓ Multi-layer connection limiting strategies")
	fmt.Println("✓ Different policies for different content types")
	fmt.Println("✓ WebSocket and streaming support")
	fmt.Println("✓ Country-based and user-based limiting")
	fmt.Println("✓ API protection and resource management")
}
