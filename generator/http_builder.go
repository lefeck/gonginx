package generator

import (
	"github.com/lefeck/gonginx/config"
)

// HTTPBuilder provides methods for building HTTP block
type HTTPBuilder struct {
	config    *config.Config
	httpBlock *config.HTTP
}

// AddDirective adds a directive to the HTTP block
func (hb *HTTPBuilder) AddDirective(name string, parameters ...string) *HTTPBuilder {
	var params []config.Parameter
	for _, param := range parameters {
		params = append(params, config.NewParameter(param))
	}

	directive := &config.Directive{
		Name:       name,
		Parameters: params,
	}

	hb.httpBlock.Directives = append(hb.httpBlock.Directives, directive)
	return hb
}

// SendFile enables/disables sendfile
func (hb *HTTPBuilder) SendFile(enabled bool) *HTTPBuilder {
	value := "off"
	if enabled {
		value = "on"
	}
	return hb.AddDirective("sendfile", value)
}

// TcpNoPush enables/disables tcp_nopush
func (hb *HTTPBuilder) TcpNoPush(enabled bool) *HTTPBuilder {
	value := "off"
	if enabled {
		value = "on"
	}
	return hb.AddDirective("tcp_nopush", value)
}

// TcpNoDelay enables/disables tcp_nodelay
func (hb *HTTPBuilder) TcpNoDelay(enabled bool) *HTTPBuilder {
	value := "off"
	if enabled {
		value = "on"
	}
	return hb.AddDirective("tcp_nodelay", value)
}

// KeepaliveTimeout sets keepalive_timeout
func (hb *HTTPBuilder) KeepaliveTimeout(timeout string) *HTTPBuilder {
	return hb.AddDirective("keepalive_timeout", timeout)
}

// ClientMaxBodySize sets client_max_body_size
func (hb *HTTPBuilder) ClientMaxBodySize(size string) *HTTPBuilder {
	return hb.AddDirective("client_max_body_size", size)
}

// Gzip enables/disables gzip compression
func (hb *HTTPBuilder) Gzip(enabled bool) *HTTPBuilder {
	value := "off"
	if enabled {
		value = "on"
	}
	return hb.AddDirective("gzip", value)
}

// GzipTypes sets gzip_types
func (hb *HTTPBuilder) GzipTypes(types ...string) *HTTPBuilder {
	return hb.AddDirective("gzip_types", types...)
}

// AccessLog sets access_log
func (hb *HTTPBuilder) AccessLog(path string, format ...string) *HTTPBuilder {
	if len(format) > 0 {
		return hb.AddDirective("access_log", path, format[0])
	}
	return hb.AddDirective("access_log", path)
}

// ErrorLog sets error_log
func (hb *HTTPBuilder) ErrorLog(path string, level ...string) *HTTPBuilder {
	if len(level) > 0 {
		return hb.AddDirective("error_log", path, level[0])
	}
	return hb.AddDirective("error_log", path)
}

// Include adds include directive
func (hb *HTTPBuilder) Include(path string) *HTTPBuilder {
	return hb.AddDirective("include", path)
}

// Server creates and returns a server block builder
func (hb *HTTPBuilder) Server() *ServerBuilder {
	serverDirective := &config.Directive{
		Name:  "server",
		Block: &config.Block{Directives: []config.IDirective{}},
	}

	serverBlock, _ := config.NewServer(serverDirective)
	hb.httpBlock.Directives = append(hb.httpBlock.Directives, serverBlock)

	return &ServerBuilder{
		config:      hb.config,
		httpBuilder: hb,
		serverBlock: serverBlock,
	}
}

// Upstream creates and returns an upstream block builder
func (hb *HTTPBuilder) Upstream(name string) *UpstreamBuilder {
	upstreamDirective := &config.Directive{
		Name:       "upstream",
		Parameters: []config.Parameter{config.NewParameter(name)},
		Block:      &config.Block{Directives: []config.IDirective{}},
	}

	upstreamBlock, _ := config.NewUpstream(upstreamDirective)
	hb.httpBlock.Directives = append(hb.httpBlock.Directives, upstreamBlock)

	return &UpstreamBuilder{
		config:        hb.config,
		httpBuilder:   hb,
		upstreamBlock: upstreamBlock,
	}
}

