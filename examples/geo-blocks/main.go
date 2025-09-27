package main

import (
	"fmt"
	"log"

	"github.com/lefeck/gonginx/dumper"
	"github.com/lefeck/gonginx/parser"
)

func main() {
	// Example nginx configuration with geo blocks
	configString := `
http {
	# Country detection based on client IP
	geo $country {
		default ZZ;
		127.0.0.0/8 US;
		10.0.0.0/8 US;
		192.168.1.0/24 CN;
		203.208.60.0/24 CN;
		84.54.0.0/16 DE;
		217.69.128.0/18 DE;
	}
	
	# Region detection with IP ranges
	geo $remote_addr $region {
		ranges;
		default unknown;
		127.0.0.1-127.0.0.255 local;
		10.0.0.1-10.0.0.100 internal;
		192.168.1.1-192.168.1.100 office;
		172.16.0.1-172.16.255.254 datacenter;
	}
	
	# City detection with proxy configuration
	geo $city {
		proxy 192.168.1.0/24;
		proxy 10.0.0.0/8;
		proxy_recursive;
		delete 127.0.0.1;
		default unknown;
		192.168.0.0/16 Beijing;
		10.0.0.0/8 Shanghai;
		172.16.0.0/12 Guangzhou;
		203.208.60.0/24 Shenzhen;
	}
	
	# ISP detection
	geo $isp {
		default other;
		1.2.4.0/22 chinanet;
		1.4.1.0/24 chinanet;
		58.14.0.0/15 unicom;
		61.135.0.0/16 unicom;
		219.128.0.0/11 cmcc;
		223.0.0.0/11 cmcc;
	}

	upstream backend_us {
		server us1.example.com:8080;
		server us2.example.com:8080;
	}
	
	upstream backend_cn {
		server cn1.example.com:8080;
		server cn2.example.com:8080;
	}
	
	upstream backend_de {
		server de1.example.com:8080;
		server de2.example.com:8080;
	}
	
	upstream backend_default {
		server global.example.com:8080;
	}

	# Map country to backend
	map $country $backend {
		default backend_default;
		US backend_us;
		CN backend_cn;
		DE backend_de;
	}

	server {
		listen 80;
		server_name _;
		
		# Use geo variables for dynamic configuration
		location / {
			# Add geo information to headers
			add_header X-Country $country;
			add_header X-Region $region;
			add_header X-City $city;
			add_header X-ISP $isp;
			
			# Route to appropriate backend based on country
			proxy_pass http://$backend;
			proxy_set_header Host $host;
			proxy_set_header X-Real-IP $remote_addr;
			proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
		}
		
		# Special handling for local access
		location /admin {
			# Only allow access from internal networks
			if ($region != "internal") {
				return 403 "Access denied from $region";
			}
			
			proxy_pass http://admin.internal.com;
		}
		
		# CDN optimization based on ISP
		location ~* \.(css|js|png|jpg|jpeg|gif|ico|svg)$ {
			expires 1y;
			add_header Cache-Control "public, immutable";
			
			# Use different CDN based on ISP
			if ($isp = "chinanet") {
				proxy_pass http://cdn-chinanet.example.com;
			}
			if ($isp = "unicom") {
				proxy_pass http://cdn-unicom.example.com;
			}
			if ($isp = "cmcc") {
				proxy_pass http://cdn-cmcc.example.com;
			}
			# Default CDN for other ISPs
			proxy_pass http://cdn-global.example.com;
		}
	}
	
	# Logging server with geo information
	server {
		listen 8080;
		server_name logs.example.com;
		
		access_log /var/log/nginx/geo_access.log combined;
		
		location /log {
			return 200 "IP: $remote_addr, Country: $country, Region: $region, City: $city, ISP: $isp\n";
			add_header Content-Type text/plain;
		}
	}
}
`

	fmt.Println("=== Geo Blocks Example ===\n")

	// Parse the configuration
	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	if err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	// Demonstrate geo block operations
	fmt.Println("1. Finding all geo blocks:")
	geos := conf.FindGeos()
	fmt.Printf("Found %d geo blocks:\n", len(geos))
	for i, geo := range geos {
		fmt.Printf("  %d. Variable: %s, Source: %s, Entries: %d\n",
			i+1, geo.Variable, geo.SourceAddress, len(geo.Entries))

		// Show special configurations
		if geo.Ranges {
			fmt.Printf("     - Uses IP ranges format\n")
		}
		if geo.ProxyRecursive {
			fmt.Printf("     - Uses proxy recursive lookup\n")
		}
		if len(geo.Proxy) > 0 {
			fmt.Printf("     - Trusted proxies: %v\n", geo.Proxy)
		}
		if len(geo.Delete) > 0 {
			fmt.Printf("     - Deleted IPs: %v\n", geo.Delete)
		}
		if geo.DefaultValue != "" {
			fmt.Printf("     - Default value: %s\n", geo.DefaultValue)
		}
	}

	fmt.Println("\n2. Finding specific geo blocks:")

	// Find geo by variable
	countryGeo := conf.FindGeoByVariable("$country")
	if countryGeo != nil {
		fmt.Printf("Found country geo: %s -> %s\n", countryGeo.SourceAddress, countryGeo.Variable)
		fmt.Printf("  Default value: %s\n", countryGeo.GetDefaultValue())
		fmt.Printf("  Network mappings:\n")
		for _, entry := range countryGeo.Entries {
			fmt.Printf("    %s -> %s\n", entry.Network, entry.Value)
		}
	}

	// Find geo by both variables
	regionGeo := conf.FindGeoByVariables("$remote_addr", "$region")
	if regionGeo != nil {
		fmt.Printf("\nFound region geo: %s -> %s\n", regionGeo.SourceAddress, regionGeo.Variable)
		fmt.Printf("  Uses ranges: %t\n", regionGeo.Ranges)
		fmt.Printf("  Range mappings:\n")
		for _, entry := range regionGeo.Entries {
			fmt.Printf("    %s -> %s\n", entry.Network, entry.Value)
		}
	}

	fmt.Println("\n3. Manipulating geo blocks:")

	// Add new entry to country geo
	if countryGeo != nil {
		fmt.Println("Adding new country mapping...")
		err := countryGeo.AddEntry("8.8.8.0/24", "US")
		if err != nil {
			fmt.Printf("Error adding entry: %v\n", err)
		} else {
			fmt.Println("Successfully added: 8.8.8.0/24 -> US")
		}

		// Update default value
		countryGeo.SetDefaultValue("UNKNOWN")
		fmt.Printf("Updated default value to: %s\n", countryGeo.GetDefaultValue())
	}

	// Create a new geo block
	fmt.Println("\n4. Adding new geo block:")
	cityGeo := conf.FindGeoByVariable("$city")
	if cityGeo != nil {
		// Add proxy configuration
		cityGeo.AddProxy("203.0.113.0/24")
		fmt.Println("Added new proxy: 203.0.113.0/24")

		// Add delete entry
		cityGeo.AddDelete("198.51.100.1")
		fmt.Println("Added delete entry: 198.51.100.1")

		// Enable proxy recursive
		cityGeo.SetProxyRecursive(true)
		fmt.Println("Enabled proxy recursive lookup")
	}

	fmt.Println("\n5. Dumping modified configuration:")

	// Dump the modified configuration
	style := &dumper.Style{
		SortDirectives:    false,
		StartIndent:       0,
		Indent:            4,
		SpaceBeforeBlocks: true,
	}

	output := dumper.DumpConfig(conf, style)
	fmt.Println("\n" + output)

	fmt.Print("\n=== Geo Block Features Demonstrated ===\n")
	fmt.Println("✓ CIDR notation support (192.168.1.0/24)")
	fmt.Println("✓ IP range support (127.0.0.1-127.0.0.255)")
	fmt.Println("✓ Default value configuration")
	fmt.Println("✓ Proxy configuration with trusted networks")
	fmt.Println("✓ Proxy recursive lookup")
	fmt.Println("✓ Delete entries for inherited geo blocks")
	fmt.Println("✓ Multiple geo blocks with different variables")
	fmt.Println("✓ Integration with map blocks and upstreams")
	fmt.Println("✓ Dynamic backend selection based on geography")
	fmt.Println("✓ ISP-based CDN optimization")
	fmt.Println("✓ Region-based access control")
}
