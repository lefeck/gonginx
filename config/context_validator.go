package config

import (
	"fmt"
	"strings"
)

// ContextValidationError represents a context validation error
type ContextValidationError struct {
	Directive string
	Context   string
	Line      int
	Message   string
}

// Error returns the error message
func (cve *ContextValidationError) Error() string {
	return fmt.Sprintf("line %d: directive '%s' is not allowed in '%s' context: %s",
		cve.Line, cve.Directive, cve.Context, cve.Message)
}

// ContextValidator validates nginx directive contexts
type ContextValidator struct {
	// directiveContexts maps directive names to their allowed contexts
	directiveContexts map[string][]string
	// blockTypes maps block directive names to their context types
	blockTypes map[string]string
}

// NewContextValidator creates a new context validator
func NewContextValidator() *ContextValidator {
	cv := &ContextValidator{
		directiveContexts: make(map[string][]string),
		blockTypes:        make(map[string]string),
	}
	cv.initializeContextRules()
	return cv
}

// initializeContextRules initializes the context validation rules
func (cv *ContextValidator) initializeContextRules() {
	// Clear any existing mappings
	cv.directiveContexts = make(map[string][]string)

	// Define directives with their exact allowed contexts
	directiveMappings := map[string][]string{
		// Main context only
		"user":                    {"main"},
		"worker_processes":        {"main"},
		"worker_rlimit_nofile":    {"main"},
		"pid":                     {"main"},
		"daemon":                  {"main"},
		"master_process":          {"main"},
		"timer_resolution":        {"main"},
		"worker_priority":         {"main"},
		"worker_cpu_affinity":     {"main"},
		"worker_shutdown_timeout": {"main"},
		"load_module":             {"main"},
		"pcre_jit":                {"main"},
		"ssl_engine":              {"main"},
		"thread_pool":             {"main"},

		// Events context only
		"worker_connections":  {"events"},
		"use":                 {"events"},
		"multi_accept":        {"events"},
		"accept_mutex":        {"events"},
		"accept_mutex_delay":  {"events"},
		"worker_aio_requests": {"events"},

		// HTTP context only
		"server":                     {"http"},
		"upstream":                   {"http"},
		"map":                        {"http"},
		"geo":                        {"http"},
		"split_clients":              {"http"},
		"limit_req_zone":             {"http"},
		"limit_conn_zone":            {"http"},
		"proxy_cache_path":           {"http"},
		"fastcgi_cache_path":         {"http"},
		"uwsgi_cache_path":           {"http"},
		"scgi_cache_path":            {"http"},
		"types_hash_bucket_size":     {"http"},
		"types_hash_max_size":        {"http"},
		"variables_hash_bucket_size": {"http"},
		"variables_hash_max_size":    {"http"},
		"server_tokens":              {"http"},

		// Server context only
		"listen":                    {"server"},
		"server_name":               {"server"},
		"location":                  {"server"},
		"ssl_certificate":           {"server"},
		"ssl_certificate_key":       {"server"},
		"ssl_protocols":             {"server"},
		"ssl_ciphers":               {"server"},
		"ssl_prefer_server_ciphers": {"server"},
		"ssl_session_cache":         {"server"},
		"ssl_session_timeout":       {"server"},

		// Location and IF context only
		"alias":             {"location"},
		"proxy_pass":        {"location", "if"},
		"fastcgi_pass":      {"location", "if"},
		"uwsgi_pass":        {"location", "if"},
		"scgi_pass":         {"location", "if"},
		"memcached_pass":    {"location", "if"},
		"break":             {"location", "if"},
		"set":               {"location", "if"},
		"internal":          {"location"},
		"proxy_cache":       {"location", "if"},
		"proxy_cache_valid": {"location", "if"},
		"proxy_cache_key":   {"location", "if"},
		"fastcgi_cache":     {"location", "if"},
		"uwsgi_cache":       {"location", "if"},
		"scgi_cache":        {"location", "if"},

		// Multiple contexts (http, server, location, if)
		"root":                 {"http", "server", "location", "if"},
		"index":                {"http", "server", "location"},
		"try_files":            {"server", "location"},
		"return":               {"server", "location", "if"},
		"rewrite":              {"server", "location", "if"},
		"error_page":           {"http", "server", "location", "if"},
		"access_log":           {"http", "server", "location", "if"},
		"error_log":            {"main", "http", "server", "location"},
		"add_header":           {"http", "server", "location", "if"},
		"expires":              {"http", "server", "location", "if"},
		"deny":                 {"http", "server", "location", "limit_except"},
		"allow":                {"http", "server", "location", "limit_except"},
		"auth_basic":           {"http", "server", "location", "limit_except"},
		"auth_basic_user_file": {"http", "server", "location", "limit_except"},
		"autoindex":            {"http", "server", "location"},
		"gzip":                 {"http", "server", "location", "if"},
		"gzip_types":           {"http", "server", "location"},
		"limit_req":            {"http", "server", "location"},
		"limit_conn":           {"http", "server", "location"},

		// HTTP and server context
		"client_max_body_size":  {"http", "server", "location"},
		"client_body_timeout":   {"http", "server", "location"},
		"client_header_timeout": {"http", "server"},
		"keepalive_timeout":     {"http", "server", "location"},
		"send_timeout":          {"http", "server", "location"},
		"sendfile":              {"http", "server", "location", "if"},
		"tcp_nodelay":           {"http", "server", "location"},
		"tcp_nopush":            {"http", "server", "location"},

		// Proxy directives
		"proxy_set_header":        {"http", "server", "location", "if"},
		"proxy_connect_timeout":   {"http", "server", "location"},
		"proxy_read_timeout":      {"http", "server", "location"},
		"proxy_send_timeout":      {"http", "server", "location"},
		"proxy_buffering":         {"http", "server", "location"},
		"proxy_buffer_size":       {"http", "server", "location"},
		"proxy_buffers":           {"http", "server", "location"},
		"proxy_busy_buffers_size": {"http", "server", "location"},

		// Upstream context directives
		"hash":               {"upstream"},
		"ip_hash":            {"upstream"},
		"least_conn":         {"upstream"},
		"least_time":         {"upstream"},
		"random":             {"upstream"},
		"keepalive":          {"upstream"},
		"keepalive_requests": {"upstream"},
		"ntlm":               {"upstream"},
		"sticky":             {"upstream"},
		"zone":               {"upstream"},

		// Can appear in main and other contexts
		"include": {"main", "events", "http", "server", "location", "upstream", "map", "geo", "split_clients"},

		// Special directives
		"if":           {"server", "location"},
		"default_type": {"http", "server", "location"},
	}

	// Apply the mappings
	for directive, contexts := range directiveMappings {
		cv.directiveContexts[directive] = contexts
	}

	// Handle special upstream directive which can appear in both http and upstream contexts
	cv.directiveContexts["server"] = []string{"http", "upstream", "stream_upstream"}

	// Define block types
	cv.blockTypes["http"] = "http"
	cv.blockTypes["server"] = "server"
	cv.blockTypes["location"] = "location"
	cv.blockTypes["upstream"] = "upstream"
	cv.blockTypes["stream"] = "stream"
	cv.blockTypes["events"] = "events"
	cv.blockTypes["map"] = "map"
	cv.blockTypes["geo"] = "geo"
	cv.blockTypes["split_clients"] = "split_clients"
	cv.blockTypes["limit_except"] = "limit_except"
	cv.blockTypes["if"] = "if"

	// Add special handling for stream blocks
	cv.blockTypes["stream_server"] = "stream_server"
	cv.blockTypes["stream_upstream"] = "stream_upstream"
}

