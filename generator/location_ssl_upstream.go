package generator

import (
	"github.com/lefeck/gonginx/config"
)

// LocationBuilder provides methods for building location block
type LocationBuilder struct {
	config        *config.Config
	httpBuilder   *HTTPBuilder
	serverBuilder *ServerBuilder
	locationBlock *config.Location
}

// AddDirective adds a directive to the location block
func (lb *LocationBuilder) AddDirective(name string, parameters ...string) *LocationBuilder {
	var params []config.Parameter
	for _, param := range parameters {
		params = append(params, config.NewParameter(param))
	}

	directive := &config.Directive{
		Name:       name,
		Parameters: params,
	}

	lb.locationBlock.Block.(*config.Block).AddDirective(directive)
	return lb
}

// ProxyPass sets proxy_pass
func (lb *LocationBuilder) ProxyPass(upstream string) *LocationBuilder {
	return lb.AddDirective("proxy_pass", upstream)
}

// ProxySetHeader sets proxy headers
func (lb *LocationBuilder) ProxySetHeader(header, value string) *LocationBuilder {
	return lb.AddDirective("proxy_set_header", header, value)
}

// ProxyTimeout sets various proxy timeouts
func (lb *LocationBuilder) ProxyTimeout(timeout string) *LocationBuilder {
	return lb.AddDirective("proxy_read_timeout", timeout).
		AddDirective("proxy_connect_timeout", timeout).
		AddDirective("proxy_send_timeout", timeout)
}

// TryFiles sets try_files
func (lb *LocationBuilder) TryFiles(files ...string) *LocationBuilder {
	return lb.AddDirective("try_files", files...)
}

// Root sets root for this location
func (lb *LocationBuilder) Root(path string) *LocationBuilder {
	return lb.AddDirective("root", path)
}

// Alias sets alias for this location
func (lb *LocationBuilder) Alias(path string) *LocationBuilder {
	return lb.AddDirective("alias", path)
}

// Index sets index files for this location
func (lb *LocationBuilder) Index(files ...string) *LocationBuilder {
	return lb.AddDirective("index", files...)
}

// Return adds a return directive
func (lb *LocationBuilder) Return(code string, url ...string) *LocationBuilder {
	if len(url) > 0 {
		return lb.AddDirective("return", code, url[0])
	}
	return lb.AddDirective("return", code)
}

// FastCGI configures FastCGI
func (lb *LocationBuilder) FastCGI(pass string) *LocationBuilder {
	return lb.AddDirective("fastcgi_pass", pass).
		AddDirective("fastcgi_index", "index.php").
		AddDirective("include", "fastcgi_params")
}

// EndLocation returns to the server builder
func (lb *LocationBuilder) EndLocation() *ServerBuilder {
	return lb.serverBuilder
}

// End returns to the main config builder
func (lb *LocationBuilder) End() *ConfigBuilder {
	return &ConfigBuilder{config: lb.config}
}

// SSLBuilder provides methods for SSL configuration
type SSLBuilder struct {
	config        *config.Config
	httpBuilder   *HTTPBuilder
	serverBuilder *ServerBuilder
}

// Certificate sets SSL certificate
func (ssl *SSLBuilder) Certificate(path string) *SSLBuilder {
	ssl.serverBuilder.AddDirective("ssl_certificate", path)
	return ssl
}

// CertificateKey sets SSL certificate key
func (ssl *SSLBuilder) CertificateKey(path string) *SSLBuilder {
	ssl.serverBuilder.AddDirective("ssl_certificate_key", path)
	return ssl
}

// Protocols sets SSL protocols
func (ssl *SSLBuilder) Protocols(protocols ...string) *SSLBuilder {
	ssl.serverBuilder.AddDirective("ssl_protocols", protocols...)
	return ssl
}

// Ciphers sets SSL ciphers
func (ssl *SSLBuilder) Ciphers(ciphers string) *SSLBuilder {
	ssl.serverBuilder.AddDirective("ssl_ciphers", ciphers)
	return ssl
}

// PreferServerCiphers enables ssl_prefer_server_ciphers
func (ssl *SSLBuilder) PreferServerCiphers(enabled bool) *SSLBuilder {
	value := "off"
	if enabled {
		value = "on"
	}
	ssl.serverBuilder.AddDirective("ssl_prefer_server_ciphers", value)
	return ssl
}

// SessionTimeout sets SSL session timeout
func (ssl *SSLBuilder) SessionTimeout(timeout string) *SSLBuilder {
	ssl.serverBuilder.AddDirective("ssl_session_timeout", timeout)
	return ssl
}

// SessionCache sets SSL session cache
func (ssl *SSLBuilder) SessionCache(cache string) *SSLBuilder {
	ssl.serverBuilder.AddDirective("ssl_session_cache", cache)
	return ssl
}

// HSTS enables HTTP Strict Transport Security
func (ssl *SSLBuilder) HSTS(maxAge string, includeSubdomains bool) *SSLBuilder {
	value := "max-age=" + maxAge
	if includeSubdomains {
		value += "; includeSubDomains"
	}
	ssl.serverBuilder.AddDirective("add_header", "Strict-Transport-Security", value)
	return ssl
}

// EndSSL returns to the server builder
func (ssl *SSLBuilder) EndSSL() *ServerBuilder {
	return ssl.serverBuilder
}

// UpstreamBuilder provides methods for building upstream block
type UpstreamBuilder struct {
	config        *config.Config
	httpBuilder   *HTTPBuilder
	upstreamBlock *config.Upstream
}

// AddDirective adds a directive to the upstream block
func (ub *UpstreamBuilder) AddDirective(name string, parameters ...string) *UpstreamBuilder {
	var params []config.Parameter
	for _, param := range parameters {
		params = append(params, config.NewParameter(param))
	}

	directive := &config.Directive{
		Name:       name,
		Parameters: params,
	}

	ub.upstreamBlock.Directives = append(ub.upstreamBlock.Directives, directive)
	return ub
}

// Server adds a server to the upstream
func (ub *UpstreamBuilder) Server(address string, options ...string) *UpstreamBuilder {
	params := []string{address}
	params = append(params, options...)
	return ub.AddDirective("server", params...)
}

// IpHash enables ip_hash
func (ub *UpstreamBuilder) IpHash() *UpstreamBuilder {
	return ub.AddDirective("ip_hash")
}

// LeastConn enables least_conn
func (ub *UpstreamBuilder) LeastConn() *UpstreamBuilder {
	return ub.AddDirective("least_conn")
}

// KeepaliveConnections sets keepalive connections
func (ub *UpstreamBuilder) KeepaliveConnections(count string) *UpstreamBuilder {
	return ub.AddDirective("keepalive", count)
}

// EndUpstream returns to the HTTP builder
func (ub *UpstreamBuilder) EndUpstream() *HTTPBuilder {
	return ub.httpBuilder
}

// End returns to the main config builder
func (ub *UpstreamBuilder) End() *ConfigBuilder {
	return &ConfigBuilder{config: ub.config}
}
