package main

import (
	"fmt"
	"log"
	"time"

	"github.com/lefeck/gonginx/dumper"
	"github.com/lefeck/gonginx/parser"
)

func main() {
	// Example nginx configuration with proxy_cache_path directives
	configString := `
http {
	# Basic cache configuration
	proxy_cache_path /var/cache/nginx/basic levels=1:2 keys_zone=basic_cache:10m max_size=1g inactive=30m use_temp_path=off;
	
	# High-performance cache for static content
	proxy_cache_path /var/cache/nginx/static 
		levels=1:2 
		keys_zone=static_cache:100m 
		max_size=50g 
		inactive=7d 
		use_temp_path=off 
		manager_files=10000 
		manager_sleep=50ms;
	
	# API cache with shorter expiration
	proxy_cache_path /var/cache/nginx/api 
		levels=2:2 
		keys_zone=api_cache:50m 
		max_size=5g 
		inactive=1h 
		use_temp_path=off 
		manager_files=5000 
		loader_files=2000;
	
	# Large file cache with purger
	proxy_cache_path /var/cache/nginx/files 
		levels=1:2:2 
		keys_zone=files_cache:200m 
		max_size=100g 
		inactive=30d 
		min_free=10g 
		use_temp_path=off 
		purger=on 
		purger_files=100 
		purger_sleep=10ms;
	
	# Microservice cache with complex configuration
	proxy_cache_path /var/cache/nginx/microservices 
		levels=2:1:2 
		keys_zone=micro_cache:128m 
		max_size=20g 
		inactive=12h 
		use_temp_path=off 
		manager_files=8000 
		manager_sleep=100ms 
		manager_threshold=500ms 
		loader_files=4000 
		loader_sleep=50ms 
		loader_threshold=200ms 
		purger=on 
		purger_files=500 
		purger_sleep=20ms 
		purger_threshold=100ms;
	
	# Development/testing cache
	proxy_cache_path /tmp/nginx/dev 
		keys_zone=dev_cache:5m 
		max_size=100m 
		inactive=10m 
		use_temp_path=on;

	upstream backend_static {
		server static1.example.com:8080;
		server static2.example.com:8080;
		server static3.example.com:8080;
	}

	upstream backend_api {
		server api1.example.com:8080;
		server api2.example.com:8080;
		keepalive 16;
	}

	upstream backend_files {
		server files1.example.com:8080;
		server files2.example.com:8080;
	}

	# Static content server with aggressive caching
	server {
		listen 80;
		server_name static.example.com;
		
		access_log /var/log/nginx/static_access.log combined;
		error_log /var/log/nginx/static_error.log;
		
		location / {
			proxy_cache static_cache;
			proxy_cache_valid 200 302 7d;
			proxy_cache_valid 404 1h;
			proxy_cache_use_stale error timeout updating http_500 http_502 http_503 http_504;
			proxy_cache_revalidate on;
			proxy_cache_lock on;
			proxy_cache_lock_timeout 10s;
			
			# Cache control headers
			add_header X-Cache-Status $upstream_cache_status;
			add_header X-Cache-Zone "static_cache";
			
			proxy_pass http://backend_static;
			proxy_set_header Host $host;
			proxy_set_header X-Real-IP $remote_addr;
			proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
		}
		
		location /images/ {
			proxy_cache static_cache;
			proxy_cache_valid 200 30d;
			proxy_cache_valid 404 1h;
			
			# Long-term caching for images
			expires 30d;
			add_header Cache-Control "public, immutable";
			add_header X-Cache-Status $upstream_cache_status;
			
			proxy_pass http://backend_static;
		}
		
		location /css/ {
			proxy_cache static_cache;
			proxy_cache_valid 200 7d;
			proxy_cache_bypass $arg_nocache;
			
			proxy_pass http://backend_static;
		}
		
		location /js/ {
			proxy_cache static_cache;
			proxy_cache_valid 200 7d;
			proxy_cache_bypass $arg_nocache;
			
			proxy_pass http://backend_static;
		}
	}
	
	# API server with moderate caching
	server {
		listen 80;
		server_name api.example.com;
		
		location /api/v1/ {
			proxy_cache api_cache;
			proxy_cache_methods GET HEAD;
			proxy_cache_valid 200 10m;
			proxy_cache_valid 404 1m;
			proxy_cache_key "$scheme$request_method$host$request_uri$is_args$args";
			proxy_cache_bypass $arg_nocache $cookie_nocache;
			proxy_no_cache $arg_nocache $cookie_nocache;
			
			# API-specific headers
			add_header X-Cache-Status $upstream_cache_status;
			add_header X-Cache-Zone "api_cache";
			add_header X-Cache-Key "$scheme$request_method$host$request_uri$is_args$args";
			
			proxy_pass http://backend_api;
			proxy_set_header Host $host;
			proxy_set_header X-Real-IP $remote_addr;
		}
		
		location /api/v1/users/ {
			# User data with shorter cache
			proxy_cache api_cache;
			proxy_cache_valid 200 5m;
			proxy_cache_valid 404 30s;
			proxy_cache_key "$host$request_uri$cookie_user_id";
			
			proxy_pass http://backend_api;
		}
		
		location /api/v1/public/ {
			# Public API with longer cache
			proxy_cache api_cache;
			proxy_cache_valid 200 1h;
			proxy_cache_valid 404 5m;
			
			proxy_pass http://backend_api;
		}
	}
	
	# File download server with large file caching
	server {
		listen 80;
		server_name files.example.com;
		
		client_max_body_size 1g;
		
		location /download/ {
			proxy_cache files_cache;
			proxy_cache_valid 200 30d;
			proxy_cache_valid 404 1h;
			proxy_cache_use_stale error timeout updating;
			proxy_cache_lock on;
			proxy_cache_lock_timeout 30s;
			
			# Large file optimization
			proxy_buffering off;
			proxy_request_buffering off;
			proxy_max_temp_file_size 0;
			
			# Headers for file downloads
			add_header X-Cache-Status $upstream_cache_status;
			add_header X-Cache-Zone "files_cache";
			add_header Content-Disposition 'attachment';
			
			proxy_pass http://backend_files;
			proxy_set_header Host $host;
			proxy_set_header Range $http_range;
		}
		
		location /streaming/ {
			# Streaming content - limited caching
			proxy_cache files_cache;
			proxy_cache_valid 200 1h;
			proxy_cache_min_uses 3;
			
			proxy_pass http://backend_files;
			proxy_http_version 1.1;
			proxy_set_header Connection "";
		}
		
		# Cache purging endpoint
		location ~ /purge(/.*) {
			allow 127.0.0.1;
			allow 10.0.0.0/8;
			deny all;
			
			proxy_cache_purge files_cache "$scheme$request_method$host$1";
			return 200 "Purged cache for $1\n";
		}
	}
	
	# Development server with minimal caching
	server {
		listen 8080;
		server_name dev.example.com;
		
		location / {
			proxy_cache dev_cache;
			proxy_cache_valid 200 1m;
			proxy_cache_bypass $arg_nocache $http_pragma $http_authorization;
			proxy_no_cache $arg_nocache $http_pragma $http_authorization;
			
			# Development headers
			add_header X-Cache-Status $upstream_cache_status;
			add_header X-Cache-Zone "dev_cache";
			add_header X-Debug-Cache-Key "$scheme$request_method$host$request_uri";
			
			proxy_pass http://backend_api;
		}
	}
	
	# Cache statistics and management
	server {
		listen 9090;
		server_name cache-admin.example.com;
		
		allow 127.0.0.1;
		allow 10.0.0.0/8;
		deny all;
		
		location /cache/status {
			return 200 "Cache Status:
Static Cache: 50GB max, 7d inactive, levels 1:2
API Cache: 5GB max, 1h inactive, levels 2:2
Files Cache: 100GB max, 30d inactive, levels 1:2:2, purger enabled
Microservices Cache: 20GB max, 12h inactive, levels 2:1:2, purger enabled
Dev Cache: 100MB max, 10m inactive, no levels
";
			add_header Content-Type text/plain;
		}
		
		location /cache/zones {
			return 200 "Cache Zones:
basic_cache: 10MB memory
static_cache: 100MB memory
api_cache: 50MB memory
files_cache: 200MB memory
micro_cache: 128MB memory
dev_cache: 5MB memory
";
			add_header Content-Type text/plain;
		}
		
		location /cache/recommendations {
			return 200 "Cache Recommendations:
- Use levels=1:2 for most cases
- Set appropriate inactive times based on content type
- Enable purger for dynamic content
- Monitor cache hit ratios
- Adjust manager/loader parameters for large caches
- Use use_temp_path=off for better performance
- Set min_free for disk space management
";
			add_header Content-Type text/plain;
		}
	}
}
`

	fmt.Println("=== Proxy Cache Path Example ===\n")

	// Parse the configuration
	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	if err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	// Demonstrate proxy_cache_path operations
	fmt.Println("1. Finding all proxy_cache_path directives:")
	proxyCachePaths := conf.FindProxyCachePaths()
	fmt.Printf("Found %d proxy_cache_path directives:\n", len(proxyCachePaths))

	for i, cachePath := range proxyCachePaths {
		fmt.Printf("  %d. Zone: %s\n", i+1, cachePath.KeysZoneName)
		fmt.Printf("     Path: %s\n", cachePath.Path)
		fmt.Printf("     Keys Zone Size: %s", cachePath.KeysZoneSize)

		// Show size in bytes
		keysZoneBytes, err := cachePath.GetKeysZoneSizeBytes()
		if err == nil {
			fmt.Printf(" (%d bytes)", keysZoneBytes)
		}
		fmt.Println()

		// Show estimated key capacity
		keyCapacity, err := cachePath.EstimateKeyCapacity()
		if err == nil {
			fmt.Printf("     Key Capacity: ~%d keys\n", keyCapacity)
		}

		if cachePath.MaxSize != "" {
			fmt.Printf("     Max Size: %s", cachePath.MaxSize)
			maxSizeBytes, err := cachePath.GetMaxSizeBytes()
			if err == nil {
				fmt.Printf(" (%d bytes)", maxSizeBytes)
			}
			fmt.Println()
		}

		if cachePath.Inactive != "" {
			fmt.Printf("     Inactive: %s", cachePath.Inactive)
			inactiveDuration, err := cachePath.GetInactiveDuration()
			if err == nil {
				fmt.Printf(" (%v)", inactiveDuration)
			}
			fmt.Println()
		}

		if cachePath.Levels != "" {
			fmt.Printf("     Levels: %s", cachePath.Levels)
			levels, err := cachePath.GetLevelsDepth()
			if err == nil {
				fmt.Printf(" (depths: %v)", levels)
			}
			fmt.Println()
		}

		if cachePath.UseTemPath != nil {
			fmt.Printf("     Use Temp Path: %t\n", *cachePath.UseTemPath)
		}

		if cachePath.Purger != nil {
			fmt.Printf("     Purger: %s\n", map[bool]string{true: "enabled", false: "disabled"}[*cachePath.Purger])
		}

		if cachePath.MinFree != "" {
			fmt.Printf("     Min Free: %s", cachePath.MinFree)
			minFreeBytes, err := cachePath.GetMinFreeBytes()
			if err == nil {
				fmt.Printf(" (%d bytes)", minFreeBytes)
			}
			fmt.Println()
		}

		fmt.Println()
	}

	fmt.Println("2. Finding specific caches:")

	// Find cache by zone name
	staticCache := conf.FindProxyCachePathByZone("static_cache")
	if staticCache != nil {
		fmt.Printf("Found static_cache: %s -> %s (%s)\n",
			staticCache.Path, staticCache.KeysZoneName, staticCache.KeysZoneSize)

		keyCapacity, _ := staticCache.EstimateKeyCapacity()
		fmt.Printf("  Key capacity: %d keys\n", keyCapacity)

		keysZoneBytes, _ := staticCache.GetKeysZoneSizeBytes()
		fmt.Printf("  Memory: %d bytes (%.1f MB)\n", keysZoneBytes, float64(keysZoneBytes)/(1024*1024))

		if staticCache.MaxSize != "" {
			maxSizeBytes, _ := staticCache.GetMaxSizeBytes()
			fmt.Printf("  Max cache size: %d bytes (%.1f GB)\n", maxSizeBytes, float64(maxSizeBytes)/(1024*1024*1024))
		}

		if staticCache.Inactive != "" {
			inactiveDuration, _ := staticCache.GetInactiveDuration()
			fmt.Printf("  Inactive time: %v\n", inactiveDuration)
		}
	}

	// Find caches by path
	varCaches := conf.FindProxyCachePathsByPath("/var/cache/nginx/basic")
	fmt.Printf("\nFound %d caches using path '/var/cache/nginx/basic'\n", len(varCaches))

	fmt.Println("\n3. Manipulating caches:")

	if staticCache != nil {
		fmt.Println("Modifying static_cache...")

		// Update max size
		originalMaxSize := staticCache.MaxSize
		err := staticCache.SetMaxSize("100g")
		if err != nil {
			fmt.Printf("Error updating max size: %v\n", err)
		} else {
			fmt.Printf("Updated max size from %s to %s\n", originalMaxSize, staticCache.MaxSize)

			// Show new capacity
			newMaxSizeBytes, _ := staticCache.GetMaxSizeBytes()
			fmt.Printf("New max cache size: %.1f GB\n", float64(newMaxSizeBytes)/(1024*1024*1024))
		}

		// Update inactive time
		err = staticCache.SetInactive("14d")
		if err != nil {
			fmt.Printf("Error updating inactive time: %v\n", err)
		} else {
			fmt.Printf("Updated inactive time to %s\n", staticCache.Inactive)
			newInactiveDuration, _ := staticCache.GetInactiveDuration()
			fmt.Printf("New inactive duration: %v\n", newInactiveDuration)
		}

		// Enable purger
		staticCache.SetPurger(true)
		fmt.Printf("Enabled purger for better cache management\n")
	}

	fmt.Println("\n4. Testing validation:")

	if staticCache != nil {
		// Test invalid size
		err := staticCache.SetMaxSize("invalid")
		if err != nil {
			fmt.Printf("✓ Validation works - rejected invalid size: %v\n", err)
		}

		// Test invalid time
		err = staticCache.SetInactive("invalid")
		if err != nil {
			fmt.Printf("✓ Validation works - rejected invalid time: %v\n", err)
		}

		// Test valid updates
		err = staticCache.SetMaxSize("200g")
		if err == nil {
			fmt.Printf("✓ Successfully set max size to 200g\n")
		}

		err = staticCache.SetInactive("21d")
		if err == nil {
			fmt.Printf("✓ Successfully set inactive time to 21d\n")
		}
	}

	fmt.Println("\n5. Cache capacity and memory analysis:")

	// Analyze caches by memory usage
	var totalKeysZoneMemory int64
	var totalCacheSize int64
	totalKeyCapacity := int64(0)

	for _, cachePath := range proxyCachePaths {
		if size, err := cachePath.GetKeysZoneSizeBytes(); err == nil {
			totalKeysZoneMemory += size
		}
		if size, err := cachePath.GetMaxSizeBytes(); err == nil {
			totalCacheSize += size
		}
		if capacity, err := cachePath.EstimateKeyCapacity(); err == nil {
			totalKeyCapacity += capacity
		}
	}

	fmt.Printf("Total keys zone memory: %.1f MB\n", float64(totalKeysZoneMemory)/(1024*1024))
	fmt.Printf("Total cache disk capacity: %.1f GB\n", float64(totalCacheSize)/(1024*1024*1024))
	fmt.Printf("Total key capacity: ~%d keys\n", totalKeyCapacity)

	// Group caches by features
	purgerEnabledCount := 0
	tempPathDisabledCount := 0
	levelsConfiguredCount := 0

	for _, cachePath := range proxyCachePaths {
		if cachePath.Purger != nil && *cachePath.Purger {
			purgerEnabledCount++
		}
		if cachePath.UseTemPath != nil && !*cachePath.UseTemPath {
			tempPathDisabledCount++
		}
		if cachePath.Levels != "" {
			levelsConfiguredCount++
		}
	}

	fmt.Printf("Caches with purger enabled: %d/%d\n", purgerEnabledCount, len(proxyCachePaths))
	fmt.Printf("Caches with use_temp_path=off: %d/%d\n", tempPathDisabledCount, len(proxyCachePaths))
	fmt.Printf("Caches with levels configured: %d/%d\n", levelsConfiguredCount, len(proxyCachePaths))

	fmt.Println("\n6. Cache performance analysis:")

	// Analyze cache configurations for performance
	for _, cachePath := range proxyCachePaths {
		fmt.Printf("\nCache: %s\n", cachePath.KeysZoneName)

		// Check performance settings
		performanceScore := 0
		recommendations := []string{}

		if cachePath.UseTemPath != nil && !*cachePath.UseTemPath {
			performanceScore += 20
		} else {
			recommendations = append(recommendations, "Enable use_temp_path=off for better performance")
		}

		if cachePath.Levels != "" {
			performanceScore += 20
		} else {
			recommendations = append(recommendations, "Configure levels for better directory structure")
		}

		if cachePath.ManagerFiles != nil {
			performanceScore += 10
		}

		if cachePath.LoaderFiles != nil {
			performanceScore += 10
		}

		if cachePath.Purger != nil && *cachePath.Purger {
			performanceScore += 15
		}

		// Check size ratios
		if keyCapacity, err := cachePath.EstimateKeyCapacity(); err == nil && keyCapacity > 1000 {
			performanceScore += 15
		}

		// Check inactive time
		if inactiveDuration, err := cachePath.GetInactiveDuration(); err == nil {
			if inactiveDuration >= 1*time.Hour {
				performanceScore += 10
			}
		}

		fmt.Printf("  Performance Score: %d/100\n", performanceScore)
		if len(recommendations) > 0 {
			fmt.Printf("  Recommendations:\n")
			for _, rec := range recommendations {
				fmt.Printf("    - %s\n", rec)
			}
		}
	}

	fmt.Println("\n7. Dumping modified configuration:")

	// Dump the modified configuration
	style := &dumper.Style{
		SortDirectives:    false,
		StartIndent:       0,
		Indent:            4,
		SpaceBeforeBlocks: true,
	}

	output := dumper.DumpConfig(conf, style)
	fmt.Println("\n" + output[:2000] + "...") // Show first 2000 characters

	fmt.Print("\n=== Proxy Cache Path Features Demonstrated ===\n")
	fmt.Println("✓ Multiple cache paths with different configurations")
	fmt.Println("✓ Various memory sizes and disk capacities")
	fmt.Println("✓ Directory levels for efficient file organization")
	fmt.Println("✓ Inactive time configuration and parsing")
	fmt.Println("✓ Cache key capacity estimation")
	fmt.Println("✓ Manager and loader process tuning")
	fmt.Println("✓ Purger configuration for cache maintenance")
	fmt.Println("✓ Use temp path optimization")
	fmt.Println("✓ Min free space management")
	fmt.Println("✓ Cache finding by zone name and path")
	fmt.Println("✓ Size and time validation")
	fmt.Println("✓ Performance analysis and recommendations")
	fmt.Println("✓ Integration with proxy_cache directives")
	fmt.Println("✓ Multiple cache strategies (static, API, files, dev)")
	fmt.Println("✓ Cache purging and management endpoints")
	fmt.Println("✓ Memory and disk usage optimization")
	fmt.Println("✓ Complex multi-parameter configurations")
}