// ValidateContext validates that a directive is used in the correct context
func (cv *ContextValidator) ValidateContext(directive IDirective, context string) error {
	directiveName := directive.GetName()

	// Skip validation for custom directives or unknown directives
	allowedContexts, exists := cv.directiveContexts[directiveName]
	if !exists {
		// For unknown directives, we can't validate context
		return nil
	}

	// Check if the current context is allowed
	for _, allowedContext := range allowedContexts {
		if context == allowedContext {
			return nil
		}
	}

	// Generate helpful error message
	allowedStr := strings.Join(allowedContexts, ", ")
	message := fmt.Sprintf("allowed in: %s", allowedStr)

	return &ContextValidationError{
		Directive: directiveName,
		Context:   context,
		Line:      directive.GetLine(),
		Message:   message,
	}
}

// ValidateBlock validates the context of all directives within a block
func (cv *ContextValidator) ValidateBlock(block IBlock, context string) []error {
	var errors []error

	if block == nil {
		return errors
	}

	for _, directive := range block.GetDirectives() {
		// Validate this directive's context
		if err := cv.ValidateContext(directive, context); err != nil {
			errors = append(errors, err)
		}

		// If this directive has a block, validate it recursively
		if directive.GetBlock() != nil {
			directiveName := directive.GetName()

			// Determine the context for the nested block
			var nestedContext string
			switch directiveName {
			case "http":
				nestedContext = "http"
			case "server":
				// Check if we're in stream context
				if context == "stream" {
					nestedContext = "stream_server"
				} else {
					nestedContext = "server"
				}
			case "location":
				nestedContext = "location"
			case "upstream":
				// Check if we're in stream context
				if context == "stream" {
					nestedContext = "stream_upstream"
				} else {
					nestedContext = "upstream"
				}
			case "stream":
				nestedContext = "stream"
			case "events":
				nestedContext = "events"
			case "map":
				nestedContext = "map"
			case "geo":
				nestedContext = "geo"
			case "split_clients":
				nestedContext = "split_clients"
			case "limit_except":
				nestedContext = "limit_except"
			case "if":
				nestedContext = "if"
			default:
				// For unknown block types, use the directive name as context
				if blockType, exists := cv.blockTypes[directiveName]; exists {
					nestedContext = blockType
				} else {
					nestedContext = directiveName
				}
			}

			// Recursively validate the nested block
			nestedErrors := cv.ValidateBlock(directive.GetBlock(), nestedContext)
			errors = append(errors, nestedErrors...)
		}
	}

	return errors
}

// ValidateConfig validates the context of all directives in a configuration
func (cv *ContextValidator) ValidateConfig(config *Config) []error {
	return cv.ValidateBlock(config, "main")
}

// GetAllowedContexts returns the allowed contexts for a directive
func (cv *ContextValidator) GetAllowedContexts(directiveName string) []string {
	if contexts, exists := cv.directiveContexts[directiveName]; exists {
		return contexts
	}
	return []string{}
}

// IsDirectiveAllowedInContext checks if a directive is allowed in a specific context
func (cv *ContextValidator) IsDirectiveAllowedInContext(directiveName, context string) bool {
	allowedContexts := cv.GetAllowedContexts(directiveName)
	for _, allowedContext := range allowedContexts {
		if allowedContext == context {
			return true
		}
	}
	return false
}
