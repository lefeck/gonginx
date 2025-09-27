package config

import (
	"strconv"
	"strings"
)

// DetectParameterType automatically detects the type of a parameter based on its value
func DetectParameterType(value string) ParameterType {
	if value == "" {
		return ParameterTypeString
	}

	// Check for variables (starts with $)
	if strings.HasPrefix(value, "$") {
		return ParameterTypeVariable
	}

	// Check for quoted strings
	if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
		(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
		return ParameterTypeQuoted
	}

	// Check for boolean values
	lowerValue := strings.ToLower(value)
	if lowerValue == "on" || lowerValue == "off" ||
		lowerValue == "yes" || lowerValue == "no" ||
		lowerValue == "true" || lowerValue == "false" ||
		lowerValue == "enable" || lowerValue == "disable" {
		return ParameterTypeBoolean
	}

	// Check for time values first (e.g., 30s, 5m, 1h, 1d)
	if isTime(value) {
		return ParameterTypeTime
	}

	// Check for size values (e.g., 1M, 512k, 1G)
	if isSize(value) {
		return ParameterTypeSize
	}

	// Check for numbers
	if isNumber(value) {
		return ParameterTypeNumber
	}

	// Check for regex patterns first (before paths, as some regex may contain /)
	if isRegex(value) {
		return ParameterTypeRegex
	}

	// Check for URLs
	if isURL(value) {
		return ParameterTypeURL
	}

	// Check for file paths
	if isPath(value) {
		return ParameterTypePath
	}

	// Default to string
	return ParameterTypeString
}

// isNumber checks if the value is a number
func isNumber(value string) bool {
	_, err := strconv.ParseFloat(value, 64)
	return err == nil
}

// isSize checks if the value is a size specification
func isSize(value string) bool {
	if len(value) == 0 {
		return false
	}

	// Check if it ends with size units
	lowerValue := strings.ToLower(value)
	if strings.HasSuffix(lowerValue, "k") || strings.HasSuffix(lowerValue, "m") ||
		strings.HasSuffix(lowerValue, "g") || strings.HasSuffix(lowerValue, "t") ||
		strings.HasSuffix(lowerValue, "kb") || strings.HasSuffix(lowerValue, "mb") ||
		strings.HasSuffix(lowerValue, "gb") || strings.HasSuffix(lowerValue, "tb") {

		// Remove the suffix and check if the remaining part is a number
		var numPart string
		if strings.HasSuffix(lowerValue, "kb") || strings.HasSuffix(lowerValue, "mb") ||
			strings.HasSuffix(lowerValue, "gb") || strings.HasSuffix(lowerValue, "tb") {
			numPart = lowerValue[:len(lowerValue)-2]
		} else {
			numPart = lowerValue[:len(lowerValue)-1]
		}

		_, err := strconv.ParseFloat(numPart, 64)
		return err == nil
	}

	return false
}

// isTime checks if the value is a time specification
func isTime(value string) bool {
	if len(value) == 0 {
		return false
	}

	// Check if it ends with time units (be more specific about units)
	lowerValue := strings.ToLower(value)

	// Specific time units that are unambiguous
	timeUnits := []string{"ms", "s", "h", "d", "w", "y"}

	for _, unit := range timeUnits {
		if strings.HasSuffix(lowerValue, unit) {
			// Remove the suffix and check if the remaining part is a number
			numPart := lowerValue[:len(lowerValue)-len(unit)]
			_, err := strconv.ParseFloat(numPart, 64)
			return err == nil
		}
	}

	// Note: We don't handle "m" for minutes here because in nginx context
	// "m" typically means megabytes, not minutes. Use "min" for minutes explicitly.

	return false
}

// isURL checks if the value looks like a URL
func isURL(value string) bool {
	return strings.HasPrefix(value, "http://") ||
		strings.HasPrefix(value, "https://") ||
		strings.HasPrefix(value, "ftp://") ||
		strings.HasPrefix(value, "ftps://") ||
		strings.HasPrefix(value, "unix:")
}

