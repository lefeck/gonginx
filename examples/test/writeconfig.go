package main

import (
	"github.com/lefeck/gonginx/dumper"
	"github.com/lefeck/gonginx/parser"
)

func writerConfig() error {
	p, err := parser.NewParser("nginx.conf")
	if err != nil {
		return err
	}

	cfg, err := p.Parse()

	if err != nil {
		return err
	}

	err = dumper.WriteConfig(cfg, dumper.IndentedStyle, true)

	if err != nil {
		return err
	}
	return nil
}

func main() {
	writerConfig()
}
