# Geo 块功能支持

本示例展示了 Gonginx 框架新增的 nginx geo 块支持功能，这是 nginx 中用于基于客户端 IP 地址进行地理位置检测和变量设置的重要功能。

## Geo 块简介

nginx 的 geo 块允许你根据客户端 IP 地址设置变量值，支持：
- CIDR 网络格式 (192.168.1.0/24)
- IP 地址范围 (127.0.0.1-127.0.0.255)
- 默认值设置
- 可信代理配置
- 递归代理查找
- 继承的 geo 块删除特定 IP

## 新增的 API

### 1. 查找 Geo 块
```go
// 查找所有 geo 块
geos := config.FindGeos()

// 按目标变量查找特定的 geo 块
geoBlock := config.FindGeoByVariable("$country")

// 按源地址和目标变量查找 geo 块
geoBlock := config.FindGeoByVariables("$remote_addr", "$region")
```

### 2. Geo 块操作
```go
// 添加新的 IP 网络映射
geoBlock.AddEntry("192.168.1.0/24", "CN")

// 获取/设置默认值
defaultValue := geoBlock.GetDefaultValue()
geoBlock.SetDefaultValue("UNKNOWN")

// 添加可信代理
geoBlock.AddProxy("10.0.0.0/8")

// 添加删除条目
geoBlock.AddDelete("127.0.0.1")

// 设置范围模式
geoBlock.SetRanges(true)

// 设置递归代理查找
geoBlock.SetProxyRecursive(true)

// 访问网络条目
for _, entry := range geoBlock.Entries {
    fmt.Printf("%s -> %s\n", entry.Network, entry.Value)
}
```

### 3. Geo 块属性
```go
type Geo struct {
    SourceAddress  string      // 源地址变量 (默认 $remote_addr)
    Variable       string      // 目标变量 (如 $country)
    Entries        []*GeoEntry // 网络条目列表
    DefaultValue   string      // 默认值
    Ranges         bool        // 是否使用范围格式
    Delete         []string    // 要删除的 IP 地址列表
    Proxy          []string    // 可信代理地址列表
    ProxyRecursive bool        // 是否使用递归代理查找
    Comment        []string    // 注释
}

type GeoEntry struct {
    Network string // 网络地址 (CIDR、IP或范围)
    Value   string // 映射值
    Comment []string // 注释
}
```

## 支持的 Geo 语法

### 基本 CIDR 映射
```nginx
geo $country {
    default ZZ;
    127.0.0.0/8 US;
    192.168.1.0/24 CN;
    203.208.60.0/24 CN;
}
```

### IP 范围映射
```nginx
geo $remote_addr $region {
    ranges;
    default unknown;
    127.0.0.1-127.0.0.255 local;
    10.0.0.1-10.0.0.100 internal;
}
```

### 代理配置
```nginx
geo $city {
    proxy 192.168.1.0/24;
    proxy 10.0.0.0/8;
    proxy_recursive;
    delete 127.0.0.1;
    default unknown;
    192.168.0.0/16 Beijing;
    10.0.0.0/8 Shanghai;
}
```

## 常见使用场景

### 1. 基于地理位置的后端选择
根据客户端国家/地区选择不同的服务器：
```nginx
geo $country {
    default ZZ;
    127.0.0.0/8 US;
    192.168.1.0/24 CN;
}

map $country $backend {
    default backend_global;
    US backend_us;
    CN backend_cn;
}

server {
    proxy_pass http://$backend;
}
```

### 2. ISP 优化
根据运营商选择最优 CDN：
```nginx
geo $isp {
    default other;
    1.2.4.0/22 chinanet;
    58.14.0.0/15 unicom;
    219.128.0.0/11 cmcc;
}

location ~* \.(css|js|png|jpg)$ {
    if ($isp = "chinanet") {
        proxy_pass http://cdn-chinanet.example.com;
    }
}
```

### 3. 访问控制
基于地理位置的访问限制：
```nginx
geo $region {
    default external;
    10.0.0.0/8 internal;
    192.168.0.0/16 office;
}

location /admin {
    if ($region != "internal") {
        return 403;
    }
    proxy_pass http://admin-backend;
}
```

### 4. 负载均衡优化
根据地理位置分配流量：
```nginx
geo $datacenter {
    default dc3;
    1.0.0.0/8 dc1;
    8.0.0.0/8 dc1;
    128.0.0.0/8 dc2;
}

upstream backend_dc1 {
    server dc1-server1.example.com;
    server dc1-server2.example.com;
}

upstream backend_dc2 {
    server dc2-server1.example.com;
    server dc2-server2.example.com;
}
```

## 高级特性

### 代理透传
当 nginx 位于代理后面时，使用真实客户端 IP：
```nginx
geo $country {
    proxy 192.168.1.0/24;  # 信任的代理网络
    proxy_recursive;        # 递归查找真实 IP
    
    default ZZ;
    127.0.0.0/8 US;
}
```

### 继承和删除
在继承的 geo 块中删除特定 IP：
```nginx
geo $inherited_country {
    include /etc/nginx/geo/countries.conf;
    delete 127.0.0.1;  # 删除继承配置中的这个 IP
    192.168.1.100 CN;  # 添加新的映射
}
```

### 网络范围
使用 IP 范围而不是 CIDR：
```nginx
geo $region {
    ranges;  # 启用范围模式
    default unknown;
    127.0.0.1-127.0.0.255 local;
    10.0.0.1-10.255.255.254 internal;
}
```

## 运行示例

```bash
cd examples/geo-blocks
go run main.go
```

示例将演示：
- 解析包含多个 geo 块的配置
- 查找和操作 geo 块
- 添加新的网络映射
- 配置代理设置
- 生成修改后的配置文件

## 与其他功能集成

Geo 块通常与以下功能结合使用：
- **Map 块**: 将地理位置映射到后端服务器
- **Upstream 块**: 基于地理位置的负载均衡
- **Location 块**: 地理位置相关的路由规则
- **限流模块**: 基于地理位置的不同限流策略
