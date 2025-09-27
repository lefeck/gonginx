# Gonginx 综合功能示例

这个示例展示了 gonginx 库的所有主要功能，包括配置生成、解析、验证、安全检查、优化和格式转换等。

## 功能概述

### 1. 配置生成
使用内置模板生成完整的 nginx 配置文件：
- 微服务网关模板
- 反向代理配置
- SSL/TLS 配置
- 负载均衡配置

### 2. 配置解析和修改
从字符串或文件解析 nginx 配置，并进行动态修改：
- 添加新的 server 块
- 修改现有指令
- 添加新的 location
- 更新 upstream 配置

### 3. 配置验证
全面的配置验证功能：
- 上下文验证（指令是否在正确的块中）
- 依赖关系验证（如 SSL 证书配对）
- 参数验证（必需参数、格式检查）
- 结构验证（重复块、冲突配置）

### 4. 安全检查
专业的安全配置分析：
- SSL/TLS 配置检查
- 安全头验证
- 访问控制检查
- 文件上传安全
- 信息泄露检测

### 5. 配置优化
智能的配置优化建议：
- 性能优化（缓冲区、keepalive 等）
- 安全优化（协议、加密等）
- 大小优化（重复指令清理）
- 维护性优化（结构改进）

### 6. 格式转换
支持多种配置格式：
- Nginx 原生格式
- JSON 格式
- YAML 格式

### 7. 高级搜索
强大的配置搜索功能：
- 按名称查找服务器
- 查找所有 SSL 证书
- 获取 upstream 服务器列表
- 按模式查找 location

## 运行示例

```bash
cd examples/comprehensive-example
go run main.go
```

## 示例输出

```
=== Gonginx 综合功能示例 ===

1. 配置生成示例:
生成的配置:
events {
    worker_connections 1024;
}
http {
    upstream api_v1_users_upstream {
        server user-service:8080;
    }
    ...

2. 配置解析和修改示例:
解析成功！
添加新 server 块后的配置:
...

3. 配置验证示例:
验证结果: Validation Summary: 6 issues (Errors: 4, Warnings: 2, Info: 0)
发现的错误:
  - [ERROR] Line 3: Invalid directive context - proxy_pass not allowed in http context
  ...

4. 配置安全检查示例:
安全评分: 65/100
安全问题总数: 5 (严重: 2, 警告: 3)
...

5. 配置优化示例:
优化建议总数: 8
优化建议:
  - [Performance] 启用 HTTP/2
  ...

6. 配置格式转换示例:
JSON 格式:
{
  "events": {
    "worker_connections": "1024"
  },
  ...

7. 高级搜索功能示例:
SSL 证书文件 (2):
  - /etc/ssl/certs/api.crt
  - /etc/ssl/certs/web.crt
...
```

## API 使用指南

### 基础解析和生成

```go
// 从字符串解析
parser := parser.NewStringParser(configContent)
conf, err := parser.Parse()

// 从文件解析
parser, err := parser.NewParser("nginx.conf")
conf, err := parser.Parse()

// 生成配置字符串
configStr := dumper.DumpConfig(conf, dumper.IndentedStyle)
```

### 配置验证

```go
// 创建验证器
validator := config.NewConfigValidator()
report := validator.ValidateConfig(conf)

// 检查结果
if report.HasErrors() {
    for _, issue := range report.GetByLevel(config.ValidationError) {
        fmt.Printf("Error: %s\n", issue.String())
    }
}
```

### 安全检查

```go
// 进行安全检查
securityReport := utils.CheckSecurity(conf)
fmt.Printf("Security Score: %d/100\n", securityReport.Summary.Score)

// 查看安全问题
for _, issue := range securityReport.Issues {
    fmt.Printf("Security Issue: %s\n", issue.Title)
}
```

### 配置优化

```go
// 获取优化建议
optimizationReport := utils.OptimizeConfig(conf)
for _, suggestion := range optimizationReport.Suggestions {
    fmt.Printf("Suggestion: %s\n", suggestion.Title)
}
```

### 格式转换

```go
// 转换为 JSON
jsonConfig, err := utils.ConvertToJSON(conf)

// 转换为 YAML
yamlConfig, err := utils.ConvertToYAML(conf)
```

### 高级搜索

```go
// 查找服务器
servers := conf.FindServersByName("example.com")

// 查找 upstream
upstream := conf.FindUpstreamByName("backend")

// 获取所有 SSL 证书
certificates := conf.GetAllSSLCertificates()

// 查找 location
locations := conf.FindLocationsByPattern("/api/")
```

## 最佳实践

1. **总是进行配置验证**：在生产环境使用配置前，确保通过所有验证检查
2. **定期安全检查**：使用安全检查功能定期审查配置
3. **应用优化建议**：根据优化报告改进配置性能
4. **使用类型化参数**：利用参数类型系统确保配置正确性
5. **集成到 CI/CD**：将验证和安全检查集成到部署流程中

## 相关文档

- [配置验证详细文档](../config-validation/README.md)
- [错误处理文档](../error-handling/README.md)
- [工具功能文档](../utils-demo/README.md)
- [配置生成器文档](../config-generator/README.md)
