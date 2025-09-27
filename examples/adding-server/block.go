package main

import "github.com/lefeck/gonginx/config"

func main() {
	block := &config.Block{}
	serverDirective := &config.Directive{
		//Block: block,
		Name: "server",
	}
	block.Directives = append(block.Directives, serverDirective)
	results := block.FindDirectives("server")
	for _, directive := range results {
		println("Found directive:", directive.GetName())
		if directive.GetParent() != nil {
			println("Parent directive:", directive.GetParent().GetName())
		} else {
			println("No parent directive")
		}
		if block.GetCodeBlock() != "" {
			println("Code block:", block.GetCodeBlock())
		} else {
			println("No code block")
		}
	}
}
