package config

import (
	"fmt"
)

// StreamServer represents a server block within a stream context
type StreamServer struct {
	*Block
	Comment []string
	DefaultInlineComment
	Parent IDirective
	Line   int
}

// NewStreamServer creates a new StreamServer from a directive
func NewStreamServer(directive *Directive) (*StreamServer, error) {
	if directive.GetName() != "server" {
		return nil, fmt.Errorf("not a server directive")
	}

	block, ok := directive.Block.(*Block)
	if !ok {
		return nil, fmt.Errorf("directive block is not a *Block")
	}

	server := &StreamServer{
		Block:   block,
		Comment: directive.Comment,
		Parent:  directive.Parent,
		Line:    directive.Line,
	}

	return server, nil
}

// GetName returns the name of the directive
func (ss *StreamServer) GetName() string {
	return "server"
}

// GetParameters returns the parameters of the directive
func (ss *StreamServer) GetParameters() []Parameter {
	return []Parameter{}
}

// GetBlock returns the block of the directive
func (ss *StreamServer) GetBlock() IBlock {
	return ss.Block
}

// GetComment returns the comment of the directive
func (ss *StreamServer) GetComment() []string {
	return ss.Comment
}

// SetComment sets the comment of the directive
func (ss *StreamServer) SetComment(comment []string) {
	ss.Comment = comment
}

// SetLine sets the line number
func (ss *StreamServer) SetLine(line int) {
	ss.Line = line
}

// GetLine returns the line number
func (ss *StreamServer) GetLine() int {
	return ss.Line
}

// SetParent sets the parent directive
func (ss *StreamServer) SetParent(parent IDirective) {
	ss.Parent = parent
}

// GetParent returns the parent directive
func (ss *StreamServer) GetParent() IDirective {
	return ss.Parent
}

// GetListenPorts returns all listen directives' parameters
func (ss *StreamServer) GetListenPorts() []string {
	var ports []string
	listenDirectives := ss.Block.FindDirectives("listen")

	for _, directive := range listenDirectives {
		params := directive.GetParameters()
		if len(params) > 0 {
			ports = append(ports, params[0].GetValue())
		}
	}

	return ports
}

// GetProxyPass returns the proxy_pass directive value
func (ss *StreamServer) GetProxyPass() string {
	proxyPassDirectives := ss.Block.FindDirectives("proxy_pass")
	if len(proxyPassDirectives) > 0 {
		params := proxyPassDirectives[0].GetParameters()
		if len(params) > 0 {
			return params[0].GetValue()
		}
	}
	return ""
}

// SetListen sets or adds a listen directive
func (ss *StreamServer) SetListen(address string) {
	// Remove existing listen directives
	ss.removeDirectivesByName("listen")

	// Add new listen directive
	listen := &Directive{
		Name:       "listen",
		Parameters: []Parameter{{Value: address}},
	}
	ss.Block.AddDirective(listen)
}

// SetProxyPass sets or adds a proxy_pass directive
func (ss *StreamServer) SetProxyPass(upstream string) {
	// Remove existing proxy_pass directives
	ss.removeDirectivesByName("proxy_pass")

	// Add new proxy_pass directive
	proxyPass := &Directive{
		Name:       "proxy_pass",
		Parameters: []Parameter{{Value: upstream}},
	}
	ss.Block.AddDirective(proxyPass)
}

// AddDirective adds a directive to the server block
func (ss *StreamServer) AddDirective(directive IDirective) {
	if ss.Block != nil {
		ss.Block.AddDirective(directive)
	}
}

// RemoveDirective removes a directive from the server block by name
func (ss *StreamServer) RemoveDirective(directiveName string) bool {
	return ss.removeDirectivesByName(directiveName)
}

// removeDirectivesByName removes all directives with the given name
func (ss *StreamServer) removeDirectivesByName(name string) bool {
	if ss.Block == nil {
		return false
	}

	directives := ss.Block.GetDirectives()
	var newDirectives []IDirective
	removed := false

	for _, directive := range directives {
		if directive.GetName() != name {
			newDirectives = append(newDirectives, directive)
		} else {
			removed = true
		}
	}

	if removed {
		ss.Block.SetDirectives(newDirectives)
	}

	return removed
}

// FindDirectives finds directives by name within this server block
func (ss *StreamServer) FindDirectives(directiveName string) []IDirective {
	if ss.Block != nil {
		return ss.Block.FindDirectives(directiveName)
	}
	return []IDirective{}
}

// GetAllDirectivesByName returns all directives with the specified name
func (ss *StreamServer) GetAllDirectivesByName(name string) []IDirective {
	return ss.FindDirectives(name)
}

// HasDirective checks if a directive exists in this server block
func (ss *StreamServer) HasDirective(directiveName string) bool {
	directives := ss.FindDirectives(directiveName)
	return len(directives) > 0
}

// GetDirectiveValue returns the first parameter value of the specified directive
func (ss *StreamServer) GetDirectiveValue(directiveName string) string {
	directives := ss.FindDirectives(directiveName)
	if len(directives) > 0 {
		params := directives[0].GetParameters()
		if len(params) > 0 {
			return params[0].GetValue()
		}
	}
	return ""
}

// SetDirective sets a directive with a single parameter value
func (ss *StreamServer) SetDirective(directiveName, value string) {
	// Remove existing directives with this name
	ss.removeDirectivesByName(directiveName)

	// Add new directive
	directive := &Directive{
		Name:       directiveName,
		Parameters: []Parameter{{Value: value}},
	}
	ss.Block.AddDirective(directive)
}

// SetDirectiveWithParams sets a directive with multiple parameter values
func (ss *StreamServer) SetDirectiveWithParams(directiveName string, values []string) {
	// Remove existing directives with this name
	ss.removeDirectivesByName(directiveName)

	// Create parameters
	var params []Parameter
	for _, value := range values {
		params = append(params, Parameter{Value: value})
	}

	// Add new directive
	directive := &Directive{
		Name:       directiveName,
		Parameters: params,
	}
	ss.Block.AddDirective(directive)
}
