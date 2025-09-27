# Map 块功能支持

本示例展示了 Gonginx 框架新增的 nginx map 块支持功能，这是 nginx 中用于变量映射的重要功能。

## Map 块简介

nginx 的 map 块允许你根据一个变量的值来设置另一个变量的值，支持：
- 精确匹配
- 正则表达式匹配  
- 默认值设置
- 复杂的条件映射

## 新增的 API

### 1. 查找 Map 块
```go
// 查找所有 map 块
maps := config.FindMaps()

// 按变量查找特定的 map 块
mapBlock := config.FindMapByVariables("$http_host", "$backend")
```

### 2. Map 块操作
```go
// 添加新的映射
mapBlock.AddMapping("pattern", "value")

// 获取/设置默认值
defaultValue := mapBlock.GetDefaultValue()
mapBlock.SetDefaultValue("new_default")

// 访问映射条目
for _, mapping := range mapBlock.Mappings {
    fmt.Printf("%s -> %s\n", mapping.Pattern, mapping.Value)
}
```

### 3. Map 块属性
```go
type Map struct {
    Variable       string      // 源变量 (如 $http_host)
    MappedVariable string      // 目标变量 (如 $backend)
    Mappings       []*MapEntry // 映射条目列表
    Comment        []string    // 注释
}

type MapEntry struct {
    Pattern string // 匹配模式
    Value   string // 映射值
    Comment []string // 注释
}
```

## 支持的 Map 语法

### 基本映射
```nginx
map $http_host $backend {
    default backend_default;
    example.com backend_main;
    api.example.com backend_api;
}
```

### 正则表达式映射
```nginx
map $uri $content_type {
    default "text/html";
    ~\.js$ "application/javascript";
    ~\.css$ "text/css";
    ~\.(jpg|jpeg|png|gif)$ "image/*";
}
```

### 条件映射
```nginx
map $request_method $limit_key {
    default "";
    POST $binary_remote_addr;
    PUT $binary_remote_addr;
    DELETE $binary_remote_addr;
}
```

## 常见使用场景

### 1. 后端选择
根据主机名选择不同的上游服务器组：
```nginx
map $http_host $backend {
    default backend_default;
    api.example.com backend_api;
    admin.example.com backend_admin;
}
```

### 2. 速率限制
根据请求方法应用不同的限制策略：
```nginx
map $request_method $limit_key {
    default "";
    POST $binary_remote_addr;
    PUT $binary_remote_addr;
}
```

### 3. SSL 重定向
根据协议决定是否重定向：
```nginx
map $scheme $redirect_https {
    default 0;
    http 1;
}
```

### 4. 内容类型设置
根据文件扩展名设置合适的内容类型：
```nginx
map $uri $content_type {
    default "text/html";
    ~\.json$ "application/json";
    ~\.(jpg|png|gif)$ "image/*";
}
```

### 5. A/B 测试
根据 cookie 或其他标识符进行用户分流：
```nginx
map $cookie_ab_test $backend {
    default backend_a;
    "version_b" backend_b;
}
```

## 运行示例

```bash
cd examples/map-blocks
go run main.go
```

## 测试

Map 块功能包含完整的测试用例：

```bash
# 运行 config 包的 map 测试
go test ./config -v -run TestMap

# 运行 dumper 包的 map 测试  
go test ./dumper -v -run TestMap
```

## 高级特性

- **注释支持**: 完整保留 map 块和映射条目的注释
- **格式化输出**: 支持多种缩进和格式化风格
- **动态操作**: 支持运行时添加、修改映射条目
- **类型安全**: 强类型的 Go API，避免配置错误
- **递归搜索**: 支持在复杂嵌套配置中查找 map 块

## 与其他功能集成

Map 块与其他 Gonginx 功能无缝集成：
- 可以与 server、upstream 块配合使用
- 支持 include 文件中的 map 块
- 与高级搜索功能兼容
- 完全支持配置解析和生成

这个功能完成了 doc.md 中优先级第二的需求：**缺失的 nginx 核心块支持 - map 块 (用于变量映射)**。
