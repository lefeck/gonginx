package main

import (
	"fmt"
	"log"

	"github.com/lefeck/gonginx/errors"
)

func main() {
	fmt.Println("=== Nginx 错误处理改进示例 ===")

	// 示例1: 语法错误处理
	fmt.Println("\n1. 语法错误处理示例:")
	testSyntaxError()

	// 示例2: 语义错误处理
	fmt.Println("\n2. 语义错误处理示例:")
	testSemanticError()

	// 示例3: 配置验证错误
	fmt.Println("\n3. 配置验证错误示例:")
	testValidationError()

	// 示例4: 文件错误处理
	fmt.Println("\n4. 文件错误处理示例:")
	testFileError()

	// 示例5: 未知指令处理
	fmt.Println("\n5. 未知指令错误处理示例:")
	testUnknownDirectiveError()

	// 示例6: 多个错误处理
	fmt.Println("\n6. 多个错误处理示例:")
	testMultipleErrors()

	fmt.Println("\n=== 错误处理示例完成 ===")
}

func testSyntaxError() {
	// 包含语法错误的配置
	configWithSyntaxError := `
server {
    listen 80
    server_name example.com  # 缺少分号
    root /var/www/html
    # 缺少关闭大括号
`

	parser := errors.NewEnhancedStringParser(configWithSyntaxError)
	_, err := parser.Parse()

	if err != nil {
		fmt.Printf("检测到语法错误:\n%s\n", err.Error())
	}
}

func testSemanticError() {
	// 包含语义错误的配置
	configWithSemanticError := `
http {
    server {
        listen 80;
        server_name example.com;
        # 同样的 server_name 出现多次
    }
    server {
        listen 80;
        server_name example.com;  # 重复的 server_name
    }
}

http {
    # 多个 http 块
    server {
        listen 8080;
    }
}
`

	parser := errors.NewEnhancedStringParser(configWithSemanticError)
	_, err := parser.ParseWithValidation()

	if err != nil {
		fmt.Printf("检测到语义错误:\n%s\n", err.Error())
	}
}

func testValidationError() {
	// 包含配置验证错误的配置
	configWithValidationError := `
http {
    upstream backend {
        # 没有任何 server 的 upstream
    }
    
    server {
        # 没有 listen 指令的 server
        server_name example.com;
        
        # SSL 配置不完整
        ssl_certificate /path/to/cert.pem;
        # 缺少 ssl_certificate_key
    }
}
`

	parser := errors.NewEnhancedStringParser(configWithValidationError)
	_, err := parser.ParseWithValidation()

	if err != nil {
		fmt.Printf("检测到配置验证错误:\n%s\n", err.Error())
	}
}

func testFileError() {
	// 尝试解析不存在的文件
	_, err := errors.NewEnhancedParser("/path/to/nonexistent/nginx.conf")

	if err != nil {
		fmt.Printf("检测到文件错误:\n%s\n", err.Error())
	}
}

func testUnknownDirectiveError() {
	// 包含未知指令的配置
	configWithUnknownDirective := `
http {
    server {
        listen 80;
        servername example.com;        # 错误: 应该是 server_name
        documentroot /var/www/html;    # 错误: 应该是 root
        proxypass http://backend;      # 错误: 应该是 proxy_pass
        workerprocesses auto;          # 错误: 在错误的上下文中
    }
}
`

	parser := errors.NewEnhancedStringParser(configWithUnknownDirective)
	_, err := parser.ParseWithValidation()

	if err != nil {
		fmt.Printf("检测到未知指令错误:\n%s\n", err.Error())
	} else {
		fmt.Println("注意: 一些未知指令可能没有被检测到，因为它们被解析为有效的自定义指令")
	}
}

func testMultipleErrors() {
	// 包含多种错误的配置
	configWithMultipleErrors := `
# 多个错误的配置示例
worker_processes auto;

events {
    worker_connections 1024;
}

events {
    # 重复的 events 块
    worker_connections 2048;
}

http {
    server {
        # 缺少 listen 指令
        server_name example.com;
        ssl_certificate /etc/ssl/cert.pem;
        # 缺少 ssl_certificate_key
    }
    
    server {
        listen 80;
        server_name example.com;  # 重复的 server_name
    }
    
    upstream empty_upstream {
        # 空的 upstream 块
    }
    
    upstream another_upstream {
        server 192.168.1.100:8080;
    }
}

http {
    # 重复的 http 块
    server {
        listen 8080;
    }
}
`

	parser := errors.NewEnhancedStringParser(configWithMultipleErrors)
	_, err := parser.ParseWithValidation()

	if err != nil {
		if errCollection, ok := err.(*errors.ErrorCollection); ok {
			fmt.Printf("检测到 %d 个错误:\n", errCollection.Count())

			// 按类型分组显示错误
			syntaxErrors := errCollection.GetByType(errors.SyntaxError)
			semanticErrors := errCollection.GetByType(errors.SemanticError)
			validationErrors := errCollection.GetByType(errors.ValidationError)

			if len(syntaxErrors) > 0 {
				fmt.Printf("\n语法错误 (%d):\n", len(syntaxErrors))
				for i, err := range syntaxErrors {
					fmt.Printf("  %d. %s\n", i+1, err.Error())
				}
			}

			if len(semanticErrors) > 0 {
				fmt.Printf("\n语义错误 (%d):\n", len(semanticErrors))
				for i, err := range semanticErrors {
					fmt.Printf("  %d. %s\n", i+1, err.Error())
				}
			}

			if len(validationErrors) > 0 {
				fmt.Printf("\n验证错误 (%d):\n", len(validationErrors))
				for i, err := range validationErrors {
					fmt.Printf("  %d. %s\n", i+1, err.Error())
				}
			}
		} else {
			fmt.Printf("检测到错误:\n%s\n", err.Error())
		}
	}
}

// 演示创建自定义错误的功能
func demonstrateCustomErrors() {
	fmt.Println("\n=== 自定义错误创建示例 ===")

	// 创建不同类型的错误
	syntaxErr := errors.NewSyntaxError("missing semicolon").
		WithFile("nginx.conf").
		WithLine(42).
		WithColumn(25).
		WithContext("server { listen 80 }").
		WithSuggestion("Add a semicolon after the listen directive")

	semanticErr := errors.NewSemanticError("invalid parameter value").
		WithDirective("worker_processes").
		WithParameter("invalid_value").
		WithSuggestion("Use 'auto' or a positive integer")

	fileErr := errors.NewFileError("include file not found").
		WithFile("/etc/nginx/sites-enabled/default").
		WithSuggestion("Check if the included file exists and has correct permissions")

	// 创建错误集合
	errCollection := errors.NewErrorCollection()
	errCollection.Add(syntaxErr)
	errCollection.Add(semanticErr)
	errCollection.Add(fileErr)

	fmt.Printf("错误集合:\n%s\n", errCollection.Error())
}

func init() {
	// 设置日志
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
