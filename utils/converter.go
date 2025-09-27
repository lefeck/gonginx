package utils

import (
	"encoding/json"
	"fmt"

	"github.com/lefeck/gonginx/config"
	"gopkg.in/yaml.v2"
)

// ConfigFormat represents the configuration format
type ConfigFormat int

const (
	// FormatNginx represents the native nginx configuration format
	FormatNginx ConfigFormat = iota
	// FormatJSON represents JSON format
	FormatJSON
	// FormatYAML represents YAML format
	FormatYAML
	// FormatTOML represents TOML format
	FormatTOML
)

// String returns the string representation of the format
func (cf ConfigFormat) String() string {
	switch cf {
	case FormatNginx:
		return "nginx"
	case FormatJSON:
		return "json"
	case FormatYAML:
		return "yaml"
	case FormatTOML:
		return "toml"
	default:
		return "unknown"
	}
}

// ConfigConverter handles conversion between different configuration formats
type ConfigConverter struct {
	config *config.Config
}

// NewConfigConverter creates a new configuration converter
func NewConfigConverter(conf *config.Config) *ConfigConverter {
	return &ConfigConverter{
		config: conf,
	}
}

// ConvertToJSON converts the configuration to JSON format
func (cc *ConfigConverter) ConvertToJSON(pretty bool) (string, error) {
	configMap := cc.configToMap()

	var data []byte
	var err error

	if pretty {
		data, err = json.MarshalIndent(configMap, "", "  ")
	} else {
		data, err = json.Marshal(configMap)
	}

	if err != nil {
		return "", fmt.Errorf("failed to marshal to JSON: %w", err)
	}

	return string(data), nil
}

// ConvertToYAML converts the configuration to YAML format
func (cc *ConfigConverter) ConvertToYAML() (string, error) {
	configMap := cc.configToMap()

	data, err := yaml.Marshal(configMap)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to YAML: %w", err)
	}

	return string(data), nil
}

// ConvertFromJSON converts JSON configuration to nginx format
func ConvertFromJSON(jsonData string) (*config.Config, error) {
	var configMap map[string]interface{}

	err := json.Unmarshal([]byte(jsonData), &configMap)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return mapToConfig(configMap)
}

// ConvertFromYAML converts YAML configuration to nginx format
func ConvertFromYAML(yamlData string) (*config.Config, error) {
	var configMap map[string]interface{}

	err := yaml.Unmarshal([]byte(yamlData), &configMap)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return mapToConfig(configMap)
}

// configToMap converts a config.Config to a map representation
func (cc *ConfigConverter) configToMap() map[string]interface{} {
	result := make(map[string]interface{})

	if cc.config.Block != nil {
		cc.processBlock(cc.config.Block, result)
	}

	return result
}

// processBlock processes a configuration block and adds it to the result map
func (cc *ConfigConverter) processBlock(block config.IBlock, result map[string]interface{}) {
	if block == nil {
		return
	}

	directives := block.GetDirectives()
	for _, directive := range directives {
		cc.processDirective(directive, result)
	}
}

