package integration_tests

import (
	"testing"

	"github.com/lefeck/gonginx/config"
	"github.com/lefeck/gonginx/dumper"
	"github.com/lefeck/gonginx/parser"
)

// TestBasicParsing 测试基础解析功能
func TestBasicParsing(t *testing.T) {
	configContent := `
events {
    worker_connections 1024;
}

http {
    server {
        listen 80;
        server_name example.com;
        root /var/www/html;
        index index.html;
        
        location / {
            try_files $uri $uri/ =404;
        }
    }
}
`

	// 解析配置
	p := parser.NewStringParser(configContent)
	conf, err := p.Parse()
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	// 验证解析结果
	servers := conf.FindDirectives("server")
	if len(servers) != 1 {
		t.Errorf("期望 1 个 server，实际 %d 个", len(servers))
	}

	locations := conf.FindDirectives("location")
	if len(locations) != 1 {
		t.Errorf("期望 1 个 location，实际 %d 个", len(locations))
	}

	// 测试导出
	configStr := dumper.DumpConfig(conf, dumper.IndentedStyle)
	if len(configStr) == 0 {
		t.Error("导出的配置为空")
	}

	// 往返测试
	roundTripParser := parser.NewStringParser(configStr)
	roundTripConf, err := roundTripParser.Parse()
	if err != nil {
		t.Fatalf("往返解析失败: %v", err)
	}

	roundTripServers := roundTripConf.FindDirectives("server")
	if len(roundTripServers) != len(servers) {
		t.Errorf("往返后 server 数量不匹配: 原始 %d, 往返后 %d",
			len(servers), len(roundTripServers))
	}

	t.Log("基础解析测试成功")
}

// TestContextValidation 测试上下文验证功能
func TestContextValidation(t *testing.T) {
	// 包含上下文错误的配置
	problematicConfig := `
events {
    worker_connections 1024;
}

http {
    server {
        listen 80;
        server_name example.com;
        
        location / {
            listen 8080;  # 错误：listen 不能在 location 中
            proxy_pass http://backend;
        }
    }
}
`

	p := parser.NewStringParser(problematicConfig)
	conf, err := p.Parse()
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	// 创建上下文验证器
	contextValidator := config.NewContextValidator()
	errors := contextValidator.ValidateConfig(conf)

	// 应该发现上下文错误
	if len(errors) == 0 {
		t.Error("应该发现上下文错误，但没有")
	}

	t.Logf("发现 %d 个上下文错误", len(errors))
	for _, err := range errors {
		t.Logf("上下文错误: %s", err.Error())
	}

	t.Log("上下文验证测试成功")
}

// TestDependencyValidation 测试依赖关系验证功能
func TestDependencyValidation(t *testing.T) {
	// 包含依赖关系错误的配置
	problematicConfig := `
events {
    worker_connections 1024;
}

http {
    server {
        listen 443 ssl;
        server_name example.com;
        
        ssl_certificate /etc/ssl/cert.pem;  # 缺少 ssl_certificate_key
        
        location / {
            proxy_pass http://backend;  # backend upstream 未定义
        }
    }
    
    upstream empty_upstream {
        # 空的 upstream - 缺少 server 指令
    }
}
`

	p := parser.NewStringParser(problematicConfig)
	conf, err := p.Parse()
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	// 创建依赖验证器
	dependencyValidator := config.NewDependencyValidator()
	errors := dependencyValidator.ValidateDependencies(conf)

	// 应该发现依赖关系错误
	if len(errors) == 0 {
		t.Error("应该发现依赖关系错误，但没有")
	}

	t.Logf("发现 %d 个依赖关系错误", len(errors))
	for _, err := range errors {
		t.Logf("依赖关系错误: %s", err.Error())
	}

	t.Log("依赖关系验证测试成功")
}

