package main

import (
	"fmt"
	"log"
	"os"

	"github.com/lefeck/gonginx/config"
	"github.com/lefeck/gonginx/dumper"
	"github.com/lefeck/gonginx/parser"
)

// Include配置文件管理示例
func main() {
	fmt.Println("=== Include 配置文件管理示例 ===")

	// 创建示例配置
	setupExampleConfig()

	// 1. 解析带有 include 的配置
	fmt.Println("\n🔍 1. 解析主配置文件")
	conf := parseMainConfig()

	// 2. 查看 include 指令
	fmt.Println("\n📖 2. 查看所有 include 指令")
	listIncludes(conf)

	// 3. 创建新的 include 指令
	fmt.Println("\n📝 3. 创建新的 include 指令")
	createIncludeDirective(conf)

	// 4. 创建被包含的配置文件
	fmt.Println("\n📄 4. 创建被包含的配置文件")
	createIncludedFiles()

	// 5. 修改被包含文件的内容
	fmt.Println("\n✏️ 5. 修改被包含文件的内容")
	modifyIncludedFile()

	// 6. 删除 include 指令
	fmt.Println("\n🗑️ 6. 删除 include 指令")
	removeIncludeDirective(conf)

	// 7. 重新解析带 include 的完整配置
	fmt.Println("\n🔄 7. 重新解析完整配置 (含 include)")
	parseFullConfigWithIncludes()

	// 清理示例文件
	//cleanup()

	fmt.Println("\n=== Include 管理示例完成 ===")
}

// 设置示例配置文件
func setupExampleConfig() {
	// 创建目录
	os.MkdirAll("conf/vhosts", 0755)
	os.MkdirAll("conf/snippets", 0755)

	// 主配置文件
	mainConfig := `worker_processes auto;

http {
    include       conf/mime.types;
    sendfile      on;
    keepalive_timeout 65;
    
    # 包含虚拟主机配置
    include       conf/vhosts/*.conf;
    
    # 包含通用配置片段
    include       conf/snippets/ssl.conf;
}`

	err := os.WriteFile("nginx.conf", []byte(mainConfig), 0644)
	if err != nil {
		log.Fatal("创建主配置文件失败:", err)
	}

	// mime.types
	mimeTypes := `types {
    text/html                             html htm shtml;
    text/css                              css;
    text/xml                              xml;
    application/javascript                js;
    application/json                      json;
}`

	err = os.WriteFile("conf/mime.types", []byte(mimeTypes), 0644)
	if err != nil {
		log.Fatal("创建 mime.types 失败:", err)
	}

	fmt.Println("✅ 示例配置文件创建完成")
}

// 解析主配置文件
func parseMainConfig() *config.Config {
	p, err := parser.NewParser("nginx.conf")
	if err != nil {
		log.Fatal("创建解析器失败:", err)
	}

	conf, err := p.Parse()
	if err != nil {
		log.Fatal("解析配置失败:", err)
	}

	fmt.Println("✅ 主配置文件解析成功")
	return conf
}

// 列出所有 include 指令
func listIncludes(conf *config.Config) {
	includes := conf.FindDirectives("include")
	fmt.Printf("   找到 %d 个 include 指令:\n", len(includes))

	for i, inc := range includes {
		params := inc.GetParameters()
		if len(params) > 0 {
			fmt.Printf("   %d. %s\n", i+1, params[0].GetValue())
		}
	}
}

// 创建新的 include 指令
func createIncludeDirective(conf *config.Config) {
	// 在 http 块中添加新的 include
	httpBlocks := conf.FindDirectives("http")
	if len(httpBlocks) > 0 {
		if httpBlock, ok := httpBlocks[0].(*config.HTTP); ok {
			// 创建新的 include 指令
			includeDirective := &config.Directive{
				Name:       "include",
				Parameters: []config.Parameter{config.NewParameter("conf/api/*.conf")},
			}

			httpBlock.Directives = append(httpBlock.Directives, includeDirective)
			fmt.Println("   ✅ 成功添加 include 'conf/api/*.conf'")
		}
	}
}

