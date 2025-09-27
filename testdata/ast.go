package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

func main() {
	src := `package main
func main() {
    a := 1 + 2
}`
	fset := token.NewFileSet()

	expr, err := parser.ParseExprFrom(fset, "", src, 0)
	if err != nil {
		panic(err)
	}

	ast.Inspect(expr, func(n ast.Node) bool {
		if n != nil {
			pos := fset.Position(n.Pos())
			fmt.Printf("Node: %-20T Pos: %s\n", n, pos)
		}
		return true
	})
}
