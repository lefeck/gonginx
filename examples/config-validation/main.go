package main

import (
	"fmt"
	"log"

	"github.com/lefeck/gonginx/config"
	"github.com/lefeck/gonginx/parser"
)

func main() {
	fmt.Println("=== Nginx 配置验证功能示例 ===")

	// 示例1: 上下文验证
	fmt.Println("\n1. 上下文验证示例:")
	testContextValidation()

	// 示例2: 依赖关系验证
	fmt.Println("\n2. 依赖关系验证示例:")
	testDependencyValidation()

	// 示例3: 综合配置验证
	fmt.Println("\n3. 综合配置验证示例:")
	testComprehensiveValidation()

	// 示例4: 参数验证
	fmt.Println("\n4. 参数验证示例:")
	testParameterValidation()

	// 示例5: 结构验证
	fmt.Println("\n5. 结构验证示例:")
	testStructuralValidation()

	fmt.Println("\n=== 配置验证示例完成 ===")
}

func testContextValidation() {
	// 包含上下文错误的配置
	configWithContextErrors := `
# 错误：proxy_pass 不能在 http 上下文中
http {
    proxy_pass http://backend;  # 应该在 location 中
    
    server {
        listen 80;
        server_name example.com;
        
        # 错误：listen 不能在 location 中
        location / {
            listen 8080;  # 不应该在这里
            proxy_pass http://backend;
        }
    }
}

# 错误：server 不能在 main 上下文中
server {
    listen 80;
}
`

	parser := parser.NewStringParser(configWithContextErrors)
	conf, err := parser.Parse()
	if err != nil {
		fmt.Printf("解析错误: %s\n", err)
		return
	}

	// 创建上下文验证器
	contextValidator := config.NewContextValidator()
	errors := contextValidator.ValidateConfig(conf)

	if len(errors) > 0 {
		fmt.Printf("发现 %d 个上下文错误:\n", len(errors))
		for _, err := range errors {
			fmt.Printf("  - %s\n", err.Error())
		}
	} else {
		fmt.Println("没有发现上下文错误")
	}
}

func testDependencyValidation() {
	// 包含依赖关系错误的配置
	configWithDependencyErrors := `
http {
    # 错误：没有定义 upstream backend
    server {
        listen 80;
        server_name example.com;
        
        location / {
            proxy_pass http://backend;  # backend upstream 未定义
        }
        
        # 错误：SSL 证书没有对应的私钥
        ssl_certificate /etc/ssl/cert.pem;
        # 缺少 ssl_certificate_key
        
        # 错误：使用 proxy_cache 但没有定义 proxy_cache_path
        location /cache/ {
            proxy_cache my_cache;  # my_cache 未定义
            proxy_pass http://backend2;
        }
    }
    
    # 错误：empty upstream
    upstream empty_upstream {
        # 没有 server 指令
    }
    
    # 正确的 upstream
    upstream backend2 {
        server 192.168.1.10:8080;
        server 192.168.1.11:8080;
    }
}
`

	parser := parser.NewStringParser(configWithDependencyErrors)
	conf, err := parser.Parse()
	if err != nil {
		fmt.Printf("解析错误: %s\n", err)
		return
	}

	// 创建依赖验证器
	dependencyValidator := config.NewDependencyValidator()
	errors := dependencyValidator.ValidateDependencies(conf)

	if len(errors) > 0 {
		fmt.Printf("发现 %d 个依赖关系错误:\n", len(errors))
		for _, err := range errors {
			fmt.Printf("  - %s\n", err.Error())
		}
	} else {
		fmt.Println("没有发现依赖关系错误")
	}
}

func testComprehensiveValidation() {
	// 包含多种错误的配置
	configWithMultipleErrors := `
http {
    # 参数验证错误
    worker_connections;  # 缺少参数
    
    server {
        # 缺少 listen 指令（依赖验证）
        server_name example.com;
        
        # SSL 配置不完整（依赖验证）
        ssl_certificate /etc/ssl/cert.pem;
        
        location / {
            # 上下文错误
            listen 80;  # listen 不应该在 location 中
            
            # 参数错误
            proxy_pass;  # 缺少参数
        }
    }
    
    # 结构错误：重复的 server_name
    server {
        listen 8080;
        server_name example.com;  # 重复的 server_name
    }
}

# 结构错误：多个 http 块
http {
    server {
        listen 9090;
    }
}
`

	parser := parser.NewStringParser(configWithMultipleErrors)
	conf, err := parser.Parse()
	if err != nil {
		fmt.Printf("解析错误: %s\n", err)
		return
	}

	// 创建综合验证器
	validator := config.NewConfigValidator()
	report := validator.ValidateConfig(conf)

	fmt.Printf("验证报告: %s\n", report.Summary.String())

	if len(report.Issues) > 0 {
		fmt.Println("\n验证问题详情:")

		// 按级别分组显示
		errors := report.GetByLevel(config.ValidationError)
		warnings := report.GetByLevel(config.ValidationWarning)
		infos := report.GetByLevel(config.ValidationInfo)

		if len(errors) > 0 {
			fmt.Printf("\n错误 (%d):\n", len(errors))
			for _, issue := range errors {
				fmt.Printf("  - %s\n", issue.String())
				if issue.Fix != "" {
					fmt.Printf("    修复建议: %s\n", issue.Fix)
				}
			}
		}

		if len(warnings) > 0 {
			fmt.Printf("\n警告 (%d):\n", len(warnings))
			for _, issue := range warnings {
				fmt.Printf("  - %s\n", issue.String())
				if issue.Fix != "" {
					fmt.Printf("    修复建议: %s\n", issue.Fix)
				}
			}
		}

		if len(infos) > 0 {
			fmt.Printf("\n信息 (%d):\n", len(infos))
			for _, issue := range infos {
				fmt.Printf("  - %s\n", issue.String())
			}
		}
	} else {
		fmt.Println("没有发现配置问题")
	}
}

