# Gonginx API 参考文档

这是 gonginx 库的完整 API 参考文档，涵盖了所有主要包和功能。

## 目录

- [Parser 包](#parser-包)
- [Config 包](#config-包)
- [Dumper 包](#dumper-包)
- [Generator 包](#generator-包)
- [Utils 包](#utils-包)
- [Errors 包](#errors-包)
- [使用示例](#使用示例)

## Parser 包

用于解析 nginx 配置文件和字符串。

### 类型

#### Parser

```go
type Parser struct {
    // 内部字段
}
```

nginx 配置解析器，支持从文件和字符串解析配置。

### 函数

#### NewParser

```go
func NewParser(filename string, options ...Option) (*Parser, error)
```

创建一个新的文件解析器。

**参数：**
- `filename`: 配置文件路径
- `options`: 可选的解析选项

**返回值：**
- `*Parser`: 解析器实例
- `error`: 错误信息

**示例：**
```go
parser, err := parser.NewParser("nginx.conf")
if err != nil {
    log.Fatal(err)
}
```

#### NewStringParser

```go
func NewStringParser(content string, options ...Option) *Parser
```

创建一个新的字符串解析器。

**参数：**
- `content`: nginx 配置内容字符串
- `options`: 可选的解析选项

**返回值：**
- `*Parser`: 解析器实例

**示例：**
```go
config := `
events {
    worker_connections 1024;
}
http {
    server {
        listen 80;
        server_name example.com;
    }
}
`
parser := parser.NewStringParser(config)
```

### 方法

#### Parse

```go
func (p *Parser) Parse() (*config.Config, error)
```

解析配置并返回配置对象。

**返回值：**
- `*config.Config`: 解析后的配置对象
- `error`: 解析错误

**示例：**
```go
conf, err := parser.Parse()
if err != nil {
    log.Fatal(err)
}
```

### 选项

#### WithSkipComments

```go
func WithSkipComments() Option
```

跳过注释的解析选项。

#### WithCustomDirectives

```go
func WithCustomDirectives(directives ...string) Option
```

允许自定义指令的解析选项。

#### WithSkipValidBlocks

```go
func WithSkipValidBlocks(blocks ...string) Option
```

跳过特定块验证的选项。

#### WithSkipValidDirectivesErr

```go
func WithSkipValidDirectivesErr() Option
```

跳过无效指令错误的选项。

## Config 包

定义了 nginx 配置的数据结构和操作方法。

### 核心接口

#### IDirective

```go
type IDirective interface {
    GetName() string
    GetParameters() []Parameter
    GetBlock() IBlock
    GetComment() []string
    SetComment(comment []string)
    SetParent(IDirective)
    GetParent() IDirective
    GetLine() int
    SetLine(int)
    InlineCommenter
}
```

表示 nginx 指令的接口。

#### IBlock

```go
type IBlock interface {
    GetDirectives() []IDirective
    FindDirectives(directiveName string) []IDirective
    GetCodeBlock() string
    SetParent(IDirective)
    GetParent() IDirective
}
```

表示 nginx 配置块的接口。

### 主要类型

#### Config

```go
type Config struct {
    Block
}
```

nginx 配置的根对象。

**方法：**

##### FindDirectives

```go
func (c *Config) FindDirectives(directiveName string) []IDirective
```

查找所有指定名称的指令。

##### FindServersByName

```go
func (c *Config) FindServersByName(name string) []*Server
```

按 server_name 查找服务器块。

##### FindUpstreamByName

```go
func (c *Config) FindUpstreamByName(name string) *Upstream
```

按名称查找 upstream 块。

##### FindLocationsByPattern

```go
func (c *Config) FindLocationsByPattern(pattern string) []*Location
```

按模式查找 location 块。

##### GetAllSSLCertificates

```go
func (c *Config) GetAllSSLCertificates() []string
```

获取所有 SSL 证书文件路径。

##### GetAllUpstreamServers

```go
func (c *Config) GetAllUpstreamServers() []*UpstreamServer
```

获取所有 upstream 服务器。

#### Directive

```go
type Directive struct {
    Block      IBlock
    Name       string
    Parameters []Parameter
    Comment    []string
    Parent     IDirective
    Line       int
}
```

nginx 指令的具体实现。

#### Parameter

```go
type Parameter struct {
    Value             string
    Type              ParameterType
    RelativeLineIndex int
}
```

指令参数。

#### ParameterType

```go
type ParameterType int

const (
    ParameterTypeString ParameterType = iota
    ParameterTypeVariable
    ParameterTypeNumber
    ParameterTypeSize
    ParameterTypeTime
    ParameterTypePath
    ParameterTypeURL
    ParameterTypeRegex
    ParameterTypeBoolean
    ParameterTypeQuoted
)
```

参数类型枚举。

### 专门类型

#### Server

```go
type Server struct {
    Directive
}
```

nginx server 块的专门类型。

**方法：**
- `GetListen() []string`: 获取 listen 指令
- `GetServerNames() []string`: 获取 server_name 指令

#### Upstream

```go
type Upstream struct {
    Directive
}
```

nginx upstream 块的专门类型。

**方法：**
- `GetServers() []*UpstreamServer`: 获取 upstream 服务器列表

#### Location

```go
type Location struct {
    Directive
}
```

nginx location 块的专门类型。

**方法：**
- `GetPath() string`: 获取 location 路径

### 验证功能

#### ContextValidator

```go
type ContextValidator struct {
    // 内部字段
}
```

上下文验证器，验证指令是否在正确的上下文中使用。

**方法：**

##### NewContextValidator

```go
func NewContextValidator() *ContextValidator
```

创建新的上下文验证器。

##### ValidateConfig

```go
func (cv *ContextValidator) ValidateConfig(config *Config) []error
```

验证配置的上下文正确性。

##### ValidateContext

```go
func (cv *ContextValidator) ValidateContext(directive IDirective, context string) error
```

验证单个指令的上下文。

##### GetAllowedContexts

```go
func (cv *ContextValidator) GetAllowedContexts(directiveName string) []string
```

获取指令允许的上下文列表。

#### DependencyValidator

```go
type DependencyValidator struct {
    // 内部字段
}
```

依赖关系验证器，检查指令间的依赖关系。

**方法：**

##### NewDependencyValidator

```go
func NewDependencyValidator() *DependencyValidator
```

创建新的依赖关系验证器。

##### ValidateDependencies

```go
func (dv *DependencyValidator) ValidateDependencies(config *Config) []error
```

验证配置的依赖关系。

#### ConfigValidator

```go
type ConfigValidator struct {
    // 内部字段
}
```

综合配置验证器，提供完整的配置验证功能。

**方法：**

##### NewConfigValidator

```go
func NewConfigValidator() *ConfigValidator
```

创建新的配置验证器。

##### ValidateConfig

```go
func (cv *ConfigValidator) ValidateConfig(config *Config) *ValidationReport
```

执行完整的配置验证。

#### ValidationReport

```go
type ValidationReport struct {
    Issues  []ValidationIssue
    Summary ValidationSummary
    Config  *Config
}
```

验证报告。

**方法：**
- `HasErrors() bool`: 是否有错误
- `GetByLevel(level ValidationLevel) []ValidationIssue`: 按级别获取问题
- `GetByCategory(category string) []ValidationIssue`: 按类别获取问题

### 参数处理

#### DetectParameterType

```go
func DetectParameterType(value string) ParameterType
```

自动检测参数类型。

#### NewParameter

```go
func NewParameter(value string) Parameter
```

创建自动检测类型的参数。

#### NewParameterWithType

```go
func NewParameterWithType(value string, paramType ParameterType) Parameter
```

创建指定类型的参数。

#### 验证函数

```go
func ValidateSize(value string) (string, bool)
func ValidateTime(value string) (string, bool)
func ValidateNumber(value string) (float64, bool)
func ValidateBoolean(value string) (bool, bool)
```

各种参数类型的验证函数。

## Dumper 包

用于将配置对象导出为字符串格式。

### 类型

#### Style

```go
type Style struct {
    StartIndent    int
    Indent         string
    ClosingBraceIndent string
}
```

导出格式样式。

### 预定义样式

```go
var (
    NoIndentStyle = Style{
        StartIndent:    0,
        Indent:         "",
        ClosingBraceIndent: "",
    }
    IndentedStyle = Style{
        StartIndent:    0,
        Indent:         "    ",
        ClosingBraceIndent: "",
    }
)
```

### 函数

#### DumpConfig

```go
func DumpConfig(c *config.Config, style Style) string
```

将配置对象导出为字符串。

**参数：**
- `c`: 配置对象
- `style`: 导出样式

**返回值：**
- `string`: 格式化的配置字符串

**示例：**
```go
configStr := dumper.DumpConfig(conf, dumper.IndentedStyle)
fmt.Println(configStr)
```

#### WriteConfig

```go
func WriteConfig(c *config.Config, filename string, style Style) error
```

将配置写入文件。

**参数：**
- `c`: 配置对象
- `filename`: 目标文件路径
- `style`: 导出样式

**返回值：**
- `error`: 写入错误

**示例：**
```go
err := dumper.WriteConfig(conf, "output.conf", dumper.IndentedStyle)
if err != nil {
    log.Fatal(err)
}
```

## Generator 包

提供配置模板和生成器功能。

### 模板接口

#### Template

```go
type Template interface {
    Generate() (*config.Config, error)
}
```

配置模板接口。

### 内置模板

#### BasicWebServerTemplate

```go
type BasicWebServerTemplate struct {
    ServerName string
    Port       int
    Root       string
    Index      string
}
```

基础 Web 服务器模板。

#### ReverseProxyTemplate

```go
type ReverseProxyTemplate struct {
    ServerName    string
    Port          int
    BackendServer string
    SSLCert       string
    SSLKey        string
    RateLimit     string
}
```

反向代理模板。

#### LoadBalancerTemplate

```go
type LoadBalancerTemplate struct {
    ServerName      string
    Port            int
    BackendServers  []string
    HealthCheck     string
    LoadBalanceMethod string
}
```

负载均衡器模板。

#### SSLWebServerTemplate

```go
type SSLWebServerTemplate struct {
    ServerName   string
    Port         int
    SSLPort      int
    Root         string
    SSLCert      string
    SSLKey       string
    ForceSSL     bool
}
```

SSL Web 服务器模板。

#### MicroservicesGatewayTemplate

```go
type MicroservicesGatewayTemplate struct {
    ServerName   string
    Port         int
    SSLPort      int
    SSLCert      string
    SSLKey       string
    Services     map[string]string
    HealthCheck  string
    RateLimit    string
}
```

微服务网关模板。

### 构建器

#### ConfigBuilder

```go
type ConfigBuilder struct {
    // 内部字段
}
```

配置构建器，使用 Builder 模式构建配置。

**方法：**

##### NewConfigBuilder

```go
func NewConfigBuilder() *ConfigBuilder
```

创建新的配置构建器。

##### Events

```go
func (cb *ConfigBuilder) Events() *EventsBuilder
```

添加 events 块。

##### HTTP

```go
func (cb *ConfigBuilder) HTTP() *HTTPBuilder
```

添加 http 块。

##### Stream

```go
func (cb *ConfigBuilder) Stream() *StreamBuilder
```

添加 stream 块。

##### Build

```go
func (cb *ConfigBuilder) Build() *config.Config
```

构建最终的配置对象。

## Utils 包

提供各种实用工具功能。

### 安全检查

#### CheckSecurity

```go
func CheckSecurity(conf *config.Config) *SecurityReport
```

执行配置安全检查。

**参数：**
- `conf`: 配置对象

**返回值：**
- `*SecurityReport`: 安全报告

#### SecurityReport

```go
type SecurityReport struct {
    Issues  []SecurityIssue
    Summary SecuritySummary
    Passed  []string
    Config  *config.Config
}
```

安全报告。

**方法：**
- `HasCriticalIssues() bool`: 是否有严重安全问题
- `GetByLevel(level SecurityLevel) []SecurityIssue`: 按级别获取问题
- `GetByCategory(category string) []SecurityIssue`: 按类别获取问题

### 配置优化

#### OptimizeConfig

```go
func OptimizeConfig(conf *config.Config) *OptimizationReport
```

分析配置并提供优化建议。

**参数：**
- `conf`: 配置对象

**返回值：**
- `*OptimizationReport`: 优化报告

#### OptimizationReport

```go
type OptimizationReport struct {
    Suggestions []OptimizationSuggestion
    Issues      []OptimizationIssue
    Summary     OptimizationSummary
}
```

优化报告。

### 格式转换

#### ConvertToJSON

```go
func ConvertToJSON(conf *config.Config) (string, error)
```

将配置转换为 JSON 格式。

#### ConvertToYAML

```go
func ConvertToYAML(conf *config.Config) (string, error)
```

将配置转换为 YAML 格式。

### 配置比较

#### CompareConfigs

```go
func CompareConfigs(conf1, conf2 *config.Config) *DiffReport
```

比较两个配置的差异。

**参数：**
- `conf1`, `conf2`: 要比较的配置对象

**返回值：**
- `*DiffReport`: 差异报告

#### DiffReport

```go
type DiffReport struct {
    Added    []DiffItem
    Removed  []DiffItem
    Modified []DiffItem
    Summary  DiffSummary
}
```

差异报告。

## Errors 包

提供增强的错误处理功能。

### 类型

#### EnhancedParser

```go
type EnhancedParser struct {
    // 内部字段
}
```

增强的解析器，提供更好的错误信息。

**方法：**

##### NewEnhancedParser

```go
func NewEnhancedParser(filename string) (*EnhancedParser, error)
```

创建文件增强解析器。

##### NewEnhancedStringParser

```go
func NewEnhancedStringParser(content string) *EnhancedParser
```

创建字符串增强解析器。

##### Parse

```go
func (ep *EnhancedParser) Parse() (*config.Config, error)
```

解析配置。

##### ParseWithValidation

```go
func (ep *EnhancedParser) ParseWithValidation() (*config.Config, error)
```

解析并验证配置。

#### ErrorCollection

```go
type ErrorCollection struct {
    // 内部字段
}
```

错误集合，用于收集多个错误。

**方法：**
- `AddError(err error)`: 添加错误
- `HasErrors() bool`: 是否有错误
- `GetErrors() []error`: 获取所有错误
- `Error() string`: 错误信息

## 使用示例

### 基础使用

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
    // 解析配置
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
    
    // 验证配置
    validator := config.NewConfigValidator()
    report := validator.ValidateConfig(conf)
    
    if report.HasErrors() {
        for _, issue := range report.GetByLevel(config.ValidationError) {
            fmt.Printf("Error: %s\n", issue.String())
        }
    }
    
    // 导出配置
    fmt.Println(dumper.DumpConfig(conf, dumper.IndentedStyle))
}
```

### 高级使用

```go
package main

import (
    "fmt"
    "log"

    "github.com/lefeck/gonginx/generator"
    "github.com/lefeck/gonginx/utils"
)

func main() {
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
    
    // 安全检查
    securityReport := utils.CheckSecurity(conf)
    fmt.Printf("Security Score: %d/100\n", securityReport.Summary.Score)
    
    // 配置优化
    optimizationReport := utils.OptimizeConfig(conf)
    fmt.Printf("Optimization Suggestions: %d\n", len(optimizationReport.Suggestions))
    
    // 格式转换
    jsonConfig, err := utils.ConvertToJSON(conf)
    if err == nil {
        fmt.Println("JSON Config:", jsonConfig)
    }
}
```

### 配置修改

```go
// 添加新的 server 块
newServer := &config.Directive{
    Name: "server",
    Block: &config.Block{
        Directives: []config.IDirective{
            &config.Directive{
                Name:       "listen",
                Parameters: []config.Parameter{{Value: "443", Type: config.ParameterTypeNumber}},
            },
            &config.Directive{
                Name:       "server_name",
                Parameters: []config.Parameter{{Value: "secure.example.com", Type: config.ParameterTypeString}},
            },
        },
    },
}

// 将新 server 添加到 http 块
httpDirective := conf.FindDirectives("http")[0]
httpDirective.GetBlock().(*config.Block).Directives = append(
    httpDirective.GetBlock().GetDirectives(),
    newServer,
)
```

### 搜索功能

```go
// 查找所有 SSL 证书
certificates := conf.GetAllSSLCertificates()
fmt.Printf("SSL Certificates: %v\n", certificates)

// 查找特定 server
servers := conf.FindServersByName("example.com")
for _, server := range servers {
    listen := server.GetListen()
    fmt.Printf("Server listens on: %v\n", listen)
}

// 查找 upstream
upstream := conf.FindUpstreamByName("backend")
if upstream != nil {
    servers := upstream.GetServers()
    fmt.Printf("Upstream has %d servers\n", len(servers))
}
```

## 错误处理

gonginx 提供了多层次的错误处理：

1. **解析错误**：语法错误、格式错误
2. **验证错误**：上下文错误、依赖关系错误
3. **运行时错误**：文件 I/O 错误、转换错误

建议使用 `errors.EnhancedParser` 来获得更好的错误信息和建议。

## 最佳实践

1. **总是进行配置验证**：使用 `ConfigValidator` 验证配置
2. **使用类型化参数**：利用参数类型系统确保正确性
3. **定期安全检查**：使用 `utils.CheckSecurity` 检查安全性
4. **应用优化建议**：使用 `utils.OptimizeConfig` 改进性能
5. **错误处理**：使用增强解析器获得更好的错误信息

## 版本兼容性

gonginx 遵循语义版本控制。主要版本变更可能包含破坏性更改，次要版本添加新功能，补丁版本修复错误。

## 贡献

欢迎贡献代码、报告问题或提出建议。请参阅项目的贡献指南。
