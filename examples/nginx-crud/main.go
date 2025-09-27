package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/lefeck/gonginx/config"
	"github.com/lefeck/gonginx/dumper"
	"github.com/lefeck/gonginx/parser"
)

// nginx 配置 CRUD 操作完整示例
func main() {
	fmt.Println("=== Nginx 配置 CRUD 操作示例 ===")

	// 准备一个示例配置文件
	sampleConfig := `
worker_processes auto;
worker_connections 1024;

http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;
    
    sendfile on;
    keepalive_timeout 65;
    
    upstream backend {
        server 192.168.1.10:8080 weight=3;
        server 192.168.1.11:8080 weight=2;
        server 192.168.1.12:8080 backup;
    }
    
    upstream api_servers {
        server 10.0.1.1:9000;
        server 10.0.1.2:9000;
    }
    
    server {
        listen 80;
        server_name example.com www.example.com;
        root /var/www/html;
        index index.html index.htm;
        
        location / {
            try_files $uri $uri/ =404;
        }
        
        location /api {
            proxy_pass http://backend;
            proxy_set_header Host $host;
        }
        
        location /health {
            return 200 "OK";
        }
    }
    
    server {
        listen 443 ssl;
        server_name secure.example.com;
        
        ssl_certificate /etc/ssl/certs/example.crt;
        ssl_certificate_key /etc/ssl/private/example.key;
        
        location / {
            proxy_pass http://api_servers;
        }
    }
}
`

	// 解析配置
	p := parser.NewStringParser(sampleConfig)
	conf, err := p.Parse()
	if err != nil {
		log.Fatal("解析配置失败:", err)
	}

	fmt.Println("✅ 配置解析成功")
	fmt.Println()

	// ===================
	// 1. CREATE (创建) 操作
	// ===================
	fmt.Println("🔨 1. CREATE (创建) 操作示例")
	fmt.Println("----------------------------------------")

	// 1.1 创建新的 upstream
	createUpstream(conf)

	// 1.2 创建新的 server
	createServer(conf)

	// 1.3 创建新的 location
	createLocation(conf)

	// 1.4 创建新的指令
	createDirective(conf)

	fmt.Println()

	// ===================
	// 2. READ (读取) 操作
	// ===================
	fmt.Println("🔍 2. READ (读取) 操作示例")
	fmt.Println("----------------------------------------")

	// 2.1 读取所有 upstream
	readUpstreams(conf)

	// 2.2 读取所有 server
	readServers(conf)

	// 2.3 读取特定 server 的信息
	readSpecificServer(conf)

	// 2.4 读取所有 location
	readLocations(conf)

	// 2.5 搜索特定指令
	searchDirectives(conf)

	fmt.Println()

	// ===================
	// 3. UPDATE (更新) 操作
	// ===================
	fmt.Println("✏️ 3. UPDATE (更新) 操作示例")
	fmt.Println("----------------------------------------")

	// 3.1 更新 upstream 服务器
	updateUpstreamServer(conf)

	// 3.2 更新 server 配置
	updateServerConfig(conf)

	// 3.3 更新 location 配置
	updateLocationConfig(conf)

	// 3.4 更新全局指令
	updateGlobalDirective(conf)

	fmt.Println()

	// ===================
	// 4. DELETE (删除) 操作
	// ===================
	fmt.Println("🗑️ 4. DELETE (删除) 操作示例")
	fmt.Println("----------------------------------------")

	// 4.1 删除 upstream 服务器
	deleteUpstreamServer(conf)

	// 4.2 删除整个 upstream
	deleteUpstream(conf)

	// 4.3 删除 location
	deleteLocation(conf)

	// 4.4 删除指令
	deleteDirective(conf)

	fmt.Println()

	// ===================
	// 5. 保存修改后的配置
	// ===================
	fmt.Println("💾 5. 保存修改后的配置")
	fmt.Println("----------------------------------------")

	// 输出最终配置
	finalConfig := dumper.DumpConfig(conf, dumper.IndentedStyle)
	fmt.Println("修改后的完整配置:")
	fmt.Println(finalConfig)

	// 可选：保存到文件
	saveConfigToFile(finalConfig, "nginx_modified.conf")

	fmt.Println()
	fmt.Println("=== CRUD 操作示例完成 ===")
}