// 创建被包含的配置文件
func createIncludedFiles() {
	// 创建目录
	os.MkdirAll("conf/api", 0755)

	// 创建 SSL 配置片段
	sslConfig := `# SSL 配置片段
ssl_protocols TLSv1.2 TLSv1.3;
ssl_ciphers HIGH:!aNULL:!MD5;
ssl_prefer_server_ciphers on;`

	err := os.WriteFile("conf/snippets/ssl.conf", []byte(sslConfig), 0644)
	if err != nil {
		log.Fatal("创建 SSL 配置失败:", err)
	}

	// 创建虚拟主机配置
	vhostConfig := `server {
    listen 80;
    server_name example.com;
    root /var/www/example.com;
    
    location / {
        try_files $uri $uri/ =404;
    }
    
    location /api {
        proxy_pass http://backend;
    }
}`

	err = os.WriteFile("conf/vhosts/example.conf", []byte(vhostConfig), 0644)
	if err != nil {
		log.Fatal("创建虚拟主机配置失败:", err)
	}

	// 创建 API 配置
	apiConfig := `upstream api_backend {
    server 127.0.0.1:8001;
    server 127.0.0.1:8002;
}

server {
    listen 8080;
    server_name api.example.com;
    
    location /v1 {
        proxy_pass http://api_backend;
        proxy_set_header X-Real-IP $remote_addr;
    }
}`

	err = os.WriteFile("conf/api/api.conf", []byte(apiConfig), 0644)
	if err != nil {
		log.Fatal("创建 API 配置失败:", err)
	}

	fmt.Println("   ✅ 成功创建被包含的配置文件:")
	fmt.Println("      - conf/snippets/ssl.conf")
	fmt.Println("      - conf/vhosts/example.conf")
	fmt.Println("      - conf/api/api.conf")
}

// 修改被包含文件的内容
func modifyIncludedFile() {
	// 读取并解析虚拟主机配置
	p, err := parser.NewParser("conf/vhosts/example.conf")
	if err != nil {
		log.Fatal("解析虚拟主机配置失败:", err)
	}

	conf, err := p.Parse()
	if err != nil {
		log.Fatal("解析虚拟主机配置失败:", err)
	}

	// 找到 server 块并添加新的 location
	servers := conf.FindDirectives("server")
	if len(servers) > 0 {
		serverBlock := servers[0].GetBlock()
		if block, ok := serverBlock.(*config.Block); ok {
			// 添加新的 location
			newLocation := &config.Directive{
				Name:       "location",
				Parameters: []config.Parameter{config.NewParameter("/health")},
				Block: &config.Block{
					Directives: []config.IDirective{
						&config.Directive{
							Name:       "return",
							Parameters: []config.Parameter{config.NewParameter("200"), config.NewParameter("'OK'")},
						},
					},
				},
			}

			block.Directives = append(block.Directives, newLocation)
		}
	}

	// 保存修改后的配置
	modifiedConfig := dumper.DumpConfig(conf, dumper.IndentedStyle)
	err = os.WriteFile("conf/vhosts/example.conf", []byte(modifiedConfig), 0644)
	if err != nil {
		log.Fatal("保存修改后的配置失败:", err)
	}

	fmt.Println("   ✅ 成功修改 conf/vhosts/example.conf，添加了 /health location")
}

// 删除 include 指令
func removeIncludeDirective(conf *config.Config) {
	httpBlocks := conf.FindDirectives("http")
	if len(httpBlocks) > 0 {
		if httpBlock, ok := httpBlocks[0].(*config.HTTP); ok {
			// 删除特定的 include 指令
			newDirectives := make([]config.IDirective, 0)
			removed := false

			for _, directive := range httpBlock.Directives {
				if directive.GetName() == "include" {
					params := directive.GetParameters()
					if len(params) > 0 && params[0].GetValue() == "conf/snippets/ssl.conf" {
						// 跳过这个 include（删除它）
						removed = true
						continue
					}
				}
				newDirectives = append(newDirectives, directive)
			}

			httpBlock.Directives = newDirectives
			if removed {
				fmt.Println("   ✅ 成功删除 include 'conf/snippets/ssl.conf'")
			}
		}
	}
}

