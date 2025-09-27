package main

import (
	"fmt"
	"strings"

	"github.com/lefeck/gonginx/config"
	"github.com/lefeck/gonginx/dumper"
	"github.com/lefeck/gonginx/generator"
)

// 示例2：配置生成 - 使用 Builder 模式创建 nginx 配置
func main() {
	fmt.Println("=== 示例2：使用 Builder 模式生成 nginx 配置 ===")

	builderConfig := generator.NewConfigBuilder().
		WorkerProcesses("auto").
		WorkerConnections("1024").
		HTTP().
		// 添加基础配置
		SendFile(true).
		TCPNoPush("on").
		KeepaliveTimeout("65").
		// 添加 upstream
		Upstream("api_backend").
		Server("10.0.1.10:8080", "weight=3").
		Server("10.0.1.11:8080", "weight=2").
		Server("10.0.1.12:8080", "backup").
		End().
		// 添加主服务器
		Server().
		Listen("80").
		ServerName("api.example.com").
		Location("/").
		ProxyPass("http://api_backend").
		ProxySetHeader("Host", "$host").
		ProxySetHeader("X-Real-IP", "$remote_addr").
		End().
		Location("/health").
		Return("200", "\"healthy\"").
		End().
		End().
		// 添加 HTTPS 服务器
		Server().
		Listen("443", "ssl", "http2").
		ServerName("api.example.com").
		SSL().
		Certificate("/etc/ssl/certs/api.crt").
		CertificateKey("/etc/ssl/private/api.key").
		Protocols("TLSv1.2", "TLSv1.3").
		End().
		Location("/").
		ProxyPass("http://api_backend").
		End().
		End().
		End().
		Build()

	if builderConfig != nil {
		fmt.Println("✅ Builder 配置创建成功")
		output := dumper.DumpConfig(builderConfig, dumper.IndentedStyle)
		fmt.Println(output)
	}

	fmt.Println("\n" + strings.Repeat("=", 50))

	// 方法2：手动创建配置对象
	fmt.Println("🔨 方法2：手动创建配置对象")

	manualConfig := createManualConfig()
	if manualConfig != nil {
		fmt.Println("✅ 手动配置创建成功")
		output := dumper.DumpConfig(manualConfig, dumper.IndentedStyle)
		fmt.Println(output)
	}

	fmt.Println("\n" + strings.Repeat("=", 50))

	// 方法3：使用模板生成特定类型的配置
	fmt.Println("🔨 方法3：使用模板生成反向代理配置")

	templateConfig := createReverseProxyTemplate()
	if templateConfig != nil {
		fmt.Println("✅ 模板配置创建成功")
		output := dumper.DumpConfig(templateConfig, dumper.IndentedStyle)
		fmt.Println(output)
	}
}

// 手动创建配置的辅助函数
func createManualConfig() *config.Config {
	// 创建主配置
	conf := &config.Config{
		Block: &config.Block{
			Directives: []config.IDirective{},
		},
		FilePath: "generated.conf",
	}

	// 添加全局指令
	conf.Block.Directives = append(conf.Block.Directives, &config.Directive{
		Name:       "worker_processes",
		Parameters: []config.Parameter{{Value: "auto"}},
	})

	// 创建 events 块
	eventsBlock := &config.Directive{
		Name: "events",
		Block: &config.Block{
			Directives: []config.IDirective{
				&config.Directive{
					Name:       "worker_connections",
					Parameters: []config.Parameter{{Value: "1024"}},
				},
				&config.Directive{
					Name:       "use",
					Parameters: []config.Parameter{{Value: "epoll"}},
				},
			},
		},
	}
	conf.Block.Directives = append(conf.Block.Directives, eventsBlock)

	// 创建 http 块
	httpBlock := &config.Directive{
		Name: "http",
		Block: &config.Block{
			Directives: []config.IDirective{
				&config.Directive{
					Name:       "include",
					Parameters: []config.Parameter{{Value: "/etc/nginx/mime.types"}},
				},
				&config.Directive{
					Name:       "default_type",
					Parameters: []config.Parameter{{Value: "application/octet-stream"}},
				},
			},
		},
	}

	// 添加 upstream
	upstreamBlock := &config.Directive{
		Name:       "upstream",
		Parameters: []config.Parameter{{Value: "web_backend"}},
		Block: &config.Block{
			Directives: []config.IDirective{
				&config.Directive{
					Name:       "server",
					Parameters: []config.Parameter{{Value: "192.168.1.100:8080"}},
				},
				&config.Directive{
					Name:       "server",
					Parameters: []config.Parameter{{Value: "192.168.1.101:8080"}},
				},
			},
		},
	}
	httpBlock.Block.(*config.Block).Directives = append(httpBlock.Block.(*config.Block).Directives, upstreamBlock)

	// 添加 server 块
	serverBlock := &config.Directive{
		Name: "server",
		Block: &config.Block{
			Directives: []config.IDirective{
				&config.Directive{
					Name:       "listen",
					Parameters: []config.Parameter{{Value: "80"}},
				},
				&config.Directive{
					Name:       "server_name",
					Parameters: []config.Parameter{{Value: "web.example.com"}},
				},
				&config.Directive{
					Name:       "location",
					Parameters: []config.Parameter{{Value: "/"}},
					Block: &config.Block{
						Directives: []config.IDirective{
							&config.Directive{
								Name:       "proxy_pass",
								Parameters: []config.Parameter{{Value: "http://web_backend"}},
							},
						},
					},
				},
			},
		},
	}
	httpBlock.Block.(*config.Block).Directives = append(httpBlock.Block.(*config.Block).Directives, serverBlock)

	conf.Block.Directives = append(conf.Block.Directives, httpBlock)
	return conf
}