// processDirective processes a single directive
func (cc *ConfigConverter) processDirective(directive config.IDirective, result map[string]interface{}) {
	name := directive.GetName()

	// Handle special directive types
	switch d := directive.(type) {
	case *config.HTTP:
		result["http"] = cc.processHTTPBlock(d)
	case *config.Server:
		if _, exists := result["servers"]; !exists {
			result["servers"] = make([]interface{}, 0)
		}
		servers := result["servers"].([]interface{})
		result["servers"] = append(servers, cc.processServerBlock(d))
	case *config.Location:
		if _, exists := result["locations"]; !exists {
			result["locations"] = make([]interface{}, 0)
		}
		locations := result["locations"].([]interface{})
		result["locations"] = append(locations, cc.processLocationBlock(d))
	case *config.Upstream:
		if _, exists := result["upstreams"]; !exists {
			result["upstreams"] = make([]interface{}, 0)
		}
		upstreams := result["upstreams"].([]interface{})
		result["upstreams"] = append(upstreams, cc.processUpstreamBlock(d))
	default:
		// Handle regular directives
		params := directive.GetParameters()
		if len(params) == 0 {
			// Directive without parameters
			if directive.GetBlock() != nil {
				blockMap := make(map[string]interface{})
				cc.processBlock(directive.GetBlock(), blockMap)
				result[name] = blockMap
			} else {
				result[name] = true
			}
		} else if len(params) == 1 {
			// Single parameter
			value := params[0].GetValue()
			if directive.GetBlock() != nil {
				blockMap := map[string]interface{}{
					"value": value,
					"block": make(map[string]interface{}),
				}
				cc.processBlock(directive.GetBlock(), blockMap["block"].(map[string]interface{}))
				result[name] = blockMap
			} else {
				// Check if this directive can have multiple values
				if existing, exists := result[name]; exists {
					switch v := existing.(type) {
					case []interface{}:
						result[name] = append(v, value)
					case string:
						result[name] = []interface{}{v, value}
					default:
						result[name] = []interface{}{existing, value}
					}
				} else {
					result[name] = value
				}
			}
		} else {
			// Multiple parameters
			values := make([]string, len(params))
			for i, param := range params {
				values[i] = param.GetValue()
			}

			if directive.GetBlock() != nil {
				blockMap := map[string]interface{}{
					"parameters": values,
					"block":      make(map[string]interface{}),
				}
				cc.processBlock(directive.GetBlock(), blockMap["block"].(map[string]interface{}))
				result[name] = blockMap
			} else {
				if existing, exists := result[name]; exists {
					switch v := existing.(type) {
					case []interface{}:
						result[name] = append(v, values)
					default:
						result[name] = []interface{}{existing, values}
					}
				} else {
					result[name] = values
				}
			}
		}
	}
}

// processHTTPBlock processes an HTTP block
func (cc *ConfigConverter) processHTTPBlock(http *config.HTTP) map[string]interface{} {
	result := make(map[string]interface{})

	// Process HTTP directives
	for _, directive := range http.Directives {
		cc.processDirective(directive, result)
	}

	// Process servers
	if len(http.Servers) > 0 {
		servers := make([]interface{}, len(http.Servers))
		for i, server := range http.Servers {
			servers[i] = cc.processServerBlock(server)
		}
		result["servers"] = servers
	}

	return result
}

// processServerBlock processes a server block
func (cc *ConfigConverter) processServerBlock(server *config.Server) map[string]interface{} {
	result := make(map[string]interface{})

	if server.Block != nil {
		cc.processBlock(server.Block, result)
	}

	return result
}

// processLocationBlock processes a location block
func (cc *ConfigConverter) processLocationBlock(location *config.Location) map[string]interface{} {
	result := make(map[string]interface{})

	// Add location pattern
	params := location.GetParameters()
	if len(params) > 0 {
		if len(params) == 1 {
			result["pattern"] = params[0].GetValue()
		} else {
			// Handle modifier + pattern
			result["modifier"] = params[0].GetValue()
			result["pattern"] = params[1].GetValue()
		}
	}

	// Process location directives
	if location.Block != nil {
		cc.processBlock(location.Block, result)
	}

	return result
}

// processUpstreamBlock processes an upstream block
func (cc *ConfigConverter) processUpstreamBlock(upstream *config.Upstream) map[string]interface{} {
	result := map[string]interface{}{
		"name": upstream.UpstreamName,
	}

	// Process upstream servers
	if len(upstream.UpstreamServers) > 0 {
		servers := make([]interface{}, len(upstream.UpstreamServers))
		for i, server := range upstream.UpstreamServers {
			servers[i] = cc.processUpstreamServer(server)
		}
		result["servers"] = servers
	}

	// Process other upstream directives
	for _, directive := range upstream.Directives {
		cc.processDirective(directive, result)
	}

	return result
}

// processUpstreamServer processes an upstream server directive
func (cc *ConfigConverter) processUpstreamServer(server *config.UpstreamServer) map[string]interface{} {
	result := map[string]interface{}{
		"address": server.Address,
	}

	// Add server parameters
	params := server.GetParameters()
	if len(params) > 1 { // First parameter is address
		options := make([]string, len(params)-1)
		for i := 1; i < len(params); i++ {
			options[i-1] = params[i].GetValue()
		}
		if len(options) > 0 {
			result["options"] = options
		}
	}

	return result
}

