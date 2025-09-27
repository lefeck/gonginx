package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/lefeck/gonginx/config"
)

// SecurityLevel represents the severity level of a security issue
type SecurityLevel int

const (
	// SecurityInfo represents informational security notices
	SecurityInfo SecurityLevel = iota
	// SecurityWarning represents security warnings that should be addressed
	SecurityWarning
	// SecurityCritical represents critical security issues that must be fixed
	SecurityCritical
)

// String returns the string representation of the security level
func (sl SecurityLevel) String() string {
	switch sl {
	case SecurityInfo:
		return "INFO"
	case SecurityWarning:
		return "WARNING"
	case SecurityCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// SecurityIssue represents a security-related issue in the configuration
type SecurityIssue struct {
	Level       SecurityLevel
	Category    string
	Title       string
	Description string
	Directive   string
	Parameter   string
	Context     string
	Fix         string
	Reference   string
}

// String returns a human-readable representation of the security issue
func (si *SecurityIssue) String() string {
	return fmt.Sprintf("[%s] %s: %s - %s",
		si.Level.String(), si.Category, si.Title, si.Description)
}

// SecurityReport contains all security issues found in a configuration
type SecurityReport struct {
	Issues  []SecurityIssue
	Summary SecuritySummary
	Passed  []string
	Config  *config.Config
}

// SecuritySummary provides summary statistics about security issues
type SecuritySummary struct {
	Total    int
	Critical int
	Warning  int
	Info     int
	Score    int // Security score out of 100
}

// String returns a summary of the security report
func (ss *SecuritySummary) String() string {
	return fmt.Sprintf("Security Score: %d/100, Issues: %d (Critical: %d, Warning: %d, Info: %d)",
		ss.Score, ss.Total, ss.Critical, ss.Warning, ss.Info)
}

// HasCriticalIssues returns true if there are critical security issues
func (sr *SecurityReport) HasCriticalIssues() bool {
	return sr.Summary.Critical > 0
}

// GetByLevel returns issues of a specific security level
func (sr *SecurityReport) GetByLevel(level SecurityLevel) []SecurityIssue {
	var filtered []SecurityIssue
	for _, issue := range sr.Issues {
		if issue.Level == level {
			filtered = append(filtered, issue)
		}
	}
	return filtered
}

// GetByCategory returns issues of a specific category
func (sr *SecurityReport) GetByCategory(category string) []SecurityIssue {
	var filtered []SecurityIssue
	for _, issue := range sr.Issues {
		if issue.Category == category {
			filtered = append(filtered, issue)
		}
	}
	return filtered
}

// SecurityChecker performs security analysis on nginx configurations
type SecurityChecker struct {
	config *config.Config
	report *SecurityReport
}

// NewSecurityChecker creates a new security checker
func NewSecurityChecker(conf *config.Config) *SecurityChecker {
	return &SecurityChecker{
		config: conf,
		report: &SecurityReport{
			Issues: make([]SecurityIssue, 0),
			Passed: make([]string, 0),
			Config: conf,
		},
	}
}

// CheckSecurity performs a comprehensive security analysis
func CheckSecurity(conf *config.Config) *SecurityReport {
	checker := NewSecurityChecker(conf)
	checker.runAllChecks()
	return checker.report
}

// runAllChecks executes all security checks
func (sc *SecurityChecker) runAllChecks() {
	// HTTP Security Checks
	sc.checkServerTokens()
	sc.checkSSLConfiguration()
	sc.checkSecurityHeaders()
	sc.checkDirectoryTraversal()
	sc.checkFileUploadSecurity()
	sc.checkRateLimiting()
	sc.checkAccessControl()

	// General Security Checks
	sc.checkErrorPages()
	sc.checkLogConfiguration()
	sc.checkWorkerProcesses()
	sc.checkFilePermissions()

	// Calculate security score
	sc.calculateSecurityScore()
	sc.calculateSummary()
}

// checkServerTokens checks if server tokens are disabled
func (sc *SecurityChecker) checkServerTokens() {
	httpBlocks := sc.config.FindDirectives("http")

	serverTokensFound := false
	for _, httpBlock := range httpBlocks {
		if http, ok := httpBlock.(*config.HTTP); ok {
			serverTokensDirs := http.FindDirectives("server_tokens")
			for _, dir := range serverTokensDirs {
				serverTokensFound = true
				if len(dir.GetParameters()) > 0 {
					value := dir.GetParameters()[0].GetValue()
					if value != "off" {
						sc.addIssue(SecurityWarning, "Information Disclosure",
							"Server tokens not disabled",
							"Server version information is exposed in HTTP headers and error pages",
							"server_tokens", value, "http",
							"Set 'server_tokens off;' in the http block",
							"https://nginx.org/en/docs/http/ngx_http_core_module.html#server_tokens")
					} else {
						sc.addPassed("Server tokens properly disabled")
					}
				}
			}
		}
	}

	if !serverTokensFound {
		sc.addIssue(SecurityWarning, "Information Disclosure",
			"Server tokens not configured",
			"Server version information may be exposed by default",
			"server_tokens", "", "http",
			"Add 'server_tokens off;' to the http block",
			"https://nginx.org/en/docs/http/ngx_http_core_module.html#server_tokens")
	}
}

// checkSSLConfiguration checks SSL/TLS configuration security
func (sc *SecurityChecker) checkSSLConfiguration() {
	httpBlocks := sc.config.FindDirectives("http")

	for _, httpBlock := range httpBlocks {
		if http, ok := httpBlock.(*config.HTTP); ok {
			for _, server := range http.Servers {
				sc.checkServerSSL(server)
			}
		}
	}
}

func (sc *SecurityChecker) checkServerSSL(server *config.Server) {
	// Check SSL protocols
	protocolDirs := server.FindDirectives("ssl_protocols")
	if len(protocolDirs) > 0 {
		for _, dir := range protocolDirs {
			protocols := make([]string, 0)
			for _, param := range dir.GetParameters() {
				protocols = append(protocols, param.GetValue())
			}

			// Check for insecure protocols
			insecureProtocols := []string{"SSLv2", "SSLv3", "TLSv1", "TLSv1.1"}
			for _, protocol := range protocols {
				for _, insecure := range insecureProtocols {
					if strings.EqualFold(protocol, insecure) {
						sc.addIssue(SecurityCritical, "SSL/TLS Security",
							"Insecure SSL/TLS protocol enabled",
							fmt.Sprintf("Protocol %s is vulnerable to attacks", protocol),
							"ssl_protocols", protocol, "server",
							"Use only TLSv1.2 and TLSv1.3",
							"https://ssl-config.mozilla.org/")
					}
				}
			}

			// Check if modern protocols are used
			hasModernProtocol := false
			for _, protocol := range protocols {
				if strings.EqualFold(protocol, "TLSv1.2") || strings.EqualFold(protocol, "TLSv1.3") {
					hasModernProtocol = true
					break
				}
			}

			if hasModernProtocol {
				sc.addPassed("Modern SSL/TLS protocols configured")
			}
		}
	} else {
		// Check if SSL is used at all
		certDirs := server.FindDirectives("ssl_certificate")
		if len(certDirs) > 0 {
			sc.addIssue(SecurityWarning, "SSL/TLS Security",
				"SSL protocols not explicitly configured",
				"Default SSL protocols may include insecure versions",
				"ssl_protocols", "", "server",
				"Explicitly configure ssl_protocols with TLSv1.2 and TLSv1.3 only",
				"https://ssl-config.mozilla.org/")
		}
	}

	// Check SSL ciphers
	cipherDirs := server.FindDirectives("ssl_ciphers")
	if len(cipherDirs) > 0 {
		for _, dir := range cipherDirs {
			if len(dir.GetParameters()) > 0 {
				ciphers := dir.GetParameters()[0].GetValue()

				// Check for weak ciphers
				weakCiphers := []string{"RC4", "DES", "3DES", "MD5", "NULL"}
				for _, weak := range weakCiphers {
					if strings.Contains(strings.ToUpper(ciphers), weak) {
						sc.addIssue(SecurityCritical, "SSL/TLS Security",
							"Weak SSL cipher enabled",
							fmt.Sprintf("Cipher suite contains weak cipher: %s", weak),
							"ssl_ciphers", ciphers, "server",
							"Use strong cipher suites only",
							"https://ssl-config.mozilla.org/")
					}
				}
			}
		}
	}

	// Check HSTS
	sc.checkHSTS(server)
}

func (sc *SecurityChecker) checkHSTS(server *config.Server) {
	hstsFound := false
	locationDirs := server.FindDirectives("location")

	// Check in server block
	addHeaderDirs := server.FindDirectives("add_header")
	for _, dir := range addHeaderDirs {
		if len(dir.GetParameters()) >= 2 {
			header := dir.GetParameters()[0].GetValue()
			if strings.EqualFold(header, "Strict-Transport-Security") {
				hstsFound = true
				value := dir.GetParameters()[1].GetValue()

				// Check HSTS value
				if !strings.Contains(value, "max-age") {
					sc.addIssue(SecurityWarning, "SSL/TLS Security",
						"Invalid HSTS configuration",
						"HSTS header missing max-age directive",
						"add_header", value, "server",
						"Include max-age in HSTS header",
						"https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Strict-Transport-Security")
				} else {
					sc.addPassed("HSTS properly configured")
				}
			}
		}
	}

	// Check in location blocks
	for _, locationDir := range locationDirs {
		if location, ok := locationDir.(*config.Location); ok {
			locationAddHeaders := location.FindDirectives("add_header")
			for _, dir := range locationAddHeaders {
				if len(dir.GetParameters()) >= 2 {
					header := dir.GetParameters()[0].GetValue()
					if strings.EqualFold(header, "Strict-Transport-Security") {
						hstsFound = true
					}
				}
			}
		}
	}

	// Check if SSL is configured but HSTS is missing
	sslCertDirs := server.FindDirectives("ssl_certificate")
	if len(sslCertDirs) > 0 && !hstsFound {
		sc.addIssue(SecurityWarning, "SSL/TLS Security",
			"HSTS not configured",
			"HTTPS server missing HTTP Strict Transport Security header",
			"add_header", "", "server",
			"Add 'add_header Strict-Transport-Security \"max-age=31536000; includeSubDomains\" always;'",
			"https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Strict-Transport-Security")
	}
}

// checkSecurityHeaders checks for important security headers
func (sc *SecurityChecker) checkSecurityHeaders() {
	httpBlocks := sc.config.FindDirectives("http")

	securityHeaders := map[string]string{
		"X-Content-Type-Options":  "nosniff",
		"X-Frame-Options":         "DENY or SAMEORIGIN",
		"X-XSS-Protection":        "1; mode=block",
		"Referrer-Policy":         "strict-origin-when-cross-origin",
		"Content-Security-Policy": "appropriate CSP directives",
	}

	for _, httpBlock := range httpBlocks {
		if http, ok := httpBlock.(*config.HTTP); ok {
			foundHeaders := make(map[string]bool)

			// Check in http block
			sc.checkHeadersInDirectives(http.Directives, foundHeaders, "http")

			// Check in servers
			for _, server := range http.Servers {
				sc.checkHeadersInDirectives(server.GetDirectives(), foundHeaders, "server")
			}

			// Report missing headers
			for header, recommendation := range securityHeaders {
				if !foundHeaders[strings.ToLower(header)] {
					sc.addIssue(SecurityWarning, "Security Headers",
						fmt.Sprintf("Missing %s header", header),
						fmt.Sprintf("Security header %s not configured", header),
						"add_header", "", "server",
						fmt.Sprintf("Add 'add_header %s \"%s\" always;'", header, recommendation),
						"https://owasp.org/www-project-secure-headers/")
				} else {
					sc.addPassed(fmt.Sprintf("Security header %s configured", header))
				}
			}
		}
	}
}

func (sc *SecurityChecker) checkHeadersInDirectives(directives []config.IDirective, foundHeaders map[string]bool, context string) {
	for _, dir := range directives {
		if dir.GetName() == "add_header" && len(dir.GetParameters()) >= 2 {
			header := strings.ToLower(dir.GetParameters()[0].GetValue())
			foundHeaders[header] = true
		}

		// Check in nested blocks (locations)
		if dir.GetBlock() != nil {
			sc.checkHeadersInDirectives(dir.GetBlock().GetDirectives(), foundHeaders, context)
		}
	}
}

// checkDirectoryTraversal checks for directory traversal vulnerabilities
func (sc *SecurityChecker) checkDirectoryTraversal() {
	httpBlocks := sc.config.FindDirectives("http")

	for _, httpBlock := range httpBlocks {
		if http, ok := httpBlock.(*config.HTTP); ok {
			for _, server := range http.Servers {
				sc.checkServerDirectoryTraversal(server)
			}
		}
	}
}

func (sc *SecurityChecker) checkServerDirectoryTraversal(server *config.Server) {
	// Check root directive
	rootDirs := server.FindDirectives("root")
	for _, dir := range rootDirs {
		if len(dir.GetParameters()) > 0 {
			rootPath := dir.GetParameters()[0].GetValue()
			if strings.Contains(rootPath, "../") {
				sc.addIssue(SecurityCritical, "Directory Traversal",
					"Dangerous root path",
					"Root path contains directory traversal sequences",
					"root", rootPath, "server",
					"Use absolute paths without .. sequences",
					"https://owasp.org/www-community/attacks/Path_Traversal")
			}
		}
	}

	// Check location blocks
	locationDirs := server.FindDirectives("location")
	for _, locationDir := range locationDirs {
		if location, ok := locationDir.(*config.Location); ok {
			sc.checkLocationSecurity(location)
		}
	}
}

func (sc *SecurityChecker) checkLocationSecurity(location *config.Location) {
	// Check for unsafe location patterns
	params := location.GetParameters()
	if len(params) > 0 {
		pattern := params[len(params)-1].GetValue() // Last parameter is usually the pattern

		// Check for dangerous patterns
		dangerousPatterns := []string{
			`\.\.`,             // Directory traversal
			`\.(php|asp|jsp)$`, // Executable files (if not properly configured)
		}

		for _, dangerous := range dangerousPatterns {
			if matched, _ := regexp.MatchString(dangerous, pattern); matched {
				sc.addIssue(SecurityWarning, "Location Security",
					"Potentially dangerous location pattern",
					fmt.Sprintf("Location pattern '%s' may allow unauthorized access", pattern),
					"location", pattern, "server",
					"Review location pattern and add proper access controls",
					"https://nginx.org/en/docs/http/ngx_http_core_module.html#location")
			}
		}
	}

	// Check for directory listing
	autoindexDirs := location.FindDirectives("autoindex")
	for _, dir := range autoindexDirs {
		if len(dir.GetParameters()) > 0 {
			value := dir.GetParameters()[0].GetValue()
			if value == "on" {
				sc.addIssue(SecurityWarning, "Information Disclosure",
					"Directory listing enabled",
					"Autoindex allows directory browsing which may expose sensitive files",
					"autoindex", value, "location",
					"Set 'autoindex off;' unless directory listing is required",
					"https://nginx.org/en/docs/http/ngx_http_autoindex_module.html")
			}
		}
	}
}

// checkFileUploadSecurity checks file upload security configurations
func (sc *SecurityChecker) checkFileUploadSecurity() {
	httpBlocks := sc.config.FindDirectives("http")

	for _, httpBlock := range httpBlocks {
		if http, ok := httpBlock.(*config.HTTP); ok {
			// Check client_max_body_size
			bodySize := sc.findDirectiveValue(http.Directives, "client_max_body_size")
			if bodySize != "" {
				if sc.isLargeSize(bodySize) {
					sc.addIssue(SecurityWarning, "File Upload Security",
						"Large file upload limit",
						fmt.Sprintf("Client max body size is set to %s which may allow large file uploads", bodySize),
						"client_max_body_size", bodySize, "http",
						"Consider reducing the file upload limit if large uploads are not required",
						"https://nginx.org/en/docs/http/ngx_http_core_module.html#client_max_body_size")
				}
			}

			// Check for file upload locations without restrictions
			for _, server := range http.Servers {
				sc.checkServerFileUpload(server)
			}
		}
	}
}

func (sc *SecurityChecker) checkServerFileUpload(server *config.Server) {
	locationDirs := server.FindDirectives("location")

	for _, locationDir := range locationDirs {
		if location, ok := locationDir.(*config.Location); ok {
			// Check if this location handles file uploads
			if sc.isUploadLocation(location) {
				// Check for upload restrictions
				if !sc.hasUploadRestrictions(location) {
					sc.addIssue(SecurityWarning, "File Upload Security",
						"Unrestricted file upload location",
						"File upload location lacks proper restrictions",
						"location", sc.getLocationPattern(location), "server",
						"Add file type restrictions and size limits",
						"https://owasp.org/www-project-web-security-testing-guide/latest/4-Web_Application_Security_Testing/10-Business_Logic_Testing/09-Test_Upload_of_Malicious_Files")
				}
			}
		}
	}
}

// checkRateLimiting checks for rate limiting configurations
func (sc *SecurityChecker) checkRateLimiting() {
	httpBlocks := sc.config.FindDirectives("http")

	rateLimitingFound := false
	for _, httpBlock := range httpBlocks {
		if http, ok := httpBlock.(*config.HTTP); ok {
			// Check for limit_req_zone
			limitReqZones := http.FindDirectives("limit_req_zone")
			if len(limitReqZones) > 0 {
				rateLimitingFound = true
				sc.addPassed("Rate limiting configured with limit_req_zone")
			}

			// Check for limit_conn_zone
			limitConnZones := http.FindDirectives("limit_conn_zone")
			if len(limitConnZones) > 0 {
				rateLimitingFound = true
				sc.addPassed("Connection limiting configured with limit_conn_zone")
			}
		}
	}

	if !rateLimitingFound {
		sc.addIssue(SecurityWarning, "Rate Limiting",
			"Rate limiting not configured",
			"No rate limiting or connection limiting configured, server may be vulnerable to DoS attacks",
			"limit_req_zone", "", "http",
			"Configure rate limiting with limit_req_zone and limit_req",
			"https://nginx.org/en/docs/http/ngx_http_limit_req_module.html")
	}
}

// checkAccessControl checks access control configurations
func (sc *SecurityChecker) checkAccessControl() {
	httpBlocks := sc.config.FindDirectives("http")

	for _, httpBlock := range httpBlocks {
		if http, ok := httpBlock.(*config.HTTP); ok {
			for _, server := range http.Servers {
				sc.checkServerAccessControl(server)
			}
		}
	}
}

func (sc *SecurityChecker) checkServerAccessControl(server *config.Server) {
	// Check for admin or sensitive locations without access control
	locationDirs := server.FindDirectives("location")

	for _, locationDir := range locationDirs {
		if location, ok := locationDir.(*config.Location); ok {
			pattern := sc.getLocationPattern(location)

			// Check for admin/sensitive paths
			sensitivePatterns := []string{
				"admin", "wp-admin", "phpmyadmin", "adminer",
				".git", ".svn", ".env", "backup",
			}

			for _, sensitive := range sensitivePatterns {
				if strings.Contains(strings.ToLower(pattern), sensitive) {
					if !sc.hasAccessControl(location) {
						sc.addIssue(SecurityCritical, "Access Control",
							"Sensitive location without access control",
							fmt.Sprintf("Sensitive location '%s' lacks proper access control", pattern),
							"location", pattern, "server",
							"Add IP restrictions, authentication, or deny directives",
							"https://nginx.org/en/docs/http/ngx_http_access_module.html")
					}
				}
			}
		}
	}
}

// checkErrorPages checks error page configurations
func (sc *SecurityChecker) checkErrorPages() {
	httpBlocks := sc.config.FindDirectives("http")

	for _, httpBlock := range httpBlocks {
		if http, ok := httpBlock.(*config.HTTP); ok {
			errorPageFound := false
			errorPages := http.FindDirectives("error_page")
			if len(errorPages) > 0 {
				errorPageFound = true
				sc.addPassed("Custom error pages configured")
			}

			for _, server := range http.Servers {
				serverErrorPages := server.FindDirectives("error_page")
				if len(serverErrorPages) > 0 {
					errorPageFound = true
				}
			}

			if !errorPageFound {
				sc.addIssue(SecurityInfo, "Information Disclosure",
					"Default error pages",
					"Using default error pages may expose server information",
					"error_page", "", "server",
					"Configure custom error pages to hide server details",
					"https://nginx.org/en/docs/http/ngx_http_core_module.html#error_page")
			}
		}
	}
}

// checkLogConfiguration checks logging security
func (sc *SecurityChecker) checkLogConfiguration() {
	httpBlocks := sc.config.FindDirectives("http")

	for _, httpBlock := range httpBlocks {
		if http, ok := httpBlock.(*config.HTTP); ok {
			// Check access log
			accessLogs := http.FindDirectives("access_log")
			if len(accessLogs) > 0 {
				for _, log := range accessLogs {
					if len(log.GetParameters()) > 0 {
						logPath := log.GetParameters()[0].GetValue()
						if logPath == "off" {
							sc.addIssue(SecurityWarning, "Logging",
								"Access logging disabled",
								"Access logging is disabled, reducing audit capabilities",
								"access_log", logPath, "http",
								"Enable access logging for security monitoring",
								"https://nginx.org/en/docs/http/ngx_http_log_module.html")
						} else {
							sc.addPassed("Access logging enabled")
						}
					}
				}
			}

			// Check error log
			errorLogs := http.FindDirectives("error_log")
			if len(errorLogs) > 0 {
				sc.addPassed("Error logging configured")
			} else {
				sc.addIssue(SecurityInfo, "Logging",
					"Error logging not explicitly configured",
					"Error logging should be explicitly configured for security monitoring",
					"error_log", "", "http",
					"Configure error_log directive",
					"https://nginx.org/en/docs/ngx_core_module.html#error_log")
			}
		}
	}
}

// checkWorkerProcesses checks worker process security
func (sc *SecurityChecker) checkWorkerProcesses() {
	workerProcessDirs := sc.config.FindDirectives("worker_processes")

	if len(workerProcessDirs) > 0 {
		for _, dir := range workerProcessDirs {
			if len(dir.GetParameters()) > 0 {
				value := dir.GetParameters()[0].GetValue()
				if value != "auto" {
					if num, err := strconv.Atoi(value); err == nil {
						if num > 32 {
							sc.addIssue(SecurityWarning, "Resource Management",
								"High number of worker processes",
								fmt.Sprintf("Worker processes set to %d, which may impact system resources", num),
								"worker_processes", value, "main",
								"Consider using 'auto' or a reasonable number based on CPU cores",
								"https://nginx.org/en/docs/ngx_core_module.html#worker_processes")
						}
					}
				} else {
					sc.addPassed("Worker processes set to auto")
				}
			}
		}
	}
}

// checkFilePermissions checks file permission related configurations
func (sc *SecurityChecker) checkFilePermissions() {
	// This is a placeholder for file permission checks
	// In a real implementation, you might check:
	// - User/group directives
	// - File paths for proper permissions
	// - Temporary directory configurations

	userDirs := sc.config.FindDirectives("user")
	if len(userDirs) > 0 {
		for _, dir := range userDirs {
			if len(dir.GetParameters()) > 0 {
				user := dir.GetParameters()[0].GetValue()
				if user == "root" {
					sc.addIssue(SecurityCritical, "Privilege Escalation",
						"Running as root user",
						"Nginx is configured to run as root, which is a security risk",
						"user", user, "main",
						"Configure nginx to run as a non-privileged user",
						"https://nginx.org/en/docs/ngx_core_module.html#user")
				} else {
					sc.addPassed("Running as non-root user")
				}
			}
		}
	}
}

// Helper methods

func (sc *SecurityChecker) addIssue(level SecurityLevel, category, title, description, directive, parameter, context, fix, reference string) {
	issue := SecurityIssue{
		Level:       level,
		Category:    category,
		Title:       title,
		Description: description,
		Directive:   directive,
		Parameter:   parameter,
		Context:     context,
		Fix:         fix,
		Reference:   reference,
	}
	sc.report.Issues = append(sc.report.Issues, issue)
}

func (sc *SecurityChecker) addPassed(description string) {
	sc.report.Passed = append(sc.report.Passed, description)
}

func (sc *SecurityChecker) findDirectiveValue(directives []config.IDirective, directiveName string) string {
	for _, dir := range directives {
		if dir.GetName() == directiveName && len(dir.GetParameters()) > 0 {
			return dir.GetParameters()[0].GetValue()
		}
	}
	return ""
}

func (sc *SecurityChecker) isLargeSize(size string) bool {
	// Simple check for large file sizes
	size = strings.ToLower(size)
	if strings.Contains(size, "g") {
		return true // Any gigabyte size is considered large
	}
	if strings.Contains(size, "m") {
		if num, err := strconv.Atoi(strings.TrimSuffix(size, "m")); err == nil {
			return num > 100 // More than 100MB
		}
	}
	return false
}

func (sc *SecurityChecker) isUploadLocation(location *config.Location) bool {
	// Check if location appears to handle uploads
	pattern := sc.getLocationPattern(location)
	uploadKeywords := []string{"upload", "file", "media", "content"}

	for _, keyword := range uploadKeywords {
		if strings.Contains(strings.ToLower(pattern), keyword) {
			return true
		}
	}
	return false
}

func (sc *SecurityChecker) hasUploadRestrictions(location *config.Location) bool {
	// Check for upload restrictions
	directives := location.GetDirectives()

	for _, dir := range directives {
		switch dir.GetName() {
		case "client_max_body_size", "limit_except", "if":
			return true
		}
	}
	return false
}

func (sc *SecurityChecker) getLocationPattern(location *config.Location) string {
	params := location.GetParameters()
	if len(params) > 0 {
		return params[len(params)-1].GetValue()
	}
	return ""
}

func (sc *SecurityChecker) hasAccessControl(location *config.Location) bool {
	directives := location.GetDirectives()

	for _, dir := range directives {
		switch dir.GetName() {
		case "allow", "deny", "auth_basic", "auth_request", "access_by_lua":
			return true
		}
	}
	return false
}

func (sc *SecurityChecker) calculateSecurityScore() {
	score := 100

	for _, issue := range sc.report.Issues {
		switch issue.Level {
		case SecurityCritical:
			score -= 20
		case SecurityWarning:
			score -= 10
		case SecurityInfo:
			score -= 2
		}
	}

	if score < 0 {
		score = 0
	}

	sc.report.Summary.Score = score
}

func (sc *SecurityChecker) calculateSummary() {
	summary := SecuritySummary{Score: sc.report.Summary.Score}

	for _, issue := range sc.report.Issues {
		summary.Total++
		switch issue.Level {
		case SecurityCritical:
			summary.Critical++
		case SecurityWarning:
			summary.Warning++
		case SecurityInfo:
			summary.Info++
		}
	}

	sc.report.Summary = summary
}
