# Limit Conn Zone 功能支持

本示例展示了 Gonginx 框架新增的 nginx limit_conn_zone 指令支持功能，这是 nginx 中用于并发连接数限制的重要功能。

## Limit Conn Zone 简介

nginx 的 limit_conn_zone 指令用于定义限制并发连接数的共享内存区域，支持：
- 基于各种变量的连接限制 (IP地址、用户ID、请求URI等)
- 内存区域大小配置 (k, m, g)
- 并发连接数控制
- 分布式环境同步 (sync参数)
- 多层限制策略

## 新增的 API

### 1. 查找 Limit Conn Zone 指令
```go
// 查找所有 limit_conn_zone 指令
zones := config.FindLimitConnZones()

// 按区域名查找特定的 limit_conn_zone
zone := config.FindLimitConnZoneByName("addr")

// 按键变量查找 limit_conn_zone
zones := config.FindLimitConnZonesByKey("$binary_remote_addr")
```

### 2. Limit Conn Zone 操作
```go
// 获取/设置内存区域大小
sizeBytes, err := zone.GetZoneSizeBytes() // 获取字节数
err = zone.SetZoneSize("20m")             // 设置新大小

// 连接容量估算
maxConnections, err := zone.EstimateMaxConnections()

// 获取推荐限制值
recommendations, err := zone.GetRecommendedLimits()

// 同步设置
zone.SetSync(true)  // 启用同步
zone.SetSync(false) // 禁用同步

// 兼容性检查
compatible := zone1.IsCompatibleWith(zone2)

// 内存使用估算
memoryInfo := zone.GetMemoryUsageEstimate()

// 访问区域属性
fmt.Printf("Key: %s\n", zone.Key)
fmt.Printf("Zone: %s\n", zone.ZoneName)
fmt.Printf("Size: %s\n", zone.ZoneSize)
fmt.Printf("Sync: %t\n", zone.Sync)
```

### 3. Limit Conn Zone 属性
```go
type LimitConnZone struct {
    Key      string   // 键变量 (如 $binary_remote_addr)
    ZoneName string   // 区域名称 (如 "addr")
    ZoneSize string   // 区域大小 (如 "10m")
    Sync     bool     // 是否启用同步
    Comment  []string // 注释
}
```

## 支持的 Limit Conn Zone 语法

### 基本连接限制
```nginx
limit_conn_zone $binary_remote_addr zone=addr:10m;
```

### 基于服务器的限制
```nginx
limit_conn_zone $server_name zone=perserver:5m;
```

### 基于路径的限制
```nginx
limit_conn_zone $request_uri zone=perpath:20m;
```

### 基于用户的限制
```nginx
limit_conn_zone $cookie_user_id zone=peruser:100m;
```

### 分布式同步
```nginx
limit_conn_zone $binary_remote_addr zone=cluster:50m sync;
```

### 复合键限制
```nginx
limit_conn_zone $binary_remote_addr$request_method zone=per_ip_method:50m;
```

## 常见使用场景

### 1. API 保护
限制每个IP的并发API连接：
```nginx
http {
    limit_conn_zone $binary_remote_addr zone=api:10m;
    
    server {
        location /api/ {
            limit_conn api 10;  # 每IP最多10个并发连接
            proxy_pass http://backend;
        }
    }
}
```

### 2. 下载保护
限制每个用户的并发下载数：
```nginx
http {
    limit_conn_zone $cookie_user_id zone=downloads:50m;
    
    server {
        location /download/ {
            limit_conn downloads 3;  # 每用户最多3个并发下载
            proxy_pass http://backend;
        }
    }
}
```

### 3. 流媒体限制
控制流媒体连接数：
```nginx
http {
    limit_conn_zone $binary_remote_addr zone=streaming:20m;
    
    server {
        location /stream/ {
            limit_conn streaming 2;  # 每IP最多2个流
            proxy_pass http://streaming_backend;
        }
    }
}
```

### 4. WebSocket 连接限制
限制WebSocket连接数：
```nginx
http {
    limit_conn_zone $binary_remote_addr zone=websockets:30m;
    limit_conn_zone $cookie_user_id zone=ws_peruser:50m;
    
    server {
        location /ws {
            limit_conn websockets 5;   # 每IP最多5个WebSocket
            limit_conn ws_peruser 3;   # 每用户最多3个WebSocket
            
            proxy_pass http://websocket_backend;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection "upgrade";
        }
    }
}
```

