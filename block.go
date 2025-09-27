package gonginx

// Block a block statement
type Block struct {
	Directives  []IDirective
	IsLuaBlock  bool
	LiteralCode string
}

// GetDirectives get all directives in this block
func (b *Block) GetDirectives() []IDirective {
	return b.Directives
}

//把相同类型的属性值的这些方法，抽象成一个interface，使用就直接调用接口就可以使用

func (b *Block) GetCodeBlock() string {
	return b.LiteralCode
}

// FindDirectives find directives in block recursively
func (b *Block) FindDirectives(directiveName string) []IDirective {
	directives := make([]IDirective, 0)
	//for _, dir :=range directives {
	//	if dir.GetName() == directiveName {
	//
	//	}
	//}
	for _, directive := range b.GetDirectives() {
		if directive.GetName() == directiveName {
			directives = append(directives, directive)
		}
		if include, ok := directive.(*Include); ok {
			for _, c := range include.Configs {
				directives = append(directives, c.FindDirectives(directiveName)...)
			}
		}
		if directive.GetBlock() != nil {
			directives = append(directives, directive.GetBlock().FindDirectives(directiveName)...)
		}
	}

	return directives
}
