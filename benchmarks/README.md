# Gonginx 性能基准测试

这个目录包含了 gonginx 库的性能基准测试，用于评估各种操作的性能表现。

## 测试覆盖

### 1. 解析性能测试 (`parse_benchmark_test.go`)

测试不同规模配置文件的解析性能：

- **小型配置**：简单的 events + http + server 结构
- **中型配置**：包含 upstream、SSL、多个 location 的真实配置
- **大型配置**：包含多个 upstream 和大量 server 块的配置
- **复杂嵌套**：包含 map、geo、split_clients、if 等复杂结构
- **Include 文件**：测试包含 include 指令的配置解析
- **内存分配**：测试解析过程中的内存分配情况

### 2. 验证性能测试 (`validation_benchmark_test.go`)

测试配置验证功能的性能：

- **上下文验证**：指令上下文正确性检查
- **依赖关系验证**：指令间依赖关系检查
- **综合验证**：完整的配置验证流程
- **大型配置验证**：大规模配置的验证性能
- **参数类型检测**：参数类型自动检测性能
- **内存分配验证**：验证过程中的内存使用

### 3. 搜索性能测试 (`search_benchmark_test.go`)

测试配置搜索和查询功能的性能：

- **指令搜索**：FindDirectives 方法性能
- **服务器搜索**：FindServersByName 方法性能
- **Upstream 搜索**：FindUpstreams 方法性能
- **Location 搜索**：FindLocationsByPattern 方法性能
- **SSL 证书搜索**：GetAllSSLCertificates 方法性能
- **Upstream 服务器搜索**：GetAllUpstreamServers 方法性能
- **深度嵌套搜索**：复杂嵌套结构中的搜索性能
- **内存分配搜索**：搜索操作的内存使用

## 运行基准测试

### 运行所有基准测试

```bash
cd /path/to/gonginx
go test -bench=. ./benchmarks/
```

### 运行特定类型的基准测试

```bash
# 只运行解析基准测试
go test -bench=BenchmarkParse ./benchmarks/

# 只运行验证基准测试
go test -bench=BenchmarkValidation ./benchmarks/

# 只运行搜索基准测试
go test -bench=BenchmarkFind ./benchmarks/
```

### 运行特定的基准测试

```bash
# 运行小型配置解析测试
go test -bench=BenchmarkParseSmallConfig ./benchmarks/

# 运行上下文验证测试
go test -bench=BenchmarkContextValidation ./benchmarks/

# 运行服务器搜索测试
go test -bench=BenchmarkFindServersByName ./benchmarks/
```

### 详细输出和内存分析

```bash
# 显示内存分配统计
go test -bench=. -benchmem ./benchmarks/

# 输出详细结果
go test -bench=. -v ./benchmarks/

# 运行多次获得更准确的结果
go test -bench=. -count=5 ./benchmarks/

# 设置运行时间
go test -bench=. -benchtime=10s ./benchmarks/
```

### 性能剖析

```bash
# CPU 性能剖析
go test -bench=BenchmarkParseLargeConfig -cpuprofile=cpu.prof ./benchmarks/

# 内存性能剖析
go test -bench=BenchmarkParseLargeConfig -memprofile=mem.prof ./benchmarks/

# 查看性能剖析结果
go tool pprof cpu.prof
go tool pprof mem.prof
```

## 基准测试结果示例

```
BenchmarkParseSmallConfig-8               10000    156789 ns/op    45231 B/op     892 allocs/op
BenchmarkParseMediumConfig-8                2000    456123 ns/op   123456 B/op    2341 allocs/op
BenchmarkParseLargeConfig-8                  500   1234567 ns/op   567890 B/op    8765 allocs/op

BenchmarkContextValidation-8              50000     23456 ns/op     8912 B/op     234 allocs/op
BenchmarkDependencyValidation-8           30000     34567 ns/op    12345 B/op     345 allocs/op
BenchmarkComprehensiveValidation-8        20000     67890 ns/op    23456 B/op     567 allocs/op

BenchmarkFindDirectives-8                100000     12345 ns/op     3456 B/op      89 allocs/op
BenchmarkFindServersByName-8              80000     15678 ns/op     4567 B/op     123 allocs/op
BenchmarkGetAllSSLCertificates-8          60000     23456 ns/op     6789 B/op     234 allocs/op
```

## 结果解读

- **ns/op**: 每次操作的纳秒数（越小越好）
- **B/op**: 每次操作分配的字节数（越小越好）
- **allocs/op**: 每次操作的内存分配次数（越小越好）

## 性能优化建议

### 1. 解析优化
- 对于大型配置文件，考虑分段解析
- 缓存经常使用的配置解析结果
- 使用流式解析减少内存占用

### 2. 验证优化
- 只在必要时进行全面验证
- 缓存验证结果避免重复计算
- 使用增量验证只检查变更部分

### 3. 搜索优化
- 为频繁搜索的字段建立索引
- 使用缓存存储搜索结果
- 避免重复的深度遍历

### 4. 内存优化
- 及时释放不需要的对象
- 使用对象池减少 GC 压力
- 优化数据结构减少内存碎片

## 持续性能监控

建议在 CI/CD 流程中集成基准测试：

```bash
# 在 CI 中运行基准测试
go test -bench=. -benchmem ./benchmarks/ > benchmark_results.txt

# 比较性能变化
benchcmp old_results.txt new_results.txt
```

## 自定义基准测试

如果需要测试特定场景，可以参考现有测试编写自定义基准测试：

```go
func BenchmarkCustomOperation(b *testing.B) {
    // 准备测试数据
    config := "your config here"
    
    // 解析配置
    p := parser.NewStringParser(config)
    conf, err := p.Parse()
    if err != nil {
        b.Fatal(err)
    }
    
    // 重置计时器
    b.ResetTimer()
    
    // 执行基准测试
    for i := 0; i < b.N; i++ {
        // 你的操作
        result := someOperation(conf)
        _ = result // 避免编译器优化
    }
}
```

通过这些基准测试，可以确保 gonginx 库在各种使用场景下都有良好的性能表现。
