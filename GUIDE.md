# Gonginx 完整使用指南

Gonginx 是一个功能强大的 Go 语言 nginx 配置文件处理库，提供解析、验证、修改、生成和优化 nginx 配置的完整解决方案。

## 📚 文档导航

- **[快速开始](#快速开始)** - 5分钟上手指南
- **[API 参考](API_REFERENCE.md)** - 完整的 API 文档
- **[功能详解](doc.md)** - 所有功能的详细说明
- **[示例代码](examples/)** - 各种使用场景的示例
- **[性能测试](benchmarks/)** - 基准测试和性能分析
- **[集成测试](integration_tests/)** - 完整的集成测试

## 🚀 快速开始

### 安装

```bash
go get github.com/lefeck/gonginx
```

### 基础解析

```go
package main

import (
    "fmt"
    "log"

    "github.com/lefeck/gonginx/config"
    "github.com/lefeck/gonginx/dumper"
    "github.com/lefeck/gonginx/parser"
)

func main() {
    // 从字符串解析
    p := parser.NewStringParser(`
        events {
            worker_connections 1024;
        }
        http {
            server {
                listen 80;
                server_name example.com;
                root /var/www/html;
            }
        }
    `)
    
    conf, err := p.Parse()
    if err != nil {
        log.Fatal(err)
    }
    
    // 输出解析后的配置
    fmt.Println(dumper.DumpConfig(conf, dumper.IndentedStyle))
}
```

### 配置验证

```go
// 创建验证器
validator := config.NewConfigValidator()
report := validator.ValidateConfig(conf)

// 检查验证结果
if report.HasErrors() {
    fmt.Println("发现配置错误:")
    for _, issue := range report.GetByLevel(config.ValidationError) {
        fmt.Printf("  - %s\n", issue.String())
        if issue.Fix != "" {
            fmt.Printf("    修复建议: %s\n", issue.Fix)
        }
    }
}
```

### 配置生成

```go
import "github.com/lefeck/gonginx/generator"

// 使用模板生成配置
template := generator.ReverseProxyTemplate{
    ServerName:    "api.example.com",
    Port:          80,
    BackendServer: "http://192.168.1.100:8080",
    SSLCert:       "/etc/ssl/certs/api.crt",
    SSLKey:        "/etc/ssl/private/api.key",
    RateLimit:     "10r/s",
}

conf, err := template.Generate()
if err != nil {
    log.Fatal(err)
}
```

## 🔧 核心功能

### 1. 解析功能

支持从文件和字符串解析 nginx 配置：

```go
// 从文件解析
parser, err := parser.NewParser("nginx.conf")
conf, err := parser.Parse()

// 从字符串解析
parser := parser.NewStringParser(configContent)
conf, err := parser.Parse()

// 使用解析选项
parser, err := parser.NewParser("nginx.conf", 
    parser.WithSkipComments(),
    parser.WithCustomDirectives("custom_directive"),
)
```

**支持的特性：**
- ✅ 所有标准 nginx 指令
- ✅ 自定义指令支持
- ✅ 注释保持和处理
- ✅ Include 文件递归解析
- ✅ 复杂嵌套结构
- ✅ 特殊块（Lua、Map、Geo 等）

### 2. 配置验证

四层验证确保配置正确性：

```go
validator := config.NewConfigValidator()
report := validator.ValidateConfig(conf)
```

**验证类型：**

#### 上下文验证
检查指令是否在正确的块中使用：
```go
contextValidator := config.NewContextValidator()
errors := contextValidator.ValidateConfig(conf)
```

#### 依赖关系验证
检查指令间的依赖关系：
```go
dependencyValidator := config.NewDependencyValidator()
errors := dependencyValidator.ValidateDependencies(conf)
```

#### 参数验证
检查参数的必需性和格式：
- 必需参数检查
- 参数类型验证
- 格式正确性检查

#### 结构验证
检查配置的整体结构：
- 重复块检测
- 冲突配置检查
- 逻辑一致性验证

### 3. 高级搜索

强大的配置搜索功能：

```go
// 查找所有 SSL 证书
certificates := conf.GetAllSSLCertificates()

// 按名称查找服务器
servers := conf.FindServersByName("example.com")

// 查找 upstream
upstream := conf.FindUpstreamByName("backend")

// 按模式查找 location
locations := conf.FindLocationsByPattern("/api/")

// 获取所有 upstream 服务器
upstreamServers := conf.GetAllUpstreamServers()
```

### 4. 配置修改

动态修改配置：

```go
// 添加新的 server 块
newServer := &config.Directive{
    Name: "server",
    Block: &config.Block{
        Directives: []config.IDirective{
            &config.Directive{
                Name:       "listen",
                Parameters: []config.Parameter{{Value: "443"}},
            },
            &config.Directive{
                Name:       "server_name",
                Parameters: []config.Parameter{{Value: "secure.example.com"}},
            },
        },
    },
}

// 添加到 http 块
httpDirective := conf.FindDirectives("http")[0]
httpDirective.GetBlock().(*config.Block).Directives = append(
    httpDirective.GetBlock().GetDirectives(),
    newServer,
)
```

### 5. 配置生成

使用内置模板快速生成配置：

#### 可用模板

```go
// 基础 Web 服务器
template := generator.BasicWebServerTemplate{
    ServerName: "example.com",
    Port:       80,
    Root:       "/var/www/html",
    Index:      "index.html",
}

// 反向代理
template := generator.ReverseProxyTemplate{
    ServerName:    "api.example.com",
    BackendServer: "http://192.168.1.100:8080",
    RateLimit:     "10r/s",
}

// 负载均衡器
template := generator.LoadBalancerTemplate{
    ServerName:     "lb.example.com",
    BackendServers: []string{"192.168.1.10:8080", "192.168.1.11:8080"},
    HealthCheck:    "/health",
}

// SSL Web 服务器
template := generator.SSLWebServerTemplate{
    ServerName: "secure.example.com",
    SSLCert:    "/etc/ssl/certs/secure.crt",
    SSLKey:     "/etc/ssl/private/secure.key",
    ForceSSL:   true,
}

// 微服务网关
template := generator.MicroservicesGatewayTemplate{
    ServerName: "gateway.example.com",
    Services: map[string]string{
        "/api/v1/users":    "http://user-service:8080",
        "/api/v1/orders":   "http://order-service:8080",
        "/api/v1/products": "http://product-service:8080",
    },
    RateLimit: "100r/s",
}
```

#### 使用构建器

```go
// 使用构建器模式
builder := generator.NewConfigBuilder()
conf := builder.
    Events().
        WorkerConnections(1024).
        Build().
    HTTP().
        Upstream("backend").
            Server("192.168.1.10:8080").
            Server("192.168.1.11:8080").
            Build().
        Server().
            Listen(80).
            ServerName("example.com").
            Location("/").
                ProxyPass("http://backend").
                Build().
            Build().
        Build().
    Build()
```

### 6. 实用工具

#### 安全检查

```go
import "github.com/lefeck/gonginx/utils"

securityReport := utils.CheckSecurity(conf)
fmt.Printf("安全评分: %d/100\n", securityReport.Summary.Score)

// 查看安全问题
for _, issue := range securityReport.Issues {
    fmt.Printf("[%s] %s: %s\n", issue.Level, issue.Category, issue.Title)
    if issue.Fix != "" {
        fmt.Printf("修复建议: %s\n", issue.Fix)
    }
}
```

**安全检查项目：**
- SSL/TLS 配置
- 安全头设置
- 访问控制
- 信息泄露检测
- 文件上传安全
- 速率限制

#### 配置优化

```go
optimizationReport := utils.OptimizeConfig(conf)

// 查看优化建议
for _, suggestion := range optimizationReport.Suggestions {
    fmt.Printf("[%s] %s\n", suggestion.Category, suggestion.Title)
    fmt.Printf("描述: %s\n", suggestion.Description)
    if suggestion.Example != "" {
        fmt.Printf("示例: %s\n", suggestion.Example)
    }
}
```

**优化类别：**
- 性能优化（缓冲区、keepalive、压缩）
- 安全优化（SSL 协议、安全头）
- 大小优化（重复指令清理）
- 维护性优化（结构改进）

#### 格式转换

```go
// 转换为 JSON
jsonConfig, err := utils.ConvertToJSON(conf)

// 转换为 YAML
yamlConfig, err := utils.ConvertToYAML(conf)

// 配置差异比较
diffReport := utils.CompareConfigs(oldConf, newConf)
```

### 7. 错误处理

增强的错误处理提供更好的调试体验：

```go
import "github.com/lefeck/gonginx/errors"

// 使用增强解析器
enhancedParser, err := errors.NewEnhancedParser("nginx.conf")
if err != nil {
    log.Fatal(err)
}

conf, err := enhancedParser.ParseWithValidation()
if err != nil {
    // 获得详细的错误信息和修复建议
    fmt.Printf("解析错误: %s\n", err.Error())
}
```

**错误类型：**
- 语法错误（缺少分号、括号等）
- 语义错误（重复配置、冲突设置）
- 上下文错误（指令在错误的块中）
- 文件错误（文件不存在、权限问题）
- 验证错误（依赖关系、参数格式）

## 🎯 使用场景

### 1. 配置管理工具

```go
// 配置文件检查工具
func validateNginxConfig(filename string) error {
    parser, err := parser.NewParser(filename)
    if err != nil {
        return err
    }
    
    conf, err := parser.Parse()
    if err != nil {
        return err
    }
    
    validator := config.NewConfigValidator()
    report := validator.ValidateConfig(conf)
    
    if report.HasErrors() {
        return fmt.Errorf("配置验证失败: %d 个错误", report.Summary.Errors)
    }
    
    return nil
}
```

### 2. 自动化部署

```go
// 生成服务配置
func generateServiceConfig(serviceName, backend string) (*config.Config, error) {
    template := generator.ReverseProxyTemplate{
        ServerName:    fmt.Sprintf("%s.example.com", serviceName),
        Port:          80,
        BackendServer: backend,
        RateLimit:     "50r/s",
    }
    
    return template.Generate()
}

// 部署配置
func deployConfig(conf *config.Config, target string) error {
    return dumper.WriteConfig(conf, target, dumper.IndentedStyle)
}
```

### 3. 配置分析

```go
// 分析配置性能
func analyzeConfiguration(conf *config.Config) {
    // 安全检查
    securityReport := utils.CheckSecurity(conf)
    fmt.Printf("安全评分: %d/100\n", securityReport.Summary.Score)
    
    // 优化建议
    optimizationReport := utils.OptimizeConfig(conf)
    fmt.Printf("优化建议: %d 条\n", len(optimizationReport.Suggestions))
    
    // 统计信息
    servers := conf.FindDirectives("server")
    upstreams := conf.FindDirectives("upstream")
    locations := conf.FindDirectives("location")
    
    fmt.Printf("配置统计: %d servers, %d upstreams, %d locations\n",
        len(servers), len(upstreams), len(locations))
}
```

### 4. 配置模板化

```go
// 环境特定配置生成
func generateEnvironmentConfig(env string, services []Service) (*config.Config, error) {
    builder := generator.NewConfigBuilder()
    
    httpBuilder := builder.HTTP()
    
    // 添加 upstream
    for _, service := range services {
        upstreamBuilder := httpBuilder.Upstream(service.Name)
        for _, endpoint := range service.Endpoints {
            upstreamBuilder.Server(endpoint)
        }
        upstreamBuilder.Build()
    }
    
    // 添加 server
    serverBuilder := httpBuilder.Server().
        Listen(80).
        ServerName(fmt.Sprintf("api-%s.example.com", env))
    
    for _, service := range services {
        serverBuilder.Location(fmt.Sprintf("/api/%s/", service.Name)).
            ProxyPass(fmt.Sprintf("http://%s", service.Name)).
            Build()
    }
    
    return builder.Build(), nil
}
```

## 📈 性能考虑

### 解析性能

| 配置大小 | 解析时间 | 内存使用 |
|---------|----------|----------|
| 小型 (< 1KB) | < 1ms | < 50KB |
| 中型 (< 50KB) | < 10ms | < 500KB |
| 大型 (< 1MB) | < 100ms | < 5MB |

### 优化建议

1. **批量操作**：一次解析多个配置比多次解析单个配置更高效
2. **缓存结果**：对于频繁访问的配置，考虑缓存解析结果
3. **增量验证**：只验证变更的部分而不是整个配置
4. **并发处理**：解析和验证操作是线程安全的，可以并发执行

### 基准测试

运行性能基准测试：

```bash
# 运行所有基准测试
go test -bench=. ./benchmarks/

# 运行特定类型的基准测试
go test -bench=BenchmarkParse ./benchmarks/
go test -bench=BenchmarkValidation ./benchmarks/
go test -bench=BenchmarkSearch ./benchmarks/

# 查看内存分配
go test -bench=. -benchmem ./benchmarks/
```

## 🧪 测试

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./config/
go test ./parser/
go test ./dumper/

# 运行集成测试
go test -v ./integration_tests/

# 运行基准测试
go test -bench=. ./benchmarks/
```

### 测试覆盖

- **单元测试**：每个包都有详细的单元测试
- **集成测试**：测试完整的工作流和真实世界的配置
- **基准测试**：性能测试和回归检测
- **示例测试**：确保文档中的示例代码可以正常工作

## 🤝 贡献

我们欢迎社区贡献！

### 贡献方式

1. **报告问题**：在 GitHub Issues 中报告 bug 或提出功能请求
2. **提交代码**：Fork 项目，创建分支，提交 Pull Request
3. **改进文档**：帮助改进文档和示例
4. **分享经验**：分享使用经验和最佳实践

### 开发指南

1. **代码规范**：遵循 Go 语言官方代码规范
2. **测试要求**：新功能必须包含测试
3. **文档更新**：更新相关文档和示例
4. **性能考虑**：确保更改不会显著影响性能

### 运行测试

```bash
# 在提交前运行完整测试
make test

# 检查代码格式
make fmt

# 运行静态分析
make vet
```

## 📚 更多资源

- **[API 参考](API_REFERENCE.md)** - 完整的 API 文档
- **[示例代码](examples/)** - 各种使用场景的完整示例
- **[配置验证](examples/config-validation/)** - 配置验证功能示例
- **[错误处理](examples/error-handling/)** - 错误处理最佳实践
- **[工具功能](examples/utils-demo/)** - 实用工具功能演示
- **[性能测试](benchmarks/)** - 基准测试和性能分析

## 📄 许可证

MIT License - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🆘 获得帮助

- **GitHub Issues**：报告问题或请求功能
- **讨论区**：参与社区讨论
- **文档**：查看完整的 API 文档和示例

---

**Gonginx** 致力于为 Go 开发者提供最好的 nginx 配置处理体验。无论你是在构建配置管理工具、自动化部署系统，还是需要分析和优化 nginx 配置，Gonginx 都能满足你的需求。