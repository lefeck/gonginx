package generator

// Template represents a configuration template
type Template struct {
	Name        string
	Description string
	Builder     func() *ConfigBuilder
}

// GetAllTemplates returns all available templates
func GetAllTemplates() []Template {
	return []Template{
		BasicWebServerTemplate(),
		ReverseProxyTemplate(),
		LoadBalancerTemplate(),
		SSLWebServerTemplate(),
		StaticFileServerTemplate(),
		PHPWebServerTemplate(),
		StreamProxyTemplate(),
		MicroservicesGatewayTemplate(),
	}
}

// BasicWebServerTemplate creates a basic web server configuration
func BasicWebServerTemplate() Template {
	return Template{
		Name:        "Basic Web Server",
		Description: "A simple web server configuration for serving static files",
		Builder: func() *ConfigBuilder {
			return NewConfigBuilder().
				WorkerProcesses("auto").
				ErrorLog("/var/log/nginx/error.log", "warn").
				PidFile("/var/run/nginx.pid").
				WorkerConnections("1024").
				HTTP().
				SendFile(true).
				TcpNoPush(true).
				TcpNoDelay(true).
				KeepaliveTimeout("65").
				Gzip(true).
				GzipTypes("text/plain", "text/css", "application/json", "application/javascript", "text/javascript").
				Include("/etc/nginx/mime.types").
				AccessLog("/var/log/nginx/access.log").
				Server().
				Listen("80").
				ServerName("example.com", "www.example.com").
				Root("/var/www/html").
				Index("index.html", "index.htm").
				Location("/").
				TryFiles("$uri", "$uri/", "=404").
				EndLocation().
				EndServer().
				End()
		},
	}
}

// ReverseProxyTemplate creates a reverse proxy configuration
func ReverseProxyTemplate() Template {
	return Template{
		Name:        "Reverse Proxy",
		Description: "Configuration for reverse proxy with upstream servers",
		Builder: func() *ConfigBuilder {
			return NewConfigBuilder().
				WorkerProcesses("auto").
				ErrorLog("/var/log/nginx/error.log", "warn").
				WorkerConnections("1024").
				HTTP().
				SendFile(true).
				TcpNoPush(true).
				TcpNoDelay(true).
				KeepaliveTimeout("65").
				ClientMaxBodySize("32m").
				Gzip(true).
				Upstream("backend").
				Server("127.0.0.1:8001", "weight=3").
				Server("127.0.0.1:8002", "weight=2").
				Server("127.0.0.1:8003", "backup").
				KeepaliveConnections("16").
				EndUpstream().
				Server().
				Listen("80").
				ServerName("api.example.com").
				Location("/").
				ProxyPass("http://backend").
				ProxySetHeader("Host", "$host").
				ProxySetHeader("X-Real-IP", "$remote_addr").
				ProxySetHeader("X-Forwarded-For", "$proxy_add_x_forwarded_for").
				ProxySetHeader("X-Forwarded-Proto", "$scheme").
				ProxyTimeout("30s").
				EndLocation().
				EndServer().
				End()
		},
	}
}

// LoadBalancerTemplate creates a load balancer configuration
func LoadBalancerTemplate() Template {
	return Template{
		Name:        "Load Balancer",
		Description: "High-performance load balancer with multiple upstream pools",
		Builder: func() *ConfigBuilder {
			return NewConfigBuilder().
				WorkerProcesses("auto").
				ErrorLog("/var/log/nginx/error.log", "warn").
				WorkerConnections("2048").
				HTTP().
				SendFile(true).
				TcpNoPush(true).
				TcpNoDelay(true).
				KeepaliveTimeout("30").
				ClientMaxBodySize("100m").
				Gzip(true).
				GzipTypes("text/plain", "text/css", "application/json", "application/javascript").
				Upstream("web_backend").
				LeastConn().
				Server("10.0.1.10:80", "max_fails=3", "fail_timeout=30s").
				Server("10.0.1.11:80", "max_fails=3", "fail_timeout=30s").
				Server("10.0.1.12:80", "max_fails=3", "fail_timeout=30s").
				KeepaliveConnections("32").
				EndUpstream().
				Upstream("api_backend").
				IpHash().
				Server("10.0.2.10:8080", "weight=3").
				Server("10.0.2.11:8080", "weight=2").
				Server("10.0.2.12:8080", "backup").
				EndUpstream().
				Server().
				Listen("80").
				ServerName("example.com", "www.example.com").
				Location("/api/").
				ProxyPass("http://api_backend").
				ProxySetHeader("Host", "$host").
				ProxySetHeader("X-Real-IP", "$remote_addr").
				ProxySetHeader("X-Forwarded-For", "$proxy_add_x_forwarded_for").
				EndLocation().
				Location("/").
				ProxyPass("http://web_backend").
				ProxySetHeader("Host", "$host").
				ProxySetHeader("X-Real-IP", "$remote_addr").
				ProxySetHeader("X-Forwarded-For", "$proxy_add_x_forwarded_for").
				EndLocation().
				EndServer().
				End()
		},
	}
}

