package generator

import (
	"github.com/lefeck/gonginx/config"
)

// StreamBuilder provides methods for building Stream block
type StreamBuilder struct {
	config      *config.Config
	streamBlock *config.Stream
}

// AddDirective adds a directive to the Stream block
func (sb *StreamBuilder) AddDirective(name string, parameters ...string) *StreamBuilder {
	var params []config.Parameter
	for _, param := range parameters {
		params = append(params, config.NewParameter(param))
	}

	directive := &config.Directive{
		Name:       name,
		Parameters: params,
	}

	sb.streamBlock.Block.AddDirective(directive)
	return sb
}

// ErrorLog sets error_log for stream
func (sb *StreamBuilder) ErrorLog(path string, level ...string) *StreamBuilder {
	if len(level) > 0 {
		return sb.AddDirective("error_log", path, level[0])
	}
	return sb.AddDirective("error_log", path)
}

// AccessLog sets access_log for stream
func (sb *StreamBuilder) AccessLog(path string, format ...string) *StreamBuilder {
	if len(format) > 0 {
		return sb.AddDirective("access_log", path, format[0])
	}
	return sb.AddDirective("access_log", path)
}

// Upstream creates and returns a stream upstream block builder
func (sb *StreamBuilder) Upstream(name string) *StreamUpstreamBuilder {
	upstreamDirective := &config.Directive{
		Name:       "upstream",
		Parameters: []config.Parameter{config.NewParameter(name)},
		Block:      &config.Block{Directives: []config.IDirective{}},
	}

	upstreamBlock, _ := config.NewStreamUpstream(upstreamDirective)
	sb.streamBlock.Block.AddDirective(upstreamBlock)

	return &StreamUpstreamBuilder{
		config:        sb.config,
		streamBuilder: sb,
		upstreamBlock: upstreamBlock,
	}
}

// Server creates and returns a stream server block builder
func (sb *StreamBuilder) Server() *StreamServerBuilder {
	serverDirective := &config.Directive{
		Name:  "server",
		Block: &config.Block{Directives: []config.IDirective{}},
	}

	serverBlock, _ := config.NewStreamServer(serverDirective)
	sb.streamBlock.Block.AddDirective(serverBlock)

	return &StreamServerBuilder{
		config:        sb.config,
		streamBuilder: sb,
		serverBlock:   serverBlock,
	}
}

// End returns to the main config builder
func (sb *StreamBuilder) End() *ConfigBuilder {
	return &ConfigBuilder{config: sb.config}
}

// StreamUpstreamBuilder provides methods for building stream upstream block
type StreamUpstreamBuilder struct {
	config        *config.Config
	streamBuilder *StreamBuilder
	upstreamBlock *config.StreamUpstream
}

// AddDirective adds a directive to the stream upstream block
func (sub *StreamUpstreamBuilder) AddDirective(name string, parameters ...string) *StreamUpstreamBuilder {
	var params []config.Parameter
	for _, param := range parameters {
		params = append(params, config.NewParameter(param))
	}

	directive := &config.Directive{
		Name:       name,
		Parameters: params,
	}

	sub.upstreamBlock.Block.AddDirective(directive)
	return sub
}

// Server adds a server to the stream upstream
func (sub *StreamUpstreamBuilder) Server(address string, options ...string) *StreamUpstreamBuilder {
	params := []string{address}
	params = append(params, options...)
	return sub.AddDirective("server", params...)
}

// LeastConn enables least_conn for stream upstream
func (sub *StreamUpstreamBuilder) LeastConn() *StreamUpstreamBuilder {
	return sub.AddDirective("least_conn")
}

// Hash enables hash load balancing
func (sub *StreamUpstreamBuilder) Hash(key string, consistent ...bool) *StreamUpstreamBuilder {
	if len(consistent) > 0 && consistent[0] {
		return sub.AddDirective("hash", key, "consistent")
	}
	return sub.AddDirective("hash", key)
}

// EndUpstream returns to the stream builder
func (sub *StreamUpstreamBuilder) EndUpstream() *StreamBuilder {
	return sub.streamBuilder
}

// End returns to the main config builder
func (sub *StreamUpstreamBuilder) End() *ConfigBuilder {
	return &ConfigBuilder{config: sub.config}
}

// StreamServerBuilder provides methods for building stream server block
type StreamServerBuilder struct {
	config        *config.Config
	streamBuilder *StreamBuilder
	serverBlock   *config.StreamServer
}

