package config

import (
	"fmt"
	"strings"
)

// DependencyValidationError represents a dependency validation error
type DependencyValidationError struct {
	Directive  string
	Dependency string
	Line       int
	Message    string
	Suggestion string
}

// Error returns the error message
func (dve *DependencyValidationError) Error() string {
	msg := fmt.Sprintf("line %d: directive '%s' requires '%s': %s",
		dve.Line, dve.Directive, dve.Dependency, dve.Message)
	if dve.Suggestion != "" {
		msg += fmt.Sprintf(" (suggestion: %s)", dve.Suggestion)
	}
	return msg
}

// DependencyRule represents a dependency rule between directives
type DependencyRule struct {
	Directive    string   // The directive that has dependencies
	Dependencies []string // Required dependencies
	Context      string   // Context where this rule applies (empty for all contexts)
	Message      string   // Custom error message
	Suggestion   string   // Suggestion for fixing the issue
}

// DependencyValidator validates nginx directive dependencies
type DependencyValidator struct {
	rules []DependencyRule
}

// NewDependencyValidator creates a new dependency validator
func NewDependencyValidator() *DependencyValidator {
	dv := &DependencyValidator{}
	dv.initializeDependencyRules()
	return dv
}

// initializeDependencyRules initializes the dependency validation rules
func (dv *DependencyValidator) initializeDependencyRules() {
	dv.rules = []DependencyRule{
		// SSL related dependencies
		{
			Directive:    "ssl_certificate",
			Dependencies: []string{"ssl_certificate_key"},
			Context:      "",
			Message:      "SSL certificate requires a private key",
			Suggestion:   "add ssl_certificate_key directive with the path to the private key file",
		},
		{
			Directive:    "ssl_certificate_key",
			Dependencies: []string{"ssl_certificate"},
			Context:      "",
			Message:      "SSL certificate key requires a certificate",
			Suggestion:   "add ssl_certificate directive with the path to the certificate file",
		},
		{
			Directive:    "ssl_trusted_certificate",
			Dependencies: []string{"ssl_certificate", "ssl_certificate_key"},
			Context:      "",
			Message:      "SSL trusted certificate requires both certificate and key",
			Suggestion:   "add ssl_certificate and ssl_certificate_key directives",
		},

		// Proxy cache dependencies
		{
			Directive:    "proxy_cache",
			Dependencies: []string{"proxy_cache_path"},
			Context:      "",
			Message:      "proxy_cache requires proxy_cache_path to be defined",
			Suggestion:   "add proxy_cache_path directive in http context",
		},
		{
			Directive:    "proxy_cache_valid",
			Dependencies: []string{"proxy_cache"},
			Context:      "",
			Message:      "proxy_cache_valid requires proxy_cache to be enabled",
			Suggestion:   "add proxy_cache directive in the same context",
		},
		{
			Directive:    "proxy_cache_key",
			Dependencies: []string{"proxy_cache"},
			Context:      "",
			Message:      "proxy_cache_key requires proxy_cache to be enabled",
			Suggestion:   "add proxy_cache directive in the same context",
		},

		// FastCGI cache dependencies
		{
			Directive:    "fastcgi_cache",
			Dependencies: []string{"fastcgi_cache_path"},
			Context:      "",
			Message:      "fastcgi_cache requires fastcgi_cache_path to be defined",
			Suggestion:   "add fastcgi_cache_path directive in http context",
		},
		{
			Directive:    "fastcgi_cache_valid",
			Dependencies: []string{"fastcgi_cache"},
			Context:      "",
			Message:      "fastcgi_cache_valid requires fastcgi_cache to be enabled",
			Suggestion:   "add fastcgi_cache directive in the same context",
		},

		// Auth dependencies
		{
			Directive:    "auth_basic",
			Dependencies: []string{"auth_basic_user_file"},
			Context:      "",
			Message:      "auth_basic requires auth_basic_user_file",
			Suggestion:   "add auth_basic_user_file directive with the path to the password file",
		},

		// Limit request dependencies
		{
			Directive:    "limit_req",
			Dependencies: []string{"limit_req_zone"},
			Context:      "",
			Message:      "limit_req requires limit_req_zone to be defined",
			Suggestion:   "add limit_req_zone directive in http context",
		},

		// Limit connection dependencies
		{
			Directive:    "limit_conn",
			Dependencies: []string{"limit_conn_zone"},
			Context:      "",
			Message:      "limit_conn requires limit_conn_zone to be defined",
			Suggestion:   "add limit_conn_zone directive in http context",
		},

		// Gzip dependencies
		{
			Directive:    "gzip_types",
			Dependencies: []string{"gzip"},
			Context:      "",
			Message:      "gzip_types requires gzip to be enabled",
			Suggestion:   "add 'gzip on;' directive",
		},
		{
			Directive:    "gzip_vary",
			Dependencies: []string{"gzip"},
			Context:      "",
			Message:      "gzip_vary requires gzip to be enabled",
			Suggestion:   "add 'gzip on;' directive",
		},

		// Upstream dependencies (server blocks should have listen, upstream should have servers)
		{
			Directive:    "proxy_pass",
			Dependencies: []string{}, // Special case - will check upstream existence
			Context:      "",
			Message:      "proxy_pass with upstream name requires the upstream to be defined",
			Suggestion:   "define the upstream block in http context or use a direct URL",
		},

		// Index dependencies
		{
			Directive:    "autoindex",
			Dependencies: []string{}, // Will check that no index is defined when autoindex is on
			Context:      "",
			Message:      "autoindex on conflicts with index directive",
			Suggestion:   "remove index directive or set autoindex off",
		},

		// Rewrite dependencies
		{
			Directive:    "rewrite_log",
			Dependencies: []string{"rewrite"},
			Context:      "",
			Message:      "rewrite_log requires at least one rewrite directive",
			Suggestion:   "add rewrite directive or remove rewrite_log",
		},
	}
}

