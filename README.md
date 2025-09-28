<h1 align="center">Gonginx</h1>

**A powerful and comprehensive nginx configuration management library for Go**

Gonginx is a production-ready nginx configuration parser and management library that provides parsing, validation, modification, generation, and optimization capabilities for nginx configurations. It's designed to make nginx configuration management programmatic, safe, and efficient.

## ✨ Key Features

-  **Complete Configuration Management** - Parse, validate, modify, and generate nginx configurations
- ️ **Advanced Validation** - Context-aware validation with detailed error reporting and fix suggestions
-  **High Performance** - Optimized parsing with benchmark-tested performance
-  **Type Safety** - Smart parameter type detection and validation
-  **Advanced Search** - Powerful configuration search and query capabilities
-  **Rich Templates** - Pre-built templates for common nginx configurations
- ️ **Utility Tools** - Security checking, optimization suggestions, and format conversion
-  **Stream Support** - Complete support for nginx stream module (TCP/UDP load balancing)
-  **Comprehensive Testing** - Full test coverage with integration and performance tests 

## Quick Start

### Installation

```bash
go get github.com/lefeck/gonginx
```

### Basic Usage

```go
package main

import (
    "fmt"
    "log"

    "github.com/lefeck/gonginx/config"
    "github.com/lefeck/gonginx/dumper"
    "github.com/lefeck/gonginx/parser"
)

func main() {
    // Parse from file
    p, err := parser.NewParser("nginx.conf")
    if err != nil {
        log.Fatal(err)
    }

    conf, err := p.Parse()
    if err != nil {
        log.Fatal(err)
    }

    // Find all servers
    servers := conf.FindDirectives("server")
    fmt.Printf("Found %d servers\n", len(servers))

    // Add a new server
    newServer := &config.Directive{
        Name: "server",
        Block: &config.Block{
            Directives: []config.IDirective{
                &config.Directive{
                    Name:       "listen",
                    Parameters: []config.Parameter{{Value: "80"}},
                },
                &config.Directive{
                    Name:       "server_name",
                    Parameters: []config.Parameter{{Value: "example.com"}},
                },
            },
        },
    }

    // Add to http block
    http := conf.FindDirectives("http")[0]
    http.GetBlock().Directives = append(http.GetBlock().Directives, newServer)

    // Output the configuration
    fmt.Println(dumper.DumpConfig(conf, dumper.IndentedStyle))
}
```

## Architecture

### Core Components

####  [Parser](/parser/parser.go)
Advanced nginx configuration parser with context-aware processing:
- **Lexical Analysis**: Tokenizes nginx configuration syntax
- **Syntax Parsing**: Builds Abstract Syntax Tree (AST)
- **Context Awareness**: Handles different directive contexts (http vs stream)
- **Include Support**: Recursive processing of include files
- **Error Recovery**: Detailed error reporting with suggestions

#### [Config](/config/config.go)
Comprehensive configuration object model:
- **Type System**: 10+ parameter types with automatic detection
- **Validation**: Multi-level configuration validation
- **Search API**: Advanced querying and filtering capabilities
- **Manipulation**: Safe configuration modification methods
- **Relationships**: Parent-child directive linking

####  [Dumper](/dumper/dumper.go)
Flexible configuration output formatting:
- **Multiple Styles**: Indented, compact, sorted output
- **Custom Formatting**: Configurable spacing and organization
- **Comment Preservation**: Maintains comments during round-trip
- **File Writing**: Direct file output with proper permissions

####  [Generator](/generator/)
Template-based configuration generation:
- **Builder Pattern**: Fluent API for configuration construction
- **Pre-built Templates**: Common configuration patterns
- **Type Safety**: Compile-time validation of configurations
- **Extensible**: Custom template and directive support

####  [Utils](/utils/)
Powerful utility functions:
- **Security Analysis**: Automated security best practice checking
- **Performance Optimization**: Configuration optimization suggestions
- **Format Conversion**: JSON/YAML export capabilities
- **Diff Analysis**: Configuration change tracking 

## Advanced Features

### Configuration Validation

```go
// Comprehensive validation with detailed reporting
validator := config.NewConfigValidator()
report := validator.ValidateConfig(conf)

if report.HasErrors() {
    for _, issue := range report.GetByLevel(config.ValidationError) {
        fmt.Printf("Error: %s\n", issue.Message)
        if issue.Fix != "" {
            fmt.Printf("Fix: %s\n", issue.Fix)
        }
    }
}
```

### Advanced Search Operations

```go
// Find servers by name
servers := conf.FindServersByName("example.com")

// Find upstream by name
upstream := conf.FindUpstreamByName("backend")

// Find locations by pattern
locations := conf.FindLocationsByPattern("/api/")

// Get all SSL certificates
certificates := conf.GetAllSSLCertificates()

// Get all upstream servers
upstreamServers := conf.GetAllUpstreamServers()
```

### Template-Based Generation

