package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/lefeck/gonginx/config"
)

// OptimizationType represents the type of optimization
type OptimizationType int

const (
	// OptimizePerformance focuses on performance improvements
	OptimizePerformance OptimizationType = iota
	// OptimizeSize focuses on reducing configuration size
	OptimizeSize
	// OptimizeSecurity focuses on security improvements
	OptimizeSecurity
	// OptimizeMaintenance focuses on maintainability
	OptimizeMaintenance
)

// String returns the string representation of the optimization type
func (ot OptimizationType) String() string {
	switch ot {
	case OptimizePerformance:
		return "Performance"
	case OptimizeSize:
		return "Size"
	case OptimizeSecurity:
		return "Security"
	case OptimizeMaintenance:
		return "Maintenance"
	default:
		return "Unknown"
	}
}

// OptimizationSuggestion represents a suggested optimization
type OptimizationSuggestion struct {
	Type           OptimizationType
	Category       string
	Title          string
	Description    string
	Impact         string
	CurrentValue   string
	SuggestedValue string
	Directive      string
	Context        string
	Reason         string
	Implementation string
}

// String returns a human-readable representation of the optimization suggestion
func (os *OptimizationSuggestion) String() string {
	return fmt.Sprintf("[%s] %s: %s -> %s (%s)",
		os.Type.String(), os.Title, os.CurrentValue, os.SuggestedValue, os.Impact)
}

// OptimizationReport contains all optimization suggestions
type OptimizationReport struct {
	Suggestions []OptimizationSuggestion
	Summary     OptimizationSummary
	Config      *config.Config
}

// OptimizationSummary provides summary statistics about optimizations
type OptimizationSummary struct {
	Total           int
	Performance     int
	Size            int
	Security        int
	Maintenance     int
	EstimatedImpact string
}

// String returns a summary of the optimization report
func (os *OptimizationSummary) String() string {
	return fmt.Sprintf("Optimizations: %d (Performance: %d, Size: %d, Security: %d, Maintenance: %d)",
		os.Total, os.Performance, os.Size, os.Security, os.Maintenance)
}

// GetByType returns suggestions of a specific optimization type
func (or *OptimizationReport) GetByType(optimizationType OptimizationType) []OptimizationSuggestion {
	var filtered []OptimizationSuggestion
	for _, suggestion := range or.Suggestions {
		if suggestion.Type == optimizationType {
			filtered = append(filtered, suggestion)
		}
	}
	return filtered
}

// GetByCategory returns suggestions of a specific category
func (or *OptimizationReport) GetByCategory(category string) []OptimizationSuggestion {
	var filtered []OptimizationSuggestion
	for _, suggestion := range or.Suggestions {
		if suggestion.Category == category {
			filtered = append(filtered, suggestion)
		}
	}
	return filtered
}

// ConfigOptimizer analyzes and optimizes nginx configurations
type ConfigOptimizer struct {
	config *config.Config
	report *OptimizationReport
}

// NewConfigOptimizer creates a new configuration optimizer
func NewConfigOptimizer(conf *config.Config) *ConfigOptimizer {
	return &ConfigOptimizer{
		config: conf,
		report: &OptimizationReport{
			Suggestions: make([]OptimizationSuggestion, 0),
			Config:      conf,
		},
	}
}

// OptimizeConfig analyzes a configuration and returns optimization suggestions
func OptimizeConfig(conf *config.Config) *OptimizationReport {
	optimizer := NewConfigOptimizer(conf)
	optimizer.runAllOptimizations()
	return optimizer.report
}

// runAllOptimizations executes all optimization checks
func (co *ConfigOptimizer) runAllOptimizations() {
	// Performance optimizations
	co.optimizeWorkerConfiguration()
	co.optimizeBuffers()
	co.optimizeKeepalive()
	co.optimizeGzip()
	co.optimizeSSL()
	co.optimizeCaching()

	// Size optimizations
	co.removeDuplicates()
	co.consolidateDirectives()
	co.removeDefaultValues()

	// Security optimizations
	co.optimizeSecurityHeaders()
	co.optimizeSSLSecurity()

	// Maintenance optimizations
	co.addMissingComments()
	co.organizeDirectives()

	co.calculateSummary()
}

