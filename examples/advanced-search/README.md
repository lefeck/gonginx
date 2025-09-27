# 高级搜索功能示例

本示例展示了 Gonginx 框架新增的高级搜索功能，这些功能让你能够更方便地查找和操作 nginx 配置中的特定元素。

## 新增的搜索方法

### 1. 按服务器名称查找服务器
```go
servers := config.FindServersByName("example.com")
```
- 查找所有包含指定 `server_name` 的服务器块
- 支持多个 `server_name` 的服务器
- 返回 `[]*config.Server` 切片

### 2. 按名称查找上游服务器组
```go
upstream := config.FindUpstreamByName("api_backend")
```
- 查找指定名称的 upstream 块
- 返回 `*config.Upstream` 或 `nil`

### 3. 按模式查找 location 块
```go
locations := config.FindLocationsByPattern("/api")
locations := config.FindLocationsByPattern("~ \\.php$")
```
- 支持精确匹配和带修饰符的模式匹配
- 递归搜索所有嵌套的 location 块
- 返回 `[]*config.Location` 切片

### 4. 获取所有 SSL 证书路径
```go
certificates := config.GetAllSSLCertificates()
```
- 递归搜索所有 `ssl_certificate` 指令
- 返回证书文件路径的字符串切片

### 5. 获取所有上游服务器
```go
allServers := config.GetAllUpstreamServers()
```
- 获取所有 upstream 块中的所有服务器
- 返回 `[]*config.UpstreamServer` 切片

## 运行示例

```bash
cd examples/advanced-search
go run main.go
```

## 测试

新功能包含完整的测试用例：

```bash
go test ./config -v -run "TestFind|TestGetAll"
```

## 使用场景

这些搜索功能特别适用于：

- **配置管理工具**: 快速查找和修改特定的服务器配置
- **监控系统**: 提取所有上游服务器进行健康检查
- **SSL 证书管理**: 获取所有证书路径进行续期管理
- **负载均衡器配置**: 管理 upstream 服务器列表
- **自动化部署**: 根据域名查找相应的服务器配置

这些功能完成了 README.md 中 TODO 列表的第一项：**实现特定搜索，如按 server_name (域名) 查找服务器或按目标查找任何 upstream 等**。
