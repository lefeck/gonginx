# Proxy Cache Path 功能支持

本示例展示了 Gonginx 框架新增的 nginx proxy_cache_path 指令支持功能，这是 nginx 代理缓存功能的核心指令。

## Proxy Cache Path 简介

nginx 的 proxy_cache_path 指令用于定义代理缓存的存储路径和相关参数，支持：
- 缓存目录路径配置
- 目录层次结构优化 (levels)
- 共享内存区域配置 (keys_zone)
- 缓存大小限制 (max_size, min_free)
- 缓存过期时间 (inactive)
- 缓存管理进程调优 (manager_*, loader_*)
- 缓存清理功能 (purger)
- 临时路径优化 (use_temp_path)

## 新增的 API

### 1. 查找 Proxy Cache Path 指令
```go
// 查找所有 proxy_cache_path 指令
cachePaths := config.FindProxyCachePaths()

// 按区域名查找特定的 proxy_cache_path
cachePath := config.FindProxyCachePathByZone("my_cache")

// 按缓存路径查找 proxy_cache_path
cachePaths := config.FindProxyCachePathsByPath("/var/cache/nginx")
```

### 2. Proxy Cache Path 操作
```go
// 获取/设置缓存大小
maxSizeBytes, err := cachePath.GetMaxSizeBytes()    // 获取最大缓存大小字节数
err = cachePath.SetMaxSize("10g")                   // 设置最大缓存大小

minFreeBytes, err := cachePath.GetMinFreeBytes()    // 获取最小空闲空间字节数

// 获取/设置keys_zone大小
keysZoneBytes, err := cachePath.GetKeysZoneSizeBytes() // 获取共享内存大小
err = cachePath.SetKeysZoneSize("50m")                // 设置共享内存大小

// 获取/设置过期时间
inactiveDuration, err := cachePath.GetInactiveDuration() // 获取过期时间
err = cachePath.SetInactive("2h")                        // 设置过期时间

// 键容量估算
keyCapacity, err := cachePath.EstimateKeyCapacity() // 估算可存储的缓存键数量

// 目录层次解析
levels, err := cachePath.GetLevelsDepth() // 获取目录层次深度

// 布尔参数设置
cachePath.SetUseTemPath(false) // 设置是否使用临时路径
cachePath.SetPurger(true)      // 启用缓存清理

// 访问缓存路径属性
fmt.Printf("Path: %s\n", cachePath.Path)
fmt.Printf("Zone: %s:%s\n", cachePath.KeysZoneName, cachePath.KeysZoneSize)
fmt.Printf("Max Size: %s\n", cachePath.MaxSize)
fmt.Printf("Inactive: %s\n", cachePath.Inactive)
fmt.Printf("Levels: %s\n", cachePath.Levels)
```

### 3. Proxy Cache Path 属性
```go
type ProxyCachePath struct {
    Path              string // 缓存目录路径
    Levels            string // 目录层次结构 (如 "1:2")
    KeysZoneName      string // 共享内存区域名称
    KeysZoneSize      string // 共享内存区域大小
    UseTemPath        *bool  // 是否使用临时路径
    Inactive          string // 过期时间 (如 "60m")
    MaxSize           string // 最大缓存大小 (如 "10g")
    MinFree           string // 最小空闲空间
    ManagerFiles      *int   // 管理进程文件数
    ManagerSleep      string // 管理进程睡眠时间
    ManagerThreshold  string // 管理进程阈值时间
    LoaderFiles       *int   // 加载进程文件数
    LoaderSleep       string // 加载进程睡眠时间
    LoaderThreshold   string // 加载进程阈值时间
    Purger            *bool  // 是否启用清理功能
    PurgerFiles       *int   // 清理进程文件数
    PurgerSleep       string // 清理进程睡眠时间
    PurgerThreshold   string // 清理进程阈值时间
}
```

## 支持的 Proxy Cache Path 语法

### 基本缓存配置
```nginx
proxy_cache_path /var/cache/nginx keys_zone=my_cache:10m;
```

### 带目录层次的缓存
```nginx
proxy_cache_path /var/cache/nginx levels=1:2 keys_zone=my_cache:10m max_size=1g;
```

### 完整配置示例
```nginx
proxy_cache_path /var/cache/nginx/complex 
    levels=1:2:2 
    keys_zone=complex_cache:100m 
    max_size=50g 
    inactive=7d 
    use_temp_path=off 
    manager_files=10000 
    purger=on;
```