// ===========================
// CREATE (创建) 操作函数
// ===========================

func createUpstream(conf *config.Config) {
	fmt.Println("📝 创建新的 upstream 'new_backend'")

	// 创建新的 upstream 块
	upstreamDirective := &config.Directive{
		Name:       "upstream",
		Parameters: []config.Parameter{config.NewParameter("new_backend")},
		Block: &config.Block{
			Directives: []config.IDirective{
				&config.Directive{
					Name:       "server",
					Parameters: []config.Parameter{config.NewParameter("10.0.2.1:8080"), config.NewParameter("weight=5")},
				},
				&config.Directive{
					Name:       "server",
					Parameters: []config.Parameter{config.NewParameter("10.0.2.2:8080"), config.NewParameter("weight=3")},
				},
				&config.Directive{
					Name:       "least_conn",
					Parameters: []config.Parameter{},
				},
			},
		},
	}

	// 将新 upstream 添加到 http 块中
	httpBlocks := conf.FindDirectives("http")
	if len(httpBlocks) > 0 {
		httpBlock := httpBlocks[0]
		if httpDirective, ok := httpBlock.(*config.HTTP); ok {
			httpDirective.Directives = append(httpDirective.Directives, upstreamDirective)
			fmt.Println("   ✅ 成功创建 upstream 'new_backend'")
		}
	}
}

func createServer(conf *config.Config) {
	fmt.Println("📝 创建新的 server 块")

	// 创建新的 server 块
	serverDirective := &config.Directive{
		Name: "server",
		Block: &config.Block{
			Directives: []config.IDirective{
				&config.Directive{
					Name:       "listen",
					Parameters: []config.Parameter{config.NewParameter("8080")},
				},
				&config.Directive{
					Name:       "server_name",
					Parameters: []config.Parameter{config.NewParameter("api.example.com")},
				},
				&config.Directive{
					Name:       "root",
					Parameters: []config.Parameter{config.NewParameter("/var/www/api")},
				},
			},
		},
	}

	// 添加到 http 块
	httpBlocks := conf.FindDirectives("http")
	if len(httpBlocks) > 0 {
		httpBlock := httpBlocks[0]
		if httpDirective, ok := httpBlock.(*config.HTTP); ok {
			httpDirective.Directives = append(httpDirective.Directives, serverDirective)
			fmt.Println("   ✅ 成功创建新的 server 块")
		}
	}
}

func createLocation(conf *config.Config) {
	fmt.Println("📝 在第一个 server 中创建新的 location")

	// 找到第一个 server 块
	servers := conf.FindServersByName("example.com")
	if len(servers) > 0 {
		firstServer := servers[0]

		// 创建新的 location
		locationDirective := &config.Directive{
			Name:       "location",
			Parameters: []config.Parameter{config.NewParameter("/new-api")},
			Block: &config.Block{
				Directives: []config.IDirective{
					&config.Directive{
						Name:       "proxy_pass",
						Parameters: []config.Parameter{config.NewParameter("http://new_backend")},
					},
					&config.Directive{
						Name:       "proxy_set_header",
						Parameters: []config.Parameter{config.NewParameter("X-Real-IP"), config.NewParameter("$remote_addr")},
					},
				},
			},
		}

		// 添加到 server 块
		firstServer.GetBlock().(*config.Block).Directives = append(
			firstServer.GetBlock().(*config.Block).Directives,
			locationDirective,
		)
		fmt.Println("   ✅ 成功创建 location '/new-api'")
	}
}

func createDirective(conf *config.Config) {
	fmt.Println("📝 在 http 块中创建新的指令")

	// 创建新的指令
	newDirective := &config.Directive{
		Name:       "client_max_body_size",
		Parameters: []config.Parameter{config.NewParameter("100M")},
	}

	// 添加到 http 块
	httpBlocks := conf.FindDirectives("http")
	if len(httpBlocks) > 0 {
		httpBlock := httpBlocks[0]
		if httpDirective, ok := httpBlock.(*config.HTTP); ok {
			httpDirective.Directives = append(httpDirective.Directives, newDirective)
			fmt.Println("   ✅ 成功创建指令 'client_max_body_size 100M'")
		}
	}
}

