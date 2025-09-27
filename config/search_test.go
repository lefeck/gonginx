package config_test

import (
	"testing"

	"github.com/lefeck/gonginx/parser"
	"gotest.tools/v3/assert"
)

func TestFindServersByName(t *testing.T) {
	t.Parallel()

	configString := `
http {
	server {
		listen 80;
		server_name example.com www.example.com;
		location / {
			root /var/www/html;
		}
	}
	server {
		listen 443;
		server_name api.example.com;
		ssl_certificate /path/to/cert.pem;
		location /api {
			proxy_pass http://backend;
		}
	}
	server {
		listen 80;
		server_name test.com;
		location / {
			return 301 https://$server_name$request_uri;
		}
	}
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	// Test finding servers by exact server_name
	servers := conf.FindServersByName("example.com")
	assert.Equal(t, len(servers), 1)

	// Test finding servers with multiple server_names
	servers = conf.FindServersByName("www.example.com")
	assert.Equal(t, len(servers), 1)

	// Test finding servers that don't exist
	servers = conf.FindServersByName("nonexistent.com")
	assert.Equal(t, len(servers), 0)

	// Test finding another server
	servers = conf.FindServersByName("api.example.com")
	assert.Equal(t, len(servers), 1)
}

func TestFindUpstreamByName(t *testing.T) {
	t.Parallel()

	configString := `
http {
	upstream backend {
		server 192.168.1.1:8080;
		server 192.168.1.2:8080 weight=2;
	}
	upstream api_backend {
		server 10.0.0.1:3000;
		server 10.0.0.2:3000;
	}
	server {
		listen 80;
		location / {
			proxy_pass http://backend;
		}
	}
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	// Test finding existing upstream
	upstream := conf.FindUpstreamByName("backend")
	assert.Assert(t, upstream != nil)
	assert.Equal(t, upstream.UpstreamName, "backend")
	assert.Equal(t, len(upstream.UpstreamServers), 2)

	// Test finding another upstream
	upstream = conf.FindUpstreamByName("api_backend")
	assert.Assert(t, upstream != nil)
	assert.Equal(t, upstream.UpstreamName, "api_backend")

	// Test finding non-existent upstream
	upstream = conf.FindUpstreamByName("nonexistent")
	assert.Assert(t, upstream == nil)
}

func TestFindLocationsByPattern(t *testing.T) {
	t.Parallel()

	configString := `
http {
	server {
		listen 80;
		location / {
			root /var/www/html;
		}
		location /api {
			proxy_pass http://backend;
		}
		location ~ \.php$ {
			fastcgi_pass 127.0.0.1:9000;
		}
	}
	server {
		listen 443;
		location /api {
			proxy_pass https://secure_backend;
		}
		location /static {
			root /var/www/static;
		}
	}
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	// Test finding locations by exact pattern
	locations := conf.FindLocationsByPattern("/")
	assert.Equal(t, len(locations), 1)

	// Test finding multiple locations with same pattern
	locations = conf.FindLocationsByPattern("/api")
	assert.Equal(t, len(locations), 2)

	// Test finding regex location
	locations = conf.FindLocationsByPattern("~ \\.php$")
	assert.Equal(t, len(locations), 1)

	// Test finding non-existent location
	locations = conf.FindLocationsByPattern("/nonexistent")
	assert.Equal(t, len(locations), 0)
}

func TestGetAllSSLCertificates(t *testing.T) {
	t.Parallel()

	configString := `
http {
	server {
		listen 443 ssl;
		ssl_certificate /path/to/cert1.pem;
		ssl_certificate_key /path/to/key1.pem;
		location / {
			root /var/www/html;
		}
	}
	server {
		listen 443 ssl;
		ssl_certificate /path/to/cert2.pem;
		ssl_certificate_key /path/to/key2.pem;
		location / {
			root /var/www/api;
		}
	}
	server {
		listen 80;
		location / {
			return 301 https://$server_name$request_uri;
		}
	}
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	// Test getting all SSL certificates
	certificates := conf.GetAllSSLCertificates()
	assert.Equal(t, len(certificates), 2)

	// Check that we got the expected certificates
	expectedCerts := []string{"/path/to/cert1.pem", "/path/to/cert2.pem"}
	for _, expectedCert := range expectedCerts {
		found := false
		for _, cert := range certificates {
			if cert == expectedCert {
				found = true
				break
			}
		}
		assert.Assert(t, found, "Expected certificate %s not found", expectedCert)
	}
}

func TestGetAllUpstreamServers(t *testing.T) {
	t.Parallel()

	configString := `
http {
	upstream backend1 {
		server 192.168.1.1:8080;
		server 192.168.1.2:8080 weight=2;
	}
	upstream backend2 {
		server 10.0.0.1:3000;
		server 10.0.0.2:3000 backup;
	}
	server {
		listen 80;
		location / {
			proxy_pass http://backend1;
		}
	}
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	// Test getting all upstream servers
	allServers := conf.GetAllUpstreamServers()
	assert.Equal(t, len(allServers), 4)

	// Check that we have servers from both upstreams
	addresses := make(map[string]bool)
	for _, server := range allServers {
		addresses[server.Address] = true
	}

	expectedAddresses := []string{
		"192.168.1.1:8080",
		"192.168.1.2:8080",
		"10.0.0.1:3000",
		"10.0.0.2:3000",
	}

	for _, expectedAddr := range expectedAddresses {
		assert.Assert(t, addresses[expectedAddr], "Expected server address %s not found", expectedAddr)
	}
}
