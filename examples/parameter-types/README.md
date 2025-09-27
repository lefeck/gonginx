# 参数类型系统示例

这个示例展示了 gonginx 库的参数类型系统功能，包括自动类型检测、类型验证和类型转换。

## 功能特性

### 支持的参数类型

1. **String**: 普通字符串
2. **Variable**: nginx 变量 (以 $ 开头)
3. **Number**: 数值类型
4. **Size**: 大小值 (如 1M, 512k, 1G)
5. **Time**: 时间值 (如 30s, 1h, 7d)
6. **Path**: 文件/目录路径
7. **URL**: URL 地址
8. **Regex**: 正则表达式
9. **Boolean**: 布尔值 (on/off, yes/no, true/false)
10. **Quoted**: 引用字符串

### 主要功能

#### 1. 自动类型检测
```go
param := config.NewParameter("1M")  // 自动检测为 ParameterTypeSize
```

#### 2. 显式类型指定
```go
param := config.NewParameterWithType("custom", config.ParameterTypeString)
```

#### 3. 类型检查方法
```go
if param.IsSize() {
    // 处理大小类型参数
}
if param.IsTime() {
    // 处理时间类型参数
}
```

#### 4. 类型验证和转换
```go
if size, valid := config.ValidateSize("1M"); valid {
    // 使用验证过的大小值
}

if val, valid := config.ValidateBoolean("on"); valid {
    // val == true
}
```

## 运行示例

```bash
cd examples/parameter-types
go run main.go
```

## 示例配置

示例解析以下配置并分析每个参数的类型：

```nginx
server {
    listen 80;                          # number
    server_name example.com;           # string
    root /var/www/html;               # path
    client_max_body_size 1M;          # size
    proxy_read_timeout 30s;           # time
    error_log /var/log/nginx/error.log; # path
    
    # Variables
    set $backend_pool $request_uri;   # variable
    
    # Boolean values
    gzip on;                          # boolean
    autoindex off;                    # boolean
    
    # URLs
    proxy_pass http://backend.example.com; # url
    
    location ~ \.php$ {               # regex
        fastcgi_pass 127.0.0.1:9000; # string (host:port)
        proxy_connect_timeout 5s;    # time
        proxy_cache_valid 200 10m;   # mixed: number, time
    }
    
    location /files {
        alias "/var/files";           # quoted string (path)
        client_body_buffer_size 8k;  # size
    }
}
```

## 输出说明

程序将输出：
1. 每个指令的参数类型分析
2. 类型验证结果
3. 参数类型统计
4. 手动创建参数的示例

## 实际应用

参数类型系统在以下场景中特别有用：

1. **配置验证**: 验证参数是否符合预期类型
2. **配置生成**: 根据类型生成正确格式的配置
3. **IDE支持**: 为开发工具提供类型提示
4. **自动补全**: 根据参数类型提供智能补全
5. **错误检查**: 在解析时发现类型不匹配的错误
