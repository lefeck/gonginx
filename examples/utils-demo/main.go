package main

import (
	"fmt"
	"log"

	"github.com/lefeck/gonginx/parser"
	"github.com/lefeck/gonginx/utils"
)

func main() {
	fmt.Println("=== Nginx 实用工具功能示例 ===")

	// 示例1: 配置差异比较
	fmt.Println("\n1. 配置差异比较示例:")
	testConfigDiff()

	// 示例2: 安全检查
	fmt.Println("\n2. 配置安全检查示例:")
	testSecurityCheck()

	// 示例3: 配置优化建议
	fmt.Println("\n3. 配置优化建议示例:")
	testConfigOptimization()

	// 示例4: 配置格式转换
	fmt.Println("\n4. 配置格式转换示例:")
	testFormatConversion()

	fmt.Println("\n=== 实用工具功能示例完成 ===")
}

func testConfigDiff() {
	// 旧配置
	oldConfig := `
worker_processes 1;

events {
    worker_connections 512;
}

http {
    sendfile on;
    keepalive_timeout 60;
    
    server {
        listen 80;
        server_name example.com;
        root /var/www/html;
        
        location / {
            try_files $uri $uri/ =404;
        }
    }
}
`

	// 新配置
	newConfig := `
worker_processes auto;

events {
    worker_connections 1024;
    use epoll;
}

http {
    sendfile on;
    tcp_nopush on;
    keepalive_timeout 65;
    gzip on;
    
    server {
        listen 80;
        server_name example.com www.example.com;
        root /var/www/html;
        
        location / {
            try_files $uri $uri/ =404;
        }
        
        location /api/ {
            proxy_pass http://backend;
        }
    }
    
    upstream backend {
        server 127.0.0.1:8080;
        server 127.0.0.1:8081;
    }
}
`

	// 比较配置
	diffResult, err := utils.CompareConfigStrings(oldConfig, newConfig)
	if err != nil {
		fmt.Printf("配置比较错误: %s\n", err)
		return
	}

	fmt.Printf("配置差异总结: %s\n", diffResult.Summary.String())

	if diffResult.HasChanges() {
		fmt.Println("\n详细差异:")
		for i, diff := range diffResult.Differences {
			fmt.Printf("%d. %s\n", i+1, diff.String())
		}

		// 按类型显示差异
		fmt.Println("\n按类型分组:")

		added := diffResult.GetByType(utils.DiffAdded)
		if len(added) > 0 {
			fmt.Printf("新增 (%d):\n", len(added))
			for _, diff := range added {
				fmt.Printf("  + %s\n", diff.String())
			}
		}

		modified := diffResult.GetByType(utils.DiffModified)
		if len(modified) > 0 {
			fmt.Printf("修改 (%d):\n", len(modified))
			for _, diff := range modified {
				fmt.Printf("  ~ %s\n", diff.String())
			}
		}

		removed := diffResult.GetByType(utils.DiffRemoved)
		if len(removed) > 0 {
			fmt.Printf("删除 (%d):\n", len(removed))
			for _, diff := range removed {
				fmt.Printf("  - %s\n", diff.String())
			}
		}
	} else {
		fmt.Println("配置无差异")
	}
}

func testSecurityCheck() {
	// 包含安全问题的配置
	unsafeConfig := `
worker_processes auto;

events {
    worker_connections 1024;
}

http {
    server_tokens on;
    
    server {
        listen 80;
        listen 443;
        server_name example.com;
        root /var/www/html;
        autoindex on;
        
        location / {
            try_files $uri $uri/ =404;
        }
        
        location /admin/ {
            # 管理员区域没有访问控制
            root /var/www/admin;
        }
        
        location ~ \.php$ {
            fastcgi_pass 127.0.0.1:9000;
        }
    }
    
    upstream backend {
        # 空的 upstream 块
    }
}
`

	// 解析配置
	p := parser.NewStringParser(unsafeConfig)
	conf, err := p.Parse()
	if err != nil {
		fmt.Printf("解析配置错误: %s\n", err)
		return
	}

	// 执行安全检查
	securityReport := utils.CheckSecurity(conf)

	fmt.Printf("安全评估结果: %s\n", securityReport.Summary.String())

	if len(securityReport.Issues) > 0 {
		fmt.Println("\n发现的安全问题:")

		// 按严重程度分组显示
		critical := securityReport.GetByLevel(utils.SecurityCritical)
		if len(critical) > 0 {
			fmt.Printf("\n严重问题 (%d):\n", len(critical))
			for i, issue := range critical {
				fmt.Printf("%d. %s\n", i+1, issue.String())
				if issue.Fix != "" {
					fmt.Printf("   修复建议: %s\n", issue.Fix)
				}
			}
		}

		warnings := securityReport.GetByLevel(utils.SecurityWarning)
		if len(warnings) > 0 {
			fmt.Printf("\n警告 (%d):\n", len(warnings))
			for i, issue := range warnings {
				fmt.Printf("%d. %s\n", i+1, issue.String())
				if issue.Fix != "" {
					fmt.Printf("   修复建议: %s\n", issue.Fix)
				}
			}
		}

		info := securityReport.GetByLevel(utils.SecurityInfo)
		if len(info) > 0 {
			fmt.Printf("\n信息 (%d):\n", len(info))
			for i, issue := range info {
				fmt.Printf("%d. %s\n", i+1, issue.String())
			}
		}
	}

	if len(securityReport.Passed) > 0 {
		fmt.Printf("\n通过的安全检查 (%d):\n", len(securityReport.Passed))
		for i, passed := range securityReport.Passed {
			fmt.Printf("%d. %s\n", i+1, passed)
		}
	}
}

