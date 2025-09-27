package benchmarks

import (
	"strings"
	"testing"

	"github.com/lefeck/gonginx/parser"
)

// 指令搜索基准测试
func BenchmarkFindDirectives(b *testing.B) {
	// 生成一个包含大量指令的配置
	var configBuilder strings.Builder

	configBuilder.WriteString(`
events {
    worker_connections 1024;
}

http {
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    gzip on;
`)

	// 添加多个 server 块，每个都有多个 location
	for i := 0; i < 20; i++ {
		configBuilder.WriteString(`
    server {
        listen 80;
        listen 443 ssl;
        server_name site` + string(rune('0'+i)) + `.example.com;
        root /var/www/site` + string(rune('0'+i)) + `;
        index index.html;
        
        access_log /var/log/nginx/site` + string(rune('0'+i)) + `.access.log;
        error_log /var/log/nginx/site` + string(rune('0'+i)) + `.error.log;
        
        location / {
            try_files $uri $uri/ =404;
        }
        
        location /api/ {
            proxy_pass http://backend;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
        }
        
        location /static/ {
            alias /var/www/static/;
            expires 30d;
            add_header Cache-Control "public";
        }
        
        location ~ \\.php$ {
            fastcgi_pass unix:/var/run/php/php7.4-fpm.sock;
            fastcgi_index index.php;
            include fastcgi_params;
        }
    }`)
	}

	configBuilder.WriteString(`
}`)

	config := configBuilder.String()

	p := parser.NewStringParser(config)
	conf, err := p.Parse()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = conf.FindDirectives("server")
	}
}

// Server 搜索基准测试
func BenchmarkFindServersByName(b *testing.B) {
	var configBuilder strings.Builder

	configBuilder.WriteString(`
events {
    worker_connections 1024;
}

http {
`)

	// 添加多个 server 块
	for i := 0; i < 50; i++ {
		configBuilder.WriteString(`
    server {
        listen 80;
        server_name site` + string(rune('0'+i)) + `.example.com www.site` + string(rune('0'+i)) + `.example.com;
        root /var/www/site` + string(rune('0'+i)) + `;
        
        location / {
            try_files $uri $uri/ =404;
        }
    }`)
	}

	configBuilder.WriteString(`
}`)

	config := configBuilder.String()

	p := parser.NewStringParser(config)
	conf, err := p.Parse()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = conf.FindServersByName("site25.example.com")
	}
}

// Upstream 搜索基准测试
func BenchmarkFindUpstreams(b *testing.B) {
	var configBuilder strings.Builder

	configBuilder.WriteString(`
events {
    worker_connections 1024;
}

http {
`)

	// 添加多个 upstream
	for i := 0; i < 30; i++ {
		configBuilder.WriteString(`
    upstream backend` + string(rune('0'+i)) + ` {
        server 192.168.1.` + string(rune('1'+i)) + `0:8080 weight=3;
        server 192.168.1.` + string(rune('1'+i)) + `1:8080 weight=2;
        server 192.168.1.` + string(rune('1'+i)) + `2:8080 backup;
    }`)
	}

	configBuilder.WriteString(`
    server {
        listen 80;
        server_name example.com;
        
        location / {
            proxy_pass http://backend15;
        }
    }
}`)

	config := configBuilder.String()

	p := parser.NewStringParser(config)
	conf, err := p.Parse()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = conf.FindUpstreams()
	}
}

// Location 搜索基准测试
func BenchmarkFindLocationsByPattern(b *testing.B) {
	var configBuilder strings.Builder

	configBuilder.WriteString(`
events {
    worker_connections 1024;
}

http {
    server {
        listen 80;
        server_name example.com;
`)

	// 添加大量 location 块
	patterns := []string{"/", "/api/", "/static/", "/admin/", "/user/", "/blog/", "/shop/", "/docs/"}
	for i := 0; i < 100; i++ {
		pattern := patterns[i%len(patterns)] + string(rune('a'+i%26))
		configBuilder.WriteString(`
        location ` + pattern + ` {
            proxy_pass http://backend;
            proxy_set_header Host $host;
        }`)
	}

	configBuilder.WriteString(`
    }
}`)

	config := configBuilder.String()

	p := parser.NewStringParser(config)
	conf, err := p.Parse()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = conf.FindLocationsByPattern("/api/")
	}
}