// mapToConfig converts a map representation back to config.Config
func mapToConfig(configMap map[string]interface{}) (*config.Config, error) {
	conf := &config.Config{
		Block: &config.Block{
			Directives: make([]config.IDirective, 0),
		},
	}

	for key, value := range configMap {
		directive, err := mapToDirective(key, value)
		if err != nil {
			return nil, fmt.Errorf("failed to convert %s: %w", key, err)
		}

		if directive != nil {
			conf.Block.AddDirective(directive)
		}
	}

	return conf, nil
}

// mapToDirective converts a map entry to a directive
func mapToDirective(name string, value interface{}) (config.IDirective, error) {
	switch name {
	case "http":
		if httpMap, ok := value.(map[string]interface{}); ok {
			return mapToHTTP(httpMap)
		}
	case "servers":
		// Handled by parent HTTP block
		return nil, nil
	case "upstreams":
		// Handled by parent HTTP block
		return nil, nil
	default:
		return mapToGenericDirective(name, value)
	}

	return nil, fmt.Errorf("unsupported directive type: %s", name)
}

// mapToHTTP converts a map to an HTTP block
func mapToHTTP(httpMap map[string]interface{}) (*config.HTTP, error) {
	http := &config.HTTP{
		Directives: make([]config.IDirective, 0),
		Servers:    make([]*config.Server, 0),
	}

	for key, value := range httpMap {
		switch key {
		case "servers":
			if servers, ok := value.([]interface{}); ok {
				for _, serverValue := range servers {
					if serverMap, ok := serverValue.(map[string]interface{}); ok {
						server, err := mapToServer(serverMap)
						if err != nil {
							return nil, err
						}
						http.Servers = append(http.Servers, server)
					}
				}
			}
		case "upstreams":
			if upstreams, ok := value.([]interface{}); ok {
				for _, upstreamValue := range upstreams {
					if upstreamMap, ok := upstreamValue.(map[string]interface{}); ok {
						upstream, err := mapToUpstream(upstreamMap)
						if err != nil {
							return nil, err
						}
						http.Directives = append(http.Directives, upstream)
					}
				}
			}
		default:
			directive, err := mapToGenericDirective(key, value)
			if err != nil {
				return nil, err
			}
			if directive != nil {
				http.Directives = append(http.Directives, directive)
			}
		}
	}

	return http, nil
}

// mapToServer converts a map to a server block
func mapToServer(serverMap map[string]interface{}) (*config.Server, error) {
	server := &config.Server{
		Block: &config.Block{
			Directives: make([]config.IDirective, 0),
		},
	}

	for key, value := range serverMap {
		directive, err := mapToGenericDirective(key, value)
		if err != nil {
			return nil, err
		}
		if directive != nil {
			server.Block.(*config.Block).Directives = append(server.Block.(*config.Block).Directives, directive)
		}
	}

	return server, nil
}

// mapToUpstream converts a map to an upstream block
func mapToUpstream(upstreamMap map[string]interface{}) (*config.Upstream, error) {
	upstream := &config.Upstream{
		Directives:      make([]config.IDirective, 0),
		UpstreamServers: make([]*config.UpstreamServer, 0),
	}

	for key, value := range upstreamMap {
		switch key {
		case "name":
			if name, ok := value.(string); ok {
				upstream.UpstreamName = name
			}
		case "servers":
			if servers, ok := value.([]interface{}); ok {
				for _, serverValue := range servers {
					if serverMap, ok := serverValue.(map[string]interface{}); ok {
						server, err := mapToUpstreamServer(serverMap)
						if err != nil {
							return nil, err
						}
						upstream.UpstreamServers = append(upstream.UpstreamServers, server)
					}
				}
			}
		default:
			directive, err := mapToGenericDirective(key, value)
			if err != nil {
				return nil, err
			}
			if directive != nil {
				upstream.Directives = append(upstream.Directives, directive)
			}
		}
	}

	return upstream, nil
}