func testConfigOptimization() {
	// 需要优化的配置
	suboptimalConfig := `
worker_processes 1;

events {
    worker_connections 512;
}

http {
    server {
        listen 80;
        server_name example.com;
        root /var/www/html;
        
        location / {
            try_files $uri $uri/ =404;
        }
        
        location ~ \.(css|js|png|jpg|gif)$ {
            # 静态文件没有缓存设置
            root /var/www/html;
        }
        
        location /api/ {
            proxy_pass http://backend;
            # 缺少代理缓冲区设置
        }
    }
    
    upstream backend {
        server 127.0.0.1:8080;
        # 缺少负载均衡算法和keepalive设置
    }
}
`

	// 解析配置
	p := parser.NewStringParser(suboptimalConfig)
	conf, err := p.Parse()
	if err != nil {
		fmt.Printf("解析配置错误: %s\n", err)
		return
	}

	// 执行优化分析
	optimizationReport := utils.OptimizeConfig(conf)

	fmt.Printf("优化分析结果: %s\n", optimizationReport.Summary.String())

	if len(optimizationReport.Suggestions) > 0 {
		fmt.Println("\n优化建议:")

		// 按优化类型分组显示
		performance := optimizationReport.GetByType(utils.OptimizePerformance)
		if len(performance) > 0 {
			fmt.Printf("\n性能优化 (%d):\n", len(performance))
			for i, suggestion := range performance {
				fmt.Printf("%d. %s\n", i+1, suggestion.Title)
				fmt.Printf("   描述: %s\n", suggestion.Description)
				fmt.Printf("   影响: %s\n", suggestion.Impact)
				if suggestion.CurrentValue != "" {
					fmt.Printf("   当前值: %s\n", suggestion.CurrentValue)
				}
				fmt.Printf("   建议值: %s\n", suggestion.SuggestedValue)
				fmt.Printf("   实现: %s\n", suggestion.Implementation)
				fmt.Println()
			}
		}

		security := optimizationReport.GetByType(utils.OptimizeSecurity)
		if len(security) > 0 {
			fmt.Printf("安全优化 (%d):\n", len(security))
			for i, suggestion := range security {
				fmt.Printf("%d. %s\n", i+1, suggestion.Title)
				fmt.Printf("   实现: %s\n", suggestion.Implementation)
			}
		}

		size := optimizationReport.GetByType(utils.OptimizeSize)
		if len(size) > 0 {
			fmt.Printf("大小优化 (%d):\n", len(size))
			for i, suggestion := range size {
				fmt.Printf("%d. %s\n", i+1, suggestion.Title)
				fmt.Printf("   实现: %s\n", suggestion.Implementation)
			}
		}

		maintenance := optimizationReport.GetByType(utils.OptimizeMaintenance)
		if len(maintenance) > 0 {
			fmt.Printf("维护性优化 (%d):\n", len(maintenance))
			for i, suggestion := range maintenance {
				fmt.Printf("%d. %s\n", i+1, suggestion.Title)
				fmt.Printf("   实现: %s\n", suggestion.Implementation)
			}
		}
	} else {
		fmt.Println("配置已经很好，没有需要优化的地方")
	}
}

func testFormatConversion() {
	// 简单的配置用于转换
	simpleConfig := `
worker_processes auto;

events {
    worker_connections 1024;
}

http {
    sendfile on;
    gzip on;
    
    server {
        listen 80;
        server_name example.com;
        root /var/www/html;
        
        location / {
            try_files $uri $uri/ =404;
        }
    }
}
`

	// 解析配置
	p := parser.NewStringParser(simpleConfig)
	conf, err := p.Parse()
	if err != nil {
		fmt.Printf("解析配置错误: %s\n", err)
		return
	}

	// 转换为JSON格式
	converter := utils.NewConfigConverter(conf)

	jsonOutput, err := converter.ConvertToJSON(true)
	if err != nil {
		fmt.Printf("转换为JSON错误: %s\n", err)
	} else {
		fmt.Println("JSON格式:")
		fmt.Println(jsonOutput)
	}

	// 转换为YAML格式
	yamlOutput, err := converter.ConvertToYAML()
	if err != nil {
		fmt.Printf("转换为YAML错误: %s\n", err)
	} else {
		fmt.Println("\nYAML格式:")
		fmt.Println(yamlOutput)
	}

	// 演示格式转换器
	fmt.Println("\n格式转换器示例:")
	formatConverter := utils.NewFormatConverter()

	// 假设我们有一个JSON配置
	jsonConfig := `{
  "worker_processes": "auto",
  "events": {
    "worker_connections": "1024"
  },
  "http": {
    "sendfile": "on",
    "gzip": "on",
    "servers": [
      {
        "listen": "80",
        "server_name": "example.com",
        "root": "/var/www/html"
      }
    ]
  }
}`

	// 从JSON转换为YAML
	yamlFromJson, err := formatConverter.Convert(jsonConfig, utils.FormatJSON, utils.FormatYAML)
	if err != nil {
		fmt.Printf("JSON到YAML转换错误: %s\n", err)
	} else {
		fmt.Println("从JSON转换的YAML:")
		fmt.Println(yamlFromJson)
	}
}

func init() {
	// 设置日志
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