// ===========================
// READ (读取) 操作函数
// ===========================

func readUpstreams(conf *config.Config) {
	fmt.Println("📖 读取所有 upstream 块")

	upstreams := conf.FindUpstreams()
	fmt.Printf("   找到 %d 个 upstream 块:\n", len(upstreams))

	for i, upstream := range upstreams {
		fmt.Printf("   %d. %s\n", i+1, upstream.UpstreamName)
		fmt.Printf("      服务器数量: %d\n", len(upstream.UpstreamServers))
		for j, server := range upstream.UpstreamServers {
			fmt.Printf("        - 服务器 %d: %s\n", j+1, server.Address)
		}
	}
}

func readServers(conf *config.Config) {
	fmt.Println("📖 读取所有 server 块")

	// 找到 http 块中的所有 server 指令
	httpBlocks := conf.FindDirectives("http")
	if len(httpBlocks) == 0 {
		fmt.Println("   未找到 http 块")
		return
	}

	httpBlock := httpBlocks[0]
	if httpDirective, ok := httpBlock.(*config.HTTP); ok {
		servers := httpDirective.FindDirectives("server")
		fmt.Printf("   找到 %d 个 server 块:\n", len(servers))

		for i, server := range servers {
			// 获取 listen 端口
			listenDirectives := server.GetBlock().FindDirectives("listen")
			var listenPorts []string
			for _, listen := range listenDirectives {
				params := listen.GetParameters()
				if len(params) > 0 {
					listenPorts = append(listenPorts, params[0].GetValue())
				}
			}
			listenStr := strings.Join(listenPorts, ", ")

			// 获取 server_name
			serverNameDirectives := server.GetBlock().FindDirectives("server_name")
			var serverNames []string
			for _, serverName := range serverNameDirectives {
				params := serverName.GetParameters()
				for _, param := range params {
					serverNames = append(serverNames, param.GetValue())
				}
			}
			serverNameStr := strings.Join(serverNames, ", ")

			fmt.Printf("   %d. Listen: [%s], Server Names: [%s]\n",
				i+1, listenStr, serverNameStr)
		}
	}
}

func readSpecificServer(conf *config.Config) {
	fmt.Println("📖 读取特定 server 的详细信息")

	// 按 server_name 查找
	servers := conf.FindServersByName("example.com")
	if len(servers) > 0 {
		server := servers[0]
		fmt.Println("   找到 server 'example.com':")

		// 读取所有 location
		locations := server.GetBlock().FindDirectives("location")
		fmt.Printf("   包含 %d 个 location:\n", len(locations))
		for i, loc := range locations {
			params := loc.GetParameters()
			if len(params) > 0 {
				fmt.Printf("     %d. %s\n", i+1, params[0].GetValue())
			}
		}
	}
}

func readLocations(conf *config.Config) {
	fmt.Println("📖 读取所有 location 块")

	// 使用高级搜索功能
	locations := conf.FindLocationsByPattern("/")
	fmt.Printf("   找到 %d 个匹配 '/' 的 location:\n", len(locations))

	for i, loc := range locations {
		fmt.Printf("   %d. Pattern: %s, Modifier: %s\n",
			i+1, loc.Match, loc.Modifier)
	}
}

func searchDirectives(conf *config.Config) {
	fmt.Println("📖 搜索特定指令")

	// 搜索所有 proxy_pass 指令
	proxyPasses := conf.FindDirectives("proxy_pass")
	fmt.Printf("   找到 %d 个 'proxy_pass' 指令:\n", len(proxyPasses))

	for i, directive := range proxyPasses {
		params := directive.GetParameters()
		if len(params) > 0 {
			fmt.Printf("   %d. %s\n", i+1, params[0].GetValue())
		}
	}

	// 获取所有 SSL 证书
	sslCerts := conf.GetAllSSLCertificates()
	fmt.Printf("   找到 %d 个 SSL 证书:\n", len(sslCerts))
	for i, cert := range sslCerts {
		fmt.Printf("   %d. %s\n", i+1, cert)
	}
}

