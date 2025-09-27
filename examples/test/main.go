package main

import (
	"fmt"
	"github.com/tufanbarisyildirim/gonginx/config"
	"github.com/tufanbarisyildirim/gonginx/dumper"
	"github.com/tufanbarisyildirim/gonginx/parser"
	"os"
)

func formattedConfigFile(file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	p := parser.NewStringParser(string(data))
	conf, err := p.Parse()
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}
	dumpString := dumper.DumpConfig(conf, dumper.IndentedStyle)
	if err := os.WriteFile(file, []byte(dumpString), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	return nil
}

func test() ([]*config.Include, error) {
	nginxConfig := "./nginx.conf"
	formattedConfigFile(nginxConfig)

	p, err := parser.NewParser(nginxConfig)
	if err != nil {
		panic(err)
	}
	conf, err := p.Parse()
	if err != nil {
		panic(err)
	}

	directives := conf.FindDirectives("include")
	for _, directive := range directives {
		//println("Found directive:", directive.GetName())
		fmt.Printf("name: %s, value: %s \n", directive.GetName(), directive.GetParameters()[0].GetValue())
		includePath := directive.GetParameters()[0].GetValue()
		fmt.Printf("Found include: %s\n", includePath)

		// 构造一个 config.Include 对象
		include := &config.Include{
			IncludePath: includePath,
		}
		fmt.Println(include.IncludePath)

		includeDirective, err := p.ParseInclude(include)

		if err != nil {
			fmt.Printf("Error parsing include %s: %v\n", includePath, err)
			continue
		}
		fmt.Printf("Included directive from %s:\n", includePath)

	}
}

func getports()  {
	// 我执行到下面这一行就报错了
	block, ok := includeDirective.(*config.Block)
	if !ok {
		fmt.Printf("Included directive %s is not a block directive\n", includePath)
		continue
	}
	servers := block.FindDirectives("server")
	ports := make([]string, 0)
	for _, server := range servers {
		parent := server.GetParent()
		if parent == nil {
			fmt.Println("server directive has no parent block")
		}
		fmt.Println("Server Name:", server.GetName())
		// 对每个server指令，查找其下的listen指令
		listens := server.GetBlock().FindDirectives("listen")
		if len(listens) > 0 {
			listenPorts := listens[0].GetParameters()
			for _, port := range listenPorts {
				ports = append(ports, port.GetValue())
			}
		}
	}
	fmt.Println("Server Ports:", ports)
}
}


func main() {
	test()
}