### 高性能缓存配置
```nginx
proxy_cache_path /var/cache/nginx/performance 
    levels=2:2 
    keys_zone=perf_cache:128m 
    max_size=100g 
    inactive=30d 
    min_free=10g 
    use_temp_path=off 
    manager_files=8000 
    manager_sleep=50ms 
    manager_threshold=200ms 
    loader_files=4000 
    loader_sleep=25ms 
    loader_threshold=100ms 
    purger=on 
    purger_files=500 
    purger_sleep=10ms 
    purger_threshold=50ms;
```

## 常见使用场景

### 1. 静态内容缓存
高效缓存静态资源：
```nginx
http {
    proxy_cache_path /var/cache/nginx/static 
        levels=1:2 
        keys_zone=static_cache:100m 
        max_size=50g 
        inactive=7d 
        use_temp_path=off;
    
    server {
        location /static/ {
            proxy_cache static_cache;
            proxy_cache_valid 200 7d;
            proxy_cache_valid 404 1h;
            
            proxy_pass http://backend;
        }
    }
}
```

### 2. API 缓存
缓存 API 响应以减少后端负载：
```nginx
http {
    proxy_cache_path /var/cache/nginx/api 
        levels=2:2 
        keys_zone=api_cache:50m 
        max_size=5g 
        inactive=1h 
        use_temp_path=off;
    
    server {
        location /api/ {
            proxy_cache api_cache;
            proxy_cache_valid 200 10m;
            proxy_cache_valid 404 1m;
            proxy_cache_key "$scheme$request_method$host$request_uri$is_args$args";
            
            proxy_pass http://api_backend;
        }
    }
}
```

### 3. 大文件缓存
优化大文件下载缓存：
```nginx
http {
    proxy_cache_path /var/cache/nginx/files 
        levels=1:2:2 
        keys_zone=files_cache:200m 
        max_size=100g 
        inactive=30d 
        min_free=10g 
        purger=on 
        purger_files=100;
    
    server {
        location /download/ {
            proxy_cache files_cache;
            proxy_cache_valid 200 30d;
            proxy_cache_lock on;
            proxy_cache_lock_timeout 30s;
            
            proxy_pass http://file_backend;
        }
    }
}
```

### 4. 微服务缓存
多服务环境的缓存策略：
```nginx
http {
    proxy_cache_path /var/cache/nginx/micro 
        levels=2:1:2 
        keys_zone=micro_cache:128m 
        max_size=20g 
        inactive=12h 
        use_temp_path=off 
        manager_files=8000 
        purger=on;
    
    server {
        location /service-a/ {
            proxy_cache micro_cache;
            proxy_cache_valid 200 30m;
            proxy_cache_key "$host$request_uri$http_x_user_id";
            
            proxy_pass http://service_a;
        }
        
        location /service-b/ {
            proxy_cache micro_cache;
            proxy_cache_valid 200 15m;
            proxy_cache_bypass $arg_nocache;
            
            proxy_pass http://service_b;
        }
    }
}
```

### 5. 开发环境缓存
开发测试环境的轻量级缓存：
```nginx
http {
    proxy_cache_path /tmp/nginx/dev 
        keys_zone=dev_cache:5m 
        max_size=100m 
        inactive=10m 
        use_temp_path=on;
    
    server {
        location / {
            proxy_cache dev_cache;
            proxy_cache_valid 200 1m;
            proxy_cache_bypass $arg_nocache;
            
            proxy_pass http://dev_backend;
        }
    }
}
```

## 高级特性

### 目录层次优化 (Levels)
合理的目录结构提高文件系统性能：
```
levels=1:2 -> /c/29/b7f54b2df7773722d382f4809d65029c
levels=2:2 -> /29/c7/b7f54b2df7773722d382f4809d6529c7
levels=1:1:2 -> /c/7/29/b7f54b2df7773722d382f4809d65c729
```

### 内存容量规划
每个缓存键大约需要256字节内存：
```
10m 内存 ≈ 40,960 个缓存键
50m 内存 ≈ 204,800 个缓存键
100m 内存 ≈ 409,600 个缓存键
1g 内存 ≈ 4,194,304 个缓存键
```

### 管理进程调优
优化缓存管理进程性能：
```nginx
proxy_cache_path /var/cache/nginx 
    keys_zone=tuned_cache:100m 
    manager_files=10000    # 每次处理的文件数
    manager_sleep=50ms     # 处理间隔
    manager_threshold=200ms # 最大处理时间
    loader_files=5000      # 启动时加载的文件数
    loader_sleep=25ms      # 加载间隔
    loader_threshold=100ms; # 最大加载时间
```

### 缓存清理 (Purger)
自动清理过期缓存：
```nginx
proxy_cache_path /var/cache/nginx 
    keys_zone=purged_cache:100m 
    purger=on              # 启用清理
    purger_files=1000      # 每次清理的文件数
    purger_sleep=10ms      # 清理间隔
    purger_threshold=50ms; # 最大清理时间
```

