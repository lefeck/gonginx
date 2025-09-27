package main

import (
	"fmt"
	"log"

	"github.com/lefeck/gonginx/config"
	"github.com/lefeck/gonginx/dumper"
	"github.com/lefeck/gonginx/parser"
)

func main() {
	// Create a simple stream configuration
	streamConfig := `
stream {
    upstream backend {
        server 192.168.1.1:8080 weight=3;
        server 192.168.1.2:8080 weight=2;
        server 192.168.1.3:8080 backup;
    }
    
    server {
        listen 80;
        proxy_pass backend;
        proxy_timeout 3s;
    }
    
    server {
        listen 443;
        proxy_pass 192.168.1.100:8443;
    }
}
`

	fmt.Println("=== 解析 Stream 配置 ===")
	p := parser.NewStringParser(streamConfig)
	conf, err := p.Parse()
	if err != nil {
		log.Fatal("解析错误:", err)
	}

	// 查找所有 stream 块
	streams := conf.FindStreams()
	fmt.Printf("找到 %d 个 stream 块\n", len(streams))

	for i, stream := range streams {
		fmt.Printf("\n--- Stream 块 %d ---\n", i+1)

		// 查找 upstream 块
		upstreams := stream.FindUpstreams()
		fmt.Printf("Upstream 数量: %d\n", len(upstreams))

		for _, upstream := range upstreams {
			fmt.Printf("  Upstream: %s\n", upstream.UpstreamName)
			fmt.Printf("  服务器数量: %d\n", len(upstream.Servers))

			for _, server := range upstream.Servers {
				fmt.Printf("    - %s", server.Address)
				if weight := server.GetWeight(); weight != "" {
					fmt.Printf(" (weight=%s)", weight)
				}
				if server.IsDown() {
					fmt.Printf(" [DOWN]")
				}
				if server.IsBackup() {
					fmt.Printf(" [BACKUP]")
				}
				fmt.Println()
			}
		}

		// 查找 server 块
		servers := stream.FindServers()
		fmt.Printf("Server 数量: %d\n", len(servers))

		for j, server := range servers {
			fmt.Printf("  Server %d:\n", j+1)

			ports := server.GetListenPorts()
			fmt.Printf("    监听端口: %v\n", ports)

			proxyPass := server.GetProxyPass()
			if proxyPass != "" {
				fmt.Printf("    代理到: %s\n", proxyPass)
			}
		}
	}

	// 全局查找功能测试
	fmt.Println("\n=== 全局查找功能测试 ===")

	// 查找所有 stream upstreams
	allStreamUpstreams := conf.FindStreamUpstreams()
	fmt.Printf("所有 Stream Upstream 数量: %d\n", len(allStreamUpstreams))

	// 按名称查找特定 upstream
	backend := conf.FindStreamUpstreamByName("backend")
	if backend != nil {
		fmt.Printf("找到 upstream 'backend', 服务器数量: %d\n", len(backend.Servers))
	}

	// 查找所有 stream servers
	allStreamServers := conf.FindStreamServers()
	fmt.Printf("所有 Stream Server 数量: %d\n", len(allStreamServers))

	// 按监听端口查找服务器
	servers80 := conf.FindStreamServersByListen("80")
	fmt.Printf("监听端口 80 的服务器数量: %d\n", len(servers80))

	// 获取所有 stream upstream servers
	allUpstreamServers := conf.GetAllStreamUpstreamServers()
	fmt.Printf("所有 Stream Upstream 服务器数量: %d\n", len(allUpstreamServers))

	// 生成配置测试
	fmt.Println("\n=== 生成配置测试 ===")

	// 添加新的服务器到 upstream
	if backend != nil {
		newServer := &config.StreamUpstreamServer{
			Directive: &config.Directive{
				Name: "server",
			},
			Address:    "192.168.1.4:8080",
			Parameters: map[string]string{"weight": "1"},
		}
		backend.AddServer(newServer)
		fmt.Println("已添加新服务器到 backend upstream")
	}

	// 输出修改后的配置
	fmt.Println("\n=== 修改后的配置 ===")
	output := dumper.DumpConfig(conf, dumper.IndentedStyle)
	fmt.Println(output)
}
