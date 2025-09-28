package main

import (
	"fmt"
	"os"

	"github.com/lefeck/gonginx/dumper"
	"github.com/lefeck/gonginx/parser"
)

func parseConfigAndGetPorts(filePath string) ([]string, error) {
	// 初始化一个新的解析器, 加载指定的nginx配置文件
	p, err := parser.NewParser(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create parser: %w", err)
	}
	// 解析nginx配置文件,获取nginx配置对象
	conf, err := p.Parse()
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	// 使用dumper将配置格式化为标准格式字符串
	dumpString := dumper.DumpConfig(conf, dumper.IndentedStyle)
	// 将转换后的字符串写入到指定的文件路径
	if err := os.WriteFile(filePath, []byte(dumpString), 0644); err != nil {
		return nil, fmt.Errorf("failed to write config file: %w", err)
	}
	// 查找所有的server指令
	servers := conf.FindDirectives("server")
	ports := make([]string, 0)
	for _, server := range servers {
		parent := server.GetParent()
		if parent == nil {
			return nil, fmt.Errorf("server directive has no parent block")
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
	return ports, nil
}

func getServer(filePath string) ([]string, error) {
	p, err := parser.NewParser(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create parser: %w", err)
	}
	conf, err := p.Parse()
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	dumpString := dumper.DumpConfig(conf, dumper.IndentedStyle)
	if err := os.WriteFile(filePath, []byte(dumpString), 0644); err != nil {
		return nil, fmt.Errorf("failed to write config file: %w", err)
	}
	upstreams := conf.FindDirectives("upstream")
	values := make([]string, 0)
	for _, upstream := range upstreams {
		//fmt.Println("Upstream Name:", upstream.GetName())
		for _, param := range upstream.GetParameters() {
			values = append(values, param.GetValue())
		}
	}
	return values, nil
}

func main() {
	ports, err := parseConfigAndGetPorts("nginx.conf")
	if err != nil {
		panic(err)
	}
	fmt.Println(ports)

	//upstreams, err := getServer("nginx.conf")
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(upstreams)
}
