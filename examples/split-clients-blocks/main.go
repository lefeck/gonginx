package main

import (
	"fmt"
	"log"

	"github.com/lefeck/gonginx/dumper"
	"github.com/lefeck/gonginx/parser"
)

func main() {
	// Example nginx configuration with split_clients blocks
	configString := `
http {
	# A/B testing for different UI versions
	split_clients $remote_addr $ui_version {
		0.5%     v1;
		2.0%     v2;
		10%      v3;
		*        v4;
	}
	
	# Feature flag testing
	split_clients $request_id $feature_flag {
		25%      enabled;
		*        disabled;
	}
	
	# Mobile app version testing
	split_clients $http_user_agent $mobile_version {
		5%       beta;
		15%      stable;
		*        legacy;
	}
	
	# Canary deployment testing
	split_clients $remote_addr $backend_version {
		1%       canary;
		9%       staging;
		*        production;
	}
	
	# Cache strategy testing
	split_clients $request_uri $cache_strategy {
		20%      aggressive;
		30%      moderate;
		*        conservative;
	}

	upstream backend_v1 {
		server v1.example.com:8080;
		server v1-backup.example.com:8080;
	}
	
	upstream backend_v2 {
		server v2.example.com:8080;
		server v2-backup.example.com:8080;
	}
	
	upstream backend_v3 {
		server v3.example.com:8080;
		server v3-backup.example.com:8080;
	}
	
	upstream backend_v4 {
		server v4.example.com:8080;
		server v4-backup.example.com:8080;
	}
	
	upstream backend_canary {
		server canary.example.com:8080;
	}
	
	upstream backend_staging {
		server staging.example.com:8080;
	}
	
	upstream backend_production {
		server prod1.example.com:8080;
		server prod2.example.com:8080;
		server prod3.example.com:8080;
	}

	# Map UI version to backend
	map $ui_version $backend_ui {
		default backend_v4;
		v1 backend_v1;
		v2 backend_v2;
		v3 backend_v3;
		v4 backend_v4;
	}
	
	# Map backend version to upstream
	map $backend_version $backend_deploy {
		default backend_production;
		canary backend_canary;
		staging backend_staging;
		production backend_production;
	}

	server {
		listen 80;
		server_name app.example.com;
		
		# Add A/B testing headers
		add_header X-UI-Version $ui_version;
		add_header X-Feature-Flag $feature_flag;
		add_header X-Mobile-Version $mobile_version;
		add_header X-Backend-Version $backend_version;
		add_header X-Cache-Strategy $cache_strategy;
		
		location / {
			# Route to appropriate backend based on UI version
			proxy_pass http://$backend_ui;
			proxy_set_header Host $host;
			proxy_set_header X-Real-IP $remote_addr;
			proxy_set_header X-UI-Version $ui_version;
		}
		
		location /api/ {
			# Use backend deployment strategy
			proxy_pass http://$backend_deploy;
			proxy_set_header Host $host;
			proxy_set_header X-Real-IP $remote_addr;
			proxy_set_header X-Backend-Version $backend_version;
		}
		
		location /mobile/ {
			# Special handling for mobile apps
			if ($mobile_version = "beta") {
				proxy_pass http://mobile-beta.example.com;
			}
			if ($mobile_version = "stable") {
				proxy_pass http://mobile-stable.example.com;
			}
			# Default to legacy
			proxy_pass http://mobile-legacy.example.com;
		}
		
		location ~* \.(css|js|png|jpg|jpeg|gif|ico|svg)$ {
			# Cache strategy based on split testing
			if ($cache_strategy = "aggressive") {
				expires 1y;
				add_header Cache-Control "public, immutable";
			}
			if ($cache_strategy = "moderate") {
				expires 30d;
				add_header Cache-Control "public";
			}
			if ($cache_strategy = "conservative") {
				expires 1d;
				add_header Cache-Control "public, must-revalidate";
			}
			
			# Serve static files
			try_files $uri @fallback;
		}
		
		location @fallback {
			proxy_pass http://$backend_ui;
		}
	}
	
	# Analytics and monitoring server
	server {
		listen 8080;
		server_name analytics.example.com;
		
		access_log /var/log/nginx/analytics.log combined;
		
		location /stats {
			return 200 "UI: $ui_version, Feature: $feature_flag, Mobile: $mobile_version, Backend: $backend_version, Cache: $cache_strategy\n";
			add_header Content-Type text/plain;
		}
		
		# Feature flag endpoint
		location /feature-check {
			if ($feature_flag = "enabled") {
				return 200 '{"feature_enabled": true}';
			}
			return 200 '{"feature_enabled": false}';
			add_header Content-Type application/json;
		}
	}
	
	# Admin server for A/B test management
	server {
		listen 9090;
		server_name admin.example.com;
		
		location /ab-test-status {
			return 200 "
A/B Test Status:
- UI Version Distribution: v1(0.5%), v2(2.0%), v3(10%), v4(87.5%)
- Feature Flag: enabled(25%), disabled(75%)
- Mobile Version: beta(5%), stable(15%), legacy(80%)
- Backend Version: canary(1%), staging(9%), production(90%)
- Cache Strategy: aggressive(20%), moderate(30%), conservative(50%)
";
			add_header Content-Type text/plain;
		}
	}
}
`

	fmt.Println("=== Split Clients Blocks Example ===\n")

	// Parse the configuration
	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	if err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	// Demonstrate split_clients block operations
	fmt.Println("1. Finding all split_clients blocks:")
	splitClients := conf.FindSplitClients()
	fmt.Printf("Found %d split_clients blocks:\n", len(splitClients))
	for i, sc := range splitClients {
		fmt.Printf("  %d. Variable: %s -> %s, Entries: %d\n",
			i+1, sc.Variable, sc.MappedVariable, len(sc.Entries))

		// Show total percentage and wildcard
		total, err := sc.GetTotalPercentage()
		if err != nil {
			fmt.Printf("     - Error calculating percentage: %v\n", err)
		} else {
			fmt.Printf("     - Total allocated: %.1f%%\n", total)
		}

		if sc.HasWildcard() {
			fmt.Printf("     - Wildcard value: %s\n", sc.GetWildcardValue())
		}

		// Show entries
		fmt.Printf("     - Entries:\n")
		for _, entry := range sc.Entries {
			fmt.Printf("       %s -> %s\n", entry.Percentage, entry.Value)
		}
	}

	fmt.Println("\n2. Finding specific split_clients blocks:")

	// Find split_clients by variable
	uiVersionSplit := conf.FindSplitClientsByVariable("$ui_version")
	if uiVersionSplit != nil {
		fmt.Printf("Found UI version split: %s -> %s\n", uiVersionSplit.Variable, uiVersionSplit.MappedVariable)
		fmt.Printf("  Total percentage allocated: ")
		total, err := uiVersionSplit.GetTotalPercentage()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("%.1f%%\n", total)
		}

		fmt.Printf("  Entries breakdown:\n")
		for _, entry := range uiVersionSplit.Entries {
			fmt.Printf("    %s: %s\n", entry.Percentage, entry.Value)
		}
	}

	// Find split_clients by both variables
	featureFlagSplit := conf.FindSplitClientsByVariables("$request_id", "$feature_flag")
	if featureFlagSplit != nil {
		fmt.Printf("\nFound feature flag split: %s -> %s\n", featureFlagSplit.Variable, featureFlagSplit.MappedVariable)

		enabledEntries := featureFlagSplit.GetEntriesByValue("enabled")
		fmt.Printf("  Entries with 'enabled' value: %d\n", len(enabledEntries))
		for _, entry := range enabledEntries {
			fmt.Printf("    %s -> %s\n", entry.Percentage, entry.Value)
		}
	}

	fmt.Println("\n3. Manipulating split_clients blocks:")

	// Add new entry to UI version split
	if uiVersionSplit != nil {
		fmt.Println("Adding new UI version mapping...")
		err := uiVersionSplit.AddEntry("0.1%", "v5_experimental")
		if err != nil {
			fmt.Printf("Error adding entry: %v\n", err)
		} else {
			fmt.Println("Successfully added: 0.1% -> v5_experimental")
		}

		// Check new total
		total, err := uiVersionSplit.GetTotalPercentage()
		if err != nil {
			fmt.Printf("Error calculating new total: %v\n", err)
		} else {
			fmt.Printf("New total percentage: %.1f%%\n", total)
		}
	}

	// Test validation
	fmt.Println("\n4. Testing validation:")
	if uiVersionSplit != nil {
		// Try to add invalid percentage
		err := uiVersionSplit.AddEntry("invalid%", "test")
		if err != nil {
			fmt.Printf("✓ Validation works - rejected invalid percentage: %v\n", err)
		}

		// Try to add percentage out of range
		err = uiVersionSplit.AddEntry("150%", "test")
		if err != nil {
			fmt.Printf("✓ Validation works - rejected out of range: %v\n", err)
		}
	}

	// Create a new split_clients block for demonstration
	fmt.Println("\n5. Creating new split_clients for language selection:")
	cacheSplit := conf.FindSplitClientsByVariable("$cache_strategy")
	if cacheSplit != nil {
		// Remove an entry
		removed := cacheSplit.RemoveEntry("20%")
		if removed {
			fmt.Println("Removed 20% entry for aggressive caching")
		}

		// Add new entries
		err := cacheSplit.AddEntry("10%", "experimental")
		if err != nil {
			fmt.Printf("Error adding experimental cache: %v\n", err)
		} else {
			fmt.Println("Added 10% -> experimental cache strategy")
		}
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

	fmt.Print("\n=== Split Clients Features Demonstrated ===\n")
	fmt.Println("✓ Percentage-based traffic splitting")
	fmt.Println("✓ Wildcard (*) for remaining traffic")
	fmt.Println("✓ A/B testing for UI versions")
	fmt.Println("✓ Feature flag distribution")
	fmt.Println("✓ Canary deployment strategies")
	fmt.Println("✓ Mobile app version testing")
	fmt.Println("✓ Cache strategy optimization")
	fmt.Println("✓ Integration with map blocks and upstreams")
	fmt.Println("✓ Percentage validation and error handling")
	fmt.Println("✓ Dynamic entry manipulation")
	fmt.Println("✓ Multiple split_clients blocks")
	fmt.Println("✓ Analytics and monitoring integration")
}