// SSLWebServerTemplate creates an SSL-enabled web server
func SSLWebServerTemplate() Template {
	return Template{
		Name:        "SSL Web Server",
		Description: "Secure web server with SSL/TLS configuration and HTTP to HTTPS redirect",
		Builder: func() *ConfigBuilder {
			return NewConfigBuilder().
				WorkerProcesses("auto").
				ErrorLog("/var/log/nginx/error.log", "warn").
				WorkerConnections("1024").
				HTTP().
				SendFile(true).
				TcpNoPush(true).
				TcpNoDelay(true).
				KeepaliveTimeout("65").
				Gzip(true).
				Server().
				Listen("80").
				ServerName("example.com", "www.example.com").
				Return("301", "https://$server_name$request_uri").
				EndServer().
				Server().
				Listen("443", "ssl", "http2").
				ServerName("example.com", "www.example.com").
				Root("/var/www/html").
				Index("index.html", "index.htm").
				SSL().
				Certificate("/etc/ssl/certs/example.com.crt").
				CertificateKey("/etc/ssl/private/example.com.key").
				Protocols("TLSv1.2", "TLSv1.3").
				Ciphers("ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384").
				PreferServerCiphers(true).
				SessionTimeout("1d").
				SessionCache("shared:SSL:50m").
				HSTS("31536000", true).
				EndSSL().
				Location("/").
				TryFiles("$uri", "$uri/", "=404").
				EndLocation().
				EndServer().
				End()
		},
	}
}

// StaticFileServerTemplate creates a static file server configuration
func StaticFileServerTemplate() Template {
	return Template{
		Name:        "Static File Server",
		Description: "Optimized configuration for serving static files with caching",
		Builder: func() *ConfigBuilder {
			return NewConfigBuilder().
				WorkerProcesses("auto").
				ErrorLog("/var/log/nginx/error.log", "warn").
				WorkerConnections("1024").
				HTTP().
				SendFile(true).
				TcpNoPush(true).
				TcpNoDelay(true).
				KeepaliveTimeout("65").
				Gzip(true).
				GzipTypes("text/plain", "text/css", "application/json", "application/javascript",
					"text/xml", "application/xml", "application/xml+rss", "text/javascript").
				Server().
				Listen("80").
				ServerName("static.example.com").
				Root("/var/www/static").
				Index("index.html").
				AccessLog("/var/log/nginx/static.access.log").
				Location("~*", "\\.(js|css|png|jpg|jpeg|gif|ico|svg)$").
				AddDirective("expires", "1y").
				AddDirective("add_header", "Cache-Control", "public, immutable").
				AddDirective("add_header", "Vary", "Accept-Encoding").
				EndLocation().
				Location("~*", "\\.(html|htm)$").
				AddDirective("expires", "1h").
				AddDirective("add_header", "Cache-Control", "public").
				EndLocation().
				Location("/").
				TryFiles("$uri", "$uri/", "=404").
				EndLocation().
				EndServer().
				End()
		},
	}
}

// PHPWebServerTemplate creates a PHP web server configuration
func PHPWebServerTemplate() Template {
	return Template{
		Name:        "PHP Web Server",
		Description: "Web server configuration for PHP applications with FastCGI",
		Builder: func() *ConfigBuilder {
			return NewConfigBuilder().
				WorkerProcesses("auto").
				ErrorLog("/var/log/nginx/error.log", "warn").
				WorkerConnections("1024").
				HTTP().
				SendFile(true).
				TcpNoPush(true).
				TcpNoDelay(true).
				KeepaliveTimeout("65").
				ClientMaxBodySize("64m").
				Gzip(true).
				Server().
				Listen("80").
				ServerName("php.example.com").
				Root("/var/www/html").
				Index("index.php", "index.html", "index.htm").
				Location("/").
				TryFiles("$uri", "$uri/", "=404").
				EndLocation().
				Location("~", "\\.php$").
				FastCGI("127.0.0.1:9000").
				AddDirective("fastcgi_param", "SCRIPT_FILENAME", "$document_root$fastcgi_script_name").
				EndLocation().
				Location("~*", "\\.(js|css|png|jpg|jpeg|gif|ico|svg)$").
				AddDirective("expires", "30d").
				AddDirective("add_header", "Cache-Control", "public").
				EndLocation().
				EndServer().
				End()
		},
	}
}

