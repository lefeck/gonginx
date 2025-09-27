# Limit Req Zone 功能支持

本示例展示了 Gonginx 框架新增的 nginx limit_req_zone 指令支持功能，这是 nginx 中用于请求频率限制的重要功能。

## Limit Req Zone 简介

nginx 的 limit_req_zone 指令用于定义请求频率限制的共享内存区域，支持：
- 基于各种变量的请求限制 (IP地址、用户ID、请求URI等)
- 内存区域大小配置 (k, m, g)
- 请求频率限制 (每秒/每分钟)
- 分布式环境同步 (sync参数)
- 多层限制策略

## 新增的 API

### 1. 查找 Limit Req Zone 指令
```go
// 查找所有 limit_req_zone 指令
zones := config.FindLimitReqZones()

// 按区域名查找特定的 limit_req_zone
zone := config.FindLimitReqZoneByName("global")

// 按键变量查找 limit_req_zone
zones := config.FindLimitReqZonesByKey("$binary_remote_addr")
```

### 2. Limit Req Zone 操作
```go
// 获取/设置频率限制
rateNum, err := zone.GetRateNumber()    // 获取数字部分
rateUnit, err := zone.GetRateUnit()     // 获取单位 (s/m)
err = zone.SetRate("10r/s")             // 设置新频率

// 获取/设置内存区域大小
sizeBytes, err := zone.GetZoneSizeBytes() // 获取字节数
err = zone.SetZoneSize("20m")             // 设置新大小

// 访问区域属性
fmt.Printf("Key: %s\n", zone.Key)
fmt.Printf("Zone: %s\n", zone.ZoneName)
fmt.Printf("Size: %s\n", zone.ZoneSize)
fmt.Printf("Rate: %s\n", zone.Rate)
fmt.Printf("Sync: %t\n", zone.Sync)
```

### 3. Limit Req Zone 属性
```go
type LimitReqZone struct {
    Key      string   // 键变量 (如 $binary_remote_addr)
    ZoneName string   // 区域名称 (如 "global")
    ZoneSize string   // 区域大小 (如 "10m")
    Rate     string   // 频率限制 (如 "1r/s")
    Sync     bool     // 是否启用同步
    Comment  []string // 注释
}
```

## 支持的 Limit Req Zone 语法

### 基本频率限制
```nginx
limit_req_zone $binary_remote_addr zone=global:10m rate=10r/s;
```

### 基于服务器的限制
```nginx
limit_req_zone $server_name zone=perserver:5m rate=50r/s;
```

### 基于路径的限制
```nginx
limit_req_zone $request_uri zone=perpath:20m rate=5r/s;
```

### 基于用户的限制
```nginx
limit_req_zone $cookie_user_id zone=peruser:100m rate=20r/s;
```

### 每分钟限制
```nginx
limit_req_zone $http_user_agent zone=peragent:1g rate=2r/m;
```

### 分布式同步
```nginx
limit_req_zone $binary_remote_addr zone=cluster:50m rate=100r/s sync;
```

### 复合键限制
```nginx
limit_req_zone $binary_remote_addr$request_method zone=per_ip_method:50m rate=30r/s;
```

## 常见使用场景

### 1. API 保护
保护API端点免受滥用：
```nginx
http {
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
    
    server {
        location /api/ {
            limit_req zone=api burst=20 nodelay;
            proxy_pass http://backend;
        }
    }
}
```

### 2. 登录保护
防止暴力破解攻击：
```nginx
http {
    limit_req_zone $binary_remote_addr zone=login:10m rate=1r/s;
    
    server {
        location /login {
            limit_req zone=login burst=3 delay=2;
            proxy_pass http://backend;
        }
    }
}
```

### 3. 用户级限制
基于用户身份的限制：
```nginx
http {
    limit_req_zone $cookie_user_id zone=peruser:50m rate=50r/s;
    
    server {
        location /dashboard {
            limit_req zone=peruser burst=10 nodelay;
            proxy_pass http://backend;
        }
    }
}
```