// ===========================
// UPDATE (更新) 操作函数
// ===========================

func updateUpstreamServer(conf *config.Config) {
	fmt.Println("✏️ 更新 upstream 服务器")

	// 找到指定的 upstream
	upstream := conf.FindUpstreamByName("backend")
	if upstream != nil {
		fmt.Println("   找到 upstream 'backend'")

		// 更新第一个服务器的权重
		if len(upstream.UpstreamServers) > 0 {
			oldAddress := upstream.UpstreamServers[0].Address
			upstream.UpstreamServers[0].Address = "192.168.1.10:8080"
			upstream.UpstreamServers[0].Parameters = map[string]string{
				"weight":    "5", // 从 weight=3 改为 weight=5
				"max_fails": "2",
			}
			fmt.Printf("   ✅ 更新服务器: %s -> %s (weight=5)\n", oldAddress, upstream.UpstreamServers[0].Address)
		}

		// 添加新的服务器
		upstream.AddServer(&config.UpstreamServer{
			Address: "192.168.1.20:8080",
			Parameters: map[string]string{
				"weight": "1",
			},
		})
		fmt.Println("   ✅ 添加新服务器: 192.168.1.20:8080")
	}
}

func updateServerConfig(conf *config.Config) {
	fmt.Println("✏️ 更新 server 配置")

	// 找到指定的 server
	servers := conf.FindServersByName("example.com")
	if len(servers) > 0 {
		server := servers[0]
		fmt.Println("   找到 server 'example.com'")

		// 更新 root 目录
		rootDirectives := server.GetBlock().FindDirectives("root")
		if len(rootDirectives) > 0 {
			rootDirective := rootDirectives[0].(*config.Directive)
			oldRoot := rootDirective.Parameters[0].GetValue()
			rootDirective.Parameters[0] = config.NewParameter("/var/www/new-html")
			fmt.Printf("   ✅ 更新 root: %s -> /var/www/new-html\n", oldRoot)
		}

		// 添加新的指令
		newDirective := &config.Directive{
			Name:       "access_log",
			Parameters: []config.Parameter{config.NewParameter("/var/log/nginx/example.log")},
		}
		server.GetBlock().(*config.Block).Directives = append(
			server.GetBlock().(*config.Block).Directives,
			newDirective,
		)
		fmt.Println("   ✅ 添加 access_log 指令")
	}
}

func updateLocationConfig(conf *config.Config) {
	fmt.Println("✏️ 更新 location 配置")

	// 找到特定的 location
	locations := conf.FindLocationsByPattern("/api")
	if len(locations) > 0 {
		location := locations[0]
		fmt.Println("   找到 location '/api'")

		// 更新 proxy_pass
		proxyPasses := location.GetBlock().FindDirectives("proxy_pass")
		if len(proxyPasses) > 0 {
			proxyPass := proxyPasses[0].(*config.Directive)
			oldTarget := proxyPass.Parameters[0].GetValue()
			proxyPass.Parameters[0] = config.NewParameter("http://new_backend")
			fmt.Printf("   ✅ 更新 proxy_pass: %s -> http://new_backend\n", oldTarget)
		}

		// 添加新的 header
		newHeader := &config.Directive{
			Name: "proxy_set_header",
			Parameters: []config.Parameter{
				config.NewParameter("X-Forwarded-For"),
				config.NewParameter("$proxy_add_x_forwarded_for"),
			},
		}
		location.GetBlock().(*config.Block).Directives = append(
			location.GetBlock().(*config.Block).Directives,
			newHeader,
		)
		fmt.Println("   ✅ 添加 X-Forwarded-For header")
	}
}