// isPath checks if the value looks like a file or directory path
func isPath(value string) bool {
	// Don't consider regex patterns as paths
	if strings.HasPrefix(value, "^") || strings.HasSuffix(value, "$") {
		return false
	}

	// Absolute paths
	if strings.HasPrefix(value, "/") {
		return true
	}

	// Windows paths
	if len(value) >= 3 && value[1] == ':' && (value[2] == '\\' || value[2] == '/') {
		return true
	}

	// Relative paths with common patterns
	if strings.HasPrefix(value, "./") || strings.HasPrefix(value, "../") {
		return true
	}

	// Common file extensions
	commonExtensions := []string{
		".conf", ".log", ".pid", ".sock", ".key", ".crt", ".pem",
		".txt", ".html", ".css", ".js", ".json", ".xml", ".yml", ".yaml",
	}

	for _, ext := range commonExtensions {
		if strings.HasSuffix(strings.ToLower(value), ext) {
			return true
		}
	}

	// Common directory patterns (but avoid regex-like patterns)
	if strings.Contains(value, "/") && !strings.Contains(value, " ") {
		// Make sure it's not a regex pattern
		regexChars := []string{"*", "+", "?", "[", "]", "(", ")", "|"}
		hasRegexChars := false
		for _, char := range regexChars {
			if strings.Contains(value, char) {
				hasRegexChars = true
				break
			}
		}
		return !hasRegexChars
	}

	return false
}

// isRegex checks if the value looks like a regular expression
func isRegex(value string) bool {
	// Explicitly marked as regex with ~ prefix
	if strings.HasPrefix(value, "~") {
		return true
	}

	// Strong regex indicators
	if strings.HasPrefix(value, "^") || strings.HasSuffix(value, "$") {
		return true
	}

	// Don't confuse Windows paths with regex
	if len(value) >= 3 && value[1] == ':' && (value[2] == '\\' || value[2] == '/') {
		return false
	}

	// Contains regex metacharacters (but be careful about backslashes in paths)
	regexChars := []string{"^", "$", "*", "+", "?", "[", "]", "(", ")", "|"}
	charCount := 0

	for _, char := range regexChars {
		if strings.Contains(value, char) {
			charCount++
		}
	}

	// Special handling for dots - only count as regex if multiple or in specific patterns
	if strings.Contains(value, ".") {
		// Patterns like .*php$ or .+ are likely regex
		if strings.Contains(value, ".*") || strings.Contains(value, ".+") {
			charCount++
		}
	}

	// If it contains multiple regex metacharacters, it's likely a regex
	return charCount >= 2
}

// NewParameter creates a new parameter with automatic type detection
func NewParameter(value string) Parameter {
	return Parameter{
		Value: value,
		Type:  DetectParameterType(value),
	}
}

// NewParameterWithType creates a new parameter with explicit type
func NewParameterWithType(value string, paramType ParameterType) Parameter {
	return Parameter{
		Value: value,
		Type:  paramType,
	}
}

// Common parameter type validators

// ValidateSize validates a size parameter and returns the normalized value
func ValidateSize(value string) (string, bool) {
	if !isSize(value) {
		return "", false
	}
	// Could add normalization logic here (e.g., convert kb to k)
	return value, true
}

// ValidateTime validates a time parameter and returns the normalized value
func ValidateTime(value string) (string, bool) {
	if !isTime(value) {
		return "", false
	}
	// Could add normalization logic here
	return value, true
}

// ValidateNumber validates a number parameter and returns the parsed value
func ValidateNumber(value string) (float64, bool) {
	if num, err := strconv.ParseFloat(value, 64); err == nil {
		return num, true
	}
	return 0, false
}

// ValidateBoolean validates a boolean parameter and returns the boolean value
func ValidateBoolean(value string) (bool, bool) {
	lowerValue := strings.ToLower(value)
	switch lowerValue {
	case "on", "yes", "true", "enable", "1":
		return true, true
	case "off", "no", "false", "disable", "0":
		return false, true
	default:
		return false, false
	}
}
