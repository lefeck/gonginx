# Stream 块示例

这个示例展示了如何使用 gonginx 库处理 nginx 的 stream 配置块。

## 功能特性

### Stream 模块支持
- **Stream 块解析**: 解析和操作 stream 配置块
- **Stream Upstream**: 支持 TCP/UDP 负载均衡的 upstream 配置
- **Stream Server**: 支持 stream 服务器配置
- **Upstream 服务器**: 支持 stream upstream 中的服务器配置，包括权重、backup、down 等参数

### 主要功能

1. **解析 Stream 配置**
   ```go
   streams := conf.FindStreams()
   ```

2. **查找 Stream Upstreams**
   ```go
   upstreams := conf.FindStreamUpstreams()
   backend := conf.FindStreamUpstreamByName("backend")
   ```

3. **查找 Stream Servers**
   ```go
   servers := conf.FindStreamServers() 
   servers80 := conf.FindStreamServersByListen("80")
   ```

4. **操作 Upstream 服务器**
   ```go
   // 添加服务器
   upstream.AddServer(server)
   
   // 获取服务器属性
   server.GetWeight()
   server.IsDown()
   server.IsBackup()
   ```

## 运行示例

```bash
cd examples/stream-blocks
go run main.go
```

## 示例配置

示例解析以下 stream 配置：

```nginx
stream {
    upstream backend {
        server 192.168.1.1:8080 weight=3;
        server 192.168.1.2:8080 weight=2;
        server 192.168.1.3:8080 backup;
    }
    
    server {
        listen 80;
        proxy_pass backend;
        proxy_timeout 3s;
    }
    
    server {
        listen 443;
        proxy_pass 192.168.1.100:8443;
    }
}
```

## 输出说明

程序将输出：
1. 解析的 stream 配置结构信息
2. 查找到的 upstream 和 server 详细信息
3. 全局查找功能的测试结果
4. 修改后的完整配置