// StreamProxyTemplate creates a TCP/UDP stream proxy configuration
func StreamProxyTemplate() Template {
	return Template{
		Name:        "Stream Proxy",
		Description: "TCP/UDP stream proxy for database and other TCP services",
		Builder: func() *ConfigBuilder {
			return NewConfigBuilder().
				WorkerProcesses("auto").
				ErrorLog("/var/log/nginx/error.log", "warn").
				WorkerConnections("1024").
				Stream().
				ErrorLog("/var/log/nginx/stream.error.log", "warn").
				Upstream("database_pool").
				LeastConn().
				Server("10.0.1.10:5432", "weight=3", "max_fails=3", "fail_timeout=30s").
				Server("10.0.1.11:5432", "weight=2", "max_fails=3", "fail_timeout=30s").
				Server("10.0.1.12:5432", "backup").
				EndUpstream().
				Upstream("redis_pool").
				Hash("$remote_addr", true).
				Server("10.0.2.10:6379").
				Server("10.0.2.11:6379").
				EndUpstream().
				Server().
				Listen("5432").
				ProxyPass("database_pool").
				ProxyTimeout("3s").
				ProxyConnectTimeout("1s").
				EndServer().
				Server().
				Listen("6379").
				ProxyPass("redis_pool").
				ProxyTimeout("1s").
				ProxyConnectTimeout("1s").
				EndServer().
				End()
		},
	}
}

// MicroservicesGatewayTemplate creates a microservices API gateway
func MicroservicesGatewayTemplate() Template {
	return Template{
		Name:        "Microservices Gateway",
		Description: "API Gateway configuration for microservices architecture",
		Builder: func() *ConfigBuilder {
			return NewConfigBuilder().
				WorkerProcesses("auto").
				ErrorLog("/var/log/nginx/error.log", "warn").
				WorkerConnections("2048").
				HTTP().
				SendFile(true).
				TcpNoPush(true).
				TcpNoDelay(true).
				KeepaliveTimeout("30").
				ClientMaxBodySize("10m").
				Gzip(true).
				AddDirective("map", "$http_upgrade", "$connection_upgrade").
				Upstream("auth_service").
				LeastConn().
				Server("auth-service:8080", "max_fails=3", "fail_timeout=30s").
				KeepaliveConnections("16").
				EndUpstream().
				Upstream("user_service").
				LeastConn().
				Server("user-service:8080", "max_fails=3", "fail_timeout=30s").
				KeepaliveConnections("16").
				EndUpstream().
				Upstream("order_service").
				LeastConn().
				Server("order-service:8080", "max_fails=3", "fail_timeout=30s").
				KeepaliveConnections("16").
				EndUpstream().
				Server().
				Listen("80").
				ServerName("api.example.com").
				AccessLog("/var/log/nginx/api.access.log").
				Location("/auth/").
				ProxyPass("http://auth_service/").
				ProxySetHeader("Host", "$host").
				ProxySetHeader("X-Real-IP", "$remote_addr").
				ProxySetHeader("X-Forwarded-For", "$proxy_add_x_forwarded_for").
				ProxySetHeader("X-Forwarded-Proto", "$scheme").
				ProxyTimeout("30s").
				EndLocation().
				Location("/users/").
				ProxyPass("http://user_service/").
				ProxySetHeader("Host", "$host").
				ProxySetHeader("X-Real-IP", "$remote_addr").
				ProxySetHeader("X-Forwarded-For", "$proxy_add_x_forwarded_for").
				ProxySetHeader("X-Forwarded-Proto", "$scheme").
				ProxyTimeout("30s").
				EndLocation().
				Location("/orders/").
				ProxyPass("http://order_service/").
				ProxySetHeader("Host", "$host").
				ProxySetHeader("X-Real-IP", "$remote_addr").
				ProxySetHeader("X-Forwarded-For", "$proxy_add_x_forwarded_for").
				ProxySetHeader("X-Forwarded-Proto", "$scheme").
				ProxyTimeout("30s").
				EndLocation().
				Location("/health").
				Return("200", "OK").
				AddDirective("add_header", "Content-Type", "text/plain").
				EndLocation().
				EndServer().
				End()
		},
	}
}
