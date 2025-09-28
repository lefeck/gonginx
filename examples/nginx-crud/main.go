package main

import (
	"fmt"
	"log"
	"os"

	"github.com/lefeck/gonginx/config"
	"github.com/lefeck/gonginx/dumper"
	"github.com/lefeck/gonginx/parser"
)

// 简化的 nginx 配置 CRUD 操作示例
func main() {
	fmt.Println("=== Nginx 配置 CRUD 操作示例 ===")

	// 准备一个示例配置文件
	sampleConfig := `
worker_processes auto;

http {
    sendfile on;
    keepalive_timeout 65;
    
    upstream backend {
        server 192.168.1.10:8080 weight=3;
        server 192.168.1.11:8080 weight=2;
    }
    
    server {
        listen 80;
        server_name example.com;
        
        location / {
            proxy_pass http://backend;
        }
        
        location /health {
            return 200 "OK";
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
	createUpstreamExample(conf)

	// 1.2 创建新的指令
	createDirectiveExample(conf)

	fmt.Println()

	// ===================
	// 2. READ (读取) 操作
	// ===================
	fmt.Println("🔍 2. READ (读取) 操作示例")
	fmt.Println("----------------------------------------")

	// 2.1 读取所有 upstream
	readUpstreamsExample(conf)

	// 2.2 读取特定指令
	readDirectivesExample(conf)

	// 2.3 使用高级搜索
	advancedSearchExample(conf)

	fmt.Println()

	// ===================
	// 3. UPDATE (更新) 操作
	// ===================
	fmt.Println("✏️ 3. UPDATE (更新) 操作示例")
	fmt.Println("----------------------------------------")

	// 3.1 更新 upstream 服务器
	updateUpstreamExample(conf)

	// 3.2 更新指令值
	updateDirectiveExample(conf)

	fmt.Println()

	// ===================
	// 4. DELETE (删除) 操作
	// ===================
	fmt.Println("🗑️ 4. DELETE (删除) 操作示例")
	fmt.Println("----------------------------------------")

	// 4.1 删除 upstream 服务器
	deleteUpstreamServerExample(conf)

	// 4.2 删除指令
	deleteDirectiveExample(conf)

	fmt.Println()

	// ===================
	// 5. 保存配置
	// ===================
	fmt.Println("💾 5. 保存修改后的配置")
	fmt.Println("----------------------------------------")

	finalConfig := dumper.DumpConfig(conf, dumper.IndentedStyle)
	fmt.Println("修改后的配置:")
	fmt.Println(finalConfig)

	// 保存到文件
	saveToFile(finalConfig, "nginx_modified.conf")

	fmt.Println("\n=== CRUD 操作示例完成 ===")
}

// ===========================
// CREATE (创建) 操作示例
// ===========================

func createUpstreamExample(conf *config.Config) {
	fmt.Println("📝 创建新的 upstream 'api_backend'")

	// 创建新的 upstream 对象
	upstream := &config.Upstream{
		UpstreamName: "api_backend",
		UpstreamServers: []*config.UpstreamServer{
			{
				Address:    "10.0.1.1:9000",
				Parameters: map[string]string{},
			},
			{
				Address:    "10.0.1.2:9000",
				Parameters: map[string]string{},
			},
		},
	}

	// 找到 http 块并添加 upstream
	httpBlocks := conf.FindDirectives("http")
	if len(httpBlocks) > 0 {
		if httpBlock, ok := httpBlocks[0].(*config.HTTP); ok {
			httpBlock.Directives = append(httpBlock.Directives, upstream)
			fmt.Println("   ✅ 成功创建 upstream 'api_backend'")
		}
	}
}

func createDirectiveExample(conf *config.Config) {
	fmt.Println("📝 创建新的指令 'gzip on'")

	// 创建新指令
	gzipDirective := &config.Directive{
		Name:       "gzip",
		Parameters: []config.Parameter{config.NewParameter("on")},
	}

	// 添加到 http 块
	httpBlocks := conf.FindDirectives("http")
	if len(httpBlocks) > 0 {
		if httpBlock, ok := httpBlocks[0].(*config.HTTP); ok {
			httpBlock.Directives = append(httpBlock.Directives, gzipDirective)
			fmt.Println("   ✅ 成功创建 gzip 指令")
		}
	}
}

// ===========================
// READ (读取) 操作示例
// ===========================

func readUpstreamsExample(conf *config.Config) {
	fmt.Println("📖 读取所有 upstream 块")

	upstreams := conf.FindUpstreams()
	fmt.Printf("   找到 %d 个 upstream 块:\n", len(upstreams))

	for i, upstream := range upstreams {
		fmt.Printf("   %d. %s (服务器数量: %d)\n",
			i+1, upstream.UpstreamName, len(upstream.UpstreamServers))

		for j, server := range upstream.UpstreamServers {
			fmt.Printf("      - 服务器 %d: %s", j+1, server.Address)

			// 显示参数
			if len(server.Parameters) > 0 {
				fmt.Print(" (")
				first := true
				for key, value := range server.Parameters {
					if !first {
						fmt.Print(", ")
					}
					fmt.Printf("%s=%s", key, value)
					first = false
				}
				fmt.Print(")")
			}
			fmt.Println()
		}
	}
}

func readDirectivesExample(conf *config.Config) {
	fmt.Println("📖 读取特定指令")

	// 读取 worker_processes 指令
	workerProcesses := conf.FindDirectives("worker_processes")
	if len(workerProcesses) > 0 {
		params := workerProcesses[0].GetParameters()
		if len(params) > 0 {
			fmt.Printf("   worker_processes: %s\n", params[0].GetValue())
		}
	}

	// 读取所有 proxy_pass 指令
	proxyPasses := conf.FindDirectives("proxy_pass")
	fmt.Printf("   找到 %d 个 proxy_pass 指令:\n", len(proxyPasses))
	for i, directive := range proxyPasses {
		params := directive.GetParameters()
		if len(params) > 0 {
			fmt.Printf("   %d. %s\n", i+1, params[0].GetValue())
		}
	}
}

func advancedSearchExample(conf *config.Config) {
	fmt.Println("📖 高级搜索示例")

	// 按名称查找 upstream
	upstream := conf.FindUpstreamByName("backend")
	if upstream != nil {
		fmt.Printf("   找到 upstream 'backend'，包含 %d 个服务器\n", len(upstream.UpstreamServers))
	}

	// 按名称查找 server
	servers := conf.FindServersByName("example.com")
	fmt.Printf("   找到 %d 个名为 'example.com' 的服务器\n", len(servers))

	// 按模式查找 location
	locations := conf.FindLocationsByPattern("/")
	fmt.Printf("   找到 %d 个匹配 '/' 的 location\n", len(locations))

	// 获取所有 upstream 服务器
	allServers := conf.GetAllUpstreamServers()
	fmt.Printf("   总共有 %d 个 upstream 服务器\n", len(allServers))
}

// ===========================
// UPDATE (更新) 操作示例
// ===========================

func updateUpstreamExample(conf *config.Config) {
	fmt.Println("✏️ 更新 upstream 服务器")

	// 找到指定的 upstream
	upstream := conf.FindUpstreamByName("backend")
	if upstream != nil && len(upstream.UpstreamServers) > 0 {
		// 更新第一个服务器的参数
		oldAddress := upstream.UpstreamServers[0].Address
		upstream.UpstreamServers[0].Parameters["weight"] = "5"
		upstream.UpstreamServers[0].Parameters["max_fails"] = "3"

		fmt.Printf("   ✅ 更新服务器 %s 的参数 (weight=5, max_fails=3)\n", oldAddress)

		// 添加新的服务器
		upstream.AddServer(&config.UpstreamServer{
			Address: "192.168.1.20:8080",
			Parameters: map[string]string{
				"weight": "1",
				"backup": "",
			},
		})
		fmt.Println("   ✅ 添加新服务器: 192.168.1.20:8080 (backup)")
	}
}

func updateDirectiveExample(conf *config.Config) {
	fmt.Println("✏️ 更新指令值")

	// 更新 keepalive_timeout
	httpBlocks := conf.FindDirectives("http")
	if len(httpBlocks) > 0 {
		if httpBlock, ok := httpBlocks[0].(*config.HTTP); ok {
			keepaliveDirectives := httpBlock.FindDirectives("keepalive_timeout")
			if len(keepaliveDirectives) > 0 {
				directive := keepaliveDirectives[0].(*config.Directive)
				oldValue := directive.Parameters[0].GetValue()
				directive.Parameters[0] = config.NewParameter("120")
				fmt.Printf("   ✅ 更新 keepalive_timeout: %s -> 120\n", oldValue)
			}
		}
	}
}

// ===========================
// DELETE (删除) 操作示例
// ===========================

func deleteUpstreamServerExample(conf *config.Config) {
	fmt.Println("🗑️ 删除 upstream 服务器")

	upstream := conf.FindUpstreamByName("backend")
	if upstream != nil && len(upstream.UpstreamServers) > 1 {
		// 删除最后一个服务器
		deletedServer := upstream.UpstreamServers[len(upstream.UpstreamServers)-1]
		upstream.UpstreamServers = upstream.UpstreamServers[:len(upstream.UpstreamServers)-1]
		fmt.Printf("   ✅ 删除服务器: %s\n", deletedServer.Address)
	}
}

func deleteDirectiveExample(conf *config.Config) {
	fmt.Println("🗑️ 删除指令")

	// 从 http 块中删除 sendfile 指令
	httpBlocks := conf.FindDirectives("http")
	if len(httpBlocks) > 0 {
		if httpBlock, ok := httpBlocks[0].(*config.HTTP); ok {
			for i, directive := range httpBlock.Directives {
				if directive.GetName() == "sendfile" {
					// 删除这个指令
					httpBlock.Directives = append(
						httpBlock.Directives[:i],
						httpBlock.Directives[i+1:]...,
					)
					fmt.Println("   ✅ 删除 'sendfile' 指令")
					break
				}
			}
		}
	}
}

// ===========================
// 辅助函数
// ===========================

func saveToFile(content, filename string) {
	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		fmt.Printf("   ❌ 保存文件失败: %v\n", err)
	} else {
		fmt.Printf("   ✅ 配置已保存到文件: %s\n", filename)
	}
}