// optimizeWorkerConfiguration suggests worker process optimizations
func (co *ConfigOptimizer) optimizeWorkerConfiguration() {
	workerProcessDirs := co.config.FindDirectives("worker_processes")

	if len(workerProcessDirs) == 0 {
		co.addSuggestion(OptimizePerformance, "Worker Configuration",
			"Add worker_processes directive",
			"Worker processes not configured",
			"High",
			"", "auto",
			"worker_processes", "main",
			"Auto-configure worker processes for optimal performance",
			"Add 'worker_processes auto;' to the main context")
	} else {
		for _, dir := range workerProcessDirs {
			if len(dir.GetParameters()) > 0 {
				value := dir.GetParameters()[0].GetValue()
				if value != "auto" {
					if num, err := strconv.Atoi(value); err == nil {
						if num == 1 {
							co.addSuggestion(OptimizePerformance, "Worker Configuration",
								"Use auto worker processes",
								"Single worker process may limit performance",
								"Medium",
								value, "auto",
								"worker_processes", "main",
								"Auto-configure based on CPU cores for better performance",
								"Change to 'worker_processes auto;'")
						}
					}
				}
			}
		}
	}

	// Check worker connections
	eventsDirs := co.config.FindDirectives("events")
	for _, eventsDir := range eventsDirs {
		if eventsDir.GetBlock() != nil {
			workerConnDirs := eventsDir.GetBlock().GetDirectives()
			hasWorkerConnections := false

			for _, dir := range workerConnDirs {
				if dir.GetName() == "worker_connections" {
					hasWorkerConnections = true
					if len(dir.GetParameters()) > 0 {
						value := dir.GetParameters()[0].GetValue()
						if num, err := strconv.Atoi(value); err == nil {
							if num < 1024 {
								co.addSuggestion(OptimizePerformance, "Worker Configuration",
									"Increase worker connections",
									"Worker connections may be too low for high traffic",
									"Medium",
									value, "1024",
									"worker_connections", "events",
									"Increase for better concurrent connection handling",
									"Change to 'worker_connections 1024;' or higher")
							}
						}
					}
				}
			}

			if !hasWorkerConnections {
				co.addSuggestion(OptimizePerformance, "Worker Configuration",
					"Add worker_connections directive",
					"Worker connections not configured",
					"High",
					"", "1024",
					"worker_connections", "events",
					"Configure maximum connections per worker",
					"Add 'worker_connections 1024;' to events block")
			}
		}
	}
}

// optimizeBuffers suggests buffer size optimizations
func (co *ConfigOptimizer) optimizeBuffers() {
	httpBlocks := co.config.FindDirectives("http")

	for _, httpBlock := range httpBlocks {
		if http, ok := httpBlock.(*config.HTTP); ok {
			co.checkHTTPBuffers(http)

			for _, server := range http.Servers {
				co.checkServerBuffers(server)
			}
		}
	}
}

func (co *ConfigOptimizer) checkHTTPBuffers(http *config.HTTP) {
	bufferDirectives := map[string]string{
		"client_body_buffer_size":     "16k",
		"client_header_buffer_size":   "1k",
		"large_client_header_buffers": "4 16k",
		"output_buffers":              "2 32k",
	}

	for directive, suggestedValue := range bufferDirectives {
		found := false
		for _, dir := range http.Directives {
			if dir.GetName() == directive {
				found = true
				break
			}
		}

		if !found {
			co.addSuggestion(OptimizePerformance, "Buffer Configuration",
				fmt.Sprintf("Add %s directive", directive),
				fmt.Sprintf("Buffer size for %s not optimized", directive),
				"Medium",
				"", suggestedValue,
				directive, "http",
				"Optimize buffer sizes for better performance",
				fmt.Sprintf("Add '%s %s;' to http block", directive, suggestedValue))
		}
	}
}