// ValidateDependencies validates directive dependencies within a configuration
func (dv *DependencyValidator) ValidateDependencies(config *Config) []error {
	var errors []error

	// First, collect all directives and their contexts
	directiveMap := dv.collectDirectives(config, "main")

	// Validate each rule
	for _, rule := range dv.rules {
		errors = append(errors, dv.validateRule(rule, directiveMap)...)
	}

	// Special validations
	errors = append(errors, dv.validateSpecialCases(config, directiveMap)...)

	return errors
}

// collectDirectives recursively collects all directives and their contexts
func (dv *DependencyValidator) collectDirectives(block IBlock, context string) map[string][]DirectiveInfo {
	directiveMap := make(map[string][]DirectiveInfo)

	if block == nil {
		return directiveMap
	}

	for _, directive := range block.GetDirectives() {
		directiveName := directive.GetName()

		// Add current directive
		if directiveMap[directiveName] == nil {
			directiveMap[directiveName] = []DirectiveInfo{}
		}
		directiveMap[directiveName] = append(directiveMap[directiveName], DirectiveInfo{
			Directive: directive,
			Context:   context,
		})

		// Process nested blocks
		if directive.GetBlock() != nil {
			var nestedContext string
			switch directiveName {
			case "http":
				nestedContext = "http"
			case "server":
				if context == "stream" {
					nestedContext = "stream_server"
				} else {
					nestedContext = "server"
				}
			case "location":
				nestedContext = "location"
			case "upstream":
				if context == "stream" {
					nestedContext = "stream_upstream"
				} else {
					nestedContext = "upstream"
				}
			case "stream":
				nestedContext = "stream"
			case "events":
				nestedContext = "events"
			case "if":
				nestedContext = "if"
			default:
				nestedContext = directiveName
			}

			nestedDirectives := dv.collectDirectives(directive.GetBlock(), nestedContext)
			for name, infos := range nestedDirectives {
				if directiveMap[name] == nil {
					directiveMap[name] = []DirectiveInfo{}
				}
				directiveMap[name] = append(directiveMap[name], infos...)
			}
		}
	}

	return directiveMap
}

