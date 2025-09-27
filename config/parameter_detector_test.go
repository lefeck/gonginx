package config

import (
	"testing"
)

func TestDetectParameterType(t *testing.T) {
	tests := []struct {
		value    string
		expected ParameterType
	}{
		// Variables
		{"$request_uri", ParameterTypeVariable},
		{"$remote_addr", ParameterTypeVariable},

		// Quoted strings
		{"\"hello world\"", ParameterTypeQuoted},
		{"'hello world'", ParameterTypeQuoted},

		// Boolean values
		{"on", ParameterTypeBoolean},
		{"off", ParameterTypeBoolean},
		{"yes", ParameterTypeBoolean},
		{"no", ParameterTypeBoolean},
		{"true", ParameterTypeBoolean},
		{"false", ParameterTypeBoolean},
		{"enable", ParameterTypeBoolean},
		{"disable", ParameterTypeBoolean},

		// Size values
		{"1M", ParameterTypeSize},
		{"512k", ParameterTypeSize},
		{"1G", ParameterTypeSize},
		{"100MB", ParameterTypeSize},
		{"2GB", ParameterTypeSize},

		// Time values
		{"30s", ParameterTypeTime},
		{"1h", ParameterTypeTime},
		{"7d", ParameterTypeTime},
		{"100ms", ParameterTypeTime},

		// Numbers
		{"123", ParameterTypeNumber},
		{"3.14", ParameterTypeNumber},
		{"0", ParameterTypeNumber},
		{"-1", ParameterTypeNumber},

		// URLs
		{"http://example.com", ParameterTypeURL},
		{"https://example.com", ParameterTypeURL},
		{"unix:/var/run/nginx.sock", ParameterTypeURL},

		// Paths
		{"/etc/nginx/nginx.conf", ParameterTypePath},
		{"/var/log/nginx.log", ParameterTypePath},
		{"./config.conf", ParameterTypePath},
		{"../logs/error.log", ParameterTypePath},
		{"C:\\nginx\\conf\\nginx.conf", ParameterTypePath},

		// Regex
		{"~.*\\.php$", ParameterTypeRegex},
		{"^/api/", ParameterTypeRegex},

		// Regular strings
		{"server_name", ParameterTypeString},
		{"localhost", ParameterTypeString},
		{"example", ParameterTypeString},
	}

	for _, test := range tests {
		result := DetectParameterType(test.value)
		if result != test.expected {
			t.Errorf("DetectParameterType(%q) = %v, expected %v", test.value, result, test.expected)
		}
	}
}

func TestParameterTypeMethods(t *testing.T) {
	// Test IsVariable
	param1 := NewParameter("$request_uri")
	if !param1.IsVariable() {
		t.Error("Expected parameter to be detected as variable")
	}

	// Test IsSize
	param2 := NewParameter("1M")
	if !param2.IsSize() {
		t.Error("Expected parameter to be detected as size")
	}

	// Test IsTime
	param3 := NewParameter("30s")
	if !param3.IsTime() {
		t.Error("Expected parameter to be detected as time")
	}

	// Test IsNumber
	param4 := NewParameter("123")
	if !param4.IsNumber() {
		t.Error("Expected parameter to be detected as number")
	}

	// Test IsBoolean
	param5 := NewParameter("on")
	if !param5.IsBoolean() {
		t.Error("Expected parameter to be detected as boolean")
	}
}

func TestValidators(t *testing.T) {
	// Test ValidateSize
	if _, valid := ValidateSize("1M"); !valid {
		t.Error("Expected '1M' to be valid size")
	}
	if _, valid := ValidateSize("invalid"); valid {
		t.Error("Expected 'invalid' to be invalid size")
	}

	// Test ValidateTime
	if _, valid := ValidateTime("30s"); !valid {
		t.Error("Expected '30s' to be valid time")
	}
	if _, valid := ValidateTime("invalid"); valid {
		t.Error("Expected 'invalid' to be invalid time")
	}

	// Test ValidateNumber
	if num, valid := ValidateNumber("123.45"); !valid || num != 123.45 {
		t.Error("Expected '123.45' to be valid number with value 123.45")
	}
	if _, valid := ValidateNumber("invalid"); valid {
		t.Error("Expected 'invalid' to be invalid number")
	}

	// Test ValidateBoolean
	if val, valid := ValidateBoolean("on"); !valid || !val {
		t.Error("Expected 'on' to be valid boolean with value true")
	}
	if val, valid := ValidateBoolean("off"); !valid || val {
		t.Error("Expected 'off' to be valid boolean with value false")
	}
	if _, valid := ValidateBoolean("invalid"); valid {
		t.Error("Expected 'invalid' to be invalid boolean")
	}
}

func TestNewParameterWithType(t *testing.T) {
	param := NewParameterWithType("custom_value", ParameterTypeString)
	if param.Value != "custom_value" {
		t.Errorf("Expected value 'custom_value', got %s", param.Value)
	}
	if param.Type != ParameterTypeString {
		t.Errorf("Expected type %v, got %v", ParameterTypeString, param.Type)
	}
}

func TestParameterTypeString(t *testing.T) {
	tests := []struct {
		paramType ParameterType
		expected  string
	}{
		{ParameterTypeString, "string"},
		{ParameterTypeVariable, "variable"},
		{ParameterTypeNumber, "number"},
		{ParameterTypeSize, "size"},
		{ParameterTypeTime, "time"},
		{ParameterTypePath, "path"},
		{ParameterTypeURL, "url"},
		{ParameterTypeRegex, "regex"},
		{ParameterTypeBoolean, "boolean"},
		{ParameterTypeQuoted, "quoted"},
	}

	for _, test := range tests {
		result := test.paramType.String()
		if result != test.expected {
			t.Errorf("ParameterType(%d).String() = %s, expected %s", test.paramType, result, test.expected)
		}
	}
}