func (co *ConfigOptimizer) checkServerBuffers(server *config.Server) {
	// Check proxy buffers if proxy is used
	proxyPassDirs := server.FindDirectives("proxy_pass")
	if len(proxyPassDirs) > 0 {
		proxyBufferDirectives := map[string]string{
			"proxy_buffer_size":       "4k",
			"proxy_buffers":           "8 4k",
			"proxy_busy_buffers_size": "8k",
		}

		for directive, suggestedValue := range proxyBufferDirectives {
			found := false
			for _, dir := range server.GetDirectives() {
				if dir.GetName() == directive {
					found = true
					break
				}
			}

			if !found {
				co.addSuggestion(OptimizePerformance, "Proxy Buffer Configuration",
					fmt.Sprintf("Add %s directive", directive),
					fmt.Sprintf("Proxy buffer %s not configured", directive),
					"Medium",
					"", suggestedValue,
					directive, "server",
					"Optimize proxy buffering for better performance",
					fmt.Sprintf("Add '%s %s;' to server block", directive, suggestedValue))
			}
		}
	}
}

// optimizeKeepalive suggests keepalive optimizations
func (co *ConfigOptimizer) optimizeKeepalive() {
	httpBlocks := co.config.FindDirectives("http")

	for _, httpBlock := range httpBlocks {
		if http, ok := httpBlock.(*config.HTTP); ok {
			keepaliveFound := false

			for _, dir := range http.Directives {
				if dir.GetName() == "keepalive_timeout" {
					keepaliveFound = true
					if len(dir.GetParameters()) > 0 {
						value := dir.GetParameters()[0].GetValue()
						if timeout, err := strconv.Atoi(strings.TrimSuffix(value, "s")); err == nil {
							if timeout > 75 {
								co.addSuggestion(OptimizePerformance, "Keepalive Configuration",
									"Reduce keepalive timeout",
									"Keepalive timeout too high may waste connections",
									"Low",
									value, "65s",
									"keepalive_timeout", "http",
									"Optimize connection reuse without wasting resources",
									"Change to 'keepalive_timeout 65s;'")
							}
						}
					}
				}
			}

			if !keepaliveFound {
				co.addSuggestion(OptimizePerformance, "Keepalive Configuration",
					"Add keepalive_timeout directive",
					"Keepalive timeout not configured",
					"Medium",
					"", "65s",
					"keepalive_timeout", "http",
					"Enable connection reuse for better performance",
					"Add 'keepalive_timeout 65s;' to http block")
			}
		}
	}
}

// optimizeGzip suggests gzip compression optimizations
func (co *ConfigOptimizer) optimizeGzip() {
	httpBlocks := co.config.FindDirectives("http")

	for _, httpBlock := range httpBlocks {
		if http, ok := httpBlock.(*config.HTTP); ok {
			gzipEnabled := false
			gzipTypes := false

			for _, dir := range http.Directives {
				if dir.GetName() == "gzip" {
					gzipEnabled = true
					if len(dir.GetParameters()) > 0 {
						value := dir.GetParameters()[0].GetValue()
						if value != "on" {
							co.addSuggestion(OptimizePerformance, "Compression",
								"Enable gzip compression",
								"Gzip compression disabled",
								"High",
								value, "on",
								"gzip", "http",
								"Enable compression to reduce bandwidth usage",
								"Change to 'gzip on;'")
						}
					}
				} else if dir.GetName() == "gzip_types" {
					gzipTypes = true
				}
			}

			if !gzipEnabled {
				co.addSuggestion(OptimizePerformance, "Compression",
					"Enable gzip compression",
					"Gzip compression not configured",
					"High",
					"", "on",
					"gzip", "http",
					"Enable compression to reduce bandwidth usage",
					"Add 'gzip on;' to http block")
			}

			if gzipEnabled && !gzipTypes {
				suggestedTypes := "text/plain text/css application/json application/javascript text/xml application/xml application/xml+rss text/javascript"
				co.addSuggestion(OptimizePerformance, "Compression",
					"Configure gzip_types",
					"Gzip types not specified",
					"Medium",
					"", suggestedTypes,
					"gzip_types", "http",
					"Specify which file types to compress",
					fmt.Sprintf("Add 'gzip_types %s;' to http block", suggestedTypes))
			}
		}
	}
}