// DirectiveInfo holds information about a directive and its context
type DirectiveInfo struct {
	Directive IDirective
	Context   string
}

// validateRule validates a specific dependency rule
func (dv *DependencyValidator) validateRule(rule DependencyRule, directiveMap map[string][]DirectiveInfo) []error {
	var errors []error

	// Find all instances of the rule directive
	directiveInfos, exists := directiveMap[rule.Directive]
	if !exists {
		return errors // No instances of this directive
	}

	for _, info := range directiveInfos {
		// Check if rule applies to this context
		if rule.Context != "" && info.Context != rule.Context {
			continue
		}

		// Check if all dependencies are satisfied
		for _, dependency := range rule.Dependencies {
			if !dv.isDependencySatisfied(dependency, info, directiveMap) {
				errors = append(errors, &DependencyValidationError{
					Directive:  rule.Directive,
					Dependency: dependency,
					Line:       info.Directive.GetLine(),
					Message:    rule.Message,
					Suggestion: rule.Suggestion,
				})
			}
		}
	}

	return errors
}

// isDependencySatisfied checks if a dependency is satisfied
func (dv *DependencyValidator) isDependencySatisfied(dependency string, directiveInfo DirectiveInfo, directiveMap map[string][]DirectiveInfo) bool {
	dependencyInfos, exists := directiveMap[dependency]
	if !exists {
		return false
	}

	// For most dependencies, we need them in the same context or a parent context
	for _, depInfo := range dependencyInfos {
		if dv.isContextCompatible(directiveInfo.Context, depInfo.Context) {
			return true
		}
	}

	return false
}

// isContextCompatible checks if two contexts are compatible for dependency checking
func (dv *DependencyValidator) isContextCompatible(directiveContext, dependencyContext string) bool {
	// Same context is always compatible
	if directiveContext == dependencyContext {
		return true
	}

	// Check parent-child relationships
	contextHierarchy := map[string][]string{
		"location":        {"server", "http", "main"},
		"server":          {"http", "main"},
		"http":            {"main"},
		"upstream":        {"http", "main"},
		"stream_server":   {"stream", "main"},
		"stream_upstream": {"stream", "main"},
		"stream":          {"main"},
		"events":          {"main"},
		"if":              {"location", "server", "http", "main"},
	}

	// Check if dependency context is a parent of directive context
	if parents, exists := contextHierarchy[directiveContext]; exists {
		for _, parent := range parents {
			if parent == dependencyContext {
				return true
			}
		}
	}

	return false
}

// validateSpecialCases handles special validation cases
func (dv *DependencyValidator) validateSpecialCases(config *Config, directiveMap map[string][]DirectiveInfo) []error {
	var errors []error

	// Validate upstream references in proxy_pass
	errors = append(errors, dv.validateUpstreamReferences(directiveMap)...)

	// Validate server blocks have listen directives
	errors = append(errors, dv.validateServerListenDirectives(directiveMap)...)

	// Validate upstream blocks have server directives
	errors = append(errors, dv.validateUpstreamServerDirectives(config)...)

	// Validate autoindex conflicts with index
	errors = append(errors, dv.validateAutoindexIndexConflict(directiveMap)...)

	return errors
}

