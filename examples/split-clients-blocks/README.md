# Split Clients 块功能支持

本示例展示了 Gonginx 框架新增的 nginx split_clients 块支持功能，这是 nginx 中用于A/B测试和流量分割的重要功能。

## Split Clients 块简介

nginx 的 split_clients 块允许你根据某个变量（如客户端IP、请求ID等）按百分比将流量分配到不同的值，支持：
- 百分比流量分割 (0.5%, 10%, 25% 等)
- 通配符匹配 (*) 处理剩余流量
- 基于各种nginx变量的分割策略
- A/B测试和金丝雀部署

## 新增的 API

### 1. 查找 Split Clients 块
```go
// 查找所有 split_clients 块
splitClients := config.FindSplitClients()

// 按目标变量查找特定的 split_clients 块
splitBlock := config.FindSplitClientsByVariable("$ui_version")

// 按源变量和目标变量查找 split_clients 块
splitBlock := config.FindSplitClientsByVariables("$remote_addr", "$variant")
```

### 2. Split Clients 块操作
```go
// 添加新的百分比分割
splitBlock.AddEntry("5.0%", "new_variant")

// 获取总分配百分比
total, err := splitBlock.GetTotalPercentage()

// 检查是否有通配符条目
hasWildcard := splitBlock.HasWildcard()

// 获取通配符的值
wildcardValue := splitBlock.GetWildcardValue()

// 删除条目
removed := splitBlock.RemoveEntry("5.0%")

// 按值查找条目
entries := splitBlock.GetEntriesByValue("variant_a")

// 访问分割条目
for _, entry := range splitBlock.Entries {
    fmt.Printf("%s -> %s\n", entry.Percentage, entry.Value)
}
```

### 3. Split Clients 块属性
```go
type SplitClients struct {
    Variable       string               // 源变量 (如 $remote_addr)
    MappedVariable string               // 目标变量 (如 $variant)
    Entries        []*SplitClientsEntry // 分割条目列表
    Comment        []string             // 注释
}

type SplitClientsEntry struct {
    Percentage string // 百分比 (如 "0.5%", "10%", "*")
    Value      string // 分配的值
    Comment    []string // 注释
}
```

## 支持的 Split Clients 语法

### 基本百分比分割
```nginx
split_clients $remote_addr $variant {
    0.5%     variant_a;
    2.0%     variant_b;
    10%      variant_c;
    *        variant_default;
}
```

### A/B 测试配置
```nginx
split_clients $request_id $ab_test {
    50%      version_a;
    *        version_b;
}
```

### 多阶段分割
```nginx
split_clients $http_user_agent $user_type {
    10%      mobile;
    20%      tablet;
    *        desktop;
}
```

### 金丝雀部署
```nginx
split_clients $remote_addr $deployment {
    1%       canary;
    9%       staging;
    *        production;
}
```

## 常见使用场景

### 1. A/B 测试
根据用户进行不同版本的测试：
```nginx
split_clients $remote_addr $ui_version {
    10%      new_ui;
    *        old_ui;
}

server {
    location / {
        if ($ui_version = "new_ui") {
            proxy_pass http://new-ui-backend;
        }
        proxy_pass http://old-ui-backend;
    }
}
```

### 2. 功能开关
按百分比启用新功能：
```nginx
split_clients $request_id $feature_flag {
    25%      enabled;
    *        disabled;
}

server {
    location /api/feature {
        if ($feature_flag = "enabled") {
            proxy_pass http://feature-enabled-backend;
        }
        return 404;
    }
}
```

### 3. 金丝雀部署
渐进式部署新版本：
```nginx
split_clients $remote_addr $backend_version {
    1%       canary;     # 1% 流量到金丝雀版本
    9%       staging;    # 9% 流量到预发布版本
    *        production; # 90% 流量到生产版本
}

upstream backend_canary {
    server canary.example.com:8080;
}

upstream backend_staging {
    server staging.example.com:8080;
}

upstream backend_production {
    server prod1.example.com:8080;
    server prod2.example.com:8080;
}

map $backend_version $backend {
    canary backend_canary;
    staging backend_staging;
    production backend_production;
}

server {
    location / {
        proxy_pass http://$backend;
    }
}
```

### 4. 缓存策略测试
测试不同的缓存策略：
```nginx
split_clients $request_uri $cache_strategy {
    20%      aggressive;
    30%      moderate;
    *        conservative;
}

server {
    location ~* \.(css|js|png|jpg)$ {
        if ($cache_strategy = "aggressive") {
            expires 1y;
        }
        if ($cache_strategy = "moderate") {
            expires 30d;
        }
        if ($cache_strategy = "conservative") {
            expires 1d;
        }
        
        try_files $uri @backend;
    }
}
```

### 5. 移动应用版本分发
控制移动应用的版本分发：
```nginx
split_clients $http_user_agent $app_version {
    5%       beta;
    15%      stable;
    *        legacy;
}

server {
    location /mobile-app/ {
        if ($app_version = "beta") {
            proxy_pass http://mobile-beta.example.com;
        }
        if ($app_version = "stable") {
            proxy_pass http://mobile-stable.example.com;
        }
        proxy_pass http://mobile-legacy.example.com;
    }
}
```

## 高级特性

### 百分比验证
框架自动验证百分比格式和范围：
```go
// 自动验证
err := splitBlock.AddEntry("150%", "invalid") // 错误：超出范围
err = splitBlock.AddEntry("abc%", "invalid")  // 错误：格式错误
err = splitBlock.AddEntry("10%", "valid")     // 成功
```

### 总百分比检查
确保分配的总百分比不超过100%：
```go
total, err := splitBlock.GetTotalPercentage()
if total > 100.0 {
    // 处理超出100%的情况
}
```

### 通配符处理
自动处理剩余流量：
```nginx
split_clients $remote_addr $variant {
    10%      variant_a;
    20%      variant_b;
    *        variant_default; # 处理剩余70%的流量
}
```

### 分割变量选择
支持多种nginx变量作为分割基础：
- `$remote_addr` - 基于客户端IP
- `$request_id` - 基于请求ID（更均匀的分布）
- `$http_user_agent` - 基于用户代理
- `$request_uri` - 基于请求URI
- `$cookie_user_id` - 基于用户Cookie

## 运行示例

```bash
cd examples/split-clients-blocks
go run main.go
```

示例将演示：
- 解析包含多个 split_clients 块的配置
- 查找和操作 split_clients 块
- 添加和删除分割条目
- 百分比验证和计算
- 生成修改后的配置文件

## 与其他功能集成

Split Clients 块通常与以下功能结合使用：
- **Map 块**: 将分割结果映射到后端服务器
- **Upstream 块**: 不同版本的负载均衡配置
- **Location 块**: 基于分割结果的路由规则
- **条件语句**: if语句进行更复杂的逻辑判断

## 最佳实践

1. **渐进式推出**: 从小百分比开始，逐步增加新版本的流量
2. **监控指标**: 配合日志和监控系统跟踪A/B测试效果
3. **回滚策略**: 准备快速回滚到稳定版本的方案
4. **用户体验**: 确保用户在会话期间保持一致的体验
5. **性能影响**: 监控不同版本的性能差异
