# Nginx 配置 CRUD 操作指南

这个示例展示了如何使用 gonginx 框架对 nginx 配置进行**增删改查**（CRUD）操作。

## 快速开始

```bash
cd examples/nginx-crud
go run main.go
```

## CRUD 操作概览

### 1. CREATE (创建) 操作

#### 创建新的 Upstream
```go
// 创建 upstream 指令
upstreamDirective := &config.Directive{
    Name:       "upstream",
    Parameters: []config.Parameter{config.NewParameter("new_backend")},
    Block: &config.Block{
        Directives: []config.IDirective{
            &config.Directive{
                Name:       "server",
                Parameters: []config.Parameter{config.NewParameter("10.0.2.1:8080")},
            },
        },
    },
}

// 添加到 http 块
httpBlock := conf.FindDirectives("http")[0]
httpBlock.GetBlock().(*config.Block).Directives = append(
    httpBlock.GetBlock().(*config.Block).Directives,
    upstreamDirective,
)
```

#### 创建新的 Server
```go
serverDirective := &config.Directive{
    Name: "server",
    Block: &config.Block{
        Directives: []config.IDirective{
            &config.Directive{
                Name:       "listen",
                Parameters: []config.Parameter{config.NewParameter("8080")},
            },
            &config.Directive{
                Name:       "server_name",
                Parameters: []config.Parameter{config.NewParameter("api.example.com")},
            },
        },
    },
}
```

#### 创建新的 Location
```go
locationDirective := &config.Directive{
    Name:       "location",
    Parameters: []config.Parameter{config.NewParameter("/api")},
    Block: &config.Block{
        Directives: []config.IDirective{
            &config.Directive{
                Name:       "proxy_pass",
                Parameters: []config.Parameter{config.NewParameter("http://backend")},
            },
        },
    },
}
```

### 2. READ (读取) 操作

#### 读取所有 Upstream
```go
upstreams := conf.FindUpstreams()
for _, upstream := range upstreams {
    fmt.Printf("Upstream: %s\n", upstream.UpstreamName)
    for _, server := range upstream.UpstreamServers {
        fmt.Printf("  Server: %s\n", server.Address)
    }
}
```

#### 读取所有 Server
```go
servers := conf.FindServers()
for _, server := range servers {
    listenPorts := server.GetListenPorts()
    serverNames := server.GetServerNames()
    fmt.Printf("Server - Listen: %v, Names: %v\n", listenPorts, serverNames)
}
```

#### 按名称查找 Server
```go
servers := conf.FindServersByName("example.com")
if len(servers) > 0 {
    server := servers[0]
    locations := server.GetLocations()
    fmt.Printf("Found %d locations\n", len(locations))
}
```

#### 搜索特定指令
```go
// 查找所有 proxy_pass 指令
proxyPasses := conf.FindDirectives("proxy_pass")
for _, directive := range proxyPasses {
    params := directive.GetParameters()
    if len(params) > 0 {
        fmt.Printf("Proxy pass to: %s\n", params[0].GetValue())
    }
}

// 获取所有 SSL 证书路径
sslCerts := conf.GetAllSSLCertificates()
for _, cert := range sslCerts {
    fmt.Printf("SSL Certificate: %s\n", cert)
}
```

#### 高级搜索
```go
// 按 upstream 名称查找
upstream := conf.FindUpstreamByName("backend")

// 按 location 模式查找
locations := conf.FindLocationsByPattern("/api")

// 获取所有 upstream 服务器
allServers := conf.GetAllUpstreamServers()
```

### 3. UPDATE (更新) 操作

#### 更新 Upstream 服务器
```go
upstream := conf.FindUpstreamByName("backend")
if upstream != nil && len(upstream.UpstreamServers) > 0 {
    // 更新第一个服务器的地址和参数
    upstream.UpstreamServers[0].Address = "192.168.1.10:8080"
    upstream.UpstreamServers[0].Parameters = []config.Parameter{
        config.NewParameter("weight=5"),
        config.NewParameter("max_fails=2"),
    }
    
    // 添加新服务器
    upstream.AddServer(&config.UpstreamServer{
        Address: "192.168.1.20:8080",
        Parameters: []config.Parameter{config.NewParameter("weight=1")},
    })
}
```

#### 更新 Server 配置
```go
servers := conf.FindServersByName("example.com")
if len(servers) > 0 {
    server := servers[0]
    
    // 更新 root 指令
    rootDirectives := server.GetBlock().FindDirectives("root")
    if len(rootDirectives) > 0 {
        rootDirective := rootDirectives[0].(*config.Directive)
        rootDirective.Parameters[0] = config.NewParameter("/var/www/new-html")
    }
    
    // 添加新指令
    newDirective := &config.Directive{
        Name:       "access_log",
        Parameters: []config.Parameter{config.NewParameter("/var/log/nginx/access.log")},
    }
    server.GetBlock().(*config.Block).Directives = append(
        server.GetBlock().(*config.Block).Directives,
        newDirective,
    )
}
```

#### 更新 Location 配置
```go
locations := conf.FindLocationsByPattern("/api")
if len(locations) > 0 {
    location := locations[0]
    
    // 更新 proxy_pass
    proxyPasses := location.GetBlock().FindDirectives("proxy_pass")
    if len(proxyPasses) > 0 {
        proxyPass := proxyPasses[0].(*config.Directive)
        proxyPass.Parameters[0] = config.NewParameter("http://new_backend")
    }
}
```