func updateGlobalDirective(conf *config.Config) {
	fmt.Println("✏️ 更新全局指令")

	// 更新 worker_processes
	workerProcesses := conf.FindDirectives("worker_processes")
	if len(workerProcesses) > 0 {
		directive := workerProcesses[0].(*config.Directive)
		oldValue := directive.Parameters[0].GetValue()
		directive.Parameters[0] = config.NewParameter("4")
		fmt.Printf("   ✅ 更新 worker_processes: %s -> 4\n", oldValue)
	}

	// 在 http 块中更新 keepalive_timeout
	httpBlocks := conf.FindDirectives("http")
	if len(httpBlocks) > 0 {
		httpBlock := httpBlocks[0]
		keepaliveDirectives := httpBlock.GetBlock().FindDirectives("keepalive_timeout")
		if len(keepaliveDirectives) > 0 {
			directive := keepaliveDirectives[0].(*config.Directive)
			oldValue := directive.Parameters[0].GetValue()
			directive.Parameters[0] = config.NewParameter("120")
			fmt.Printf("   ✅ 更新 keepalive_timeout: %s -> 120\n", oldValue)
		}
	}
}

// ===========================
// DELETE (删除) 操作函数
// ===========================

func deleteUpstreamServer(conf *config.Config) {
	fmt.Println("🗑️ 删除 upstream 服务器")

	// 找到指定的 upstream
	upstream := conf.FindUpstreamByName("backend")
	if upstream != nil && len(upstream.UpstreamServers) > 2 {
		// 删除最后一个服务器（backup 服务器）
		deletedServer := upstream.UpstreamServers[len(upstream.UpstreamServers)-1]
		upstream.UpstreamServers = upstream.UpstreamServers[:len(upstream.UpstreamServers)-1]
		fmt.Printf("   ✅ 删除服务器: %s\n", deletedServer.Address)
	}
}

func deleteUpstream(conf *config.Config) {
	fmt.Println("🗑️ 删除整个 upstream")

	// 找到 http 块
	httpBlocks := conf.FindDirectives("http")
	if len(httpBlocks) > 0 {
		httpBlock := httpBlocks[0]
		directives := httpBlock.GetBlock().(*config.Block).Directives

		// 查找并删除指定的 upstream
		for i, directive := range directives {
			if directive.GetName() == "upstream" {
				params := directive.GetParameters()
				if len(params) > 0 && params[0].GetValue() == "api_servers" {
					// 删除这个 upstream
					httpBlock.GetBlock().(*config.Block).Directives = append(
						directives[:i],
						directives[i+1:]...,
					)
					fmt.Println("   ✅ 删除 upstream 'api_servers'")
					break
				}
			}
		}
	}
}

func deleteLocation(conf *config.Config) {
	fmt.Println("🗑️ 删除 location")

	// 找到第一个 server
	servers := conf.FindServersByName("example.com")
	if len(servers) > 0 {
		server := servers[0]
		directives := server.GetBlock().(*config.Block).Directives

		// 查找并删除 /health location
		for i, directive := range directives {
			if directive.GetName() == "location" {
				params := directive.GetParameters()
				if len(params) > 0 && params[0].GetValue() == "/health" {
					// 删除这个 location
					server.GetBlock().(*config.Block).Directives = append(
						directives[:i],
						directives[i+1:]...,
					)
					fmt.Println("   ✅ 删除 location '/health'")
					break
				}
			}
		}
	}
}

func deleteDirective(conf *config.Config) {
	fmt.Println("🗑️ 删除指令")

	// 从 http 块中删除 sendfile 指令
	httpBlocks := conf.FindDirectives("http")
	if len(httpBlocks) > 0 {
		httpBlock := httpBlocks[0]
		directives := httpBlock.GetBlock().(*config.Block).Directives

		// 查找并删除 sendfile 指令
		for i, directive := range directives {
			if directive.GetName() == "sendfile" {
				// 删除这个指令
				httpBlock.GetBlock().(*config.Block).Directives = append(
					directives[:i],
					directives[i+1:]...,
				)
				fmt.Println("   ✅ 删除 'sendfile' 指令")
				break
			}
		}
	}
}

// ===========================
// 辅助函数
// ===========================

func saveConfigToFile(configContent, filename string) {
	err := os.WriteFile(filename, []byte(configContent), 0644)
	if err != nil {
		fmt.Printf("   ❌ 保存文件失败: %v\n", err)
	} else {
		fmt.Printf("   ✅ 配置已保存到文件: %s\n", filename)
	}
}