// TestParameterTypeDetection 测试参数类型检测功能
func TestParameterTypeDetection(t *testing.T) {
	testCases := []struct {
		value    string
		expected config.ParameterType
	}{
		{"on", config.ParameterTypeBoolean},
		{"off", config.ParameterTypeBoolean},
		{"1024", config.ParameterTypeNumber},
		{"10m", config.ParameterTypeSize},
		{"30s", config.ParameterTypeTime},
		{"/var/www/html", config.ParameterTypePath},
		{"http://example.com", config.ParameterTypeURL},
		{"$remote_addr", config.ParameterTypeVariable},
		{"\"quoted string\"", config.ParameterTypeQuoted},
		{"~^/api/(.*)$", config.ParameterTypeRegex},
	}

	for _, tc := range testCases {
		t.Run(tc.value, func(t *testing.T) {
			detected := config.DetectParameterType(tc.value)
			if detected != tc.expected {
				t.Errorf("参数 '%s': 期望类型 %v, 实际类型 %v",
					tc.value, tc.expected, detected)
			}
		})
	}

	t.Log("参数类型检测测试成功")
}

// TestComplexConfiguration 测试复杂配置的处理
func TestComplexConfiguration(t *testing.T) {
	complexConfig := `
events {
    worker_connections 1024;
    use epoll;
    multi_accept on;
}

http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;
    
    upstream backend {
        server 192.168.1.10:8080 weight=3;
        server 192.168.1.11:8080 weight=2;
        server 192.168.1.12:8080 backup;
    }
    
    map $request_uri $redirect_uri {
        ~^/old/(.*)$ /new/$1;
        default "";
    }
    
    server {
        listen 80;
        listen 443 ssl;
        server_name example.com www.example.com;
        
        ssl_certificate /etc/ssl/certs/example.crt;
        ssl_certificate_key /etc/ssl/private/example.key;
        
        if ($redirect_uri != "") {
            return 301 $redirect_uri;
        }
        
        location / {
            try_files $uri $uri/ @fallback;
        }
        
        location @fallback {
            proxy_pass http://backend;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
        }
        
        location /static/ {
            alias /var/www/static/;
            expires 30d;
        }
        
        location ~ \.php$ {
            fastcgi_pass unix:/var/run/php/php7.4-fpm.sock;
            fastcgi_index index.php;
            include fastcgi_params;
        }
    }
}
`

	// 解析复杂配置
	p := parser.NewStringParser(complexConfig)
	conf, err := p.Parse()
	if err != nil {
		t.Fatalf("复杂配置解析失败: %v", err)
	}

	// 验证解析结果
	servers := conf.FindDirectives("server")
	upstreams := conf.FindDirectives("upstream")
	locations := conf.FindDirectives("location")
	maps := conf.FindDirectives("map")
	ifs := conf.FindDirectives("if")

	t.Logf("解析结果: %d servers, %d upstreams, %d locations, %d maps, %d ifs",
		len(servers), len(upstreams), len(locations), len(maps), len(ifs))

	if len(servers) != 1 {
		t.Errorf("期望 1 个 server，实际 %d 个", len(servers))
	}

	if len(upstreams) != 1 {
		t.Errorf("期望 1 个 upstream，实际 %d 个", len(upstreams))
	}

	if len(locations) < 3 {
		t.Errorf("期望至少 3 个 location，实际 %d 个", len(locations))
	}

	// 测试上下文验证
	contextValidator := config.NewContextValidator()
	contextErrors := contextValidator.ValidateConfig(conf)

	t.Logf("上下文验证: %d 个错误", len(contextErrors))
	for _, err := range contextErrors {
		t.Logf("上下文错误: %s", err.Error())
	}

	// 测试依赖关系验证
	dependencyValidator := config.NewDependencyValidator()
	dependencyErrors := dependencyValidator.ValidateDependencies(conf)

	t.Logf("依赖关系验证: %d 个错误", len(dependencyErrors))
	for _, err := range dependencyErrors {
		t.Logf("依赖关系错误: %s", err.Error())
	}

	t.Log("复杂配置测试成功")
}
