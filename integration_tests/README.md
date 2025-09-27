# Gonginx 集成测试

这个目录包含了 gonginx 库的集成测试，用于验证各种功能在实际使用场景中的正确性和可靠性。

## 测试覆盖

### 1. 完整工作流测试 (`full_workflow_test.go`)

测试从配置生成到验证、修改、导出的完整工作流：

#### TestCompleteWorkflow
- **配置生成**：使用模板生成基础配置
- **配置验证**：验证生成的配置正确性
- **配置修改**：动态添加新的 location 块
- **重新验证**：确保修改后的配置仍然有效
- **安全检查**：评估配置的安全性
- **配置优化**：获取优化建议
- **配置导出**：将配置导出为字符串
- **往返测试**：重新解析导出的配置并验证一致性
- **格式转换**：测试 JSON 和 YAML 格式转换

#### TestConfigValidationWorkflow
- 测试各种配置错误的检测能力
- 验证上下文错误、依赖关系错误、参数错误和结构错误的发现
- 确保验证器能正确分类和报告不同类型的问题

#### TestSearchAndModifyWorkflow
- 测试高级搜索功能（SSL 证书、upstream 服务器、location 等）
- 验证搜索结果的准确性和完整性
- 测试配置修改功能和修改后的验证

#### TestFileIOWorkflow
- 测试配置文件的读写操作
- 验证文件 I/O 的正确性
- 确保往返文件操作的一致性

#### TestErrorHandlingWorkflow
- 测试各种错误情况的处理
- 验证语法错误、无效指令、嵌套错误等场景
- 确保错误信息的准确性和有用性

### 2. 真实世界配置测试 (`real_world_configs_test.go`)

测试真实生产环境中常见的复杂配置：

#### TestWordPressConfig
测试典型的 WordPress 网站配置：
- **多协议支持**：HTTP 和 HTTPS
- **PHP 处理**：FastCGI 配置
- **静态文件优化**：缓存和压缩
- **安全配置**：安全头和访问控制
- **日志配置**：访问日志和错误日志

#### TestMicroservicesGatewayConfig
测试微服务 API 网关配置：
- **多服务路由**：用户、订单、产品、通知服务
- **负载均衡**：多种负载均衡策略
- **限流控制**：不同端点的速率限制
- **健康检查**：服务健康监控
- **WebSocket 支持**：实时通信
- **CORS 配置**：跨域请求处理

#### TestHighTrafficWebsiteConfig
测试高流量网站的性能优化配置：
- **性能调优**：连接管理、缓冲区优化
- **缓存策略**：多级缓存配置
- **压缩优化**：Gzip 压缩配置
- **A/B 测试**：基于 Cookie 的流量分配
- **地理路由**：基于地理位置的路由
- **静态文件优化**：长期缓存和压缩

## 运行集成测试

### 运行所有集成测试

```bash
cd /path/to/gonginx
go test -v ./integration_tests/
```

### 运行特定测试文件

```bash
# 运行完整工作流测试
go test -v ./integration_tests/ -run TestComplete

# 运行真实世界配置测试
go test -v ./integration_tests/ -run TestWordPress
go test -v ./integration_tests/ -run TestMicroservices
go test -v ./integration_tests/ -run TestHighTraffic
```

### 运行特定测试函数

```bash
# 运行配置验证工作流测试
go test -v ./integration_tests/ -run TestConfigValidationWorkflow

# 运行搜索和修改工作流测试
go test -v ./integration_tests/ -run TestSearchAndModifyWorkflow

# 运行文件 I/O 工作流测试
go test -v ./integration_tests/ -run TestFileIOWorkflow
```

### 详细输出和调试

```bash
# 显示详细日志
go test -v ./integration_tests/ -args -test.v

# 运行时显示测试进度
go test -v ./integration_tests/ -count=1

# 并行运行测试
go test -v ./integration_tests/ -parallel 4
```

## 测试结果示例

