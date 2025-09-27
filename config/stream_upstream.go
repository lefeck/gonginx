package config

import (
	"fmt"
	"strings"
)

// StreamUpstream represents an upstream block within a stream context
type StreamUpstream struct {
	*Block
	Comment      []string
	UpstreamName string
	Servers      []*StreamUpstreamServer
	DefaultInlineComment
	Parent IDirective
	Line   int
}

// NewStreamUpstream creates a new StreamUpstream from a directive
func NewStreamUpstream(directive *Directive) (*StreamUpstream, error) {
	if directive.GetName() != "upstream" {
		return nil, fmt.Errorf("not an upstream directive")
	}

	upstreamName := ""
	if len(directive.Parameters) > 0 {
		upstreamName = directive.Parameters[0].GetValue()
	}

	block, ok := directive.Block.(*Block)
	if !ok {
		return nil, fmt.Errorf("directive block is not a *Block")
	}

	upstream := &StreamUpstream{
		Block:        block,
		Comment:      directive.Comment,
		UpstreamName: upstreamName,
		Servers:      []*StreamUpstreamServer{},
		Parent:       directive.Parent,
		Line:         directive.Line,
	}

	// Parse existing server directives
	if upstream.Block != nil {
		for _, dir := range upstream.Block.GetDirectives() {
			if dir.GetName() == "server" {
				if server, ok := dir.(*StreamUpstreamServer); ok {
					upstream.Servers = append(upstream.Servers, server)
				} else {
					// Convert regular directive to StreamUpstreamServer
					if streamServer, err := NewStreamUpstreamServer(dir.(*Directive)); err == nil {
						upstream.Servers = append(upstream.Servers, streamServer)
					}
				}
			}
		}
	}

	return upstream, nil
}

// GetName returns the name of the directive
func (su *StreamUpstream) GetName() string {
	return "upstream"
}

// GetParameters returns the parameters of the directive
func (su *StreamUpstream) GetParameters() []Parameter {
	var params []Parameter
	if su.UpstreamName != "" {
		params = append(params, Parameter{Value: su.UpstreamName})
	}
	return params
}

// GetBlock returns the block of the directive
func (su *StreamUpstream) GetBlock() IBlock {
	return su.Block
}

// GetComment returns the comment of the directive
func (su *StreamUpstream) GetComment() []string {
	return su.Comment
}

// SetComment sets the comment of the directive
func (su *StreamUpstream) SetComment(comment []string) {
	su.Comment = comment
}

// SetLine sets the line number
func (su *StreamUpstream) SetLine(line int) {
	su.Line = line
}

// GetLine returns the line number
func (su *StreamUpstream) GetLine() int {
	return su.Line
}

// SetParent sets the parent directive
func (su *StreamUpstream) SetParent(parent IDirective) {
	su.Parent = parent
}

// GetParent returns the parent directive
func (su *StreamUpstream) GetParent() IDirective {
	return su.Parent
}

// AddServer adds a server to the upstream block
func (su *StreamUpstream) AddServer(server *StreamUpstreamServer) {
	su.Servers = append(su.Servers, server)
	if su.Block != nil {
		su.Block.AddDirective(server)
	}
}

// RemoveServer removes a server from the upstream block by address
func (su *StreamUpstream) RemoveServer(address string) bool {
	for i, server := range su.Servers {
		if server.Address == address {
			// Remove from slice
			su.Servers = append(su.Servers[:i], su.Servers[i+1:]...)

			// Remove from block
			if su.Block != nil {
				directives := su.Block.GetDirectives()
				for j, dir := range directives {
					if dir.GetName() == "server" {
						if streamServer, ok := dir.(*StreamUpstreamServer); ok && streamServer.Address == address {
							// Remove directive from block
							newDirectives := append(directives[:j], directives[j+1:]...)
							su.Block.SetDirectives(newDirectives)
							break
						}
					}
				}
			}
			return true
		}
	}
	return false
}

// FindServerByAddress finds a server by its address
func (su *StreamUpstream) FindServerByAddress(address string) *StreamUpstreamServer {
	for _, server := range su.Servers {
		if server.Address == address {
			return server
		}
	}
	return nil
}