```go
// Generate reverse proxy configuration
template := generator.ReverseProxyTemplate{
    ServerName:    "api.example.com",
    BackendServer: "http://192.168.1.100:8080",
    SSLCert:       "/etc/ssl/certs/api.crt",
    SSLKey:        "/etc/ssl/private/api.key",
    RateLimit:     "10r/s",
}

conf, err := template.Generate()
```

### Builder Pattern API

```go
// Fluent configuration building
config := generator.NewConfigBuilder().
    WorkerProcesses("auto").
    HTTP().
        Upstream("backend").
            Server("127.0.0.1:8001", "weight=3").
            Server("127.0.0.1:8002", "weight=2").
            End().
        Server().
            Listen("443", "ssl").
            ServerName("example.com").
            SSL().
                Certificate("/path/to/cert.pem").
                CertificateKey("/path/to/key.pem").
                End().
            Location("/").
                ProxyPass("http://backend").
                End().
            End().
        End().
    Build()
```

### Utility Functions

```go
// Security analysis
securityReport := utils.CheckSecurity(conf)
fmt.Printf("Security score: %d/100\n", securityReport.Summary.Score)

// Configuration optimization
optimizationReport := utils.OptimizeConfig(conf)
for _, suggestion := range optimizationReport.Suggestions {
    fmt.Printf("Optimization: %s\n", suggestion.Title)
}

// Format conversion
jsonConfig, err := utils.ConvertToJSON(conf)
yamlConfig, err := utils.ConvertToYAML(conf)

// Configuration diff
diffReport := utils.CompareConfigs(oldConf, newConf)
```

## Examples

### Basic Examples
- **[Basic Parsing](examples/parse-nginx-conf-get-listen-port/)** - Parse configuration and extract listen ports
- **[Configuration Updates](examples/update-directive/)** - Modify directives and regenerate
- **[Server Management](examples/adding-server/)** - Add and manage server blocks
- **[Custom Directives](examples/add-custom-directive/)** - Handle custom nginx directives

### Advanced Examples
- **[Configuration Validation](examples/config-validation/)** - Comprehensive validation examples
- **[Advanced Search](examples/advanced-search/)** - Complex configuration queries
- **[Error Handling](examples/error-handling/)** - Best practices for error management
- **[Configuration Generation](examples/config-generator/)** - Template-based generation
- **[Parameter Types](examples/parameter-types/)** - Parameter type system usage
- **[Utility Functions](examples/utils-demo/)** - Security, optimization, and conversion tools

### Specialized Blocks
- **[Stream Configuration](examples/stream-blocks/)** - TCP/UDP load balancing
- **[Map Blocks](examples/map-blocks/)** - Variable mapping configurations
- **[Geo Blocks](examples/geo-blocks/)** - Geographic-based configurations
- **[Rate Limiting](examples/limit-req-zone/)** - Request rate limiting setup
- **[Connection Limiting](examples/limit-conn-zone/)** - Connection limiting configuration
- **[Cache Configuration](examples/proxy-cache-path/)** - Proxy caching setup
- **[Split Testing](examples/split-clients-blocks/)** - A/B testing configurations

## Performance

Gonginx is designed for high performance with comprehensive benchmarking:

- **Small configs** (~1KB): ~6μs parsing time
- **Medium configs** (~50KB): ~110μs parsing time  
- **Large configs** (~1MB): ~400μs parsing time
- **Memory efficient**: ~10KB per config, 146 allocations average

Run benchmarks:
```bash
go test -bench=. ./benchmarks/
```

## Testing

### Run Tests
```bash
# All tests
go test ./...

# Specific packages
go test ./config/
go test ./parser/ 
go test ./dumper/

# Integration tests
go test ./integration_tests/

# Benchmarks
go test -bench=. ./benchmarks/
```

### Test Coverage
- **Unit Tests**: Complete coverage of core functionality
- **Integration Tests**: End-to-end workflow testing
- **Performance Tests**: Benchmark testing for all operations
- **Example Tests**: Validation of all documentation examples

##  Documentation

- **[Complete Guide](GUIDE.md)** - Comprehensive usage guide with examples
- **[API Reference](API_REFERENCE.md)** - Full API documentation
- **[Implementation Summary](IMPLEMENTATION_SUMMARY.md)** - Technical implementation details
- **[Test Summary](test/TEST_SUMMARY.md)** - Testing strategy and results

## Contributing

We welcome contributions! Please see our [Contributing Guide](integration_tests/CONTRIBUTING.md) for details.

### Development Setup
```bash
git clone https://github.com/lefeck/gonginx.git
cd gonginx
go mod download
make test
```

## Production Ready

Gonginx is production-ready with:

- ✅ **Comprehensive validation** - Multi-level configuration checking
- ✅ **Error recovery** - Detailed error reporting with fix suggestions  
- ✅ **Performance tested** - Benchmarked for various configuration sizes
- ✅ **Memory efficient** - Optimized for low memory usage
- ✅ **Thread safe** - Safe for concurrent use
- ✅ **Backward compatible** - Maintains API compatibility
- ✅ **Well documented** - Complete documentation and examples
- ✅ **Fully tested** - Comprehensive test suite

## License

[MIT License](LICENSE) - see the license file for details.
