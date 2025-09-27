package config

import (
	"fmt"
)

// ValidationLevel represents the severity level of a validation issue
type ValidationLevel int

const (
	// ValidationInfo represents informational validation notices
	ValidationInfo ValidationLevel = iota
	// ValidationWarning represents validation warnings that should be addressed
	ValidationWarning
	// ValidationError represents validation errors that must be fixed
	ValidationError
)

// String returns the string representation of the validation level
func (vl ValidationLevel) String() string {
	switch vl {
	case ValidationInfo:
		return "INFO"
	case ValidationWarning:
		return "WARNING"
	case ValidationError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ValidationIssue represents a configuration validation issue
type ValidationIssue struct {
	Level       ValidationLevel
	Category    string
	Title       string
	Description string
	Line        int
	Directive   string
	Context     string
	Fix         string
}

// String returns a human-readable representation of the validation issue
func (vi *ValidationIssue) String() string {
	return fmt.Sprintf("[%s] Line %d: %s - %s (Context: %s)",
		vi.Level.String(), vi.Line, vi.Title, vi.Description, vi.Context)
}

// ValidationReport contains all validation issues found in a configuration
type ValidationReport struct {
	Issues  []ValidationIssue
	Summary ValidationSummary
	Config  *Config
}

// ValidationSummary provides summary statistics about validation issues
type ValidationSummary struct {
	Total   int
	Errors  int
	Warning int
	Info    int
	Passed  []string
}

// String returns a summary of the validation report
func (vs *ValidationSummary) String() string {
	return fmt.Sprintf("Validation Summary: %d issues (Errors: %d, Warnings: %d, Info: %d)",
		vs.Total, vs.Errors, vs.Warning, vs.Info)
}

// HasErrors returns true if there are validation errors
func (vr *ValidationReport) HasErrors() bool {
	return vr.Summary.Errors > 0
}

// GetByLevel returns issues of a specific validation level
func (vr *ValidationReport) GetByLevel(level ValidationLevel) []ValidationIssue {
	var filtered []ValidationIssue
	for _, issue := range vr.Issues {
		if issue.Level == level {
			filtered = append(filtered, issue)
		}
	}
	return filtered
}

// GetByCategory returns issues of a specific category
func (vr *ValidationReport) GetByCategory(category string) []ValidationIssue {
	var filtered []ValidationIssue
	for _, issue := range vr.Issues {
		if issue.Category == category {
			filtered = append(filtered, issue)
		}
	}
	return filtered
}

// ConfigValidator performs comprehensive validation of nginx configurations
type ConfigValidator struct {
	contextValidator    *ContextValidator
	dependencyValidator *DependencyValidator
	enableAllChecks     bool
}

// NewConfigValidator creates a new configuration validator
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{
		contextValidator:    NewContextValidator(),
		dependencyValidator: NewDependencyValidator(),
		enableAllChecks:     true,
	}
}

// NewConfigValidatorWithOptions creates a new configuration validator with options
func NewConfigValidatorWithOptions(enableAllChecks bool) *ConfigValidator {
	return &ConfigValidator{
		contextValidator:    NewContextValidator(),
		dependencyValidator: NewDependencyValidator(),
		enableAllChecks:     enableAllChecks,
	}
}

// ValidateConfig performs comprehensive validation of a configuration
func (cv *ConfigValidator) ValidateConfig(config *Config) *ValidationReport {
	report := &ValidationReport{
		Issues: []ValidationIssue{},
		Config: config,
	}

	// Context validation
	if cv.enableAllChecks {
		contextErrors := cv.contextValidator.ValidateConfig(config)
		for _, err := range contextErrors {
			if contextErr, ok := err.(*ContextValidationError); ok {
				report.Issues = append(report.Issues, ValidationIssue{
					Level:       ValidationError,
					Category:    "Context",
					Title:       "Invalid directive context",
					Description: contextErr.Message,
					Line:        contextErr.Line,
					Directive:   contextErr.Directive,
					Context:     contextErr.Context,
					Fix:         "Move directive to allowed context or remove it",
				})
			}
		}
	}

	// Dependency validation
	if cv.enableAllChecks {
		dependencyErrors := cv.dependencyValidator.ValidateDependencies(config)
		for _, err := range dependencyErrors {
			if depErr, ok := err.(*DependencyValidationError); ok {
				level := ValidationError
				if depErr.Directive == "server" && depErr.Dependency == "listen" {
					level = ValidationWarning // Server without listen is a warning, not an error
				}

				report.Issues = append(report.Issues, ValidationIssue{
					Level:       level,
					Category:    "Dependency",
					Title:       "Missing dependency",
					Description: depErr.Message,
					Line:        depErr.Line,
					Directive:   depErr.Directive,
					Context:     "", // Will be filled by context analysis
					Fix:         depErr.Suggestion,
				})
			}
		}
	}

	// Structural validation
	structuralIssues := cv.validateStructure(config)
	report.Issues = append(report.Issues, structuralIssues...)

	// Parameter validation
	parameterIssues := cv.validateParameters(config)
	report.Issues = append(report.Issues, parameterIssues...)

	// Generate summary
	report.Summary = cv.generateSummary(report.Issues)

	return report
}

// validateStructure performs structural validation of the configuration
func (cv *ConfigValidator) validateStructure(config *Config) []ValidationIssue {
	var issues []ValidationIssue

	// Check for multiple http blocks
	httpBlocks := config.FindDirectives("http")
	if len(httpBlocks) > 1 {
		for i, httpBlock := range httpBlocks {
			if i > 0 { // First one is OK
				issues = append(issues, ValidationIssue{
					Level:       ValidationError,
					Category:    "Structure",
					Title:       "Multiple HTTP blocks",
					Description: "Only one http block is allowed in nginx configuration",
					Line:        httpBlock.GetLine(),
					Directive:   "http",
					Context:     "main",
					Fix:         "Merge all http configurations into a single http block",
				})
			}
		}
	}

	// Check for multiple events blocks
	eventsBlocks := config.FindDirectives("events")
	if len(eventsBlocks) > 1 {
		for i, eventsBlock := range eventsBlocks {
			if i > 0 { // First one is OK
				issues = append(issues, ValidationIssue{
					Level:       ValidationError,
					Category:    "Structure",
					Title:       "Multiple events blocks",
					Description: "Only one events block is allowed in nginx configuration",
					Line:        eventsBlock.GetLine(),
					Directive:   "events",
					Context:     "main",
					Fix:         "Merge all events configurations into a single events block",
				})
			}
		}
	}

	// Check for server_name conflicts
	issues = append(issues, cv.validateServerNameConflicts(config)...)

	// Check for listen conflicts
	issues = append(issues, cv.validateListenConflicts(config)...)

	return issues
}

// validateServerNameConflicts checks for conflicting server_name directives
func (cv *ConfigValidator) validateServerNameConflicts(config *Config) []ValidationIssue {
	var issues []ValidationIssue

	serverNames := make(map[string][]IDirective)

	// Collect all server_name directives
	servers := config.FindDirectives("server")
	for _, server := range servers {
		if server.GetBlock() != nil {
			for _, directive := range server.GetBlock().GetDirectives() {
				if directive.GetName() == "server_name" {
					params := directive.GetParameters()
					for _, param := range params {
						name := param.Value
						if name != "_" && name != "localhost" { // Skip default names
							serverNames[name] = append(serverNames[name], directive)
						}
					}
				}
			}
		}
	}

	// Check for conflicts
	for name, directives := range serverNames {
		if len(directives) > 1 {
			for i, directive := range directives {
				if i > 0 { // First occurrence is OK
					issues = append(issues, ValidationIssue{
						Level:       ValidationWarning,
						Category:    "Structure",
						Title:       "Duplicate server name",
						Description: fmt.Sprintf("Server name '%s' is defined multiple times", name),
						Line:        directive.GetLine(),
						Directive:   "server_name",
						Context:     "server",
						Fix:         "Use unique server names or configure proper server priority",
					})
				}
			}
		}
	}

	return issues
}

// validateListenConflicts checks for conflicting listen directives
func (cv *ConfigValidator) validateListenConflicts(config *Config) []ValidationIssue {
	var issues []ValidationIssue

	listenAddresses := make(map[string][]IDirective)

	// Collect all listen directives
	servers := config.FindDirectives("server")
	for _, server := range servers {
		if server.GetBlock() != nil {
			for _, directive := range server.GetBlock().GetDirectives() {
				if directive.GetName() == "listen" {
					params := directive.GetParameters()
					if len(params) > 0 {
						address := params[0].Value
						listenAddresses[address] = append(listenAddresses[address], directive)
					}
				}
			}
		}
	}

	// Check for conflicts (same address without proper server_name distinction)
	for address, directives := range listenAddresses {
		if len(directives) > 1 {
			// This might be OK if server_name directives are different
			// For now, just issue a warning
			for i, directive := range directives {
				if i > 0 { // First occurrence is OK
					issues = append(issues, ValidationIssue{
						Level:       ValidationInfo,
						Category:    "Structure",
						Title:       "Multiple servers on same address",
						Description: fmt.Sprintf("Multiple servers listening on '%s'", address),
						Line:        directive.GetLine(),
						Directive:   "listen",
						Context:     "server",
						Fix:         "Ensure proper server_name distinction or use different ports",
					})
				}
			}
		}
	}

	return issues
}

// validateParameters performs parameter validation
func (cv *ConfigValidator) validateParameters(config *Config) []ValidationIssue {
	var issues []ValidationIssue

	// Validate parameters recursively
	issues = append(issues, cv.validateBlockParameters(config, "main")...)

	return issues
}

// validateBlockParameters validates parameters in a block recursively
func (cv *ConfigValidator) validateBlockParameters(block IBlock, context string) []ValidationIssue {
	var issues []ValidationIssue

	if block == nil {
		return issues
	}

	for _, directive := range block.GetDirectives() {
		// Validate parameters for this directive
		issues = append(issues, cv.validateDirectiveParameters(directive, context)...)

		// Recursively validate nested blocks
		if directive.GetBlock() != nil {
			var nestedContext string
			switch directive.GetName() {
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
			default:
				nestedContext = directive.GetName()
			}

			nestedIssues := cv.validateBlockParameters(directive.GetBlock(), nestedContext)
			issues = append(issues, nestedIssues...)
		}
	}

	return issues
}

// validateDirectiveParameters validates parameters for a specific directive
func (cv *ConfigValidator) validateDirectiveParameters(directive IDirective, context string) []ValidationIssue {
	var issues []ValidationIssue

	directiveName := directive.GetName()
	params := directive.GetParameters()

	// Validate based on directive type
	switch directiveName {
	case "listen":
		issues = append(issues, cv.validateListenParameters(directive, params)...)
	case "ssl_certificate", "ssl_certificate_key":
		issues = append(issues, cv.validateSSLFileParameters(directive, params)...)
	case "proxy_pass":
		issues = append(issues, cv.validateProxyPassParameters(directive, params)...)
	case "root", "alias":
		issues = append(issues, cv.validatePathParameters(directive, params)...)
	case "worker_processes":
		issues = append(issues, cv.validateWorkerProcessesParameters(directive, params)...)
	case "worker_connections":
		issues = append(issues, cv.validateWorkerConnectionsParameters(directive, params)...)
	}

	return issues
}

// validateListenParameters validates listen directive parameters
func (cv *ConfigValidator) validateListenParameters(directive IDirective, params []Parameter) []ValidationIssue {
	var issues []ValidationIssue

	if len(params) == 0 {
		issues = append(issues, ValidationIssue{
			Level:       ValidationError,
			Category:    "Parameter",
			Title:       "Missing listen address/port",
			Description: "listen directive requires at least one parameter",
			Line:        directive.GetLine(),
			Directive:   "listen",
			Fix:         "Add port number or address:port (e.g., 'listen 80;')",
		})
	}

	return issues
}

// validateSSLFileParameters validates SSL file parameters
func (cv *ConfigValidator) validateSSLFileParameters(directive IDirective, params []Parameter) []ValidationIssue {
	var issues []ValidationIssue

	if len(params) == 0 {
		issues = append(issues, ValidationIssue{
			Level:       ValidationError,
			Category:    "Parameter",
			Title:       "Missing SSL file path",
			Description: fmt.Sprintf("%s directive requires a file path", directive.GetName()),
			Line:        directive.GetLine(),
			Directive:   directive.GetName(),
			Fix:         "Add path to SSL certificate or key file",
		})
	} else {
		// Check if parameter looks like a file path
		for _, param := range params {
			if param.Type != ParameterTypePath && param.Type != ParameterTypeString {
				issues = append(issues, ValidationIssue{
					Level:       ValidationWarning,
					Category:    "Parameter",
					Title:       "SSL file path format",
					Description: fmt.Sprintf("Parameter '%s' may not be a valid file path", param.Value),
					Line:        directive.GetLine(),
					Directive:   directive.GetName(),
					Fix:         "Ensure parameter is a valid file path",
				})
			}
		}
	}

	return issues
}

// validateProxyPassParameters validates proxy_pass directive parameters
func (cv *ConfigValidator) validateProxyPassParameters(directive IDirective, params []Parameter) []ValidationIssue {
	var issues []ValidationIssue

	if len(params) == 0 {
		issues = append(issues, ValidationIssue{
			Level:       ValidationError,
			Category:    "Parameter",
			Title:       "Missing proxy target",
			Description: "proxy_pass directive requires a target URL or upstream name",
			Line:        directive.GetLine(),
			Directive:   "proxy_pass",
			Fix:         "Add target URL (e.g., 'http://backend') or upstream name",
		})
	}

	return issues
}

// validatePathParameters validates path-related directive parameters
func (cv *ConfigValidator) validatePathParameters(directive IDirective, params []Parameter) []ValidationIssue {
	var issues []ValidationIssue

	if len(params) == 0 {
		issues = append(issues, ValidationIssue{
			Level:       ValidationError,
			Category:    "Parameter",
			Title:       "Missing path",
			Description: fmt.Sprintf("%s directive requires a path parameter", directive.GetName()),
			Line:        directive.GetLine(),
			Directive:   directive.GetName(),
			Fix:         "Add file system path",
		})
	}

	return issues
}

// validateWorkerProcessesParameters validates worker_processes directive parameters
func (cv *ConfigValidator) validateWorkerProcessesParameters(directive IDirective, params []Parameter) []ValidationIssue {
	var issues []ValidationIssue

	if len(params) == 0 {
		issues = append(issues, ValidationIssue{
			Level:       ValidationError,
			Category:    "Parameter",
			Title:       "Missing worker processes count",
			Description: "worker_processes directive requires a number or 'auto'",
			Line:        directive.GetLine(),
			Directive:   "worker_processes",
			Fix:         "Add number of worker processes (e.g., 'auto' or '4')",
		})
	} else if len(params) > 1 {
		issues = append(issues, ValidationIssue{
			Level:       ValidationWarning,
			Category:    "Parameter",
			Title:       "Too many parameters",
			Description: "worker_processes directive should have only one parameter",
			Line:        directive.GetLine(),
			Directive:   "worker_processes",
			Fix:         "Use only one parameter",
		})
	}

	return issues
}

// validateWorkerConnectionsParameters validates worker_connections directive parameters
func (cv *ConfigValidator) validateWorkerConnectionsParameters(directive IDirective, params []Parameter) []ValidationIssue {
	var issues []ValidationIssue

	if len(params) == 0 {
		issues = append(issues, ValidationIssue{
			Level:       ValidationError,
			Category:    "Parameter",
			Title:       "Missing worker connections count",
			Description: "worker_connections directive requires a number",
			Line:        directive.GetLine(),
			Directive:   "worker_connections",
			Fix:         "Add number of worker connections (e.g., '1024')",
		})
	} else {
		for _, param := range params {
			if param.Type != ParameterTypeNumber {
				issues = append(issues, ValidationIssue{
					Level:       ValidationWarning,
					Category:    "Parameter",
					Title:       "Invalid worker connections format",
					Description: fmt.Sprintf("Parameter '%s' should be a number", param.Value),
					Line:        directive.GetLine(),
					Directive:   "worker_connections",
					Fix:         "Use a numeric value",
				})
			}
		}
	}

	return issues
}

// generateSummary generates a summary of validation issues
func (cv *ConfigValidator) generateSummary(issues []ValidationIssue) ValidationSummary {
	summary := ValidationSummary{
		Total:  len(issues),
		Passed: []string{},
	}

	for _, issue := range issues {
		switch issue.Level {
		case ValidationError:
			summary.Errors++
		case ValidationWarning:
			summary.Warning++
		case ValidationInfo:
			summary.Info++
		}
	}

	// Add passed checks
	if summary.Errors == 0 {
		summary.Passed = append(summary.Passed, "No validation errors found")
	}
	if summary.Warning == 0 {
		summary.Passed = append(summary.Passed, "No validation warnings found")
	}

	return summary
}
