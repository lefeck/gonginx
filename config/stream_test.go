package config

import (
	"testing"
)

func TestStream(t *testing.T) {
	// Test stream block creation
	directive := &Directive{
		Name:  "stream",
		Block: &Block{},
	}

	stream, err := NewStream(directive)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if stream.GetName() != "stream" {
		t.Errorf("Expected name 'stream', got %s", stream.GetName())
	}

	if stream.GetBlock() == nil {
		t.Error("Expected block to not be nil")
	}
}

func TestStreamUpstream(t *testing.T) {
	// Test stream upstream creation
	directive := &Directive{
		Name:       "upstream",
		Parameters: []Parameter{{Value: "backend"}},
		Block:      &Block{},
	}

	upstream, err := NewStreamUpstream(directive)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if upstream.GetName() != "upstream" {
		t.Errorf("Expected name 'upstream', got %s", upstream.GetName())
	}

	if upstream.UpstreamName != "backend" {
		t.Errorf("Expected upstream name 'backend', got %s", upstream.UpstreamName)
	}

	// Test adding server
	serverDirective := &Directive{
		Name:       "server",
		Parameters: []Parameter{{Value: "192.168.1.1:8080"}, {Value: "weight=3"}},
	}

	server, err := NewStreamUpstreamServer(serverDirective)
	if err != nil {
		t.Errorf("Expected no error creating stream upstream server, got %v", err)
	}

	upstream.AddServer(server)
	if len(upstream.Servers) != 1 {
		t.Errorf("Expected 1 server, got %d", len(upstream.Servers))
	}

	if upstream.Servers[0].Address != "192.168.1.1:8080" {
		t.Errorf("Expected address '192.168.1.1:8080', got %s", upstream.Servers[0].Address)
	}

	if upstream.Servers[0].Parameters["weight"] != "3" {
		t.Errorf("Expected weight '3', got %s", upstream.Servers[0].Parameters["weight"])
	}
}

func TestStreamServer(t *testing.T) {
	// Test stream server creation
	directive := &Directive{
		Name:  "server",
		Block: &Block{},
	}

	server, err := NewStreamServer(directive)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if server.GetName() != "server" {
		t.Errorf("Expected name 'server', got %s", server.GetName())
	}

	// Test setting listen
	server.SetListen("8080")
	listenPorts := server.GetListenPorts()
	if len(listenPorts) != 1 || listenPorts[0] != "8080" {
		t.Errorf("Expected listen port '8080', got %v", listenPorts)
	}

	// Test setting proxy_pass
	server.SetProxyPass("backend")
	proxyPass := server.GetProxyPass()
	if proxyPass != "backend" {
		t.Errorf("Expected proxy_pass 'backend', got %s", proxyPass)
	}
}

func TestStreamUpstreamServer(t *testing.T) {
	// Test stream upstream server with various parameters
	directive := &Directive{
		Name: "server",
		Parameters: []Parameter{
			{Value: "192.168.1.1:8080"},
			{Value: "weight=5"},
			{Value: "max_fails=3"},
			{Value: "fail_timeout=30s"},
			{Value: "down"},
		},
	}

	server, err := NewStreamUpstreamServer(directive)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if server.Address != "192.168.1.1:8080" {
		t.Errorf("Expected address '192.168.1.1:8080', got %s", server.Address)
	}

	if server.GetWeight() != "5" {
		t.Errorf("Expected weight '5', got %s", server.GetWeight())
	}

	if server.GetMaxFails() != "3" {
		t.Errorf("Expected max_fails '3', got %s", server.GetMaxFails())
	}

	if server.GetFailTimeout() != "30s" {
		t.Errorf("Expected fail_timeout '30s', got %s", server.GetFailTimeout())
	}

	if !server.IsDown() {
		t.Error("Expected server to be marked as down")
	}

	// Test setting backup
	server.SetBackup(true)
	if !server.IsBackup() {
		t.Error("Expected server to be marked as backup")
	}

	// Test removing down status
	server.SetDown(false)
	if server.IsDown() {
		t.Error("Expected server to not be marked as down")
	}
}

func TestConfigStreamMethods(t *testing.T) {
	// Create a config with stream block
	config := &Config{
		Block: &Block{
			Directives: []IDirective{},
		},
	}

	// Create a stream block
	streamDirective := &Directive{
		Name:  "stream",
		Block: &Block{},
	}
	stream, _ := NewStream(streamDirective)
	config.Block.AddDirective(stream)

	// Test finding streams
	streams := config.FindStreams()
	if len(streams) != 1 {
		t.Errorf("Expected 1 stream, got %d", len(streams))
	}

	// Add upstream to stream
	upstreamDirective := &Directive{
		Name:       "upstream",
		Parameters: []Parameter{{Value: "backend"}},
		Block:      &Block{},
	}
	upstream, _ := NewStreamUpstream(upstreamDirective)
	stream.AddUpstream(upstream)

	// Test finding stream upstreams
	upstreams := config.FindStreamUpstreams()
	if len(upstreams) != 1 {
		t.Errorf("Expected 1 stream upstream, got %d", len(upstreams))
	}

	// Test finding upstream by name
	foundUpstream := config.FindStreamUpstreamByName("backend")
	if foundUpstream == nil {
		t.Error("Expected to find upstream 'backend'")
	}

	if foundUpstream.UpstreamName != "backend" {
		t.Errorf("Expected upstream name 'backend', got %s", foundUpstream.UpstreamName)
	}
}
