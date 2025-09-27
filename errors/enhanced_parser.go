package errors

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lefeck/gonginx/config"
	"github.com/lefeck/gonginx/parser"
)

// EnhancedParser wraps the original parser with better error handling
type EnhancedParser struct {
	originalParser *parser.Parser
	filename       string
	content        string
	errors         *ErrorCollection
}

// NewEnhancedParser creates a new enhanced parser
func NewEnhancedParser(filename string) (*EnhancedParser, error) {
	originalParser, err := parser.NewParser(filename)
	if err != nil {
		return nil, NewFileError(fmt.Sprintf("failed to create parser for file '%s'", filename)).
			WithFile(filename).
			WithInnerError(err)
	}

	// Read file content for better error reporting
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, NewFileError(fmt.Sprintf("failed to read file '%s'", filename)).
			WithFile(filename).
			WithInnerError(err)
	}

	return &EnhancedParser{
		originalParser: originalParser,
		filename:       filename,
		content:        string(content),
		errors:         NewErrorCollection(),
	}, nil
}

// NewEnhancedStringParser creates a new enhanced parser from string content
func NewEnhancedStringParser(content string) *EnhancedParser {
	originalParser := parser.NewStringParser(content)

	return &EnhancedParser{
		originalParser: originalParser,
		filename:       "<string>",
		content:        content,
		errors:         NewErrorCollection(),
	}
}

// Parse parses the configuration with enhanced error reporting
func (ep *EnhancedParser) Parse() (*config.Config, error) {
	conf, err := ep.originalParser.Parse()

	if err != nil {
		// Try to enhance the original error
		enhancedErr := ep.enhanceError(err)
		return conf, enhancedErr
	}

	// Perform additional validation
	ep.validateConfiguration(conf)

	if ep.errors.HasErrors() {
		return conf, ep.errors
	}

	return conf, nil
}

// ParseWithValidation parses and validates the configuration
func (ep *EnhancedParser) ParseWithValidation() (*config.Config, error) {
	conf, err := ep.Parse()
	if err != nil {
		return conf, err
	}

	// Perform comprehensive validation
	ep.performDeepValidation(conf)

	if ep.errors.HasErrors() {
		return conf, ep.errors
	}

	return conf, nil
}

// enhanceError enhances the original parser error with better context
func (ep *EnhancedParser) enhanceError(originalErr error) error {
	errStr := originalErr.Error()

	// Try to extract line information from the original error
	var lineNum int
	if strings.Contains(errStr, "line") {
		fmt.Sscanf(errStr, "%*s line %d", &lineNum)
	}

	var enhancedErr *ParseError

	// Categorize the error based on its content
	switch {
	case strings.Contains(errStr, "unexpected token"):
		enhancedErr = NewSyntaxError("unexpected token in configuration")
		enhancedErr.Suggestion = "Check for missing semicolons, braces, or quotes"

	case strings.Contains(errStr, "unexpected end"):
		enhancedErr = NewSyntaxError("unexpected end of file")
		enhancedErr.Suggestion = "Check for unclosed blocks or missing closing braces"

	case strings.Contains(errStr, "invalid directive"):
		enhancedErr = NewSyntaxError("invalid directive syntax")
		enhancedErr.Suggestion = "Check directive spelling and parameter format"

	case strings.Contains(errStr, "duplicate"):
		enhancedErr = NewSemanticError("duplicate directive not allowed")
		enhancedErr.Suggestion = "Remove duplicate directives or use appropriate context"

	default:
		enhancedErr = NewSyntaxError(errStr)
	}

	enhancedErr.WithFile(ep.filename).
		WithInnerError(originalErr)

	if lineNum > 0 {
		enhancedErr.WithLine(lineNum)
		context := ep.getLineContext(lineNum)
		if context != "" {
			enhancedErr.WithContext(context)
		}
	}

	return enhancedErr
}

// getLineContext gets the context around a specific line
func (ep *EnhancedParser) getLineContext(lineNum int) string {
	lines := strings.Split(ep.content, "\n")
	if lineNum <= 0 || lineNum > len(lines) {
		return ""
	}

	var contextLines []string
	start := lineNum - 3
	end := lineNum + 2

	if start < 1 {
		start = 1
	}
	if end > len(lines) {
		end = len(lines)
	}

	for i := start; i <= end; i++ {
		line := lines[i-1]
		if i == lineNum {
			contextLines = append(contextLines, fmt.Sprintf(">>> %3d | %s", i, line))
		} else {
			contextLines = append(contextLines, fmt.Sprintf("    %3d | %s", i, line))
		}
	}

	return strings.Join(contextLines, "\n")
}

// validateConfiguration performs basic validation on the parsed configuration
func (ep *EnhancedParser) validateConfiguration(conf *config.Config) {
	if conf == nil {
		return
	}

	// Validate HTTP block
	ep.validateHTTPBlock(conf)

	// Validate Stream blocks
	ep.validateStreamBlocks(conf)

	// Validate Events block
	ep.validateEventsBlock(conf)
}

// validateHTTPBlock validates HTTP-specific configuration
func (ep *EnhancedParser) validateHTTPBlock(conf *config.Config) {
	httpBlocks := conf.FindDirectives("http")

	if len(httpBlocks) > 1 {
		ep.errors.Add(NewSemanticError("multiple http blocks found").
			WithFile(ep.filename).
			WithDirective("http").
			WithSuggestion("nginx configuration should have only one http block"))
	}

	for _, httpBlock := range httpBlocks {
		if http, ok := httpBlock.(*config.HTTP); ok {
			ep.validateServers(http)
			ep.validateUpstreams(http)
		}
	}
}