// optimizeSSL suggests SSL performance optimizations
func (co *ConfigOptimizer) optimizeSSL() {
	httpBlocks := co.config.FindDirectives("http")

	for _, httpBlock := range httpBlocks {
		if http, ok := httpBlock.(*config.HTTP); ok {
			for _, server := range http.Servers {
				co.checkSSLOptimization(server)
			}
		}
	}
}

func (co *ConfigOptimizer) checkSSLOptimization(server *config.Server) {
	sslCertDirs := server.FindDirectives("ssl_certificate")
	if len(sslCertDirs) == 0 {
		return // No SSL configured
	}

	// Check SSL session cache
	sslSessionCacheDirs := server.FindDirectives("ssl_session_cache")
	if len(sslSessionCacheDirs) == 0 {
		co.addSuggestion(OptimizePerformance, "SSL Performance",
			"Add SSL session cache",
			"SSL session cache not configured",
			"High",
			"", "shared:SSL:10m",
			"ssl_session_cache", "server",
			"Cache SSL sessions for better performance",
			"Add 'ssl_session_cache shared:SSL:10m;' to server block")
	}

	// Check SSL session timeout
	sslSessionTimeoutDirs := server.FindDirectives("ssl_session_timeout")
	if len(sslSessionTimeoutDirs) == 0 {
		co.addSuggestion(OptimizePerformance, "SSL Performance",
			"Add SSL session timeout",
			"SSL session timeout not configured",
			"Medium",
			"", "1d",
			"ssl_session_timeout", "server",
			"Configure SSL session timeout for optimal caching",
			"Add 'ssl_session_timeout 1d;' to server block")
	}

	// Check SSL stapling
	sslStaplingDirs := server.FindDirectives("ssl_stapling")
	if len(sslStaplingDirs) == 0 {
		co.addSuggestion(OptimizePerformance, "SSL Performance",
			"Enable SSL stapling",
			"SSL stapling not enabled",
			"Medium",
			"", "on",
			"ssl_stapling", "server",
			"Enable OCSP stapling for faster SSL handshakes",
			"Add 'ssl_stapling on;' and 'ssl_stapling_verify on;' to server block")
	}
}

// optimizeCaching suggests caching optimizations
func (co *ConfigOptimizer) optimizeCaching() {
	httpBlocks := co.config.FindDirectives("http")

	for _, httpBlock := range httpBlocks {
		if http, ok := httpBlock.(*config.HTTP); ok {
			for _, server := range http.Servers {
				co.checkCaching(server)
			}
		}
	}
}

func (co *ConfigOptimizer) checkCaching(server *config.Server) {
	locationDirs := server.FindDirectives("location")

	for _, locationDir := range locationDirs {
		if location, ok := locationDir.(*config.Location); ok {
			pattern := co.getLocationPattern(location)

			// Check for static file locations
			if co.isStaticFileLocation(pattern) {
				expiresDirs := location.FindDirectives("expires")
				if len(expiresDirs) == 0 {
					co.addSuggestion(OptimizePerformance, "Caching",
						"Add expires directive for static files",
						fmt.Sprintf("Static file location '%s' lacks caching headers", pattern),
						"Medium",
						"", "1y",
						"expires", "location",
						"Cache static files for better performance",
						"Add 'expires 1y;' to location block")
				}
			}
		}
	}
}

// removeDuplicates suggests removing duplicate directives
func (co *ConfigOptimizer) removeDuplicates() {
	// This is a simplified implementation
	// In a real implementation, you would track directive occurrences
	// and suggest consolidation
}