### 4. 路径级限制
不同路径不同限制：
```nginx
http {
    limit_req_zone $request_uri zone=perpath:20m rate=5r/s;
    
    server {
        location /upload {
            limit_req zone=perpath burst=2 delay=5;
            proxy_pass http://backend;
        }
        
        location /download {
            limit_req zone=perpath burst=10 nodelay;
            proxy_pass http://backend;
        }
    }
}
```

### 5. 多层限制
组合多种限制策略：
```nginx
http {
    limit_req_zone $binary_remote_addr zone=global:10m rate=100r/s;
    limit_req_zone $request_uri zone=perpath:20m rate=10r/s;
    limit_req_zone $cookie_user_id zone=peruser:50m rate=50r/s;
    
    server {
        location /api/ {
            # 全局限制
            limit_req zone=global burst=50 nodelay;
            # 路径限制
            limit_req zone=perpath burst=5 delay=2;
            # 用户限制
            limit_req zone=peruser burst=20 nodelay;
            
            proxy_pass http://backend;
        }
    }
}
```

### 6. 地理位置限制
基于国家的不同限制：
```nginx
http {
    limit_req_zone $geoip_country_code zone=percountry:10m rate=1000r/s;
    
    server {
        location / {
            limit_req zone=percountry burst=100 nodelay;
            proxy_pass http://backend;
        }
    }
}
```

## 高级特性

### 内存大小配置
支持多种内存单位：
```nginx
limit_req_zone $binary_remote_addr zone=small:1k rate=1r/s;    # 1KB
limit_req_zone $binary_remote_addr zone=medium:10m rate=10r/s; # 10MB
limit_req_zone $binary_remote_addr zone=large:1g rate=100r/s;  # 1GB
```

### 频率单位
支持秒和分钟级别的限制：
```nginx
limit_req_zone $binary_remote_addr zone=persec:10m rate=10r/s;  # 每秒10次
limit_req_zone $binary_remote_addr zone=permin:10m rate=600r/m; # 每分钟600次
```

### 分布式同步
在多服务器环境中同步限制状态：
```nginx
limit_req_zone $binary_remote_addr zone=cluster:50m rate=100r/s sync;
```

### 参数验证
框架自动验证参数格式：
```go
// 自动验证
err := zone.SetRate("invalid")     // 错误：格式无效
err := zone.SetRate("10r/h")       // 错误：不支持小时单位
err := zone.SetSize("10x")         // 错误：无效的大小单位
err := zone.SetRate("10.5r/s")     // 成功：支持小数
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

### 内存估算
每个唯一键大约需要64字节内存：
```
10m 内存 ≈ 160,000 个唯一IP地址
50m 内存 ≈ 800,000 个唯一IP地址
1g 内存 ≈ 16,000,000 个唯一IP地址
```

### 键选择优化
- `$binary_remote_addr` 比 `$remote_addr` 节省内存
- 复合键会增加内存使用量
- 选择合适的粒度避免过度细分

## 与 limit_req 指令集成

limit_req_zone 定义区域，limit_req 使用区域：
```nginx
http {
    # 定义区域
    limit_req_zone $binary_remote_addr zone=global:10m rate=10r/s;
    
    server {
        location / {
            # 使用区域
            limit_req zone=global burst=20 nodelay;
            proxy_pass http://backend;
        }
    }
}
```

## 运行示例

```bash
cd examples/limit-req-zone
go run main.go
```

示例将演示：
- 解析包含多个 limit_req_zone 指令的配置
- 查找和操作 limit_req_zone 指令
- 修改频率和内存大小
- 参数验证和错误处理
- 内存使用分析
- 生成修改后的配置文件

## 最佳实践

1. **内存规划**: 根据预期用户数量合理分配内存
2. **键选择**: 使用 `$binary_remote_addr` 而不是 `$remote_addr`
3. **频率设置**: 从宽松开始，根据监控数据调整
4. **多层限制**: 组合全局、路径、用户限制提供细粒度控制
5. **监控**: 监控429错误和限制效果
6. **测试**: 在生产环境部署前充分测试限制效果
