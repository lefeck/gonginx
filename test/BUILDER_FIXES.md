# Builder 模式修复总结

## 问题描述

在 `test/test.go` 中发现了以下问题：

1. **缺少方法**：`TCPNoPush("on")` 和 `KeepaliveTimeout("65")` 方法不存在或被注释
2. **链式调用问题**：`Upstream("api_backend").End()` 后不能继续调用 `Server()`，因为 `End()` 方法返回了错误的类型

## 修复内容

### 1. 添加缺失的方法

在 `generator/http_builder.go` 中添加了：

```go
// TCPNoPush is an alias for TcpNoPush for consistency with nginx directive naming
func (hb *HTTPBuilder) TCPNoPush(value string) *HTTPBuilder {
	return hb.AddDirective("tcp_nopush", value)
}

// DefaultType sets default_type
func (hb *HTTPBuilder) DefaultType(mimeType string) *HTTPBuilder {
	return hb.AddDirective("default_type", mimeType)
}
```

### 2. 修复链式调用问题

**问题原因**：各个 Builder 的 `End()` 方法返回类型不正确，导致链式调用中断。

**修复方案**：

#### UpstreamBuilder.End()
```go
// 修复前：返回 *ConfigBuilder
func (ub *UpstreamBuilder) End() *ConfigBuilder {
	return &ConfigBuilder{config: ub.config}
}

// 修复后：返回 *HTTPBuilder，可以继续在 HTTP 块中添加其他元素
func (ub *UpstreamBuilder) End() *HTTPBuilder {
	return ub.httpBuilder
}
```

#### ServerBuilder.End()
```go
// 修复前：返回 *ConfigBuilder
func (sb *ServerBuilder) End() *ConfigBuilder {
	return &ConfigBuilder{config: sb.config}
}

// 修复后：返回 *HTTPBuilder，可以继续在 HTTP 块中添加其他元素
func (sb *ServerBuilder) End() *HTTPBuilder {
	return sb.httpBuilder
}
```

#### LocationBuilder.End()
```go
// 修复前：返回 *ConfigBuilder
func (lb *LocationBuilder) End() *ConfigBuilder {
	return &ConfigBuilder{config: lb.config}
}

// 修复后：返回 *ServerBuilder，可以继续在 Server 块中添加其他元素
func (lb *LocationBuilder) End() *ServerBuilder {
	return lb.serverBuilder
}
```

#### SSLBuilder.End()
```go
// 新增：为 SSLBuilder 添加 End() 方法
func (ssl *SSLBuilder) End() *ServerBuilder {
	return ssl.serverBuilder
}
```

### 3. 修复配置结构问题

修复了 `test.go` 中 `Config` 结构体的使用：

```go
// 修复前：
conf := &config.Config{
	FilePath:   "generated.conf",
	Directives: []config.IDirective{}, // 错误：Config 没有 Directives 字段
}

// 修复后：
conf := &config.Config{
	Block: &config.Block{
		Directives: []config.IDirective{},
	},
	FilePath: "generated.conf",
}
```

### 4. 修复 Go 版本问题

将 `go.mod` 中的版本从 `go 1.24`（不存在）修改为 `go 1.21`。

## 修复效果

### 修复前的问题代码：
```go
builderConfig := generator.NewConfigBuilder().
	WorkerProcesses("auto").
	WorkerConnections("1024").
	HTTP().
	SendFile(true).
	//TCPNoPush("on").        // 被注释，因为方法不存在
	//KeepaliveTimeout("65"). // 被注释，因为方法不存在
	Upstream("api_backend").
	Server("10.0.1.10:8080", "weight=3").
	End().
	// Server().  // 错误：End() 返回 ConfigBuilder，没有 Server() 方法
```

### 修复后的工作代码：
```go
builderConfig := generator.NewConfigBuilder().
	WorkerProcesses("auto").
	WorkerConnections("1024").
	HTTP().
	SendFile(true).
	TCPNoPush("on").        // ✅ 现在可以工作
	KeepaliveTimeout("65"). // ✅ 现在可以工作
	Upstream("api_backend").
	Server("10.0.1.10:8080", "weight=3").
	Server("10.0.1.11:8080", "weight=2").
	End().                  // ✅ 返回 HTTPBuilder
	Server().               // ✅ 现在可以继续添加 Server
	Listen("80").
	ServerName("api.example.com").
	// ... 更多配置
	End().                  // ✅ 返回 HTTPBuilder
	Build()
```

## 测试验证

1. **原始测试**：`test/test.go` 现在可以正常运行
2. **增强测试**：`test/test_fixed.go` 验证了所有修复的功能
3. **复杂场景**：测试了多个 upstream、多个 server、SSL 配置等复杂链式调用

## 链式调用流程图

```
ConfigBuilder
    ↓ HTTP()
HTTPBuilder
    ↓ Upstream("name")
UpstreamBuilder
    ↓ End()
HTTPBuilder ← (修复：原来返回 ConfigBuilder)
    ↓ Server()
ServerBuilder
    ↓ Location("/")
LocationBuilder
    ↓ End()
ServerBuilder ← (修复：原来返回 ConfigBuilder)
    ↓ End()
HTTPBuilder ← (修复：原来返回 ConfigBuilder)
    ↓ End()
ConfigBuilder
    ↓ Build()
*Config
```

## 总结

通过这些修复，Builder 模式现在支持：

1. ✅ 完整的方法覆盖（TCPNoPush、KeepaliveTimeout 等）
2. ✅ 正确的链式调用（Upstream().End().Server() 等）
3. ✅ 复杂的嵌套配置构建
4. ✅ 多个同类型块的添加（多个 upstream、server 等）
5. ✅ 灵活的上下文切换（在不同的 builder 之间正确跳转）

现在用户可以使用流畅的 Builder API 来构建复杂的 nginx 配置，而不会遇到链式调用中断的问题。