// End returns to the main config builder
func (hb *HTTPBuilder) End() *ConfigBuilder {
	return &ConfigBuilder{config: hb.config}
}

// ServerBuilder provides methods for building server block
type ServerBuilder struct {
	config      *config.Config
	httpBuilder *HTTPBuilder
	serverBlock *config.Server
}

// AddDirective adds a directive to the server block
func (sb *ServerBuilder) AddDirective(name string, parameters ...string) *ServerBuilder {
	var params []config.Parameter
	for _, param := range parameters {
		params = append(params, config.NewParameter(param))
	}

	directive := &config.Directive{
		Name:       name,
		Parameters: params,
	}

	sb.serverBlock.Block.(*config.Block).AddDirective(directive)
	return sb
}

// Listen adds a listen directive
func (sb *ServerBuilder) Listen(port string, options ...string) *ServerBuilder {
	params := []string{port}
	params = append(params, options...)
	return sb.AddDirective("listen", params...)
}

// ServerName sets server_name
func (sb *ServerBuilder) ServerName(names ...string) *ServerBuilder {
	return sb.AddDirective("server_name", names...)
}

// Root sets document root
func (sb *ServerBuilder) Root(path string) *ServerBuilder {
	return sb.AddDirective("root", path)
}

// Index sets index files
func (sb *ServerBuilder) Index(files ...string) *ServerBuilder {
	return sb.AddDirective("index", files...)
}

// AccessLog sets access_log for this server
func (sb *ServerBuilder) AccessLog(path string, format ...string) *ServerBuilder {
	if len(format) > 0 {
		return sb.AddDirective("access_log", path, format[0])
	}
	return sb.AddDirective("access_log", path)
}

// ErrorLog sets error_log for this server
func (sb *ServerBuilder) ErrorLog(path string, level ...string) *ServerBuilder {
	if len(level) > 0 {
		return sb.AddDirective("error_log", path, level[0])
	}
	return sb.AddDirective("error_log", path)
}

// SSL enables SSL configuration
func (sb *ServerBuilder) SSL() *SSLBuilder {
	sb.AddDirective("listen", "443", "ssl")
	return &SSLBuilder{
		config:        sb.config,
		httpBuilder:   sb.httpBuilder,
		serverBuilder: sb,
	}
}

// Location creates and returns a location block builder
func (sb *ServerBuilder) Location(pattern string, modifier ...string) *LocationBuilder {
	params := []config.Parameter{}
	if len(modifier) > 0 {
		params = append(params, config.NewParameter(modifier[0]))
	}
	params = append(params, config.NewParameter(pattern))

	locationDirective := &config.Directive{
		Name:       "location",
		Parameters: params,
		Block:      &config.Block{Directives: []config.IDirective{}},
	}

	locationBlock, _ := config.NewLocation(locationDirective)
	sb.serverBlock.Block.(*config.Block).AddDirective(locationBlock)

	return &LocationBuilder{
		config:        sb.config,
		httpBuilder:   sb.httpBuilder,
		serverBuilder: sb,
		locationBlock: locationBlock,
	}
}

// ProxyPass sets up reverse proxy
func (sb *ServerBuilder) ProxyPass(upstream string) *ServerBuilder {
	return sb.AddDirective("proxy_pass", upstream)
}

// Return adds a return directive
func (sb *ServerBuilder) Return(code string, url ...string) *ServerBuilder {
	if len(url) > 0 {
		return sb.AddDirective("return", code, url[0])
	}
	return sb.AddDirective("return", code)
}

// EndServer returns to the HTTP builder
func (sb *ServerBuilder) EndServer() *HTTPBuilder {
	return sb.httpBuilder
}

// End returns to the main config builder
func (sb *ServerBuilder) End() *ConfigBuilder {
	return &ConfigBuilder{config: sb.config}
}
