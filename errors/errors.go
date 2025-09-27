package errors

import (
	"fmt"
	"strings"
)

// ErrorType represents the type of parsing error
type ErrorType int

const (
	// SyntaxError represents a syntax error in the configuration
	SyntaxError ErrorType = iota
	// SemanticError represents a semantic error (e.g., invalid parameter)
	SemanticError
	// ContextError represents a context-related error (e.g., directive in wrong block)
	ContextError
	// FileError represents a file-related error (e.g., include file not found)
	FileError
	// ValidationError represents a validation error
	ValidationError
	// UnknownDirectiveError represents an unknown directive error
	UnknownDirectiveError
)

// String returns the string representation of the error type
func (et ErrorType) String() string {
	switch et {
	case SyntaxError:
		return "Syntax Error"
	case SemanticError:
		return "Semantic Error"
	case ContextError:
		return "Context Error"
	case FileError:
		return "File Error"
	case ValidationError:
		return "Validation Error"
	case UnknownDirectiveError:
		return "Unknown Directive Error"
	default:
		return "Unknown Error"
	}
}

// ParseError represents a detailed parsing error with context
type ParseError struct {
	Type       ErrorType
	Message    string
	File       string
	Line       int
	Column     int
	Context    string
	Suggestion string
	Directive  string
	Parameter  string
	InnerError error
}

// Error implements the error interface
func (pe *ParseError) Error() string {
	var parts []string

	// Add error type and main message
	parts = append(parts, fmt.Sprintf("[%s] %s", pe.Type.String(), pe.Message))

	// Add location information
	if pe.File != "" {
		if pe.Line > 0 {
			if pe.Column > 0 {
				parts = append(parts, fmt.Sprintf("at %s:%d:%d", pe.File, pe.Line, pe.Column))
			} else {
				parts = append(parts, fmt.Sprintf("at %s:%d", pe.File, pe.Line))
			}
		} else {
			parts = append(parts, fmt.Sprintf("in file %s", pe.File))
		}
	}

	// Add directive context
	if pe.Directive != "" {
		parts = append(parts, fmt.Sprintf("in directive '%s'", pe.Directive))
	}

	// Add parameter context
	if pe.Parameter != "" {
		parts = append(parts, fmt.Sprintf("with parameter '%s'", pe.Parameter))
	}

	// Add context if available
	if pe.Context != "" {
		parts = append(parts, fmt.Sprintf("\nContext: %s", pe.Context))
	}

	// Add suggestion if available
	if pe.Suggestion != "" {
		parts = append(parts, fmt.Sprintf("\nSuggestion: %s", pe.Suggestion))
	}

	// Add inner error if available
	if pe.InnerError != nil {
		parts = append(parts, fmt.Sprintf("\nCaused by: %s", pe.InnerError.Error()))
	}

	return strings.Join(parts, " ")
}

// Unwrap returns the inner error for error wrapping
func (pe *ParseError) Unwrap() error {
	return pe.InnerError
}

// NewParseError creates a new parse error
func NewParseError(errorType ErrorType, message string) *ParseError {
	return &ParseError{
		Type:    errorType,
		Message: message,
	}
}

// NewSyntaxError creates a new syntax error
func NewSyntaxError(message string) *ParseError {
	return &ParseError{
		Type:    SyntaxError,
		Message: message,
	}
}

// NewSemanticError creates a new semantic error
func NewSemanticError(message string) *ParseError {
	return &ParseError{
		Type:    SemanticError,
		Message: message,
	}
}

// NewContextError creates a new context error
func NewContextError(message string) *ParseError {
	return &ParseError{
		Type:    ContextError,
		Message: message,
	}
}

// NewFileError creates a new file error
func NewFileError(message string) *ParseError {
	return &ParseError{
		Type:    FileError,
		Message: message,
	}
}

// NewValidationError creates a new validation error
func NewValidationError(message string) *ParseError {
	return &ParseError{
		Type:    ValidationError,
		Message: message,
	}
}

// NewUnknownDirectiveError creates a new unknown directive error
func NewUnknownDirectiveError(directive string) *ParseError {
	return &ParseError{
		Type:       UnknownDirectiveError,
		Message:    fmt.Sprintf("unknown directive '%s'", directive),
		Directive:  directive,
		Suggestion: getSuggestionForUnknownDirective(directive),
	}
}

// WithFile adds file information to the error
func (pe *ParseError) WithFile(file string) *ParseError {
	pe.File = file
	return pe
}

// WithLine adds line information to the error
func (pe *ParseError) WithLine(line int) *ParseError {
	pe.Line = line
	return pe
}

