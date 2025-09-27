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

// 基础的 nginx 配置 CRUD 操作示例
func main() {
	fmt.Println("=== Nginx 配置基础 CRUD 操作示例 ===")

	// 准备一个示例配置文件
	sampleConfig := `
worker_processes auto;

http {
    sendfile on;
    keepalive_timeout 65;
    
    upstream backend {
        server 192.168.1.10:8080;
        server 192.168.1.11:8080;
    }
    
    server {
        listen 80;
        server_name example.com;
        
        location / {
            proxy_pass http://backend;
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
	fmt.Println("🔨 1. CREATE (创建) 操作")
	fmt.Println("----------------------------------------")
	createOperations(conf)
	fmt.Println()

	// ===================
	// 2. READ (读取) 操作
	// ===================
	fmt.Println("🔍 2. READ (读取) 操作")
	fmt.Println("----------------------------------------")
	readOperations(conf)
	fmt.Println()

	// ===================
	// 3. UPDATE (更新) 操作
	// ===================
	fmt.Println("✏️ 3. UPDATE (更新) 操作")
	fmt.Println("----------------------------------------")
	updateOperations(conf)
	fmt.Println()

	// ===================
	// 4. DELETE (删除) 操作
	// ===================
	fmt.Println("🗑️ 4. DELETE (删除) 操作")
	fmt.Println("----------------------------------------")
	deleteOperations(conf)
	fmt.Println()

	// ===================
	// 5. 保存配置
	// ===================
	fmt.Println("💾 5. 保存配置")
	fmt.Println("----------------------------------------")
	finalConfig := dumper.DumpConfig(conf, dumper.IndentedStyle)
	fmt.Println("修改后的配置:")
	fmt.Println(finalConfig)

	// 保存到文件
	err = os.WriteFile("nginx_modified.conf", []byte(finalConfig), 0644)
	if err != nil {
		fmt.Printf("   ❌ 保存失败: %v\n", err)
	} else {
		fmt.Println("   ✅ 配置已保存到 nginx_modified.conf")
	}

	fmt.Println("\n=== CRUD 操作完成 ===")
}

// CREATE 操作示例
func createOperations(conf *config.Config) {
	// 1. 创建新的全局指令
	fmt.Println("📝 创建全局指令 'error_log /var/log/nginx/error.log'")
	errorLogDirective := &config.Directive{
		Name:       "error_log",
		Parameters: []config.Parameter{config.NewParameter("/var/log/nginx/error.log")},
	}
	conf.Block.Directives = append(conf.Block.Directives, errorLogDirective)
	fmt.Println("   ✅ 成功创建全局指令")

	// 2. 在 http 块中创建新指令
	fmt.Println("📝 在 http 块中创建 'gzip on' 指令")
	httpBlocks := conf.FindDirectives("http")
	if len(httpBlocks) > 0 {
		if httpBlock, ok := httpBlocks[0].(*config.HTTP); ok {
			gzipDirective := &config.Directive{
				Name:       "gzip",
				Parameters: []config.Parameter{config.NewParameter("on")},
			}
			httpBlock.Directives = append(httpBlock.Directives, gzipDirective)
			fmt.Println("   ✅ 成功创建 gzip 指令")
		}
	}

	// 3. 创建新的 upstream 块
	fmt.Println("📝 创建新的 upstream 块 'api_backend'")
	upstreamDirective := &config.Directive{
		Name:       "upstream",
		Parameters: []config.Parameter{config.NewParameter("api_backend")},
		Block: &config.Block{
			Directives: []config.IDirective{
				&config.Directive{
					Name:       "server",
					Parameters: []config.Parameter{config.NewParameter("10.0.1.1:9000")},
				},
				&config.Directive{
					Name:       "server",
					Parameters: []config.Parameter{config.NewParameter("10.0.1.2:9000")},
				},
			},
		},
	}

	if len(httpBlocks) > 0 {
		if httpBlock, ok := httpBlocks[0].(*config.HTTP); ok {
			httpBlock.Directives = append(httpBlock.Directives, upstreamDirective)
			fmt.Println("   ✅ 成功创建 upstream 'api_backend'")
		}
	}
}

// READ 操作示例
func readOperations(conf *config.Config) {
	// 1. 读取全局指令
	fmt.Println("📖 读取全局指令")
	workerProcesses := conf.FindDirectives("worker_processes")
	if len(workerProcesses) > 0 {
		params := workerProcesses[0].GetParameters()
		if len(params) > 0 {
			fmt.Printf("   worker_processes: %s\n", params[0].GetValue())
		}
	}

	// 2. 读取 http 块中的指令
	fmt.Println("📖 读取 http 块中的指令")
	httpBlocks := conf.FindDirectives("http")
	if len(httpBlocks) > 0 {
		if httpBlock, ok := httpBlocks[0].(*config.HTTP); ok {
			fmt.Printf("   http 块包含 %d 个指令\n", len(httpBlock.Directives))

			// 列出所有指令名称
			var directiveNames []string
			for _, directive := range httpBlock.Directives {
				directiveNames = append(directiveNames, directive.GetName())
			}
			fmt.Printf("   指令列表: %s\n", strings.Join(directiveNames, ", "))
		}
	}

	// 3. 读取 upstream 块
	fmt.Println("📖 读取 upstream 块")
	upstreamDirectives := conf.FindDirectives("upstream")
	fmt.Printf("   找到 %d 个 upstream 块\n", len(upstreamDirectives))

	for i, upstream := range upstreamDirectives {
		params := upstream.GetParameters()
		if len(params) > 0 {
			fmt.Printf("   %d. %s\n", i+1, params[0].GetValue())

			// 读取 upstream 中的服务器
			if upstream.GetBlock() != nil {
				servers := upstream.GetBlock().FindDirectives("server")
				fmt.Printf("      包含 %d 个服务器:\n", len(servers))
				for j, server := range servers {
					serverParams := server.GetParameters()
					if len(serverParams) > 0 {
						fmt.Printf("        %d. %s\n", j+1, serverParams[0].GetValue())
					}
				}
			}
		}
	}

	// 4. 读取 server 块
	fmt.Println("📖 读取 server 块")
	serverDirectives := conf.FindDirectives("server")
	fmt.Printf("   找到 %d 个 server 块\n", len(serverDirectives))

	for i, server := range serverDirectives {
		fmt.Printf("   %d. server 块:\n", i+1)
		if server.GetBlock() != nil {
			// 读取 listen 指令
			listens := server.GetBlock().FindDirectives("listen")
			for _, listen := range listens {
				params := listen.GetParameters()
				if len(params) > 0 {
					fmt.Printf("      listen: %s\n", params[0].GetValue())
				}
			}

			// 读取 server_name 指令
			serverNames := server.GetBlock().FindDirectives("server_name")
			for _, serverName := range serverNames {
				params := serverName.GetParameters()
				if len(params) > 0 {
					fmt.Printf("      server_name: %s\n", params[0].GetValue())
				}
			}
		}
	}

	// 5. 使用高级搜索（如果可用）
	fmt.Println("📖 高级搜索")

	// 搜索所有 proxy_pass 指令
	proxyPasses := conf.FindDirectives("proxy_pass")
	fmt.Printf("   找到 %d 个 proxy_pass 指令\n", len(proxyPasses))
	for i, proxyPass := range proxyPasses {
		params := proxyPass.GetParameters()
		if len(params) > 0 {
			fmt.Printf("   %d. %s\n", i+1, params[0].GetValue())
		}
	}
}

// UPDATE 操作示例
func updateOperations(conf *config.Config) {
	// 1. 更新全局指令
	fmt.Println("✏️ 更新全局指令")
	workerProcesses := conf.FindDirectives("worker_processes")
	if len(workerProcesses) > 0 {
		directive := workerProcesses[0].(*config.Directive)
		oldValue := directive.Parameters[0].GetValue()
		directive.Parameters[0] = config.NewParameter("4")
		fmt.Printf("   ✅ 更新 worker_processes: %s -> 4\n", oldValue)
	}

	// 2. 更新 http 块中的指令
	fmt.Println("✏️ 更新 http 块中的指令")
	httpBlocks := conf.FindDirectives("http")
	if len(httpBlocks) > 0 {
		if httpBlock, ok := httpBlocks[0].(*config.HTTP); ok {
			// 查找 keepalive_timeout 指令
			for _, directive := range httpBlock.Directives {
				if directive.GetName() == "keepalive_timeout" {
					if dir, ok := directive.(*config.Directive); ok {
						oldValue := dir.Parameters[0].GetValue()
						dir.Parameters[0] = config.NewParameter("120")
						fmt.Printf("   ✅ 更新 keepalive_timeout: %s -> 120\n", oldValue)
						break
					}
				}
			}
		}
	}

	// 3. 在 upstream 中添加新服务器
	fmt.Println("✏️ 在 upstream 中添加新服务器")
	upstreamDirectives := conf.FindDirectives("upstream")
	for _, upstream := range upstreamDirectives {
		params := upstream.GetParameters()
		if len(params) > 0 && params[0].GetValue() == "backend" {
			if upstream.GetBlock() != nil {
				newServer := &config.Directive{
					Name:       "server",
					Parameters: []config.Parameter{config.NewParameter("192.168.1.20:8080")},
				}
				upstream.GetBlock().(*config.Block).Directives = append(
					upstream.GetBlock().(*config.Block).Directives,
					newServer,
				)
				fmt.Println("   ✅ 添加新服务器: 192.168.1.20:8080")
			}
			break
		}
	}
}

// DELETE 操作示例
func deleteOperations(conf *config.Config) {
	// 1. 删除 http 块中的指令
	fmt.Println("🗑️ 删除 http 块中的指令")
	httpBlocks := conf.FindDirectives("http")
	if len(httpBlocks) > 0 {
		if httpBlock, ok := httpBlocks[0].(*config.HTTP); ok {
			// 删除 sendfile 指令
			for i, directive := range httpBlock.Directives {
				if directive.GetName() == "sendfile" {
					httpBlock.Directives = append(
						httpBlock.Directives[:i],
						httpBlock.Directives[i+1:]...,
					)
					fmt.Println("   ✅ 删除 sendfile 指令")
					break
				}
			}
		}
	}

	// 2. 从 upstream 中删除服务器
	fmt.Println("🗑️ 从 upstream 中删除服务器")
	upstreamDirectives := conf.FindDirectives("upstream")
	for _, upstream := range upstreamDirectives {
		params := upstream.GetParameters()
		if len(params) > 0 && params[0].GetValue() == "backend" {
			if upstream.GetBlock() != nil {
				servers := upstream.GetBlock().(*config.Block).Directives
				if len(servers) > 1 {
					// 删除最后一个服务器
					lastServer := servers[len(servers)-1]
					upstream.GetBlock().(*config.Block).Directives = servers[:len(servers)-1]

					// 获取被删除服务器的地址
					if serverParams := lastServer.GetParameters(); len(serverParams) > 0 {
						fmt.Printf("   ✅ 删除服务器: %s\n", serverParams[0].GetValue())
					}
				}
			}
			break
		}
	}
}
