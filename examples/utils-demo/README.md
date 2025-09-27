# 实用工具功能示例

这个示例展示了 gonginx 库的实用工具功能，包括配置差异比较、安全检查、优化建议和格式转换。

## 功能特性

### 1. 配置差异比较 (diff.go)
- **配置比较**: 比较两个nginx配置文件的差异
- **差异类型**: 支持添加、删除、修改、移动四种差异类型
- **详细报告**: 提供路径、指令名称、旧值、新值等详细信息
- **分类显示**: 按差异类型分组显示结果

#### 支持的比较功能
- 字符串配置比较
- 配置对象比较
- 配置映射比较
- 差异统计和汇总

### 2. 安全检查 (security.go)
- **安全等级**: 信息、警告、严重三个等级
- **全面检查**: 涵盖SSL、访问控制、信息泄露等方面
- **智能建议**: 提供具体的修复建议和参考文档
- **评分系统**: 基于问题严重程度的安全评分

#### 检查项目
- **信息泄露**: server_tokens、错误页面、目录列表
- **SSL/TLS安全**: 协议版本、密码套件、HSTS
- **访问控制**: 敏感路径保护、IP限制
- **安全头**: X-Frame-Options、CSP、XSS保护等
- **文件上传**: 上传限制、文件类型检查
- **速率限制**: DoS防护、连接限制
- **日志配置**: 访问日志、错误日志
- **权限配置**: 用户权限、文件权限

### 3. 配置优化 (optimizer.go)
- **多维度优化**: 性能、大小、安全、维护性
- **智能分析**: 自动检测配置中的优化机会
- **影响评估**: 评估优化建议的影响程度
- **实施指导**: 提供具体的实施步骤

#### 优化类别
- **性能优化**: 
  - Worker进程配置
  - 缓冲区大小优化
  - Keepalive设置
  - Gzip压缩配置
  - SSL性能优化
  - 缓存配置

- **安全优化**:
  - SSL协议升级
  - 安全头添加
  - 访问控制改进

- **大小优化**:
  - 重复指令合并
  - 默认值移除
  - 指令整理

- **维护性优化**:
  - 注释添加
  - 指令组织
  - 配置结构化

### 4. 格式转换 (converter.go)
- **多格式支持**: JSON、YAML、TOML
- **双向转换**: 支持格式间的相互转换
- **结构保持**: 保持配置的层次结构
- **类型识别**: 智能识别参数类型

#### 转换功能
- Nginx配置 → JSON
- Nginx配置 → YAML
- JSON → YAML
- YAML → JSON
- 配置对象序列化

## API 使用

### 配置差异比较
```go
import "github.com/lefeck/gonginx/utils"

// 比较配置字符串
diffResult, err := utils.CompareConfigStrings(oldConfig, newConfig)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("差异总结: %s\n", diffResult.Summary.String())

// 按类型获取差异
added := diffResult.GetByType(utils.DiffAdded)
modified := diffResult.GetByType(utils.DiffModified)
removed := diffResult.GetByType(utils.DiffRemoved)
```

### 安全检查
```go
// 执行安全检查
securityReport := utils.CheckSecurity(config)

fmt.Printf("安全评分: %d/100\n", securityReport.Summary.Score)

// 获取严重问题
criticalIssues := securityReport.GetByLevel(utils.SecurityCritical)
for _, issue := range criticalIssues {
    fmt.Printf("严重问题: %s\n", issue.Title)
    fmt.Printf("修复建议: %s\n", issue.Fix)
}

// 按类别获取问题
sslIssues := securityReport.GetByCategory("SSL/TLS Security")
```

### 配置优化
```go
// 执行优化分析
optimizationReport := utils.OptimizeConfig(config)

fmt.Printf("优化建议: %s\n", optimizationReport.Summary.String())

// 获取性能优化建议
performanceOpts := optimizationReport.GetByType(utils.OptimizePerformance)
for _, suggestion := range performanceOpts {
    fmt.Printf("建议: %s\n", suggestion.Title)
    fmt.Printf("当前: %s → 建议: %s\n", suggestion.CurrentValue, suggestion.SuggestedValue)
}
```

### 格式转换
```go
// 创建转换器
converter := utils.NewConfigConverter(config)

// 转换为JSON
jsonStr, err := converter.ConvertToJSON(true) // true为美化输出
if err != nil {
    log.Fatal(err)
}

// 转换为YAML
yamlStr, err := converter.ConvertToYAML()
if err != nil {
    log.Fatal(err)
}

// 格式间转换
formatConverter := utils.NewFormatConverter()
yamlOutput, err := formatConverter.Convert(jsonInput, utils.FormatJSON, utils.FormatYAML)
```

## 运行示例