#### 更新全局指令
```go
// 更新 worker_processes
workerProcesses := conf.FindDirectives("worker_processes")
if len(workerProcesses) > 0 {
    directive := workerProcesses[0].(*config.Directive)
    directive.Parameters[0] = config.NewParameter("4")
}
```

### 4. DELETE (删除) 操作

#### 删除 Upstream 服务器
```go
upstream := conf.FindUpstreamByName("backend")
if upstream != nil && len(upstream.UpstreamServers) > 1 {
    // 删除最后一个服务器
    upstream.UpstreamServers = upstream.UpstreamServers[:len(upstream.UpstreamServers)-1]
}
```

#### 删除整个 Upstream
```go
httpBlocks := conf.FindDirectives("http")
if len(httpBlocks) > 0 {
    httpBlock := httpBlocks[0]
    directives := httpBlock.GetBlock().(*config.Block).Directives
    
    for i, directive := range directives {
        if directive.GetName() == "upstream" {
            params := directive.GetParameters()
            if len(params) > 0 && params[0].GetValue() == "api_servers" {
                // 删除这个 upstream
                httpBlock.GetBlock().(*config.Block).Directives = append(
                    directives[:i],
                    directives[i+1:]...,
                )
                break
            }
        }
    }
}
```

#### 删除 Location
```go
servers := conf.FindServers()
if len(servers) > 0 {
    server := servers[0]
    directives := server.GetBlock().(*config.Block).Directives
    
    for i, directive := range directives {
        if directive.GetName() == "location" {
            params := directive.GetParameters()
            if len(params) > 0 && params[0].GetValue() == "/health" {
                // 删除这个 location
                server.GetBlock().(*config.Block).Directives = append(
                    directives[:i],
                    directives[i+1:]...,
                )
                break
            }
        }
    }
}
```

#### 删除指令
```go
httpBlocks := conf.FindDirectives("http")
if len(httpBlocks) > 0 {
    httpBlock := httpBlocks[0]
    directives := httpBlock.GetBlock().(*config.Block).Directives
    
    for i, directive := range directives {
        if directive.GetName() == "sendfile" {
            // 删除指令
            httpBlock.GetBlock().(*config.Block).Directives = append(
                directives[:i],
                directives[i+1:]...,
            )
            break
        }
    }
}
```

### 5. 保存配置

#### 生成配置字符串
```go
configContent := dumper.DumpConfig(conf, dumper.IndentedStyle)
fmt.Println(configContent)
```

#### 保存到文件
```go
import "os"

err := os.WriteFile("nginx.conf", []byte(configContent), 0644)
if err != nil {
    log.Fatal("保存失败:", err)
}
```

## 高级功能

### 使用高级搜索 API
```go
// 按名称查找服务器
servers := conf.FindServersByName("example.com")

// 按名称查找 upstream
upstream := conf.FindUpstreamByName("backend")

// 按模式查找 location
locations := conf.FindLocationsByPattern("/api")

// 获取所有 SSL 证书
sslCerts := conf.GetAllSSLCertificates()

// 获取所有 upstream 服务器
allServers := conf.GetAllUpstreamServers()
```

### 配置验证
```go
import "github.com/lefeck/gonginx/config"

// 创建验证器
validator := config.NewConfigValidator()
report := validator.ValidateConfig(conf)

if report.HasErrors() {
    fmt.Println("配置验证失败")
    for _, issue := range report.GetByLevel(config.ValidationError) {
        fmt.Printf("错误: %s\n", issue.String())
    }
}
```

### 从文件加载配置
```go
// 从文件解析
p, err := parser.NewParser("/etc/nginx/nginx.conf")
if err != nil {
    log.Fatal(err)
}

conf, err := p.Parse()
if err != nil {
    log.Fatal(err)
}
```

## 常用模式

### 批量操作
```go
// 批量添加服务器到 upstream
upstream := conf.FindUpstreamByName("backend")
if upstream != nil {
    servers := []string{
        "10.0.1.1:8080",
        "10.0.1.2:8080", 
        "10.0.1.3:8080",
    }
    
    for _, serverAddr := range servers {
        upstream.AddServer(&config.UpstreamServer{
            Address: serverAddr,
            Parameters: []config.Parameter{config.NewParameter("weight=1")},
        })
    }
}
```

### 条件操作
```go
// 只有当 upstream 不存在时才创建
if conf.FindUpstreamByName("new_backend") == nil {
    // 创建新的 upstream
    createUpstream(conf)
}
```

### 安全删除
```go
// 删除前检查依赖
upstream := conf.FindUpstreamByName("backend")
if upstream != nil {
    // 检查是否有 proxy_pass 引用这个 upstream
    proxyPasses := conf.FindDirectives("proxy_pass")
    hasReference := false
    
    for _, directive := range proxyPasses {
        params := directive.GetParameters()
        if len(params) > 0 && strings.Contains(params[0].GetValue(), "backend") {
            hasReference = true
            break
        }
    }
    
    if !hasReference {
        // 安全删除 upstream
        deleteUpstream(conf, "backend")
    } else {
        fmt.Println("警告: upstream 'backend' 仍被引用，无法删除")
    }
}
```

这个框架提供了完整的 nginx 配置 CRUD 操作能力，你可以根据需要进行配置的动态管理和修改。