// 使用 include 解析功能解析完整配置
func parseFullConfigWithIncludes() {
	// 使用 WithIncludeParsing 选项解析配置
	p, err := parser.NewParser("nginx.conf", parser.WithIncludeParsing())
	if err != nil {
		log.Fatal("创建解析器失败:", err)
	}

	conf, err := p.Parse()
	if err != nil {
		log.Fatal("解析完整配置失败:", err)
	}

	fmt.Println("   ✅ 成功解析完整配置 (包含 include 文件)")

	// 显示解析后的完整配置结构
	fmt.Println("\n   📊 配置统计:")

	// 统计各种指令数量
	allUpstreams := conf.FindUpstreams()
	allServers := conf.FindDirectives("server")
	allLocations := conf.FindDirectives("location")

	fmt.Printf("      - Upstream 块: %d 个\n", len(allUpstreams))
	fmt.Printf("      - Server 块: %d 个\n", len(allServers))
	fmt.Printf("      - Location 块: %d 个\n", len(allLocations))

	// 显示所有 upstream 的详细信息
	if len(allUpstreams) > 0 {
		fmt.Println("\n   🔍 Upstream 详情:")
		for i, upstream := range allUpstreams {
			fmt.Printf("      %d. %s (服务器数量: %d)\n",
				i+1, upstream.UpstreamName, len(upstream.UpstreamServers))
		}
	}

	// 显示具体包含的文件内容
	fmt.Println("\n   📁 Include 文件分析:")
	includes := conf.FindDirectives("include")
	for _, inc := range includes {
		if includeDir, ok := inc.(*config.Include); ok {
			params := inc.GetParameters()
			if len(params) > 0 {
				fmt.Printf("      - %s: 包含 %d 个配置\n", params[0].GetValue(), len(includeDir.Configs))
			}
		}
	}

	// 显示完整的配置内容
	fmt.Println("\n   📄 完整解析后的配置:")
	fullConfig := dumper.DumpConfig(conf, dumper.IndentedStyle)
	fmt.Println(fullConfig)
}

// 清理示例文件
func cleanup() {
	fmt.Println("\n🧹 清理示例文件...")

	// 删除创建的文件和目录
	os.RemoveAll("conf")
	os.Remove("nginx.conf")

	fmt.Println("   ✅ 清理完成")
}

// IncludeManager 结构体 - 推荐的 Include 管理接口设计
type IncludeManager struct {
	config     *config.Config
	configPath string
}

// NewIncludeManager 创建 Include 管理器
func NewIncludeManager(configPath string) (*IncludeManager, error) {
	p, err := parser.NewParser(configPath, parser.WithIncludeParsing())
	if err != nil {
		return nil, err
	}

	conf, err := p.Parse()
	if err != nil {
		return nil, err
	}

	return &IncludeManager{
		config:     conf,
		configPath: configPath,
	}, nil
}

// 推荐的 API 接口设计 (未实现，仅作为设计参考)

// AddInclude 添加新的 include 指令
func (im *IncludeManager) AddInclude(blockName, includePath string) error {
	// 实现逻辑...
	return nil
}

// RemoveInclude 删除 include 指令
func (im *IncludeManager) RemoveInclude(includePath string) error {
	// 实现逻辑...
	return nil
}

// CreateIncludedFile 创建被包含的配置文件
func (im *IncludeManager) CreateIncludedFile(filePath string, content interface{}) error {
	// 实现逻辑...
	return nil
}

// UpdateIncludedFile 更新被包含的配置文件
func (im *IncludeManager) UpdateIncludedFile(filePath string, updater func(*config.Config) error) error {
	// 实现逻辑...
	return nil
}

// DeleteIncludedFile 删除被包含的配置文件
func (im *IncludeManager) DeleteIncludedFile(filePath string) error {
	// 实现逻辑...
	return nil
}

// ListIncludes 列出所有 include 指令
func (im *IncludeManager) ListIncludes() []string {
	// 实现逻辑...
	return nil
}

// ValidateIncludes 验证所有 include 文件是否存在和有效
func (im *IncludeManager) ValidateIncludes() []error {
	// 实现逻辑...
	return nil
}

// SaveAll 保存主配置和所有被包含的文件
func (im *IncludeManager) SaveAll() error {
	// 实现逻辑...
	return nil
}
