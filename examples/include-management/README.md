# Include 配置文件管理指南

这个示例展示了如何使用 gonginx 框架对 nginx 配置中的 `include` 指令进行完整的**增删改查**操作。

## 快速开始

```bash
cd examples/include-management
go run main.go
```

## Include 指令的使用场景

在 nginx 配置中，`include` 指令常用于：

1. **模块化配置** - 将不同功能的配置分离到不同文件
2. **虚拟主机管理** - 每个站点一个配置文件
3. **配置片段复用** - 通用的配置片段可以被多个地方引用
4. **环境隔离** - 开发、测试、生产环境的配置分离

## 典型的配置结构

```
nginx.conf                 # 主配置文件
├── conf/
│   ├── mime.types         # MIME 类型定义
│   ├── snippets/          # 配置片段
│   │   ├── ssl.conf       # SSL 通用配置
│   │   └── proxy.conf     # 代理通用配置
│   ├── vhosts/            # 虚拟主机配置
│   │   ├── example.conf   # 站点配置
│   │   └── api.conf       # API 站点配置
│   └── api/               # API 相关配置
│       └── backends.conf  # 后端服务配置
```

## Include 操作方法

### 1. 创建 Include 指令

```go
// 在 http 块中添加 include
httpBlocks := conf.FindDirectives("http")
if len(httpBlocks) > 0 {
    if httpBlock, ok := httpBlocks[0].(*config.HTTP); ok {
        includeDirective := &config.Directive{
            Name:       "include",
            Parameters: []config.Parameter{config.NewParameter("conf/vhosts/*.conf")},
        }
        httpBlock.Directives = append(httpBlock.Directives, includeDirective)
    }
}
```

### 2. 查找和读取 Include 指令

```go
// 查找所有 include 指令
includes := conf.FindDirectives("include")
for _, inc := range includes {
    params := inc.GetParameters()
    if len(params) > 0 {
        includePath := params[0].GetValue()
        fmt.Printf("Found include: %s\n", includePath)
    }
}
```

### 3. 解析 Include 文件内容

```go
// 使用 WithIncludeParsing 选项自动解析 include 文件
p, err := parser.NewParser("nginx.conf", parser.WithIncludeParsing())
conf, err := p.Parse()

// 现在 conf 包含了所有 include 文件的内容
upstreams := conf.FindUpstreams()  // 包括被包含文件中的 upstream
servers := conf.FindDirectives("server")  // 包括被包含文件中的 server
```

### 4. 操作被包含的文件

```go
// 解析单个被包含文件
p, err := parser.NewParser("conf/vhosts/example.conf")
conf, err := p.Parse()

// 修改内容
servers := conf.FindDirectives("server")
if len(servers) > 0 {
    serverBlock := servers[0].GetBlock()
    // 添加新的 location 等...
}

// 保存修改
modifiedConfig := dumper.DumpConfig(conf, dumper.IndentedStyle)
err = os.WriteFile("conf/vhosts/example.conf", []byte(modifiedConfig), 0644)
```

### 5. 删除 Include 指令

```go
// 在 http 块中删除特定的 include
httpBlocks := conf.FindDirectives("http")
if len(httpBlocks) > 0 {
    if httpBlock, ok := httpBlocks[0].(*config.HTTP); ok {
        newDirectives := make([]config.IDirective, 0)
        
        for _, directive := range httpBlock.Directives {
            if directive.GetName() == "include" {
                params := directive.GetParameters()
                if len(params) > 0 && params[0].GetValue() == "unwanted.conf" {
                    continue // 跳过这个 include（删除它）
                }
            }
            newDirectives = append(newDirectives, directive)
        }
        
        httpBlock.Directives = newDirectives
    }
}
```

## 推荐的 API 接口设计

基于实际使用需求，建议设计以下 API 接口：

```go
type IncludeManager struct {
    config     *config.Config
    configPath string
}

// 基础操作
func (im *IncludeManager) AddInclude(blockName, includePath string) error
func (im *IncludeManager) RemoveInclude(includePath string) error
func (im *IncludeManager) ListIncludes() []string

// 文件操作
func (im *IncludeManager) CreateIncludedFile(filePath string, content interface{}) error
func (im *IncludeManager) UpdateIncludedFile(filePath string, updater func(*config.Config) error) error
func (im *IncludeManager) DeleteIncludedFile(filePath string) error

// 验证和保存
func (im *IncludeManager) ValidateIncludes() []error
func (im *IncludeManager) SaveAll() error
```

## 最佳实践

### 1. 文件组织

- **按功能分离**: 将不同功能的配置分到不同文件
- **命名规范**: 使用有意义的文件名，如 `ssl.conf`, `proxy.conf`
- **目录结构**: 建立清晰的目录层次结构

### 2. Include 路径管理

- **使用相对路径**: 便于配置文件的移植和部署
- **通配符使用**: 合理使用 `*.conf` 等通配符
- **避免循环引用**: 确保 include 文件不会相互引用形成循环

### 3. 配置验证

```go
// 验证 include 文件是否存在
func validateIncludePaths(conf *config.Config) []error {
    var errors []error
    includes := conf.FindDirectives("include")
    
    for _, inc := range includes {
        params := inc.GetParameters()
        if len(params) > 0 {
            path := params[0].GetValue()
            if _, err := os.Stat(path); os.IsNotExist(err) {
                errors = append(errors, fmt.Errorf("include file not found: %s", path))
            }
        }
    }
    
    return errors
}
```

### 4. 配置合并策略

- **上下文感知**: 确保被包含的配置在正确的上下文中
- **指令优先级**: 了解 nginx 指令的继承和覆盖规则
- **冲突处理**: 处理相同指令的冲突情况

## 注意事项

1. **上下文限制**: 
   - `conf/vhosts/*.conf` 中的文件只能包含 server 块和 http 级别的指令
   - 不能在 server 块的 include 文件中定义 http 块

2. **解析顺序**: 
   - include 指令按出现顺序解析
   - 通配符按文件名字母顺序解析

3. **性能考虑**:
   - 过多的 include 文件可能影响解析性能
   - 建议合理组织文件数量和层级

4. **错误处理**:
   - include 文件不存在时的处理策略
   - 语法错误时的容错机制

## 完整示例

运行 `go run main.go` 查看完整的 include 操作示例，包括：

- ✅ 创建主配置文件和 include 结构
- ✅ 添加和删除 include 指令
- ✅ 创建和修改被包含的配置文件
- ✅ 使用 WithIncludeParsing 解析完整配置
- ✅ 配置验证和错误处理

这个示例展示了在实际项目中如何优雅地管理复杂的 nginx 配置结构。
