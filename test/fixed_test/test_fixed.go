package main

import (
	"fmt"
	"strings"

	"gonginx/dumper"
	"gonginx/generator"
)

// 测试修复后的 Builder 功能
func main() {
	fmt.Println("=== 测试修复后的 Builder 功能 ===")

	// 使用修复后的 ConfigBuilder 流式 API
	fmt.Println("🔨 测试修复后的链式调用")

	builderConfig := generator.NewConfigBuilder().
		WorkerProcesses("auto").
		WorkerConnections("1024").
		HTTP().
		// 添加基础配置 - 现在这些方法应该可以工作了
		SendFile(true).
		TCPNoPush("on").        // 修复：现在可以使用 TCPNoPush
		KeepaliveTimeout("65"). // 修复：这个方法已经存在
		// 添加 upstream
		Upstream("api_backend").
		Server("10.0.1.10:8080", "weight=3").
		Server("10.0.1.11:8080", "weight=2").
		Server("10.0.1.12:8080", "backup").
		End(). // 修复：现在 End() 返回 HTTPBuilder，可以继续链式调用
		// 添加主服务器 - 现在这应该可以工作了
		Server().
		Listen("80").
		ServerName("api.example.com").
		Location("/").
		ProxyPass("http://api_backend").
		ProxySetHeader("Host", "$host").
		ProxySetHeader("X-Real-IP", "$remote_addr").
		End(). // 返回到 ServerBuilder
		Location("/health").
		Return("200", "\"healthy\"").
		End(). // 返回到 ServerBuilder
		End(). // 返回到 HTTPBuilder
		// 添加另一个 upstream
		Upstream("web_backend").
		Server("192.168.1.100:8080").
		Server("192.168.1.101:8080").
		End(). // 返回到 HTTPBuilder
		// 添加 HTTPS 服务器
		Server().
		Listen("443", "ssl", "http2").
		ServerName("api.example.com").
		SSL().
		Certificate("/etc/ssl/certs/api.crt").
		CertificateKey("/etc/ssl/private/api.key").
		Protocols("TLSv1.2", "TLSv1.3").
		EndSSL(). // 返回到 ServerBuilder
		Location("/").
		ProxyPass("http://api_backend").
		End(). // 返回到 ServerBuilder
		End(). // 返回到 HTTPBuilder
		End(). // 返回到 ConfigBuilder
		Build()

	if builderConfig != nil {
		fmt.Println("✅ 修复后的 Builder 配置创建成功")
		output := dumper.DumpConfig(builderConfig, dumper.IndentedStyle)
		fmt.Println(output)

		// 验证配置中包含了我们期望的内容
		fmt.Println("\n🔍 验证配置内容:")
		if strings.Contains(output, "tcp_nopush on") {
			fmt.Println("✅ TCPNoPush 方法工作正常")
		} else {
			fmt.Println("❌ TCPNoPush 方法有问题")
		}

		if strings.Contains(output, "keepalive_timeout 65") {
			fmt.Println("✅ KeepaliveTimeout 方法工作正常")
		} else {
			fmt.Println("❌ KeepaliveTimeout 方法有问题")
		}

		if strings.Contains(output, "upstream api_backend") && strings.Contains(output, "upstream web_backend") {
			fmt.Println("✅ 多个 upstream 块创建成功")
		} else {
			fmt.Println("❌ upstream 块创建有问题")
		}

		serverCount := strings.Count(output, "server {")
		fmt.Printf("✅ 创建了 %d 个 server 块\n", serverCount)

	} else {
		fmt.Println("❌ Builder 配置创建失败")
	}

	fmt.Println("\n" + strings.Repeat("=", 50))

	// 测试更复杂的链式调用
	fmt.Println("🔨 测试复杂的链式调用场景")

	complexConfig := generator.NewConfigBuilder().
		WorkerProcesses("auto").
		ErrorLog("/var/log/nginx/error.log", "warn").
		HTTP().
		Include("/etc/nginx/mime.types").
		DefaultType("application/octet-stream").
		SendFile(true).
		TCPNoPush("on").
		TcpNoDelay(true).
		KeepaliveTimeout("65").
		ClientMaxBodySize("100M").
		Gzip(true).
		GzipTypes("text/plain", "text/css", "application/json", "application/javascript").
		// 第一个 upstream
		Upstream("backend_pool_1").
		IpHash().
		Server("10.0.1.1:8080", "max_fails=3", "fail_timeout=30s").
		Server("10.0.1.2:8080", "max_fails=3", "fail_timeout=30s").
		KeepaliveConnections("32").
		End().
		// 第二个 upstream
		Upstream("backend_pool_2").
		LeastConn().
		Server("10.0.2.1:8080", "weight=3").
		Server("10.0.2.2:8080", "weight=2").
		Server("10.0.2.3:8080", "backup").
		End().
		// 第一个服务器
		Server().
		Listen("80").
		ServerName("www.example.com", "example.com").
		Root("/var/www/html").
		Index("index.html", "index.htm").
		AccessLog("/var/log/nginx/access.log").
		Location("/").
		TryFiles("$uri", "$uri/", "/index.html").
		End().
		Location("/api/v1").
		ProxyPass("http://backend_pool_1").
		ProxySetHeader("Host", "$host").
		ProxySetHeader("X-Real-IP", "$remote_addr").
		ProxySetHeader("X-Forwarded-For", "$proxy_add_x_forwarded_for").
		End().
		Location("/api/v2").
		ProxyPass("http://backend_pool_2").
		ProxyTimeout("30s").
		End().
		End().
		// 第二个服务器 (HTTPS)
		Server().
		Listen("443", "ssl", "http2").
		ServerName("secure.example.com").
		SSL().
		Certificate("/etc/ssl/certs/secure.crt").
		CertificateKey("/etc/ssl/private/secure.key").
		Protocols("TLSv1.2", "TLSv1.3").
		SessionTimeout("1d").
		SessionCache("shared:SSL:50m").
		HSTS("31536000", true).
		EndSSL().
		Location("/").
		ProxyPass("http://backend_pool_1").
		End().
		End().
		End().
		Build()

	if complexConfig != nil {
		fmt.Println("✅ 复杂链式调用配置创建成功")
		output := dumper.DumpConfig(complexConfig, dumper.IndentedStyle)

		// 只显示配置的前几行来验证结构
		lines := strings.Split(output, "\n")
		fmt.Println("📋 配置结构预览 (前20行):")
		for i, line := range lines {
			if i >= 20 {
				fmt.Println("... (更多内容)")
				break
			}
			fmt.Println(line)
		}

		// 统计验证
		fmt.Println("\n📊 配置统计:")
		fmt.Printf("   - Upstream 块数量: %d\n", strings.Count(output, "upstream "))
		fmt.Printf("   - Server 块数量: %d\n", strings.Count(output, "server {"))
		fmt.Printf("   - Location 块数量: %d\n", strings.Count(output, "location "))
		fmt.Printf("   - SSL 配置: %s\n", map[bool]string{true: "已启用", false: "未启用"}[strings.Contains(output, "ssl_certificate")])

	} else {
		fmt.Println("❌ 复杂链式调用配置创建失败")
	}

	fmt.Println("\n=== 测试完成 ===")
}
