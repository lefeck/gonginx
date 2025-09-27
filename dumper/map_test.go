package dumper_test

import (
	"strings"
	"testing"

	"github.com/lefeck/gonginx/dumper"
	"github.com/lefeck/gonginx/parser"
	"gotest.tools/v3/assert"
)

func TestMapDumping(t *testing.T) {
	t.Parallel()

	configString := `http {
	map $http_host $backend {
		default backend1;
		example.com backend2;
		api.example.com backend3;
		~^www\. backend4;
	}
}`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	// Test dumping with indented style
	result := dumper.DumpConfig(conf, dumper.IndentedStyle)

	// Verify the map block is properly formatted
	assert.Assert(t, strings.Contains(result, "map $http_host $backend"))
	assert.Assert(t, strings.Contains(result, "default backend1;"))
	assert.Assert(t, strings.Contains(result, "example.com backend2;"))
	assert.Assert(t, strings.Contains(result, "api.example.com backend3;"))
	assert.Assert(t, strings.Contains(result, "~^www\\. backend4;"))

	// Verify proper block structure
	assert.Assert(t, strings.Contains(result, "map $http_host $backend {"))
	assert.Assert(t, strings.Count(result, "{") == strings.Count(result, "}"))
}

func TestMapDumpingWithComments(t *testing.T) {
	t.Parallel()

	configString := `http {
	# Backend selection map
	map $http_host $backend { # inline comment
		default backend1; # default backend
		example.com backend2;
		# Pattern for www subdomains  
		~^www\. backend4;
	}
}`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	// Test dumping with indented style
	result := dumper.DumpConfig(conf, dumper.IndentedStyle)

	// Verify comments are preserved
	assert.Assert(t, strings.Contains(result, "# Backend selection map"))
	assert.Assert(t, strings.Contains(result, "# inline comment"))
	assert.Assert(t, strings.Contains(result, "# default backend"))
	assert.Assert(t, strings.Contains(result, "# Pattern for www subdomains"))
}

func TestMapDumpingNoIndent(t *testing.T) {
	t.Parallel()

	configString := `http {
	map $http_host $backend {
		default backend1;
		example.com backend2;
	}
}`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	// Test dumping with no indent style
	result := dumper.DumpConfig(conf, dumper.NoIndentStyle)

	// Verify the map block is properly formatted without indentation
	assert.Assert(t, strings.Contains(result, "map $http_host $backend"))
	assert.Assert(t, strings.Contains(result, "default backend1;"))
	assert.Assert(t, strings.Contains(result, "example.com backend2;"))
}

func TestMultipleMapsDumping(t *testing.T) {
	t.Parallel()

	configString := `http {
	map $http_host $backend {
		default backend1;
		example.com backend2;
	}
	
	map $request_method $limit {
		default "";
		POST $binary_remote_addr;
	}
}`

	p := parser.NewStringParser(configString)
	conf, err := p.Parse()
	assert.NilError(t, err)

	// Test dumping with indented style
	result := dumper.DumpConfig(conf, dumper.IndentedStyle)

	// Verify both maps are present
	assert.Assert(t, strings.Contains(result, "map $http_host $backend"))
	assert.Assert(t, strings.Contains(result, "map $request_method $limit"))
	assert.Assert(t, strings.Contains(result, "POST $binary_remote_addr;"))
}
