# 配置生成器示例

这个示例展示了 gonginx 库的配置生成器功能，包括 Builder 模式和预定义模板。

## 功能特性

### Builder 模式
- **流式接口**: 使用链式调用构建配置
- **类型安全**: 编译时检查配置正确性
- **易于使用**: 直观的 API 设计
- **可扩展**: 支持自定义指令和块

### 预定义模板
1. **Basic Web Server** - 基础静态文件服务器
2. **Reverse Proxy** - 反向代理配置
3. **Load Balancer** - 负载均衡器
4. **SSL Web Server** - SSL/TLS 安全服务器
5. **Static File Server** - 优化的静态文件服务器
6. **PHP Web Server** - PHP 应用服务器
7. **Stream Proxy** - TCP/UDP 流代理
8. **Microservices Gateway** - 微服务 API 网关

## Builder API 结构

### 主要构建器
- **ConfigBuilder**: 主配置构建器
- **HTTPBuilder**: HTTP 块构建器
- **StreamBuilder**: Stream 块构建器
- **ServerBuilder**: Server 块构建器
- **LocationBuilder**: Location 块构建器
- **UpstreamBuilder**: Upstream 块构建器
- **SSLBuilder**: SSL 配置构建器

### 使用模式

#### 基础配置
```go
config := generator.NewConfigBuilder().
    WorkerProcesses("auto").
    WorkerConnections("1024").
    HTTP().
    SendFile(true).
    Gzip(true).
    Server().
    Listen("80").
    ServerName("example.com").
    Root("/var/www/html").
    EndServer().
    End().
    Build()
```

#### 反向代理配置
```go
config := generator.NewConfigBuilder().
    HTTP().
    Upstream("backend").
    Server("127.0.0.1:8001", "weight=3").
    Server("127.0.0.1:8002", "weight=2").
    EndUpstream().
    Server().
    Listen("80").
    Location("/").
    ProxyPass("http://backend").
    ProxySetHeader("Host", "$host").
    EndLocation().
    EndServer().
    End().
    Build()
```

#### SSL 配置
```go
config := generator.NewConfigBuilder().
    HTTP().
    Server().
    Listen("443", "ssl", "http2").
    SSL().
    Certificate("/path/to/cert.pem").
    CertificateKey("/path/to/key.pem").
    Protocols("TLSv1.2", "TLSv1.3").
    HSTS("31536000", true).
    EndSSL().
    EndServer().
    End().
    Build()
```

#### Stream 配置
```go
config := generator.NewConfigBuilder().
    Stream().
    Upstream("database_pool").
    Server("10.0.1.10:5432", "weight=3").
    Server("10.0.1.11:5432", "weight=2").
    EndUpstream().
    Server().
    Listen("5432").
    ProxyPass("database_pool").
    EndServer().
    End().
    Build()
```

## 预定义模板使用

### 获取所有模板
```go
templates := generator.GetAllTemplates()
for _, template := range templates {
    fmt.Printf("模板: %s - %s\n", template.Name, template.Description)
    config := template.Builder().Build()
    // 使用生成的配置
}
```

### 使用特定模板
```go
// 基础 Web 服务器
template := generator.BasicWebServerTemplate()
config := template.Builder().Build()

// SSL Web 服务器
sslTemplate := generator.SSLWebServerTemplate()
sslConfig := sslTemplate.Builder().Build()

// 微服务网关
gatewayTemplate := generator.MicroservicesGatewayTemplate()
gatewayConfig := gatewayTemplate.Builder().Build()
```

## 运行示例

```bash
cd examples/config-generator
go run main.go
```

## 示例输出

程序将输出：
1. 手动构建的基础 Web 服务器配置
2. 反向代理配置
3. Stream 代理配置
4. 所有预定义模板的列表和示例
5. 复杂的 SSL 配置
6. 动态生成的不同环境配置

## 实际应用场景

1. **DevOps 自动化**: 根据环境动态生成配置
2. **配置管理**: 标准化的配置模板
3. **微服务部署**: 自动生成服务网关配置
4. **负载均衡**: 动态添加/移除后端服务器
5. **SSL 证书管理**: 自动化 SSL 配置
6. **容器化部署**: 为容器应用生成代理配置

## 扩展功能

### 自定义模板
```go
func CustomTemplate() generator.Template {
    return generator.Template{
        Name:        "Custom Template",
        Description: "My custom configuration",
        Builder: func() *generator.ConfigBuilder {
            return generator.NewConfigBuilder().
                // 自定义配置逻辑
                Build()
        },
    }
}
```

### 添加自定义指令
```go
config := generator.NewConfigBuilder().
    HTTP().
    AddDirective("custom_directive", "value1", "value2").
    Server().
    AddDirective("custom_server_directive", "value").
    EndServer().
    End().
    Build()
```
