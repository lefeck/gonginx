package benchmarks

import (
	"strings"
	"testing"

	"github.com/lefeck/gonginx/parser"
)

// 小型配置基准测试
func BenchmarkParseSmallConfig(b *testing.B) {
	config := `
events {
    worker_connections 1024;
}

http {
    server {
        listen 80;
        server_name example.com;
        root /var/www/html;
        index index.html;
    }
}
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := parser.NewStringParser(config)
		_, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// 中型配置基准测试
func BenchmarkParseMediumConfig(b *testing.B) {
	config := `
events {
    worker_connections 1024;
    use epoll;
    multi_accept on;
}

http {
    include       /etc/nginx/mime.types;
    default_type  application/octet-stream;
    
    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                    '$status $body_bytes_sent "$http_referer" '
                    '"$http_user_agent" "$http_x_forwarded_for"';
    
    access_log /var/log/nginx/access.log main;
    error_log /var/log/nginx/error.log warn;
    
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    types_hash_max_size 2048;
    
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_types text/plain text/css text/xml text/javascript application/javascript application/xml+rss application/json;
    
    upstream backend {
        server 192.168.1.10:8080 weight=3 max_fails=3 fail_timeout=30s;
        server 192.168.1.11:8080 weight=2 max_fails=3 fail_timeout=30s;
        server 192.168.1.12:8080 backup;
    }
    
    server {
        listen 80;
        listen [::]:80;
        server_name example.com www.example.com;
        root /var/www/html;
        index index.html index.htm;
        
        access_log /var/log/nginx/example.access.log main;
        error_log /var/log/nginx/example.error.log warn;
        
        location / {
            try_files $uri $uri/ @fallback;
        }
        
        location @fallback {
            proxy_pass http://backend;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_connect_timeout 30;
            proxy_send_timeout 30;
            proxy_read_timeout 30;
        }
        
        location /static/ {
            alias /var/www/static/;
            expires 30d;
            add_header Cache-Control "public, immutable";
        }
        
        location /api/ {
            proxy_pass http://backend;
            proxy_buffering on;
            proxy_buffer_size 128k;
            proxy_buffers 4 256k;
            proxy_busy_buffers_size 256k;
        }
        
        location ~ \\.php$ {
            fastcgi_pass unix:/var/run/php/php7.4-fpm.sock;
            fastcgi_index index.php;
            fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
            include fastcgi_params;
        }
    }
    
    server {
        listen 443 ssl http2;
        listen [::]:443 ssl http2;
        server_name example.com www.example.com;
        
        ssl_certificate /etc/ssl/certs/example.com.crt;
        ssl_certificate_key /etc/ssl/private/example.com.key;
        ssl_session_cache shared:SSL:1m;
        ssl_session_timeout 10m;
        ssl_ciphers HIGH:!aNULL:!MD5;
        ssl_prefer_server_ciphers on;
        
        root /var/www/html;
        index index.html index.htm;
        
        location / {
            try_files $uri $uri/ @fallback;
        }
        
        location @fallback {
            proxy_pass http://backend;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }
}
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := parser.NewStringParser(config)
		_, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// 大型配置基准测试
func BenchmarkParseLargeConfig(b *testing.B) {
	// 生成一个包含多个 server 块的大型配置
	var configBuilder strings.Builder

	configBuilder.WriteString(`
events {
    worker_connections 1024;
    use epoll;
    multi_accept on;
}

http {
    include       /etc/nginx/mime.types;
    default_type  application/octet-stream;
    
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml application/xml+rss text/javascript;
`)

	// 添加多个 upstream
	for i := 0; i < 10; i++ {
		configBuilder.WriteString(`
    upstream backend` + string(rune('0'+i)) + ` {
        server 192.168.1.` + string(rune('1'+i)) + `0:8080 weight=3;
        server 192.168.1.` + string(rune('1'+i)) + `1:8080 weight=2;
        server 192.168.1.` + string(rune('1'+i)) + `2:8080 backup;
    }`)
	}

	// 添加多个 server 块
	for i := 0; i < 10; i++ { // 减少到 10 个以避免字符编码问题
		siteNum := string(rune('0' + i))
		configBuilder.WriteString(`
    server {
        listen 80;
        server_name site` + siteNum + `.example.com;
        root /var/www/site` + siteNum + `;
        index index.html;
        
        location / {
            try_files $uri $uri/ @fallback` + siteNum + `;
        }
        
        location @fallback` + siteNum + ` {
            proxy_pass http://backend` + siteNum + `;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        }
        
        location /static/ {
            alias /var/www/site` + siteNum + `/static/;
            expires 30d;
        }
        
        location /api/ {
            proxy_pass http://backend` + siteNum + `;
            proxy_buffering on;
        }
        
        location ~ \\.php$ {
            fastcgi_pass unix:/var/run/php/php7.4-fpm.sock;
            fastcgi_index index.php;
            include fastcgi_params;
        }
        
        location /admin/ {
            auth_basic "Admin Area";
            auth_basic_user_file /etc/nginx/.htpasswd` + siteNum + `;
            try_files $uri $uri/ @fallback` + siteNum + `;
        }
    }`)
	}

	configBuilder.WriteString(`
}`)

	config := configBuilder.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := parser.NewStringParser(config)
		_, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// 解析复杂嵌套结构的基准测试
func BenchmarkParseComplexNesting(b *testing.B) {
	config := `
events {
    worker_connections 1024;
}

http {
    map $request_uri $redirect_uri {
        ~^/old/(.*)$ /new/$1;
        ~^/legacy/(.*)$ /modern/$1;
        default "";
    }
    
    geo $remote_addr $geo {
        default US;
        127.0.0.1 local;
        10.0.0.0/8 internal;
        192.168.0.0/16 internal;
    }
    
    split_clients $remote_addr $variant {
        50% .v1;
        40% .v2;
        *   .v3;
    }
    
    limit_req_zone $binary_remote_addr zone=login:10m rate=1r/s;
    limit_conn_zone $binary_remote_addr zone=addr:10m;
    
    upstream api_backend {
        server api1.example.com:8080 weight=3;
        server api2.example.com:8080 weight=2;
        server api3.example.com:8080 backup;
    }
    
    server {
        listen 80;
        server_name example.com;
        
        if ($redirect_uri != "") {
            return 301 $redirect_uri;
        }
        
        location / {
            set $backend "api_backend";
            
            if ($geo = "internal") {
                set $backend "internal_backend";
            }
            
            if ($variant = ".v1") {
                proxy_pass http://$backend/v1;
            }
            if ($variant = ".v2") {
                proxy_pass http://$backend/v2;
            }
            if ($variant = ".v3") {
                proxy_pass http://$backend/v3;
            }
        }
        
        location /login {
            limit_req zone=login burst=5 nodelay;
            limit_conn addr 1;
            
            proxy_pass http://api_backend;
        }
        
        location ~ ^/api/v([0-9]+)/(.*)$ {
            set $version $1;
            set $path $2;
            
            if ($version = "1") {
                proxy_pass http://api_backend/legacy/$path;
            }
            if ($version = "2") {
                proxy_pass http://api_backend/current/$path;
            }
            if ($version = "3") {
                proxy_pass http://api_backend/beta/$path;
            }
        }
    }
}
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := parser.NewStringParser(config)
		_, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Include 文件解析基准测试
func BenchmarkParseWithIncludes(b *testing.B) {
	config := `
events {
    worker_connections 1024;
}

http {
    include /etc/nginx/mime.types;
    include /etc/nginx/conf.d/*.conf;
    include /etc/nginx/sites-enabled/*;
    
    server {
        listen 80;
        server_name example.com;
        
        include /etc/nginx/snippets/ssl-params.conf;
        include /etc/nginx/snippets/fastcgi-php.conf;
        
        location / {
            include /etc/nginx/snippets/proxy-params.conf;
            proxy_pass http://backend;
        }
    }
}
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := parser.NewStringParser(config)
		_, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// 内存分配基准测试
func BenchmarkParseMemoryAllocation(b *testing.B) {
	config := `
events {
    worker_connections 1024;
}

http {
    server {
        listen 80;
        server_name example.com;
        root /var/www/html;
        
        location / {
            proxy_pass http://backend;
        }
    }
}
`

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := parser.NewStringParser(config)
		_, err := p.Parse()
		if err != nil {
			b.Fatal(err)
		}
	}
}