// GetServerAddresses returns all server addresses in the upstream
func (su *StreamUpstream) GetServerAddresses() []string {
	var addresses []string
	for _, server := range su.Servers {
		addresses = append(addresses, server.Address)
	}
	return addresses
}

// StreamUpstreamServer represents a server directive within a stream upstream block
type StreamUpstreamServer struct {
	*Directive
	Address    string
	Parameters map[string]string // weight, max_fails, fail_timeout, etc.
}

// NewStreamUpstreamServer creates a new StreamUpstreamServer from a directive
func NewStreamUpstreamServer(directive *Directive) (*StreamUpstreamServer, error) {
	if directive.GetName() != "server" {
		return nil, fmt.Errorf("not a server directive")
	}

	params := directive.GetParameters()
	if len(params) == 0 {
		return nil, fmt.Errorf("server directive must have at least one parameter (address)")
	}

	server := &StreamUpstreamServer{
		Directive:  directive,
		Address:    params[0].GetValue(),
		Parameters: make(map[string]string),
	}

	// Parse additional parameters (weight, max_fails, etc.)
	for i := 1; i < len(params); i++ {
		param := params[i].GetValue()
		if strings.Contains(param, "=") {
			parts := strings.SplitN(param, "=", 2)
			if len(parts) == 2 {
				server.Parameters[parts[0]] = parts[1]
			}
		} else {
			// Boolean parameters like "down", "backup"
			server.Parameters[param] = "true"
		}
	}

	return server, nil
}

// GetName returns the name of the directive
func (sus *StreamUpstreamServer) GetName() string {
	return "server"
}

// GetParameters returns the parameters of the directive
func (sus *StreamUpstreamServer) GetParameters() []Parameter {
	var params []Parameter
	params = append(params, Parameter{Value: sus.Address})

	for key, value := range sus.Parameters {
		if value == "true" {
			// Boolean parameter
			params = append(params, Parameter{Value: key})
		} else {
			// Key=value parameter
			params = append(params, Parameter{Value: fmt.Sprintf("%s=%s", key, value)})
		}
	}

	return params
}

// GetBlock returns the block of the directive (should be nil for server directives)
func (sus *StreamUpstreamServer) GetBlock() IBlock {
	return nil
}

// GetComment returns the comment of the directive
func (sus *StreamUpstreamServer) GetComment() []string {
	return sus.Directive.Comment
}

// SetComment sets the comment of the directive
func (sus *StreamUpstreamServer) SetComment(comment []string) {
	sus.Directive.Comment = comment
}

// SetWeight sets the weight parameter
func (sus *StreamUpstreamServer) SetWeight(weight string) {
	sus.Parameters["weight"] = weight
}

// SetMaxFails sets the max_fails parameter
func (sus *StreamUpstreamServer) SetMaxFails(maxFails string) {
	sus.Parameters["max_fails"] = maxFails
}

// SetFailTimeout sets the fail_timeout parameter
func (sus *StreamUpstreamServer) SetFailTimeout(failTimeout string) {
	sus.Parameters["fail_timeout"] = failTimeout
}

// SetDown marks the server as down
func (sus *StreamUpstreamServer) SetDown(down bool) {
	if down {
		sus.Parameters["down"] = "true"
	} else {
		delete(sus.Parameters, "down")
	}
}

// SetBackup marks the server as backup
func (sus *StreamUpstreamServer) SetBackup(backup bool) {
	if backup {
		sus.Parameters["backup"] = "true"
	} else {
		delete(sus.Parameters, "backup")
	}
}

// IsDown returns true if the server is marked as down
func (sus *StreamUpstreamServer) IsDown() bool {
	return sus.Parameters["down"] == "true"
}

// IsBackup returns true if the server is marked as backup
func (sus *StreamUpstreamServer) IsBackup() bool {
	return sus.Parameters["backup"] == "true"
}

// GetWeight returns the weight of the server
func (sus *StreamUpstreamServer) GetWeight() string {
	return sus.Parameters["weight"]
}

// GetMaxFails returns the max_fails of the server
func (sus *StreamUpstreamServer) GetMaxFails() string {
	return sus.Parameters["max_fails"]
}

// GetFailTimeout returns the fail_timeout of the server
func (sus *StreamUpstreamServer) GetFailTimeout() string {
	return sus.Parameters["fail_timeout"]
}