// consolidateDirectives suggests consolidating similar directives
func (co *ConfigOptimizer) consolidateDirectives() {
	httpBlocks := co.config.FindDirectives("http")

	for _, httpBlock := range httpBlocks {
		if http, ok := httpBlock.(*config.HTTP); ok {
			// Check for multiple add_header directives that could be consolidated
			addHeaderCount := 0
			for _, dir := range http.Directives {
				if dir.GetName() == "add_header" {
					addHeaderCount++
				}
			}

			if addHeaderCount > 3 {
				co.addSuggestion(OptimizeSize, "Directive Consolidation",
					"Consolidate add_header directives",
					fmt.Sprintf("Found %d add_header directives that could be consolidated", addHeaderCount),
					"Low",
					fmt.Sprintf("%d directives", addHeaderCount), "consolidated block",
					"add_header", "http",
					"Group related headers together for better maintainability",
					"Consider grouping headers in a single location or include file")
			}
		}
	}
}

// removeDefaultValues suggests removing default value declarations
func (co *ConfigOptimizer) removeDefaultValues() {
	defaultValues := map[string]string{
		"sendfile":             "on",
		"tcp_nopush":           "off",
		"tcp_nodelay":          "on",
		"keepalive_timeout":    "75s",
		"client_max_body_size": "1m",
		"gzip":                 "off",
	}

	httpBlocks := co.config.FindDirectives("http")
	for _, httpBlock := range httpBlocks {
		if http, ok := httpBlock.(*config.HTTP); ok {
			for _, dir := range http.Directives {
				if defaultValue, isDefault := defaultValues[dir.GetName()]; isDefault {
					if len(dir.GetParameters()) > 0 {
						currentValue := dir.GetParameters()[0].GetValue()
						if currentValue == defaultValue {
							co.addSuggestion(OptimizeSize, "Default Values",
								fmt.Sprintf("Remove default %s directive", dir.GetName()),
								fmt.Sprintf("Directive %s is set to default value", dir.GetName()),
								"Low",
								currentValue, "remove",
								dir.GetName(), "http",
								"Remove unnecessary default value declarations",
								fmt.Sprintf("Remove '%s %s;' as it's the default", dir.GetName(), currentValue))
						}
					}
				}
			}
		}
	}
}

// optimizeSecurityHeaders suggests security header optimizations
func (co *ConfigOptimizer) optimizeSecurityHeaders() {
	// This integrates with the security checker
	// In a real implementation, you would check for missing security headers
	// and suggest adding them
}

// optimizeSSLSecurity suggests SSL security optimizations
func (co *ConfigOptimizer) optimizeSSLSecurity() {
	httpBlocks := co.config.FindDirectives("http")

	for _, httpBlock := range httpBlocks {
		if http, ok := httpBlock.(*config.HTTP); ok {
			for _, server := range http.Servers {
				co.checkSSLSecurityOptimization(server)
			}
		}
	}
}

func (co *ConfigOptimizer) checkSSLSecurityOptimization(server *config.Server) {
	sslCertDirs := server.FindDirectives("ssl_certificate")
	if len(sslCertDirs) == 0 {
		return // No SSL configured
	}

	// Check SSL protocols
	protocolDirs := server.FindDirectives("ssl_protocols")
	if len(protocolDirs) > 0 {
		for _, dir := range protocolDirs {
			protocols := make([]string, 0)
			for _, param := range dir.GetParameters() {
				protocols = append(protocols, param.GetValue())
			}

			// Check for old protocols
			hasOldProtocol := false
			for _, protocol := range protocols {
				if protocol == "TLSv1" || protocol == "TLSv1.1" {
					hasOldProtocol = true
					break
				}
			}

			if hasOldProtocol {
				co.addSuggestion(OptimizeSecurity, "SSL Security",
					"Update SSL protocols",
					"SSL configuration includes outdated protocols",
					"High",
					strings.Join(protocols, " "), "TLSv1.2 TLSv1.3",
					"ssl_protocols", "server",
					"Use only modern secure protocols",
					"Change to 'ssl_protocols TLSv1.2 TLSv1.3;'")
			}
		}
	}
}

