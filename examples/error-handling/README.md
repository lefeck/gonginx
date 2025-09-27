# 错误处理改进示例

这个示例展示了 gonginx 库的增强错误处理功能，提供了更好的错误信息、行号报告和建议。

## 功能特性

### 错误类型
1. **SyntaxError** - 语法错误 (缺少分号、括号等)
2. **SemanticError** - 语义错误 (重复指令、逻辑错误等)
3. **ContextError** - 上下文错误 (指令在错误的块中)
4. **FileError** - 文件错误 (文件不存在、权限问题等)
5. **ValidationError** - 验证错误 (参数值无效等)
6. **UnknownDirectiveError** - 未知指令错误

### 增强的错误信息
- **位置信息**: 文件名、行号、列号
- **上下文**: 错误周围的代码行
- **建议**: 如何修复错误的建议
- **指令信息**: 相关的指令和参数
- **错误分类**: 按类型组织错误

### 智能建议系统
- **拼写检查**: 检测常见的指令拼写错误
- **语法提示**: 提供正确的语法格式
- **配置建议**: 给出最佳实践建议
- **修复指导**: 具体的修复步骤

## API 使用

### 基础用法
```go
import "github.com/lefeck/gonginx/errors"

// 从文件解析
parser, err := errors.NewEnhancedParser("nginx.conf")
if err != nil {
    log.Fatal(err)
}

config, err := parser.Parse()
if err != nil {
    fmt.Printf("解析错误: %s\n", err.Error())
}
```

### 从字符串解析
```go
parser := errors.NewEnhancedStringParser(configContent)
config, err := parser.ParseWithValidation()
if err != nil {
    // 处理错误
}
```

### 错误集合处理
```go
if errCollection, ok := err.(*errors.ErrorCollection); ok {
    fmt.Printf("发现 %d 个错误\n", errCollection.Count())
    
    // 按类型获取错误
    syntaxErrors := errCollection.GetByType(errors.SyntaxError)
    semanticErrors := errCollection.GetByType(errors.SemanticError)
    
    // 处理每种类型的错误
}
```

### 创建自定义错误
```go
err := errors.NewSyntaxError("missing semicolon").
    WithFile("nginx.conf").
    WithLine(42).
    WithColumn(25).
    WithContext("server { listen 80 }").
    WithSuggestion("Add a semicolon after the listen directive")
```

## 错误示例

### 1. 语法错误
```nginx
server {
    listen 80        # 缺少分号
    server_name example.com;
# 缺少关闭大括号
```

**错误信息**:
```
[Syntax Error] missing semicolon at <string>:3
Context:
    1 | server {
    2 |     listen 80        # 缺少分号
>>> 3 |     server_name example.com;
    4 | # 缺少关闭大括号
Suggestion: Add a semicolon after the listen directive
```

### 2. 语义错误
```nginx
http {
    server {
        listen 80;
        server_name example.com;
    }
    server {
        listen 80;
        server_name example.com;  # 重复的 server_name
    }
}
```

**错误信息**:
```
[Semantic Error] duplicate server_name 'example.com' found
in directive 'server_name' with parameter 'example.com'
Suggestion: Each server_name should be unique unless using default_server
```

### 3. 配置验证错误
```nginx
http {
    upstream backend {
        # 空的 upstream 块
    }
    
    server {
        # 没有 listen 指令
        server_name example.com;
        ssl_certificate /path/to/cert.pem;
        # 缺少 ssl_certificate_key
    }
}
```

**错误信息**:
```
Multiple errors (3):
1. [Semantic Error] upstream 'backend' has no servers
2. [Semantic Error] server block has no listen directive
3. [Semantic Error] ssl_certificate specified without ssl_certificate_key
```

### 4. 未知指令错误
```nginx
server {
    listen 80;
    servername example.com;     # 拼写错误
    documentroot /var/www;      # 应该是 root
    proxypass http://backend;   # 应该是 proxy_pass
}
```

**错误信息**:
```
[Unknown Directive Error] unknown directive 'servername'
Suggestion: Did you mean 'server_name'?
```

## 验证功能

### HTTP 块验证
- 检查多个 HTTP 块
- 验证服务器配置
- 检查重复的 server_name
- 验证 SSL 配置

### SSL 配置验证
- 检查证书文件是否存在
- 验证证书和密钥配对
- 检查 SSL 指令完整性

### Upstream 验证
- 检查空的 upstream 块
- 验证服务器配置
- 检查负载均衡算法

### Stream 块验证
- 检查多个 Stream 块
- 验证流服务器配置
- 检查代理配置

## 运行示例

```bash
cd examples/error-handling
go run main.go
```

## 示例输出

程序将演示：
1. 各种类型的错误检测
2. 详细的错误信息和建议
3. 多错误处理
4. 智能拼写检查
5. 配置验证功能

## 实际应用

### DevOps 工具集成
```go
func validateNginxConfig(configPath string) error {
    parser, err := errors.NewEnhancedParser(configPath)
    if err != nil {
        return err
    }
    
    _, err = parser.ParseWithValidation()
    return err
}
```

### CI/CD 管道
```go
func checkConfigInPipeline(configContent string) {
    parser := errors.NewEnhancedStringParser(configContent)
    _, err := parser.ParseWithValidation()
    
    if err != nil {
        if errCollection, ok := err.(*errors.ErrorCollection); ok {
            for _, e := range errCollection.Errors {
                fmt.Printf("::error file=%s,line=%d::%s\n", 
                    e.File, e.Line, e.Message)
            }
            os.Exit(1)
        }
    }
}
```

### IDE 插件支持
```go
func getConfigErrors(content string) []*errors.ParseError {
    parser := errors.NewEnhancedStringParser(content)
    _, err := parser.ParseWithValidation()
    
    if errCollection, ok := err.(*errors.ErrorCollection); ok {
        return errCollection.Errors
    }
    
    return nil
}
```

## 扩展功能

### 自定义验证规则
可以扩展 `EnhancedParser` 添加自定义验证规则：

```go
func (ep *EnhancedParser) addCustomValidation(rule ValidationRule) {
    // 添加自定义验证逻辑
}
```

### 错误处理插件
可以创建错误处理插件来扩展功能：

```go
type ErrorHandler interface {
    HandleError(*errors.ParseError) *errors.ParseError
}
```
