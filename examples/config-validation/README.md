# Nginx 配置验证功能示例

这个示例演示了 gonginx 库的配置验证功能，包括上下文验证、依赖关系检查、参数验证和结构验证。

## 功能特性

### 1. 上下文验证 (Context Validation)
验证指令是否在正确的上下文中使用：
- 检查指令是否在允许的块中
- 验证块的嵌套关系
- 提供详细的错误信息和修复建议

### 2. 依赖关系验证 (Dependency Validation)
检查指令之间的依赖关系：
- SSL 证书和私钥的配对
- proxy_cache 和 proxy_cache_path 的依赖
- auth_basic 和 auth_basic_user_file 的依赖
- upstream 引用的有效性
- 限流配置的依赖关系

### 3. 参数验证 (Parameter Validation)
验证指令参数的正确性：
- 检查必需参数是否存在
- 验证参数格式和类型
- 文件路径有效性检查

### 4. 结构验证 (Structural Validation)
验证配置的整体结构：
- 检查重复的块（如多个 http 块）
- 验证 server_name 冲突
- 检查 listen 端口冲突
- upstream 块的完整性检查

## 使用方法

### 基本使用

```go
package main

import (
    "github.com/lefeck/gonginx/config"
    "github.com/lefeck/gonginx/parser"
)

func main() {
    // 解析配置
    p := parser.NewStringParser(configContent)
    conf, err := p.Parse()
    if err != nil {
        // 处理解析错误
    }
    
    // 创建验证器
    validator := config.NewConfigValidator()
    
    // 进行验证
    report := validator.ValidateConfig(conf)
    
    // 检查结果
    if report.HasErrors() {
        // 处理验证错误
        for _, issue := range report.GetByLevel(config.ValidationError) {
            fmt.Printf("错误: %s\n", issue.String())
        }
    }
}
```

### 分别进行不同类型的验证

```go
// 只进行上下文验证
contextValidator := config.NewContextValidator()
contextErrors := contextValidator.ValidateConfig(conf)

// 只进行依赖关系验证
dependencyValidator := config.NewDependencyValidator()
dependencyErrors := dependencyValidator.ValidateDependencies(conf)
```

### 自定义验证选项

```go
// 创建验证器时禁用某些检查
validator := config.NewConfigValidatorWithOptions(false) // 禁用所有检查
report := validator.ValidateConfig(conf)
```

## 验证类型

### ValidationLevel
- `ValidationError`: 必须修复的错误
- `ValidationWarning`: 建议修复的警告
- `ValidationInfo`: 信息性提示

### 错误类别
- `Context`: 上下文相关错误
- `Dependency`: 依赖关系错误
- `Parameter`: 参数相关错误
- `Structure`: 结构相关错误

## 运行示例

```bash
cd examples/config-validation
go run main.go
```

## 示例输出

```
=== Nginx 配置验证功能示例 ===

1. 上下文验证示例:
发现 3 个上下文错误:
  - line 4: directive 'proxy_pass' is not allowed in 'http' context: allowed in: http, server, location, if
  - line 11: directive 'listen' is not allowed in 'location' context: allowed in: http, server
  - line 17: directive 'server' is not allowed in 'main' context: allowed in: http

2. 依赖关系验证示例:
发现 4 个依赖关系错误:
  - line 9: directive 'proxy_pass' requires 'upstream backend': upstream 'backend' is not defined (suggestion: define upstream backend in http context or use a direct URL)
  - line 14: directive 'ssl_certificate' requires 'ssl_certificate_key': SSL certificate requires a private key (suggestion: add ssl_certificate_key directive with the path to the private key file)
  - line 19: directive 'proxy_cache' requires 'proxy_cache_path': proxy_cache requires proxy_cache_path to be defined (suggestion: add proxy_cache_path directive in http context)
  - line 25: directive 'upstream' requires 'server': upstream 'empty_upstream' has no server directives (suggestion: add at least one server directive (e.g., 'server backend1.example.com;'))

...
```

## 集成到现有代码

这个验证功能可以轻松集成到现有的 nginx 配置处理流程中：

1. **在配置加载后进行验证**
2. **根据验证结果决定是否继续处理**
3. **为用户提供详细的错误信息和修复建议**
4. **支持不同级别的验证严格性**

## 最佳实践

1. **总是在生产环境使用配置前进行验证**
2. **将验证集成到 CI/CD 流程中**
3. **根据项目需求调整验证级别**
4. **使用验证报告改进配置质量**