```bash
cd examples/utils-demo
go run main.go
```

## 示例输出

### 1. 配置差异比较
```
配置差异总结: Total: 6, Added: 3, Removed: 0, Modified: 2, Moved: 0

详细差异:
1. + [worker_processes] worker_processes: auto
2. ~ [events/worker_connections] worker_connections: 512 -> 1024
3. + [events/use] use: epoll
4. + [http/tcp_nopush] tcp_nopush: on
5. + [http/gzip] gzip: on
6. + [http/upstream] backend: added
```

### 2. 安全检查
```
安全评估结果: Security Score: 45/100, Issues: 8 (Critical: 2, Warning: 4, Info: 2)

严重问题 (2):
1. [CRITICAL] SSL/TLS Security: Insecure SSL/TLS protocol enabled
   修复建议: Use only TLSv1.2 and TLSv1.3
2. [CRITICAL] Access Control: Sensitive location without access control
   修复建议: Add IP restrictions, authentication, or deny directives
```

### 3. 配置优化
```
优化分析结果: Optimizations: 12 (Performance: 6, Size: 2, Security: 2, Maintenance: 2)

性能优化 (6):
1. Add worker_connections directive
   描述: Worker connections not configured
   影响: High
   建议值: 1024
   实现: Add 'worker_connections 1024;' to events block
```

### 4. 格式转换
```
JSON格式:
{
  "worker_processes": "auto",
  "events": {
    "worker_connections": "1024"
  },
  "http": {
    "sendfile": "on",
    "gzip": "on",
    "servers": [...]
  }
}

YAML格式:
worker_processes: auto
events:
  worker_connections: "1024"
http:
  sendfile: "on"
  gzip: "on"
  servers: [...]
```

## 实际应用场景

### DevOps 自动化
```go
// 配置部署前检查
func validateConfigBeforeDeploy(configPath string) error {
    config, err := parseConfig(configPath)
    if err != nil {
        return err
    }
    
    // 安全检查
    securityReport := utils.CheckSecurity(config)
    if securityReport.HasCriticalIssues() {
        return fmt.Errorf("存在严重安全问题，部署被阻止")
    }
    
    // 优化建议
    optimizationReport := utils.OptimizeConfig(config)
    logOptimizationSuggestions(optimizationReport)
    
    return nil
}
```

### 配置管理
```go
// 配置变更管理
func trackConfigChanges(oldConfig, newConfig string) {
    diffResult, err := utils.CompareConfigStrings(oldConfig, newConfig)
    if err != nil {
        log.Printf("配置比较失败: %v", err)
        return
    }
    
    // 记录变更日志
    for _, diff := range diffResult.Differences {
        auditLog.Record(diff.Type, diff.Path, diff.DirectiveName, diff.OldValue, diff.NewValue)
    }
}
```

### 安全审计
```go
// 定期安全审计
func performSecurityAudit(configs []string) {
    for _, configPath := range configs {
        config, err := parseConfig(configPath)
        if err != nil {
            continue
        }
        
        securityReport := utils.CheckSecurity(config)
        generateSecurityReport(configPath, securityReport)
        
        if securityReport.Summary.Score < 70 {
            alertSecurityTeam(configPath, securityReport)
        }
    }
}
```

### 配置标准化
```go
// 配置标准化检查
func enforceConfigStandards(configPath string) {
    config, err := parseConfig(configPath)
    if err != nil {
        return
    }
    
    optimizationReport := utils.OptimizeConfig(config)
    
    // 强制执行关键优化
    for _, suggestion := range optimizationReport.Suggestions {
        if suggestion.Impact == "High" {
            autoApplyOptimization(suggestion)
        }
    }
}
```

## 扩展功能

### 自定义安全规则
```go
type CustomSecurityRule struct {
    Name        string
    Check       func(*config.Config) []utils.SecurityIssue
    Level       utils.SecurityLevel
    Category    string
}

func addCustomSecurityRules(checker *utils.SecurityChecker) {
    // 添加自定义安全检查规则
}
```

### 自定义优化规则
```go
type CustomOptimizationRule struct {
    Name        string
    Analyze     func(*config.Config) []utils.OptimizationSuggestion
    Type        utils.OptimizationType
    Category    string
}
```

### 输出格式扩展
可以扩展支持更多输出格式：
- XML
- HCL (HashiCorp Configuration Language)
- INI
- 自定义格式

### 集成示例
```go
// CI/CD 管道集成
func integrateCIPipeline() {
    // 在构建管道中集成配置检查
    if !passesSecurityCheck(configFile) {
        os.Exit(1)
    }
    
    if !meetsPerformanceStandards(configFile) {
        log.Println("警告: 配置可能影响性能")
    }
}
```
