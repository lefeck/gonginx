package config

import (
	"fmt"
)

// Stream represents a stream block in nginx configuration
// Stream module is used for TCP/UDP load balancing and proxying
type Stream struct {
	*Block
	Comment []string
	DefaultInlineComment
	Parent IDirective
	Line   int
}

// NewStream creates a new Stream from a directive
func NewStream(directive *Directive) (*Stream, error) {
	if directive.GetName() != "stream" {
		return nil, fmt.Errorf("not a stream directive")
	}

	block, ok := directive.Block.(*Block)
	if !ok {
		return nil, fmt.Errorf("directive block is not a *Block")
	}

	stream := &Stream{
		Block:   block,
		Comment: directive.Comment,
		Parent:  directive.Parent,
		Line:    directive.Line,
	}

	return stream, nil
}

// GetName returns the name of the directive
func (s *Stream) GetName() string {
	return "stream"
}

// GetParameters returns the parameters of the directive
func (s *Stream) GetParameters() []Parameter {
	return []Parameter{}
}

// GetBlock returns the block of the directive
func (s *Stream) GetBlock() IBlock {
	return s.Block
}

// GetComment returns the comment of the directive
func (s *Stream) GetComment() []string {
	return s.Comment
}

// SetComment sets the comment of the directive
func (s *Stream) SetComment(comment []string) {
	s.Comment = comment
}

// SetLine sets the line number
func (s *Stream) SetLine(line int) {
	s.Line = line
}

// GetLine returns the line number
func (s *Stream) GetLine() int {
	return s.Line
}

// SetParent sets the parent directive
func (s *Stream) SetParent(parent IDirective) {
	s.Parent = parent
}

// GetParent returns the parent directive
func (s *Stream) GetParent() IDirective {
	return s.Parent
}

// FindUpstreams finds all upstream blocks within this stream block
func (s *Stream) FindUpstreams() []*StreamUpstream {
	var upstreams []*StreamUpstream
	directives := s.Block.FindDirectives("upstream")
	for _, directive := range directives {
		if upstream, ok := directive.(*StreamUpstream); ok {
			upstreams = append(upstreams, upstream)
		}
	}
	return upstreams
}

// FindUpstreamByName finds an upstream by name within this stream block
func (s *Stream) FindUpstreamByName(name string) *StreamUpstream {
	upstreams := s.FindUpstreams()
	for _, upstream := range upstreams {
		if upstream.UpstreamName == name {
			return upstream
		}
	}
	return nil
}

// FindServers finds all server blocks within this stream block
func (s *Stream) FindServers() []*StreamServer {
	var servers []*StreamServer
	directives := s.Block.FindDirectives("server")
	for _, directive := range directives {
		if server, ok := directive.(*StreamServer); ok {
			servers = append(servers, server)
		}
	}
	return servers
}

// FindServersByListen finds servers by their listen directive
func (s *Stream) FindServersByListen(listen string) []*StreamServer {
	var servers []*StreamServer
	allServers := s.FindServers()

	for _, server := range allServers {
		listenDirectives := server.FindDirectives("listen")
		for _, listenDir := range listenDirectives {
			params := listenDir.GetParameters()
			for _, param := range params {
				if param.GetValue() == listen {
					servers = append(servers, server)
					break
				}
			}
		}
	}
	return servers
}

// AddUpstream adds a new upstream block to the stream
func (s *Stream) AddUpstream(upstream *StreamUpstream) {
	s.Block.AddDirective(upstream)
}

// AddServer adds a new server block to the stream
func (s *Stream) AddServer(server *StreamServer) {
	s.Block.AddDirective(server)
}