// validateUpstreamReferences validates that proxy_pass references to upstreams exist
func (dv *DependencyValidator) validateUpstreamReferences(directiveMap map[string][]DirectiveInfo) []error {
	var errors []error

	proxyPassInfos, exists := directiveMap["proxy_pass"]
	if !exists {
		return errors
	}

	// Collect upstream names
	upstreamNames := make(map[string]bool)
	if upstreamInfos, exists := directiveMap["upstream"]; exists {
		for _, upstreamInfo := range upstreamInfos {
			params := upstreamInfo.Directive.GetParameters()
			if len(params) > 0 {
				upstreamNames[params[0].Value] = true
			}
		}
	}

	// Check proxy_pass directives
	for _, proxyPassInfo := range proxyPassInfos {
		params := proxyPassInfo.Directive.GetParameters()
		if len(params) > 0 {
			target := params[0].Value
			// If it looks like an upstream name (not a URL), check if it exists
			if !strings.Contains(target, "://") && !strings.HasPrefix(target, "unix:") {
				// Remove potential path part
				upstreamName := strings.Split(target, "/")[0]
				if !upstreamNames[upstreamName] {
					errors = append(errors, &DependencyValidationError{
						Directive:  "proxy_pass",
						Dependency: fmt.Sprintf("upstream %s", upstreamName),
						Line:       proxyPassInfo.Directive.GetLine(),
						Message:    fmt.Sprintf("upstream '%s' is not defined", upstreamName),
						Suggestion: fmt.Sprintf("define upstream %s in http context or use a direct URL", upstreamName),
					})
				}
			}
		}
	}

	return errors
}

// validateServerListenDirectives validates that server blocks have listen directives
func (dv *DependencyValidator) validateServerListenDirectives(directiveMap map[string][]DirectiveInfo) []error {
	var errors []error

	serverInfos, exists := directiveMap["server"]
	if !exists {
		return errors
	}

	for _, serverInfo := range serverInfos {
		// Check if this server block has a listen directive
		hasListen := false
		if serverInfo.Directive.GetBlock() != nil {
			for _, directive := range serverInfo.Directive.GetBlock().GetDirectives() {
				if directive.GetName() == "listen" {
					hasListen = true
					break
				}
			}
		}

		if !hasListen {
			errors = append(errors, &DependencyValidationError{
				Directive:  "server",
				Dependency: "listen",
				Line:       serverInfo.Directive.GetLine(),
				Message:    "server block should have at least one listen directive",
				Suggestion: "add listen directive (e.g., 'listen 80;' or 'listen 443 ssl;')",
			})
		}
	}

	return errors
}

// validateUpstreamServerDirectives validates that upstream blocks have server directives
func (dv *DependencyValidator) validateUpstreamServerDirectives(config *Config) []error {
	var errors []error

	// Find all upstream blocks
	upstreams := config.FindDirectives("upstream")
	for _, upstream := range upstreams {
		if upstream.GetBlock() != nil {
			hasServer := false
			for _, directive := range upstream.GetBlock().GetDirectives() {
				if directive.GetName() == "server" {
					hasServer = true
					break
				}
			}

			if !hasServer {
				params := upstream.GetParameters()
				upstreamName := "unnamed"
				if len(params) > 0 {
					upstreamName = params[0].Value
				}

				errors = append(errors, &DependencyValidationError{
					Directive:  "upstream",
					Dependency: "server",
					Line:       upstream.GetLine(),
					Message:    fmt.Sprintf("upstream '%s' has no server directives", upstreamName),
					Suggestion: "add at least one server directive (e.g., 'server backend1.example.com;')",
				})
			}
		}
	}

	return errors
}

// validateAutoindexIndexConflict validates autoindex and index directive conflicts
func (dv *DependencyValidator) validateAutoindexIndexConflict(directiveMap map[string][]DirectiveInfo) []error {
	var errors []error

	autoindexInfos, hasAutoindex := directiveMap["autoindex"]
	indexInfos, hasIndex := directiveMap["index"]

	if !hasAutoindex || !hasIndex {
		return errors
	}

	// Check for conflicts in the same context
	for _, autoindexInfo := range autoindexInfos {
		params := autoindexInfo.Directive.GetParameters()
		if len(params) > 0 && strings.ToLower(params[0].Value) == "on" {
			// Autoindex is on, check for index in same context
			for _, indexInfo := range indexInfos {
				if indexInfo.Context == autoindexInfo.Context {
					errors = append(errors, &DependencyValidationError{
						Directive:  "autoindex",
						Dependency: "index",
						Line:       autoindexInfo.Directive.GetLine(),
						Message:    "autoindex on conflicts with index directive in the same context",
						Suggestion: "remove index directive or set autoindex off",
					})
					break
				}
			}
		}
	}

	return errors
}
