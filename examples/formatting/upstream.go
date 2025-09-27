package main

import (
	"fmt"
	"github.com/tufanbarisyildirim/gonginx/config"
	"github.com/tufanbarisyildirim/gonginx/dumper"
	"github.com/tufanbarisyildirim/gonginx/parser"
)

func upstream(file string) {
	p, err := parser.NewParser(file)
	if err != nil {
		panic(err)
	}
	conf, err := p.Parse()
	if err != nil {
		panic(err)
	}

	upstreams := conf.FindUpstreams()

	upstreams[0].AddServer(&config.UpstreamServer{
		Address: "127.0.0.1:443",
		Parameters: map[string]string{
			"weight": "5",
		},
		Flags: []string{"down"},
	})
	fmt.Println(dumper.DumpBlock(conf.Block, dumper.IndentedStyle))

	//conf.GetParent()
	//
	//up := &config.Upstream{
	//	UpstreamName: "example_upstream",
	//	Comment:      []string{"This is an example upstream block"},
	//}
	//
	//server := &config.UpstreamServer{
	//	Address: "192.19.2.11:8080",
	//}
	//server.SetParent(up)
	//parent := server.GetParent()
	//fmt.Println("server parent:", parent.GetName())
}

func main() {
	upstream("./examples/parse-nginx-conf-get-listen-port/nginx.conf")

}