// mapToUpstreamServer converts a map to an upstream server
func mapToUpstreamServer(serverMap map[string]interface{}) (*config.UpstreamServer, error) {
	server := &config.UpstreamServer{}

	for key, value := range serverMap {
		switch key {
		case "address":
			if address, ok := value.(string); ok {
				server.Address = address
			}
		case "options":
			// Handle server options
			if options, ok := value.([]interface{}); ok {
				params := []config.Parameter{
					{Value: server.Address},
				}
				for _, option := range options {
					if optionStr, ok := option.(string); ok {
						params = append(params, config.Parameter{Value: optionStr})
					}
				}
				// Set parameters (this is a simplified approach)
			}
		}
	}

	return server, nil
}

// mapToGenericDirective converts a map entry to a generic directive
func mapToGenericDirective(name string, value interface{}) (config.IDirective, error) {
	directive := &config.Directive{
		Name:       name,
		Parameters: make([]config.Parameter, 0),
	}

	switch v := value.(type) {
	case string:
		directive.Parameters = append(directive.Parameters, config.Parameter{Value: v})
	case bool:
		if v {
			directive.Parameters = append(directive.Parameters, config.Parameter{Value: "on"})
		} else {
			directive.Parameters = append(directive.Parameters, config.Parameter{Value: "off"})
		}
	case []interface{}:
		for _, item := range v {
			if str, ok := item.(string); ok {
				directive.Parameters = append(directive.Parameters, config.Parameter{Value: str})
			} else if arr, ok := item.([]interface{}); ok {
				// Handle arrays of arrays (multiple directive instances)
				newDirective := &config.Directive{
					Name:       name,
					Parameters: make([]config.Parameter, 0),
				}
				for _, subItem := range arr {
					if str, ok := subItem.(string); ok {
						newDirective.Parameters = append(newDirective.Parameters, config.Parameter{Value: str})
					}
				}
				return newDirective, nil
			}
		}
	case map[string]interface{}:
		// Handle directives with blocks
		if valueStr, hasValue := v["value"]; hasValue {
			if str, ok := valueStr.(string); ok {
				directive.Parameters = append(directive.Parameters, config.Parameter{Value: str})
			}
		}
		if parameters, hasParams := v["parameters"]; hasParams {
			if params, ok := parameters.([]interface{}); ok {
				for _, param := range params {
					if str, ok := param.(string); ok {
						directive.Parameters = append(directive.Parameters, config.Parameter{Value: str})
					}
				}
			}
		}
		if block, hasBlock := v["block"]; hasBlock {
			if blockMap, ok := block.(map[string]interface{}); ok {
				blockDirectives := make([]config.IDirective, 0)
				for blockKey, blockValue := range blockMap {
					blockDirective, err := mapToGenericDirective(blockKey, blockValue)
					if err != nil {
						return nil, err
					}
					if blockDirective != nil {
						blockDirectives = append(blockDirectives, blockDirective)
					}
				}
				directive.Block = &config.Block{Directives: blockDirectives}
			}
		}
	default:
		return nil, fmt.Errorf("unsupported value type for directive %s: %T", name, value)
	}

	return directive, nil
}

// FormatConverter provides high-level conversion functions
type FormatConverter struct{}

// NewFormatConverter creates a new format converter
func NewFormatConverter() *FormatConverter {
	return &FormatConverter{}
}

// Convert converts between different configuration formats
func (fc *FormatConverter) Convert(input string, fromFormat, toFormat ConfigFormat) (string, error) {
	var conf *config.Config
	var err error

	// Parse input based on source format
	switch fromFormat {
	case FormatJSON:
		conf, err = ConvertFromJSON(input)
	case FormatYAML:
		conf, err = ConvertFromYAML(input)
	case FormatNginx:
		return "", fmt.Errorf("nginx to other format conversion not implemented")
	default:
		return "", fmt.Errorf("unsupported source format: %s", fromFormat)
	}

	if err != nil {
		return "", err
	}

	// Convert to target format
	converter := NewConfigConverter(conf)
	switch toFormat {
	case FormatJSON:
		return converter.ConvertToJSON(true)
	case FormatYAML:
		return converter.ConvertToYAML()
	case FormatNginx:
		return "", fmt.Errorf("conversion to nginx format not implemented")
	default:
		return "", fmt.Errorf("unsupported target format: %s", toFormat)
	}
}