// addMissingComments suggests adding comments for maintainability
func (co *ConfigOptimizer) addMissingComments() {
	// Check for complex configurations without comments
	httpBlocks := co.config.FindDirectives("http")

	for _, httpBlock := range httpBlocks {
		if http, ok := httpBlock.(*config.HTTP); ok {
			for _, server := range http.Servers {
				// Check if server has complex configuration but no comments
				directiveCount := len(server.GetDirectives())
				commentCount := len(server.GetComment())

				if directiveCount > 10 && commentCount == 0 {
					co.addSuggestion(OptimizeMaintenance, "Documentation",
						"Add comments to complex server block",
						"Complex server block lacks documentation",
						"Low",
						"no comments", "descriptive comments",
						"# comment", "server",
						"Add comments for better maintainability",
						"Add comments describing the purpose of this server block")
				}
			}
		}
	}
}

// organizeDirectives suggests better directive organization
func (co *ConfigOptimizer) organizeDirectives() {
	// Check for directive organization
	httpBlocks := co.config.FindDirectives("http")

	for _, httpBlock := range httpBlocks {
		if http, ok := httpBlock.(*config.HTTP); ok {
			directiveNames := make([]string, 0)
			for _, dir := range http.Directives {
				directiveNames = append(directiveNames, dir.GetName())
			}

			// Check if directives are organized logically
			if !co.areDirectivesOrganized(directiveNames) {
				co.addSuggestion(OptimizeMaintenance, "Organization",
					"Reorganize directives",
					"Directives could be better organized",
					"Low",
					"current order", "logical grouping",
					"organization", "http",
					"Group related directives together for better readability",
					"Organize directives by category (basic settings, compression, security, etc.)")
			}
		}
	}
}

// Helper methods

func (co *ConfigOptimizer) addSuggestion(optimizationType OptimizationType, category, title, description, impact, currentValue, suggestedValue, directive, context, reason, implementation string) {
	suggestion := OptimizationSuggestion{
		Type:           optimizationType,
		Category:       category,
		Title:          title,
		Description:    description,
		Impact:         impact,
		CurrentValue:   currentValue,
		SuggestedValue: suggestedValue,
		Directive:      directive,
		Context:        context,
		Reason:         reason,
		Implementation: implementation,
	}
	co.report.Suggestions = append(co.report.Suggestions, suggestion)
}

func (co *ConfigOptimizer) getLocationPattern(location *config.Location) string {
	params := location.GetParameters()
	if len(params) > 0 {
		return params[len(params)-1].GetValue()
	}
	return ""
}

func (co *ConfigOptimizer) isStaticFileLocation(pattern string) bool {
	staticExtensions := []string{".css", ".js", ".png", ".jpg", ".jpeg", ".gif", ".ico", ".svg", ".woff", ".woff2", ".ttf", ".eot"}

	for _, ext := range staticExtensions {
		if strings.Contains(pattern, ext) {
			return true
		}
	}
	return false
}

func (co *ConfigOptimizer) areDirectivesOrganized(directives []string) bool {
	// Simple heuristic: check if related directives are grouped
	// This is a simplified implementation
	return true // Placeholder
}

func (co *ConfigOptimizer) calculateSummary() {
	summary := OptimizationSummary{}

	for _, suggestion := range co.report.Suggestions {
		summary.Total++
		switch suggestion.Type {
		case OptimizePerformance:
			summary.Performance++
		case OptimizeSize:
			summary.Size++
		case OptimizeSecurity:
			summary.Security++
		case OptimizeMaintenance:
			summary.Maintenance++
		}
	}

	// Calculate estimated impact
	highImpact := 0
	for _, suggestion := range co.report.Suggestions {
		if suggestion.Impact == "High" {
			highImpact++
		}
	}

	if highImpact > 0 {
		summary.EstimatedImpact = "High"
	} else {
		summary.EstimatedImpact = "Medium"
	}

	co.report.Summary = summary
}