```
=== RUN   TestCompleteWorkflow
    full_workflow_test.go:28: 步骤1: 生成基础配置
    full_workflow_test.go:38: 步骤2: 验证生成的配置
    full_workflow_test.go:46: 步骤3: 修改配置
    full_workflow_test.go:79: 步骤4: 验证修改后的配置
    full_workflow_test.go:87: 步骤5: 安全检查
    full_workflow_test.go:95: 步骤6: 配置优化
    full_workflow_test.go:101: 步骤7: 导出配置
    full_workflow_test.go:108: 步骤8: 往返测试
    full_workflow_test.go:125: 步骤9: 格式转换测试
    full_workflow_test.go:140: 完整工作流测试成功
--- PASS: TestCompleteWorkflow (0.05s)

=== RUN   TestWordPressConfig
    real_world_configs_test.go:403: 测试 WordPress 配置
    real_world_configs_test.go:425: WordPress 安全评分: 78/100
    real_world_configs_test.go:436: WordPress 配置有 3 个优化建议
    real_world_configs_test.go:444: WordPress 配置统计: 2 个 server, 0 个 upstream, 8 个 location
    real_world_configs_test.go:490: WordPress 配置测试完成
--- PASS: TestWordPressConfig (0.03s)

=== RUN   TestMicroservicesGatewayConfig
    real_world_configs_test.go:403: 测试 Microservices Gateway 配置
    real_world_configs_test.go:425: Microservices Gateway 安全评分: 85/100
    real_world_configs_test.go:436: Microservices Gateway 配置有 2 个优化建议
    real_world_configs_test.go:444: Microservices Gateway 配置统计: 1 个 server, 4 个 upstream, 12 个 location
    real_world_configs_test.go:490: Microservices Gateway 配置测试完成
--- PASS: TestMicroservicesGatewayConfig (0.04s)
```

## 测试数据和配置

### 配置复杂度统计

| 配置类型 | Server 块 | Upstream 块 | Location 块 | 特殊指令 |
|---------|----------|-------------|-------------|----------|
| WordPress | 2 | 0 | 8 | FastCGI, SSL |
| 微服务网关 | 1 | 4 | 12 | 限流, WebSocket |
| 高流量网站 | 1 | 2 | 6 | 缓存, A/B测试 |

### 测试覆盖的功能

✅ **解析功能**
- 基础指令解析
- 块嵌套结构
- 参数类型检测
- Include 指令处理

✅ **验证功能**
- 上下文验证
- 依赖关系检查
- 参数验证
- 结构完整性

✅ **搜索功能**
- 指令搜索
- 服务器查找
- Upstream 查找
- Location 模式匹配

✅ **修改功能**
- 动态添加指令
- 修改参数值
- 添加新块
- 删除指令

✅ **导出功能**
- 格式化输出
- 往返一致性
- 多格式支持

✅ **安全检查**
- SSL 配置验证
- 安全头检查
- 访问控制验证

✅ **性能优化**
- 配置优化建议
- 性能问题检测
- 最佳实践建议

## 故障排除

### 常见问题

1. **Include 指令错误**
   ```
   Error: include file not found
   ```
   - 原因：测试环境中 include 的文件不存在
   - 解决：使用相对路径或跳过 include 验证

2. **上下文验证错误**
   ```
   Error: directive not allowed in context
   ```
   - 原因：真实配置可能包含扩展模块的指令
   - 解决：更新上下文验证规则或标记为已知问题

3. **往返测试失败**
   ```
   Error: round-trip configuration mismatch
   ```
   - 原因：格式化或解析差异
   - 解决：标准化空白字符和注释处理

### 调试技巧

1. **启用详细日志**
   ```bash
   go test -v ./integration_tests/ -args -test.v
   ```

2. **单独运行失败的测试**
   ```bash
   go test -v ./integration_tests/ -run TestSpecificFunction
   ```

3. **保存中间结果**
   ```go
   // 在测试中添加调试输出
   t.Logf("Configuration: %s", dumper.DumpConfig(conf, dumper.IndentedStyle))
   ```

## 扩展测试

### 添加新的真实世界配置

1. 在 `real_world_configs_test.go` 中添加新的测试函数
2. 提供完整的配置字符串
3. 调用 `testRealWorldConfig` 函数
4. 添加特定于配置的验证逻辑

### 自定义工作流测试

```go
func TestCustomWorkflow(t *testing.T) {
    // 1. 准备测试数据
    config := "your config here"
    
    // 2. 执行操作
    p := parser.NewStringParser(config)
    conf, err := p.Parse()
    if err != nil {
        t.Fatalf("解析失败: %v", err)
    }
    
    // 3. 验证结果
    validator := config.NewConfigValidator()
    report := validator.ValidateConfig(conf)
    
    // 4. 断言
    if report.HasErrors() {
        t.Errorf("验证失败")
    }
}
```

这些集成测试确保 gonginx 库在真实使用场景中的可靠性和正确性，为用户提供稳定的配置处理体验。