// WithColumn adds column information to the error
func (pe *ParseError) WithColumn(column int) *ParseError {
	pe.Column = column
	return pe
}

// WithContext adds context information to the error
func (pe *ParseError) WithContext(context string) *ParseError {
	pe.Context = context
	return pe
}

// WithSuggestion adds a suggestion to the error
func (pe *ParseError) WithSuggestion(suggestion string) *ParseError {
	pe.Suggestion = suggestion
	return pe
}

// WithDirective adds directive information to the error
func (pe *ParseError) WithDirective(directive string) *ParseError {
	pe.Directive = directive
	return pe
}

// WithParameter adds parameter information to the error
func (pe *ParseError) WithParameter(parameter string) *ParseError {
	pe.Parameter = parameter
	return pe
}

// WithInnerError adds an inner error to wrap
func (pe *ParseError) WithInnerError(err error) *ParseError {
	pe.InnerError = err
	return pe
}

// ErrorCollection represents a collection of errors
type ErrorCollection struct {
	Errors []*ParseError
}

// Error implements the error interface
func (ec *ErrorCollection) Error() string {
	if len(ec.Errors) == 0 {
		return "no errors"
	}

	if len(ec.Errors) == 1 {
		return ec.Errors[0].Error()
	}

	var parts []string
	parts = append(parts, fmt.Sprintf("Multiple errors (%d):", len(ec.Errors)))

	for i, err := range ec.Errors {
		parts = append(parts, fmt.Sprintf("%d. %s", i+1, err.Error()))
	}

	return strings.Join(parts, "\n")
}

// Add adds an error to the collection
func (ec *ErrorCollection) Add(err *ParseError) {
	ec.Errors = append(ec.Errors, err)
}

// HasErrors returns true if there are any errors
func (ec *ErrorCollection) HasErrors() bool {
	return len(ec.Errors) > 0
}

// Count returns the number of errors
func (ec *ErrorCollection) Count() int {
	return len(ec.Errors)
}

// Clear clears all errors
func (ec *ErrorCollection) Clear() {
	ec.Errors = nil
}

// GetByType returns errors of a specific type
func (ec *ErrorCollection) GetByType(errorType ErrorType) []*ParseError {
	var filtered []*ParseError
	for _, err := range ec.Errors {
		if err.Type == errorType {
			filtered = append(filtered, err)
		}
	}
	return filtered
}

// NewErrorCollection creates a new error collection
func NewErrorCollection() *ErrorCollection {
	return &ErrorCollection{
		Errors: make([]*ParseError, 0),
	}
}

// getSuggestionForUnknownDirective provides suggestions for common misspellings
func getSuggestionForUnknownDirective(directive string) string {
	suggestions := map[string]string{
		"servername":           "server_name",
		"server-name":          "server_name",
		"servernames":          "server_name",
		"listenport":           "listen",
		"listen_port":          "listen",
		"documentroot":         "root",
		"document_root":        "root",
		"doc_root":             "root",
		"proxypass":            "proxy_pass",
		"proxy-pass":           "proxy_pass",
		"proxy_timeout":        "proxy_read_timeout",
		"keepalive":            "keepalive_timeout",
		"keep_alive":           "keepalive_timeout",
		"gzip_enable":          "gzip",
		"enable_gzip":          "gzip",
		"ssl_cert":             "ssl_certificate",
		"ssl_certificate_file": "ssl_certificate",
		"ssl_key":              "ssl_certificate_key",
		"ssl_private_key":      "ssl_certificate_key",
		"workerprocesses":      "worker_processes",
		"worker-processes":     "worker_processes",
		"workerconnections":    "worker_connections",
		"worker-connections":   "worker_connections",
		"clientmaxbodysize":    "client_max_body_size",
		"client-max-body-size": "client_max_body_size",
		"max_body_size":        "client_max_body_size",
		"upstream_server":      "server",
		"backend_server":       "server",
	}

	if suggestion, exists := suggestions[strings.ToLower(directive)]; exists {
		return fmt.Sprintf("Did you mean '%s'?", suggestion)
	}

	// Check for common patterns
	lower := strings.ToLower(directive)
	if strings.Contains(lower, "ssl") && !strings.Contains(lower, "ssl_") {
		return "SSL directives typically use underscore format (e.g., ssl_certificate)"
	}

	if strings.Contains(lower, "proxy") && !strings.Contains(lower, "proxy_") {
		return "Proxy directives typically use underscore format (e.g., proxy_pass)"
	}

	if strings.Contains(lower, "worker") && !strings.Contains(lower, "worker_") {
		return "Worker directives typically use underscore format (e.g., worker_processes)"
	}

	return "Check the nginx documentation for valid directives"
}