### 5. 多层限制
组合多种连接限制策略：
```nginx
http {
    limit_conn_zone $binary_remote_addr zone=global:10m;
    limit_conn_zone $server_name zone=perserver:5m;
    limit_conn_zone $cookie_user_id zone=peruser:50m;
    
    server {
        location / {
            # 全局IP限制
            limit_conn global 100;
            # 服务器级限制
            limit_conn perserver 500;
            # 用户级限制
            limit_conn peruser 10;
            
            proxy_pass http://backend;
        }
    }
}
```

### 6. 基于地理位置的限制
不同国家不同连接限制：
```nginx
http {
    limit_conn_zone $geoip_country_code zone=percountry:20m;
    
    server {
        location / {
            limit_conn percountry 1000;  # 每国家1000连接
            proxy_pass http://backend;
        }
    }
}
```

## 高级特性

### 内存容量规划
每个连接大约需要64字节内存：
```
1m 内存 ≈ 16,384 个并发连接
10m 内存 ≈ 163,840 个并发连接
100m 内存 ≈ 1,638,400 个并发连接
1g 内存 ≈ 16,777,216 个并发连接
```

### 推荐限制值
框架提供基于内存大小的推荐限制：
```go
recommendations, err := zone.GetRecommendedLimits()
// 返回：
// conservative: 70% of max capacity
// moderate: 85% of max capacity  
// aggressive: 95% of max capacity
```

### 分布式同步
在多服务器环境中同步连接状态：
```nginx
limit_conn_zone $binary_remote_addr zone=cluster:100m sync;
```

### 兼容性检查
检查两个区域是否可以合并：
```go
if zone1.IsCompatibleWith(zone2) {
    // 相同键变量的区域可以合并
}
```

### 键变量选择
支持多种nginx变量作为限制键：
- `$binary_remote_addr` - 客户端IP (推荐，内存效率高)
- `$remote_addr` - 客户端IP (文本格式)
- `$server_name` - 服务器名称
- `$request_uri` - 请求URI
- `$cookie_user_id` - 用户ID cookie
- `$http_user_agent` - 用户代理
- `$geoip_country_code` - 地理位置国家代码
- `$http_x_real_ip` - 真实IP (代理环境)

## 性能和内存考虑

### 内存使用优化
- 使用 `$binary_remote_addr` 而不是 `$remote_addr` 节省内存
- 根据实际并发需求合理分配内存
- 考虑连接的平均持续时间

### 限制值设置
- 从保守值开始，根据监控数据调整
- 考虑服务器的CPU和内存资源
- 为突发流量预留一定缓冲

### 监控指标
- 监控503错误（连接数超限）
- 跟踪实际并发连接数
- 分析连接持续时间模式

## 与 limit_conn 指令集成

limit_conn_zone 定义区域，limit_conn 使用区域：
```nginx
http {
    # 定义区域
    limit_conn_zone $binary_remote_addr zone=addr:10m;
    
    server {
        location / {
            # 使用区域
            limit_conn addr 10;  # 每IP最多10个连接
            proxy_pass http://backend;
        }
    }
}
```

## 错误处理

当连接数超限时，nginx返回503错误：
```nginx
server {
    # 自定义503错误页面
    error_page 503 /503.html;
    
    location = /503.html {
        internal;
        return 503 "Too many connections. Please try again later.";
    }
}
```

## 与其他限制功能对比

| 功能 | limit_conn_zone | limit_req_zone |
|------|----------------|----------------|
| 限制对象 | 并发连接数 | 请求频率 |
| 适用场景 | 长连接、下载、流媒体 | API调用、防暴力破解 |
| 内存使用 | ~64字节/连接 | ~64字节/IP |
| 时间窗口 | 连接期间 | 时间间隔 |

## 运行示例

```bash
cd examples/limit-conn-zone
go run main.go
```

示例将演示：
- 解析包含多个 limit_conn_zone 指令的配置
- 查找和操作 limit_conn_zone 指令
- 修改内存大小和同步设置
- 连接容量估算和推荐
- 兼容性检查和内存分析
- 生成修改后的配置文件

## 最佳实践

1. **内存规划**: 根据预期并发连接数合理分配内存
2. **键选择**: 使用 `$binary_remote_addr` 提高内存效率
3. **限制设置**: 根据服务器性能和业务需求设置合理限制
4. **多层限制**: 组合IP、用户、路径限制提供细粒度控制
5. **监控**: 监控503错误和实际连接使用情况
6. **测试**: 在生产环境部署前充分测试连接限制效果
7. **同步**: 在多服务器环境中启用sync参数
8. **错误处理**: 提供友好的连接数超限错误页面