// 创建反向代理模板配置
func createReverseProxyTemplate() *config.Config {
	// 这里模拟一个反向代理模板的创建
	// 实际使用中，你可能需要查看 gonginx 是否提供了具体的模板功能

	conf := &config.Config{
		Block: &config.Block{
			Directives: []config.IDirective{},
		},
		FilePath: "reverse_proxy.conf",
	}

	// 添加优化的反向代理配置
	httpBlock := &config.Directive{
		Name: "http",
		Block: &config.Block{
			Directives: []config.IDirective{
				// 基础配置
				&config.Directive{
					Name:       "sendfile",
					Parameters: []config.Parameter{{Value: "on"}},
				},
				&config.Directive{
					Name:       "tcp_nopush",
					Parameters: []config.Parameter{{Value: "on"}},
				},
				&config.Directive{
					Name:       "tcp_nodelay",
					Parameters: []config.Parameter{{Value: "on"}},
				},
				// 代理缓冲区设置
				&config.Directive{
					Name:       "proxy_buffering",
					Parameters: []config.Parameter{{Value: "on"}},
				},
				&config.Directive{
					Name:       "proxy_buffer_size",
					Parameters: []config.Parameter{{Value: "4k"}},
				},
				&config.Directive{
					Name:       "proxy_buffers",
					Parameters: []config.Parameter{{Value: "8"}, {Value: "4k"}},
				},
				// upstream 定义
				&config.Directive{
					Name:       "upstream",
					Parameters: []config.Parameter{{Value: "app_servers"}},
					Block: &config.Block{
						Directives: []config.IDirective{
							&config.Directive{
								Name:       "least_conn",
								Parameters: []config.Parameter{},
							},
							&config.Directive{
								Name:       "server",
								Parameters: []config.Parameter{{Value: "app1.internal:8080"}, {Value: "max_fails=3"}, {Value: "fail_timeout=30s"}},
							},
							&config.Directive{
								Name:       "server",
								Parameters: []config.Parameter{{Value: "app2.internal:8080"}, {Value: "max_fails=3"}, {Value: "fail_timeout=30s"}},
							},
						},
					},
				},
				// 服务器配置
				&config.Directive{
					Name: "server",
					Block: &config.Block{
						Directives: []config.IDirective{
							&config.Directive{
								Name:       "listen",
								Parameters: []config.Parameter{{Value: "80"}},
							},
							&config.Directive{
								Name:       "server_name",
								Parameters: []config.Parameter{{Value: "proxy.example.com"}},
							},
							// 主要 location
							&config.Directive{
								Name:       "location",
								Parameters: []config.Parameter{{Value: "/"}},
								Block: &config.Block{
									Directives: []config.IDirective{
										&config.Directive{
											Name:       "proxy_pass",
											Parameters: []config.Parameter{{Value: "http://app_servers"}},
										},
										&config.Directive{
											Name:       "proxy_set_header",
											Parameters: []config.Parameter{{Value: "Host"}, {Value: "$host"}},
										},
										&config.Directive{
											Name:       "proxy_set_header",
											Parameters: []config.Parameter{{Value: "X-Real-IP"}, {Value: "$remote_addr"}},
										},
										&config.Directive{
											Name:       "proxy_set_header",
											Parameters: []config.Parameter{{Value: "X-Forwarded-For"}, {Value: "$proxy_add_x_forwarded_for"}},
										},
										&config.Directive{
											Name:       "proxy_set_header",
											Parameters: []config.Parameter{{Value: "X-Forwarded-Proto"}, {Value: "$scheme"}},
										},
										&config.Directive{
											Name:       "proxy_connect_timeout",
											Parameters: []config.Parameter{{Value: "30s"}},
										},
										&config.Directive{
											Name:       "proxy_send_timeout",
											Parameters: []config.Parameter{{Value: "30s"}},
										},
										&config.Directive{
											Name:       "proxy_read_timeout",
											Parameters: []config.Parameter{{Value: "30s"}},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	conf.Block.Directives = append(conf.Block.Directives, httpBlock)
	return conf
}