### 性能优化建议
- **use_temp_path=off**: 避免临时文件复制，提高性能
- **levels**: 使用合适的目录层次，避免单目录文件过多
- **min_free**: 预留足够的磁盘空间
- **manager/loader调优**: 根据文件数量和磁盘性能调整参数

## 缓存策略最佳实践

### 1. 内容类型分类缓存
```nginx
# 静态资源 - 长期缓存
proxy_cache_path /var/cache/static levels=1:2 keys_zone=static:100m max_size=50g inactive=30d;

# API 响应 - 短期缓存
proxy_cache_path /var/cache/api levels=2:2 keys_zone=api:50m max_size=5g inactive=1h;

# 用户内容 - 中期缓存
proxy_cache_path /var/cache/user levels=1:1:2 keys_zone=user:200m max_size=20g inactive=7d;
```

### 2. 多层缓存架构
```nginx
# L1 缓存 - 热点内容
proxy_cache_path /var/cache/l1 keys_zone=l1:50m max_size=1g inactive=30m;

# L2 缓存 - 常用内容
proxy_cache_path /var/cache/l2 keys_zone=l2:200m max_size=20g inactive=24h;

# L3 缓存 - 归档内容
proxy_cache_path /var/cache/l3 keys_zone=l3:500m max_size=100g inactive=30d;
```

### 3. 环境特定配置
```nginx
# 生产环境 - 高性能配置
proxy_cache_path /var/cache/prod 
    levels=1:2 
    keys_zone=prod:256m 
    max_size=100g 
    inactive=7d 
    use_temp_path=off 
    manager_files=10000 
    purger=on;

# 开发环境 - 快速刷新
proxy_cache_path /tmp/cache/dev 
    keys_zone=dev:10m 
    max_size=100m 
    inactive=5m 
    use_temp_path=on;
```

## 监控和维护

### 缓存状态监控
```nginx
location /cache/status {
    return 200 "Cache Statistics:
Hit Rate: Monitor via $upstream_cache_status
Memory Usage: Check /proc/meminfo
Disk Usage: Monitor cache directories
Cache Size: Track with du commands
";
}
```

### 缓存清理端点
```nginx
location ~ /purge(/.*) {
    allow 127.0.0.1;
    deny all;
    proxy_cache_purge my_cache "$scheme$request_method$host$1";
}
```

### 性能指标
- **缓存命中率**: 通过 `$upstream_cache_status` 监控
- **内存使用率**: 监控 keys_zone 内存使用
- **磁盘使用率**: 监控缓存目录大小
- **清理效率**: 跟踪 purger 性能

## 故障排除

### 常见问题
1. **缓存不工作**: 检查 keys_zone 配置和权限
2. **磁盘空间满**: 调整 max_size 和 min_free
3. **性能问题**: 优化 levels 和 manager 参数
4. **内存不足**: 增加 keys_zone 大小

### 调试技巧
```nginx
# 添加缓存状态头
add_header X-Cache-Status $upstream_cache_status;
add_header X-Cache-Key "$scheme$request_method$host$request_uri";

# 缓存绕过调试
proxy_cache_bypass $arg_nocache $http_pragma;
proxy_no_cache $arg_nocache $http_pragma;
```

## 运行示例

```bash
cd examples/proxy-cache-path
go run main.go
```

示例将演示：
- 解析包含多个 proxy_cache_path 指令的配置
- 查找和操作 proxy_cache_path 指令
- 修改缓存大小、过期时间等参数
- 缓存容量估算和性能分析
- 目录层次和内存优化
- 生成修改后的配置文件

## 与其他缓存指令集成

proxy_cache_path 与其他缓存指令配合使用：

| 指令 | 作用 |
|------|------|
| proxy_cache | 指定使用的缓存区域 |
| proxy_cache_valid | 设置缓存有效期 |
| proxy_cache_key | 定义缓存键 |
| proxy_cache_bypass | 缓存绕过条件 |
| proxy_cache_use_stale | 使用过期缓存的条件 |
| proxy_cache_lock | 缓存锁定机制 |
| proxy_cache_purge | 缓存清理 |

## 最佳实践总结

1. **容量规划**: 根据业务需求合理分配内存和磁盘空间
2. **目录优化**: 使用适当的 levels 避免文件系统瓶颈
3. **性能调优**: 启用 use_temp_path=off 和 purger
4. **监控告警**: 监控缓存命中率和资源使用情况
5. **分层策略**: 为不同类型内容配置不同的缓存策略
6. **定期维护**: 监控磁盘使用和清理无效缓存
7. **测试验证**: 在生产环境部署前充分测试缓存效果