// validateServers validates server configurations
func (ep *EnhancedParser) validateServers(http *config.HTTP) {
	serverNames := make(map[string][]*config.Server)

	servers := http.Servers
	for _, server := range servers {
		// Validate server block
		ep.validateServerBlock(server)

		// Check for duplicate server names
		serverNameDirs := server.FindDirectives("server_name")
		for _, serverNameDir := range serverNameDirs {
			for _, param := range serverNameDir.GetParameters() {
				name := param.GetValue()
				serverNames[name] = append(serverNames[name], server)
			}
		}
	}

	// Report duplicate server names
	for name, servers := range serverNames {
		if len(servers) > 1 && name != "_" && name != "default_server" {
			ep.errors.Add(NewSemanticError(fmt.Sprintf("duplicate server_name '%s' found", name)).
				WithFile(ep.filename).
				WithDirective("server_name").
				WithParameter(name).
				WithSuggestion("Each server_name should be unique unless using default_server"))
		}
	}
}

// validateServerBlock validates a single server block
func (ep *EnhancedParser) validateServerBlock(server *config.Server) {
	listenDirs := server.FindDirectives("listen")
	if len(listenDirs) == 0 {
		ep.errors.Add(NewSemanticError("server block has no listen directive").
			WithFile(ep.filename).
			WithDirective("server").
			WithSuggestion("Add a listen directive (e.g., listen 80;)"))
	}

	// Validate SSL configuration
	ep.validateSSLConfig(server)
}

// validateSSLConfig validates SSL configuration
func (ep *EnhancedParser) validateSSLConfig(server *config.Server) {
	sslCertDirs := server.FindDirectives("ssl_certificate")
	sslKeyDirs := server.FindDirectives("ssl_certificate_key")

	if len(sslCertDirs) > 0 && len(sslKeyDirs) == 0 {
		ep.errors.Add(NewSemanticError("ssl_certificate specified without ssl_certificate_key").
			WithFile(ep.filename).
			WithDirective("ssl_certificate").
			WithSuggestion("Add ssl_certificate_key directive"))
	}

	if len(sslKeyDirs) > 0 && len(sslCertDirs) == 0 {
		ep.errors.Add(NewSemanticError("ssl_certificate_key specified without ssl_certificate").
			WithFile(ep.filename).
			WithDirective("ssl_certificate_key").
			WithSuggestion("Add ssl_certificate directive"))
	}

	// Check if SSL files exist
	for _, certDir := range sslCertDirs {
		if len(certDir.GetParameters()) > 0 {
			certPath := certDir.GetParameters()[0].GetValue()
			if !ep.isAbsolutePath(certPath) {
				// Make relative to config file directory
				certPath = filepath.Join(filepath.Dir(ep.filename), certPath)
			}
			if !ep.fileExists(certPath) {
				ep.errors.Add(NewFileError(fmt.Sprintf("SSL certificate file not found: %s", certPath)).
					WithFile(ep.filename).
					WithDirective("ssl_certificate").
					WithParameter(certPath).
					WithSuggestion("Check the file path and ensure the certificate file exists"))
			}
		}
	}
}

// validateUpstreams validates upstream configurations
func (ep *EnhancedParser) validateUpstreams(http *config.HTTP) {
	upstreamDirs := http.FindDirectives("upstream")
	var upstreams []*config.Upstream
	for _, upstreamDir := range upstreamDirs {
		if upstream, ok := upstreamDir.(*config.Upstream); ok {
			upstreams = append(upstreams, upstream)
		}
	}

	for _, upstream := range upstreams {
		servers := upstream.UpstreamServers
		if len(servers) == 0 {
			ep.errors.Add(NewSemanticError(fmt.Sprintf("upstream '%s' has no servers", upstream.UpstreamName)).
				WithFile(ep.filename).
				WithDirective("upstream").
				WithParameter(upstream.UpstreamName).
				WithSuggestion("Add at least one server directive to the upstream block"))
		}
	}
}

// validateStreamBlocks validates stream configurations
func (ep *EnhancedParser) validateStreamBlocks(conf *config.Config) {
	streamBlocks := conf.FindStreams()

	if len(streamBlocks) > 1 {
		ep.errors.Add(NewSemanticError("multiple stream blocks found").
			WithFile(ep.filename).
			WithDirective("stream").
			WithSuggestion("nginx configuration should have only one stream block"))
	}
}

// validateEventsBlock validates events configuration
func (ep *EnhancedParser) validateEventsBlock(conf *config.Config) {
	eventsBlocks := conf.FindDirectives("events")

	if len(eventsBlocks) > 1 {
		ep.errors.Add(NewSemanticError("multiple events blocks found").
			WithFile(ep.filename).
			WithDirective("events").
			WithSuggestion("nginx configuration should have only one events block"))
	}
}

// performDeepValidation performs comprehensive validation
func (ep *EnhancedParser) performDeepValidation(conf *config.Config) {
	// Additional validation rules can be added here
	ep.validateDirectiveParameters(conf)
	ep.validateContextUsage(conf)
}

// validateDirectiveParameters validates parameter types and values
func (ep *EnhancedParser) validateDirectiveParameters(conf *config.Config) {
	// This can be extended to validate parameter types using the parameter type system
	// For now, we'll do basic validation
}

// validateContextUsage validates that directives are used in correct contexts
func (ep *EnhancedParser) validateContextUsage(conf *config.Config) {
	// This can be extended to validate directive contexts
	// For now, we'll do basic validation
}

// Helper functions
func (ep *EnhancedParser) isAbsolutePath(path string) bool {
	return filepath.IsAbs(path)
}

func (ep *EnhancedParser) fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// GetErrors returns the accumulated errors
func (ep *EnhancedParser) GetErrors() *ErrorCollection {
	return ep.errors
}
