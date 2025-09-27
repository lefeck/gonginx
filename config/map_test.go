package config_test

import (
	"testing"

	"github.com/lefeck/gonginx/parser"
	"gotest.tools/v3/assert"
)

func TestMapParsing(t *testing.T) {
	t.Parallel()

	configString := `
http {
	map $http_host $backend {
		default backend1;
		example.com backend2;
		api.example.com backend3;
		~^www\. backend4;
	}
	
	map $request_method $limit {
		default "";
		POST $binary_remote_addr;
		PUT $binary_remote_addr;
	}
	
	server {
		listen 80;
		proxy_pass http://$backend;
	}
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	// Test finding maps
	maps := conf.FindMaps()
	assert.Equal(t, len(maps), 2)

	// Test first map
	hostMap := maps[0]
	assert.Equal(t, hostMap.Variable, "$http_host")
	assert.Equal(t, hostMap.MappedVariable, "$backend")
	assert.Equal(t, len(hostMap.Mappings), 4)

	// Test default mapping
	assert.Equal(t, hostMap.GetDefaultValue(), "backend1")

	// Test specific mappings
	expectedMappings := map[string]string{
		"default":         "backend1",
		"example.com":     "backend2",
		"api.example.com": "backend3",
		"~^www\\.":        "backend4",
	}

	for _, mapping := range hostMap.Mappings {
		expectedValue, exists := expectedMappings[mapping.Pattern]
		assert.Assert(t, exists, "Unexpected pattern: %s", mapping.Pattern)
		assert.Equal(t, mapping.Value, expectedValue)
	}

	// Test second map
	methodMap := maps[1]
	assert.Equal(t, methodMap.Variable, "$request_method")
	assert.Equal(t, methodMap.MappedVariable, "$limit")
	assert.Equal(t, len(methodMap.Mappings), 3)
}

func TestFindMapByVariables(t *testing.T) {
	t.Parallel()

	configString := `
http {
	map $http_host $backend {
		default backend1;
		example.com backend2;
	}
	
	map $request_method $limit {
		default "";
		POST $binary_remote_addr;
	}
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	// Test finding existing map
	hostMap := conf.FindMapByVariables("$http_host", "$backend")
	assert.Assert(t, hostMap != nil)
	assert.Equal(t, hostMap.Variable, "$http_host")
	assert.Equal(t, hostMap.MappedVariable, "$backend")

	// Test finding another map
	methodMap := conf.FindMapByVariables("$request_method", "$limit")
	assert.Assert(t, methodMap != nil)
	assert.Equal(t, methodMap.Variable, "$request_method")

	// Test finding non-existent map
	nonExistent := conf.FindMapByVariables("$nonexistent", "$also_nonexistent")
	assert.Assert(t, nonExistent == nil)
}

func TestMapManipulation(t *testing.T) {
	t.Parallel()

	configString := `
http {
	map $http_host $backend {
		default backend1;
		example.com backend2;
	}
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	maps := conf.FindMaps()
	assert.Equal(t, len(maps), 1)

	hostMap := maps[0]

	// Test adding new mapping
	originalCount := len(hostMap.Mappings)
	hostMap.AddMapping("test.com", "backend3")
	assert.Equal(t, len(hostMap.Mappings), originalCount+1)

	// Verify the new mapping was added
	found := false
	for _, mapping := range hostMap.Mappings {
		if mapping.Pattern == "test.com" && mapping.Value == "backend3" {
			found = true
			break
		}
	}
	assert.Assert(t, found, "New mapping was not added correctly")

	// Test updating default value
	hostMap.SetDefaultValue("new_default_backend")
	assert.Equal(t, hostMap.GetDefaultValue(), "new_default_backend")
}

func TestMapWithComments(t *testing.T) {
	t.Parallel()

	configString := `
http {
	# Map configuration for backend selection
	map $http_host $backend { # inline comment
		default backend1; # default backend
		example.com backend2; # main site
		# regex pattern for www subdomains
		~^www\. backend4;
	}
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	maps := conf.FindMaps()
	assert.Equal(t, len(maps), 1)

	hostMap := maps[0]

	// Test that comments are preserved
	assert.Assert(t, len(hostMap.GetComment()) > 0)

	// Test that mapping comments are preserved
	for _, mapping := range hostMap.Mappings {
		if mapping.Pattern == "~^www\\." {
			assert.Assert(t, len(mapping.GetComment()) > 0)
			break
		}
	}
}

func TestEmptyMapBlock(t *testing.T) {
	t.Parallel()

	configString := `
http {
	map $http_host $backend {
	}
}
`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	maps := conf.FindMaps()
	assert.Equal(t, len(maps), 1)

	hostMap := maps[0]
	assert.Equal(t, len(hostMap.Mappings), 0)
	assert.Equal(t, hostMap.GetDefaultValue(), "") // No default mapping
}

func TestInvalidMapDirective(t *testing.T) {
	t.Parallel()

	// Test map directive with insufficient parameters
	configString := `
http {
	map $http_host {
		default backend1;
	}
}
`

	p := parser.NewStringParser(configString)
	_, err := p.Parse()
	// Should return an error due to insufficient parameters
	assert.Error(t, err, "map directive requires at least 2 parameters: source and target variables")
}
