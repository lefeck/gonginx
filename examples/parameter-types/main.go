package main

import (
	"fmt"
	"log"

	"github.com/lefeck/gonginx/config"
	"github.com/lefeck/gonginx/parser"
)

func main() {
	// Create a configuration with various parameter types
	configContent := `
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
`

	fmt.Println("=== 参数类型系统示例 ===")
	fmt.Println("解析配置并分析参数类型...")

	p := parser.NewStringParser(configContent)
	conf, err := p.Parse()
	if err != nil {
		log.Fatal("解析错误:", err)
	}

	// 递归分析配置中的所有参数
	var analyzeDirective func(directive config.IDirective, indent string)
	analyzeDirective = func(directive config.IDirective, indent string) {
		fmt.Printf("%s指令: %s\n", indent, directive.GetName())

		params := directive.GetParameters()
		if len(params) > 0 {
			for i, param := range params {
				fmt.Printf("%s  参数 %d: \"%s\" (类型: %s)\n",
					indent, i+1, param.GetValue(), param.GetType().String())

				// 展示类型检查方法
				if param.IsSize() {
					if size, valid := config.ValidateSize(param.GetValue()); valid {
						fmt.Printf("%s    ✓ 有效的大小值: %s\n", indent, size)
					}
				}

				if param.IsTime() {
					if time, valid := config.ValidateTime(param.GetValue()); valid {
						fmt.Printf("%s    ✓ 有效的时间值: %s\n", indent, time)
					}
				}

				if param.IsNumber() {
					if num, valid := config.ValidateNumber(param.GetValue()); valid {
						fmt.Printf("%s    ✓ 有效的数值: %.0f\n", indent, num)
					}
				}

				if param.IsBoolean() {
					if val, valid := config.ValidateBoolean(param.GetValue()); valid {
						fmt.Printf("%s    ✓ 布尔值: %t\n", indent, val)
					}
				}

				if param.IsVariable() {
					fmt.Printf("%s    ✓ nginx 变量\n", indent)
				}

				if param.IsPath() {
					fmt.Printf("%s    ✓ 文件/目录路径\n", indent)
				}

				if param.IsURL() {
					fmt.Printf("%s    ✓ URL 地址\n", indent)
				}

				if param.IsRegex() {
					fmt.Printf("%s    ✓ 正则表达式\n", indent)
				}

				if param.IsQuoted() {
					fmt.Printf("%s    ✓ 引用字符串\n", indent)
				}
			}
		}

		// 递归处理子块
		if directive.GetBlock() != nil {
			for _, subDirective := range directive.GetBlock().GetDirectives() {
				analyzeDirective(subDirective, indent+"  ")
			}
		}
	}

	// 分析所有顶级指令
	for _, directive := range conf.Block.GetDirectives() {
		analyzeDirective(directive, "")
		fmt.Println()
	}

	// 展示参数类型统计
	fmt.Println("=== 参数类型统计 ===")
	typeCount := make(map[config.ParameterType]int)

	var countParameters func(directive config.IDirective)
	countParameters = func(directive config.IDirective) {
		for _, param := range directive.GetParameters() {
			typeCount[param.GetType()]++
		}
		if directive.GetBlock() != nil {
			for _, subDirective := range directive.GetBlock().GetDirectives() {
				countParameters(subDirective)
			}
		}
	}

	for _, directive := range conf.Block.GetDirectives() {
		countParameters(directive)
	}

	for paramType, count := range typeCount {
		fmt.Printf("%s: %d 个\n", paramType.String(), count)
	}

	// 展示手动创建参数的功能
	fmt.Println("\n=== 手动创建参数示例 ===")

	// 自动检测类型
	param1 := config.NewParameter("512M")
	fmt.Printf("NewParameter(\"512M\"): %s (类型: %s)\n", param1.GetValue(), param1.GetType().String())

	// 显式指定类型
	param2 := config.NewParameterWithType("custom_value", config.ParameterTypeString)
	fmt.Printf("NewParameterWithType(\"custom_value\", String): %s (类型: %s)\n", param2.GetValue(), param2.GetType().String())

	// 类型转换示例
	param3 := config.NewParameter("on")
	if param3.IsBoolean() {
		if val, valid := config.ValidateBoolean(param3.GetValue()); valid {
			fmt.Printf("布尔参数 \"%s\" 的值为: %t\n", param3.GetValue(), val)
		}
	}

	fmt.Println("\n=== 参数类型系统功能完成 ===")
}