func testParameterValidation() {
	// 包含参数错误的配置
	configWithParameterErrors := `
events {
    worker_connections;  # 缺少参数
}

http {
    server {
        listen;  # 缺少端口
        server_name example.com;
        
        # SSL 文件路径验证
        ssl_certificate;  # 缺少文件路径
        ssl_certificate_key "invalid-key";  # 可能的格式问题
        
        location / {
            proxy_pass;  # 缺少目标
            root;  # 缺少路径
        }
    }
}
`

	parser := parser.NewStringParser(configWithParameterErrors)
	conf, err := parser.Parse()
	if err != nil {
		fmt.Printf("解析错误: %s\n", err)
		return
	}

	validator := config.NewConfigValidator()
	report := validator.ValidateConfig(conf)

	parameterIssues := report.GetByCategory("Parameter")
	if len(parameterIssues) > 0 {
		fmt.Printf("发现 %d 个参数错误:\n", len(parameterIssues))
		for _, issue := range parameterIssues {
			fmt.Printf("  - %s\n", issue.String())
			if issue.Fix != "" {
				fmt.Printf("    修复建议: %s\n", issue.Fix)
			}
		}
	} else {
		fmt.Println("没有发现参数错误")
	}
}

func testStructuralValidation() {
	// 包含结构错误的配置
	configWithStructuralErrors := `
# 多个 events 块
events {
    worker_connections 1024;
}

events {
    worker_connections 2048;  # 重复的 events 块
}

# 多个 http 块
http {
    server {
        listen 80;
        server_name example.com;
    }
}

http {
    server {
        listen 8080;
        server_name test.com;
    }
}

# 在第一个 http 块中的服务器
http {
    server {
        listen 80;
        server_name example.com;  # 重复的 server_name
    }
    
    server {
        listen 8080;
        server_name example.com;  # 又一个重复的 server_name
    }
    
    server {
        listen 80;  # 重复的端口
        server_name different.com;
    }
}
`

	parser := parser.NewStringParser(configWithStructuralErrors)
	conf, err := parser.Parse()
	if err != nil {
		fmt.Printf("解析错误: %s\n", err)
		return
	}

	validator := config.NewConfigValidator()
	report := validator.ValidateConfig(conf)

	structuralIssues := report.GetByCategory("Structure")
	if len(structuralIssues) > 0 {
		fmt.Printf("发现 %d 个结构错误:\n", len(structuralIssues))
		for _, issue := range structuralIssues {
			fmt.Printf("  - %s\n", issue.String())
			if issue.Fix != "" {
				fmt.Printf("    修复建议: %s\n", issue.Fix)
			}
		}
	} else {
		fmt.Println("没有发现结构错误")
	}
}

// 演示如何将验证集成到现有的错误处理中
func demonstrateIntegration() {
	fmt.Println("\n=== 集成验证到现有流程 ===")

	configFile := "nginx.conf"

	// 1. 使用增强解析器进行基础验证
	enhancedParser, err := parser.NewParser(configFile)
	if err != nil {
		log.Fatalf("创建解析器失败: %v", err)
	}

	conf, err := enhancedParser.Parse()
	if err != nil {
		log.Fatalf("解析配置失败: %v", err)
	}

	// 2. 使用新的配置验证器进行深度验证
	validator := config.NewConfigValidator()
	report := validator.ValidateConfig(conf)

	// 3. 根据验证结果决定下一步操作
	if report.HasErrors() {
		fmt.Printf("配置验证失败: %s\n", report.Summary.String())

		// 显示所有错误
		for _, issue := range report.GetByLevel(config.ValidationError) {
			fmt.Printf("错误: %s\n", issue.String())
		}

		// 不继续处理配置
		return
	}

	// 4. 显示警告但继续处理
	warnings := report.GetByLevel(config.ValidationWarning)
	if len(warnings) > 0 {
		fmt.Printf("发现 %d 个警告，但配置可以使用:\n", len(warnings))
		for _, warning := range warnings {
			fmt.Printf("警告: %s\n", warning.String())
		}
	}

	fmt.Println("配置验证通过，可以安全使用")
}