// SSL 证书搜索基准测试
func BenchmarkGetAllSSLCertificates(b *testing.B) {
	var configBuilder strings.Builder

	configBuilder.WriteString(`
events {
    worker_connections 1024;
}

http {
`)

	// 添加多个带 SSL 的 server
	for i := 0; i < 25; i++ {
		configBuilder.WriteString(`
    server {
        listen 443 ssl;
        server_name site` + string(rune('0'+i)) + `.example.com;
        
        ssl_certificate /etc/ssl/certs/site` + string(rune('0'+i)) + `.crt;
        ssl_certificate_key /etc/ssl/private/site` + string(rune('0'+i)) + `.key;
        ssl_trusted_certificate /etc/ssl/certs/site` + string(rune('0'+i)) + `-chain.crt;
        
        location / {
            proxy_pass http://backend;
        }
    }`)
	}

	configBuilder.WriteString(`
}`)

	config := configBuilder.String()

	p := parser.NewStringParser(config)
	conf, err := p.Parse()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = conf.GetAllSSLCertificates()
	}
}

// Upstream 服务器搜索基准测试
func BenchmarkGetAllUpstreamServers(b *testing.B) {
	var configBuilder strings.Builder

	configBuilder.WriteString(`
events {
    worker_connections 1024;
}

http {
`)

	// 添加多个 upstream，每个有多个 server
	for i := 0; i < 15; i++ {
		configBuilder.WriteString(`
    upstream backend` + string(rune('0'+i)) + ` {`)

		// 每个 upstream 有多个 server
		for j := 0; j < 5; j++ {
			configBuilder.WriteString(`
        server 192.168.` + string(rune('1'+i)) + `.` + string(rune('1'+j)) + `:8080 weight=` + string(rune('1'+j)) + `;`)
		}

		configBuilder.WriteString(`
    }`)
	}

	configBuilder.WriteString(`
    server {
        listen 80;
        server_name example.com;
        
        location / {
            proxy_pass http://backend5;
        }
    }
}`)

	config := configBuilder.String()

	p := parser.NewStringParser(config)
	conf, err := p.Parse()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = conf.GetAllUpstreamServers()
	}
}

// 深度嵌套搜索基准测试
func BenchmarkDeepNestedSearch(b *testing.B) {
	config := `
events {
    worker_connections 1024;
}

http {
    server {
        listen 80;
        server_name example.com;
        
        if ($request_method = POST) {
            set $post_request 1;
        }
        
        location / {
            if ($post_request = 1) {
                return 200 "POST request";
            }
            
            location ~ ^/api/v([0-9]+)/(.*)$ {
                set $version $1;
                set $path $2;
                
                if ($version = "1") {
                    proxy_pass http://api_v1_backend/$path;
                }
                if ($version = "2") {
                    proxy_pass http://api_v2_backend/$path;
                }
            }
            
            try_files $uri $uri/ =404;
        }
        
        location /admin/ {
            if ($remote_addr !~ ^192\.168\.1\.) {
                return 403;
            }
            
            location ~ ^/admin/api/(.*)$ {
                proxy_pass http://admin_api_backend/$1;
            }
            
            try_files $uri $uri/ =404;
        }
    }
}
`

	p := parser.NewStringParser(config)
	conf, err := p.Parse()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 搜索深度嵌套的指令
		_ = conf.FindDirectives("if")
		_ = conf.FindDirectives("set")
		_ = conf.FindDirectives("proxy_pass")
		_ = conf.FindDirectives("location")
	}
}

// 内存分配搜索基准测试
func BenchmarkSearchMemoryAllocation(b *testing.B) {
	config := `
events {
    worker_connections 1024;
}

http {
    upstream backend {
        server 192.168.1.10:8080;
        server 192.168.1.11:8080;
    }
    
    server {
        listen 80;
        server_name example.com;
        
        ssl_certificate /etc/ssl/certs/example.crt;
        ssl_certificate_key /etc/ssl/private/example.key;
        
        location / {
            proxy_pass http://backend;
        }
        
        location /api/ {
            proxy_pass http://backend;
        }
    }
}
`

	p := parser.NewStringParser(config)
	conf, err := p.Parse()
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = conf.FindDirectives("server")
		_ = conf.GetAllSSLCertificates()
		_ = conf.FindUpstreams()
		_ = conf.GetAllUpstreamServers()
	}
}