// AddDirective adds a directive to the stream server block
func (ssb *StreamServerBuilder) AddDirective(name string, parameters ...string) *StreamServerBuilder {
	var params []config.Parameter
	for _, param := range parameters {
		params = append(params, config.NewParameter(param))
	}

	directive := &config.Directive{
		Name:       name,
		Parameters: params,
	}

	ssb.serverBlock.Block.AddDirective(directive)
	return ssb
}

// Listen adds a listen directive
func (ssb *StreamServerBuilder) Listen(port string, options ...string) *StreamServerBuilder {
	params := []string{port}
	params = append(params, options...)
	return ssb.AddDirective("listen", params...)
}

// ProxyPass sets proxy_pass for stream server
func (ssb *StreamServerBuilder) ProxyPass(upstream string) *StreamServerBuilder {
	return ssb.AddDirective("proxy_pass", upstream)
}

// ProxyTimeout sets proxy timeout
func (ssb *StreamServerBuilder) ProxyTimeout(timeout string) *StreamServerBuilder {
	return ssb.AddDirective("proxy_timeout", timeout)
}

// ProxyConnectTimeout sets proxy connect timeout
func (ssb *StreamServerBuilder) ProxyConnectTimeout(timeout string) *StreamServerBuilder {
	return ssb.AddDirective("proxy_connect_timeout", timeout)
}

// ProxyBind sets proxy bind address
func (ssb *StreamServerBuilder) ProxyBind(address string) *StreamServerBuilder {
	return ssb.AddDirective("proxy_bind", address)
}

// SSL configures SSL for stream server
func (ssb *StreamServerBuilder) SSL() *StreamSSLBuilder {
	return &StreamSSLBuilder{
		config:              ssb.config,
		streamBuilder:       ssb.streamBuilder,
		streamServerBuilder: ssb,
	}
}

// EndServer returns to the stream builder
func (ssb *StreamServerBuilder) EndServer() *StreamBuilder {
	return ssb.streamBuilder
}

// End returns to the main config builder
func (ssb *StreamServerBuilder) End() *ConfigBuilder {
	return &ConfigBuilder{config: ssb.config}
}

// StreamSSLBuilder provides methods for stream SSL configuration
type StreamSSLBuilder struct {
	config              *config.Config
	streamBuilder       *StreamBuilder
	streamServerBuilder *StreamServerBuilder
}

// Certificate sets SSL certificate for stream
func (ssl *StreamSSLBuilder) Certificate(path string) *StreamSSLBuilder {
	ssl.streamServerBuilder.AddDirective("ssl_certificate", path)
	return ssl
}

// CertificateKey sets SSL certificate key for stream
func (ssl *StreamSSLBuilder) CertificateKey(path string) *StreamSSLBuilder {
	ssl.streamServerBuilder.AddDirective("ssl_certificate_key", path)
	return ssl
}

// Protocols sets SSL protocols for stream
func (ssl *StreamSSLBuilder) Protocols(protocols ...string) *StreamSSLBuilder {
	ssl.streamServerBuilder.AddDirective("ssl_protocols", protocols...)
	return ssl
}

// Ciphers sets SSL ciphers for stream
func (ssl *StreamSSLBuilder) Ciphers(ciphers string) *StreamSSLBuilder {
	ssl.streamServerBuilder.AddDirective("ssl_ciphers", ciphers)
	return ssl
}

// PreferServerCiphers enables ssl_prefer_server_ciphers for stream
func (ssl *StreamSSLBuilder) PreferServerCiphers(enabled bool) *StreamSSLBuilder {
	value := "off"
	if enabled {
		value = "on"
	}
	ssl.streamServerBuilder.AddDirective("ssl_prefer_server_ciphers", value)
	return ssl
}

// SessionTimeout sets SSL session timeout for stream
func (ssl *StreamSSLBuilder) SessionTimeout(timeout string) *StreamSSLBuilder {
	ssl.streamServerBuilder.AddDirective("ssl_session_timeout", timeout)
	return ssl
}

// Preread enables SSL preread
func (ssl *StreamSSLBuilder) Preread(enabled bool) *StreamSSLBuilder {
	value := "off"
	if enabled {
		value = "on"
	}
	ssl.streamServerBuilder.AddDirective("ssl_preread", value)
	return ssl
}

// EndSSL returns to the stream server builder
func (ssl *StreamSSLBuilder) EndSSL() *StreamServerBuilder {
	return ssl.streamServerBuilder
}
