package main

import (
	"fmt"

	"github.com/lefeck/gonginx/dumper"
	"github.com/lefeck/gonginx/generator"
)

func main() {
	fmt.Println("=== Nginx 配置生成器示例 ===")

	// 示例1: 使用构建器手动创建配置
	fmt.Println("\n1. 手动构建基础 Web 服务器配置:")
	basicConfig := generator.NewConfigBuilder().
		WorkerProcesses("auto").
		ErrorLog("/var/log/nginx/error.log", "warn").
		PidFile("/var/run/nginx.pid").
		WorkerConnections("1024").
		HTTP().
		SendFile(true).
		TcpNoPush(true).
		KeepaliveTimeout("65").
		Gzip(true).
		Server().
		Listen("80").
		ServerName("example.com").
		Root("/var/www/html").
		Index("index.html").
		Location("/").
		TryFiles("$uri", "$uri/", "=404").
		EndLocation().
		EndServer().
		End().
		Build()

	fmt.Println(dumper.DumpConfig(basicConfig, dumper.IndentedStyle))

	// 示例2: 使用构建器创建反向代理配置
	fmt.Println("\n2. 手动构建反向代理配置:")
	proxyConfig := generator.NewConfigBuilder().
		WorkerProcesses("auto").
		WorkerConnections("1024").
		HTTP().
		Upstream("backend").
		Server("127.0.0.1:8001", "weight=3").
		Server("127.0.0.1:8002", "weight=2").
		Server("127.0.0.1:8003", "backup").
		LeastConn().
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
		EndLocation().
		EndServer().
		End().
		Build()

	fmt.Println(dumper.DumpConfig(proxyConfig, dumper.IndentedStyle))

	// 示例3: 使用构建器创建 Stream 配置
	fmt.Println("\n3. 手动构建 Stream 代理配置:")
	streamConfig := generator.NewConfigBuilder().
		WorkerProcesses("auto").
		WorkerConnections("1024").
		Stream().
		Upstream("database_pool").
		LeastConn().
		Server("10.0.1.10:5432", "weight=3").
		Server("10.0.1.11:5432", "weight=2").
		EndUpstream().
		Server().
		Listen("5432").
		ProxyPass("database_pool").
		ProxyTimeout("3s").
		EndServer().
		End().
		Build()

	fmt.Println(dumper.DumpConfig(streamConfig, dumper.IndentedStyle))

	// 示例4: 使用预定义模板
	fmt.Println("\n4. 使用预定义模板:")
	templates := generator.GetAllTemplates()

	for i, template := range templates {
		fmt.Printf("\n--- 模板 %d: %s ---\n", i+1, template.Name)
		fmt.Printf("描述: %s\n", template.Description)

		if i < 3 { // 只显示前3个模板的内容，避免输出过长
			templateConfig := template.Builder().Build()
			fmt.Println("生成的配置:")
			fmt.Println(dumper.DumpConfig(templateConfig, dumper.IndentedStyle))
		} else {
			fmt.Println("(配置内容省略...)")
		}
	}

	// 示例5: 复杂的 SSL 配置
	fmt.Println("\n5. 复杂的 SSL Web 服务器配置:")
	sslConfig := generator.NewConfigBuilder().
		WorkerProcesses("auto").
		ErrorLog("/var/log/nginx/error.log", "warn").
		WorkerConnections("1024").
		HTTP().
		SendFile(true).
		Gzip(true).
		GzipTypes("text/plain", "text/css", "application/json").
		Server().
		Listen("80").
		ServerName("example.com").
		Return("301", "https://$server_name$request_uri").
		EndServer().
		Server().
		Listen("443", "ssl", "http2").
		ServerName("example.com").
		Root("/var/www/html").
		SSL().
		Certificate("/etc/ssl/certs/example.com.crt").
		CertificateKey("/etc/ssl/private/example.com.key").
		Protocols("TLSv1.2", "TLSv1.3").
		Ciphers("ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384").
		PreferServerCiphers(true).
		SessionTimeout("1d").
		HSTS("31536000", true).
		EndSSL().
		Location("/").
		TryFiles("$uri", "$uri/", "=404").
		EndLocation().
		Location("~*", "\\.(css|js|png|jpg|jpeg|gif|ico)$").
		AddDirective("expires", "1y").
		AddDirective("add_header", "Cache-Control", "public").
		EndLocation().
		EndServer().
		End().
		Build()

	fmt.Println(dumper.DumpConfig(sslConfig, dumper.IndentedStyle))

	// 示例6: 动态构建不同类型的配置
	fmt.Println("\n6. 动态配置生成示例:")

	// 根据不同条件生成不同配置
	environments := []string{"development", "production"}

	for _, env := range environments {
		fmt.Printf("\n--- %s 环境配置 ---\n", env)

		builder := generator.NewConfigBuilder().
			WorkerProcesses("auto").
			WorkerConnections("1024")

		if env == "development" {
			builder = builder.ErrorLog("/var/log/nginx/error.log", "debug")
		} else {
			builder = builder.ErrorLog("/var/log/nginx/error.log", "error")
		}

		config := builder.
			HTTP().
			Gzip(true).
			Server().
			Listen("80").
			ServerName(fmt.Sprintf("%s.example.com", env)).
			Root(fmt.Sprintf("/var/www/%s", env)).
			Index("index.html").
			Location("/").
			TryFiles("$uri", "$uri/", "=404").
			EndLocation().
			EndServer().
			End().
			Build()

		fmt.Println(dumper.DumpConfig(config, dumper.IndentedStyle))
	}

	fmt.Println("\n=== 配置生成器示例完成 ===")
}
