package parser

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lefeck/gonginx/config"
	"github.com/lefeck/gonginx/parser/token"
)

// Option parsing option
type Option func(*Parser)

type options struct {
	parseInclude               bool
	skipIncludeParsingErr      bool
	skipComments               bool
	customDirectives           map[string]string
	skipValidSubDirectiveBlock map[string]struct{}
	skipValidDirectivesErr     bool
}

func defaultOptions() options {
	return options{
		parseInclude:               false,
		skipIncludeParsingErr:      false,
		skipComments:               false,
		customDirectives:           map[string]string{},
		skipValidSubDirectiveBlock: map[string]struct{}{},
		skipValidDirectivesErr:     false,
	}
}

// Parser is an nginx config parser
type Parser struct {
	opts              options
	configRoot        string // TODO: confirmation needed (whether this is the parent of nginx.conf)
	lexer             *lexer
	currentToken      token.Token
	followingToken    token.Token
	parsedIncludes    map[*config.Include]*config.Config
	statementParsers  map[string]func() (config.IDirective, error)
	blockWrappers     map[string]func(*config.Directive) (config.IDirective, error)
	directiveWrappers map[string]func(*config.Directive) (config.IDirective, error)
	includeWrappers   map[string]func(*config.Directive) (config.IDirective, error)

	commentBuffer []string
	file          *os.File
	contextStack  []string // Track parsing context (e.g., "stream", "http")
}

// WithSameOptions copy options from another parser
func WithSameOptions(p *Parser) Option {
	return func(curr *Parser) {
		curr.opts = p.opts
	}
}

func withParsedIncludes(parsedIncludes map[*config.Include]*config.Config) Option {
	return func(p *Parser) {
		p.parsedIncludes = parsedIncludes
	}
}

func withConfigRoot(configRoot string) Option {
	return func(p *Parser) {
		p.configRoot = configRoot
	}
}

// WithSkipIncludeParsingErr ignores include parsing errors
func WithSkipIncludeParsingErr() Option {
	return func(p *Parser) {
		p.opts.skipIncludeParsingErr = true
	}
}

// WithDefaultOptions default options
func WithDefaultOptions() Option {
	return func(p *Parser) {
		p.opts = defaultOptions()
	}
}

// WithSkipComments default options
func WithSkipComments() Option {
	return func(p *Parser) {
		p.opts.skipComments = true
	}
}

// WithIncludeParsing enable parsing included files
func WithIncludeParsing() Option {
	return func(p *Parser) {
		p.opts.parseInclude = true
	}
}

// WithCustomDirectives add your custom directives as valid directives
func WithCustomDirectives(directives ...string) Option {
	return func(p *Parser) {
		for _, directive := range directives {
			p.opts.customDirectives[directive] = directive
		}
	}
}

// WithSkipValidBlocks add your custom block as valid
func WithSkipValidBlocks(directives ...string) Option {
	return func(p *Parser) {
		for _, directive := range directives {
			p.opts.skipValidSubDirectiveBlock[directive] = struct{}{}
		}
	}
}

// WithSkipValidDirectivesErr ignores unknown directive errors
func WithSkipValidDirectivesErr() Option {
	return func(p *Parser) {
		p.opts.skipValidDirectivesErr = true
	}
}

// NewStringParser parses nginx conf from string
func NewStringParser(str string, opts ...Option) *Parser {
	return NewParserFromLexer(lex(str), opts...)
}

// NewParser create new parser
func NewParser(filePath string, opts ...Option) (*Parser, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	l := newLexer(bufio.NewReader(f))
	l.file = filePath
	p := NewParserFromLexer(l, opts...)
	p.file = f
	return p, nil
}

// NewParserFromLexer initilizes a new Parser
func NewParserFromLexer(lexer *lexer, opts ...Option) *Parser {
	configRoot, _ := filepath.Split(lexer.file)
	parser := &Parser{
		lexer:          lexer,
		opts:           defaultOptions(),
		parsedIncludes: make(map[*config.Include]*config.Config),
		configRoot:     configRoot,
	}

	for _, o := range opts {
		o(parser)
	}

	parser.nextToken()
	parser.nextToken()

	parser.blockWrappers = config.BlockWrappers
	parser.directiveWrappers = config.DirectiveWrappers
	parser.includeWrappers = config.IncludeWrappers
	return parser
}

func (p *Parser) nextToken() {
	p.currentToken = p.followingToken
	p.followingToken = p.lexer.scan()
}

func (p *Parser) curTokenIs(t token.Type) bool {
	return p.currentToken.Type == t
}

func (p *Parser) followingTokenIs(t token.Type) bool {
	return p.followingToken.Type == t
}

// Parse the gonginx.
func (p *Parser) Parse() (*config.Config, error) {
	parsedBlock, err := p.parseBlock(false, false)
	if err != nil {
		return nil, err
	}
	c := &config.Config{
		FilePath: p.lexer.file, //TODO: set filepath here,
		Block:    parsedBlock,
	}
	err = p.Close()
	return c, err
}

// ParseBlock parse a block statement
func (p *Parser) parseBlock(inBlock bool, isSkipValidDirective bool) (*config.Block, error) {

	context := &config.Block{
		Directives: make([]config.IDirective, 0),
	}
	var s config.IDirective
	var err error
	var line int
parsingLoop:
	for {
		switch {
		case p.curTokenIs(token.EOF):
			if inBlock {
				return nil, errors.New("unexpected eof in block")
			}
			break parsingLoop
		case p.curTokenIs(token.LuaCode):
			context.IsLuaBlock = true
			context.LiteralCode = p.currentToken.Literal
		case p.curTokenIs(token.BlockEnd):
			break parsingLoop
		case p.curTokenIs(token.Keyword) || p.curTokenIs(token.QuotedString):
			s, err = p.parseStatement(isSkipValidDirective)
			if err != nil {
				return nil, err
			}
			if s.GetBlock() == nil {
				s.SetParent(s)
			} else {
				// each directive should have a parent directive, not a block
				// find each directive in the block and set the parent directive
				b := s.GetBlock()
				for _, dir := range b.GetDirectives() {
					dir.SetParent(s)
				}
			}
			line = p.currentToken.Line
			s.SetLine(line)
			context.Directives = append(context.Directives, s)
		case p.curTokenIs(token.Comment):
			if p.opts.skipComments {
				break
			}
			// outline comment
			p.commentBuffer = append(p.commentBuffer, p.currentToken.Literal)
		}
		p.nextToken()
	}

	return context, nil
}

func (p *Parser) parseStatement(isSkipValidDirective bool) (config.IDirective, error) {
	d := &config.Directive{
		Name: p.currentToken.Literal,
	}

	if !p.opts.skipValidDirectivesErr && !isSkipValidDirective {
		_, ok := ValidDirectives[d.Name]
		_, ok2 := p.opts.customDirectives[d.Name]

		if !ok && !ok2 {
			return nil, fmt.Errorf("unknown directive '%s' on line %d, column %d", d.Name, p.currentToken.Line, p.currentToken.Column)
		}
	}

	//if we have a special parser for the directive, we use it.
	if sp, ok := p.statementParsers[d.Name]; ok {
		return sp()
	}

	// set outline comment
	if len(p.commentBuffer) > 0 {
		d.Comment = p.commentBuffer
		p.commentBuffer = make([]string, 0)
	}

	directiveLineIndex := p.currentToken.Line // keep track of the line index of the directive
	// Parse parameters until reaching the semicolon that ends the directive.
	for {
		p.nextToken()
		if p.currentToken.IsParameterEligible() {
			param := config.Parameter{
				Value:             p.currentToken.Literal,
				Type:              config.DetectParameterType(p.currentToken.Literal),
				RelativeLineIndex: p.currentToken.Line - directiveLineIndex,
			}
			d.Parameters = append(d.Parameters, param)
			if p.currentToken.Is(token.BlockEnd) {
				return d, nil
			}
		} else if p.curTokenIs(token.Semicolon) {
			// inline comment in following token
			if !p.opts.skipComments {
				if p.followingTokenIs(token.Comment) && p.followingToken.Line == p.currentToken.Line {
					// if following token is a comment, then it is an inline comment, fetch next token
					p.nextToken()
					d.SetInlineComment(config.InlineComment{
						Value:             p.currentToken.Literal,
						RelativeLineIndex: p.currentToken.Line - directiveLineIndex,
					})
				}
			}
			if iw, ok := p.includeWrappers[d.Name]; ok {
				include, err := iw(d)
				if err != nil {
					return nil, err
				}
				return p.ParseInclude(include.(*config.Include))
			} else if dw, ok := p.directiveWrappers[p.getContextAwareWrapperKey(d.Name)]; ok {
				return dw(d)
			}
			return d, nil
		} else if p.curTokenIs(token.Comment) {
			// param comment
			d.SetInlineComment(config.InlineComment{
				Value:             p.currentToken.Literal,
				RelativeLineIndex: p.currentToken.Line - directiveLineIndex,
			})
		} else if p.curTokenIs(token.BlockStart) {
			_, blockSkip1 := SkipValidBlocks[d.Name]
			_, blockSkip2 := p.opts.skipValidSubDirectiveBlock[d.Name]
			isSkipBlockSubDirective := blockSkip1 || blockSkip2 || isSkipValidDirective

			// Special handling for *_by_lua_block directives
			if strings.HasSuffix(d.Name, "_by_lua_block") {
				// For Lua blocks, we need to capture the content without parsing it as nginx directives
				b := &config.Block{
					IsLuaBlock:  true,
					Directives:  []config.IDirective{},
					LiteralCode: "",
				}

				// Skip past the opening brace
				p.nextToken()

				// Collect all content until the matching closing brace
				// We need to count braces to handle nested blocks within Lua code
				braceCount := 1
				var luaCode strings.Builder

				for braceCount > 0 && !p.curTokenIs(token.EOF) {
					if p.curTokenIs(token.BlockStart) {
						braceCount++
					} else if p.curTokenIs(token.BlockEnd) {
						braceCount--
						if braceCount == 0 {
							// This is the closing brace of the Lua block
							break
						}
					}

					// Append token to Lua code if it's not the closing brace
					if !(p.curTokenIs(token.BlockEnd) && braceCount == 0) {
						luaCode.WriteString(p.currentToken.Literal)
						// Add space between tokens for readability
						if p.followingToken.Type != token.BlockEnd &&
							p.followingToken.Type != token.Semicolon &&
							p.followingToken.Type != token.EndOfLine {
							luaCode.WriteString(" ")
						}
					}

					p.nextToken()
				}

				b.LiteralCode = strings.TrimSpace(luaCode.String())
				d.Block = b

				// Use the appropriate wrapper based on the directive name
				if strings.HasSuffix(d.Name, "_by_lua_block") {
					return p.blockWrappers["_by_lua_block"](d)
				}
				return d, nil
			}

			// Only push context for certain block directives that define context
			shouldPushContext := false
			switch d.Name {
			case "stream", "http", "events", "mail", "upstream":
				shouldPushContext = true
			}

			if shouldPushContext {
				p.pushContext(d.Name)
			}

			b, err := p.parseBlock(true, isSkipBlockSubDirective)
			if err != nil {
				if shouldPushContext {
					p.popContext()
				}
				return nil, err
			}
			d.Block = b

			// Context-aware wrapper selection
			wrapperKey := p.getContextAwareWrapperKey(d.Name)
			if bw, ok := p.blockWrappers[wrapperKey]; ok {
				result, err := bw(d)
				if shouldPushContext {
					p.popContext()
				}
				return result, err
			}

			if shouldPushContext {
				p.popContext()
			}
			return d, nil
		} else if p.currentToken.Is(token.EndOfLine) {
			continue
		} else {
			return nil, fmt.Errorf("unexpected token %s (%s) on line %d, column %d", p.currentToken.Type.String(), p.currentToken.Literal, p.currentToken.Line, p.currentToken.Column)
		}
	}
}

// 中文解释: ParseInclude 解析 include 指令
// 如果配置选项 parseInclude 为 true，则会解析 include 指令
// 如果 include 路径不是绝对路径，则将其与配置根目录拼接
// 然后使用 filepath.Glob 查找匹配的文件路径
// 如果解析过程中发生错误且配置选项 skipIncludeParsingErr 为 false，则返回错误
// 对于每个匹配的 include 路径，创建一个新的 Parser 实例
// 并调用 Parse 方法解析配置
// 如果解析成功，将解析结果添加到 include.Configs 中
// 如果解析过程中发生错误且配置选项 skipIncludeParsingErr 为 false，则返回错误
func (p *Parser) ParseInclude(include *config.Include) (config.IDirective, error) {
	if p.opts.parseInclude {
		includePath := include.IncludePath
		if !filepath.IsAbs(includePath) {
			includePath = filepath.Join(p.configRoot, include.IncludePath)
		}
		includePaths, err := filepath.Glob(includePath)
		if err != nil && !p.opts.skipIncludeParsingErr {
			return nil, err
		}
		for _, includePath := range includePaths {
			if conf, ok := p.parsedIncludes[include]; ok {
				// same file includes itself? don't blow up the parser
				if conf == nil {
					continue
				}
			} else {
				p.parsedIncludes[include] = nil
			}

			parser, err := NewParser(includePath,
				WithSameOptions(p),
				withParsedIncludes(p.parsedIncludes),
				withConfigRoot(p.configRoot),
			)

			if err != nil {
				if p.opts.skipIncludeParsingErr {
					continue
				}
				return nil, err
			}

			config, err := parser.Parse()
			if err != nil {
				return nil, err
			}
			//TODO: link parent config or include direcitve?

			p.parsedIncludes[include] = config
			include.Configs = append(include.Configs, config)
		}
	}
	return include, nil
}

// Close closes the file handler and releases the resources
func (p *Parser) Close() (err error) {
	if p.file != nil {
		err = p.file.Close()
	}
	return err
}

// pushContext pushes a new context to the context stack
func (p *Parser) pushContext(context string) {
	p.contextStack = append(p.contextStack, context)
}

// popContext pops the top context from the context stack
func (p *Parser) popContext() {
	if len(p.contextStack) > 0 {
		p.contextStack = p.contextStack[:len(p.contextStack)-1]
	}
}

// getCurrentContext returns the current parsing context
func (p *Parser) getCurrentContext() string {
	if len(p.contextStack) > 0 {
		return p.contextStack[len(p.contextStack)-1]
	}
	return ""
}

// getContextAwareWrapperKey returns the appropriate wrapper key based on current context
func (p *Parser) getContextAwareWrapperKey(directiveName string) string {
	contextStack := p.contextStack

	// For directives that behave differently in different contexts
	switch directiveName {
	case "upstream":
		// Check if we're in stream context (either current or previous in stack)
		inStream := false
		for _, ctx := range contextStack {
			if ctx == "stream" {
				inStream = true
				break
			}
		}

		if inStream {
			return "stream_upstream"
		}
		return "upstream"
	case "server":
		// Check if we're in stream context
		inStream := false
		inUpstream := false
		for _, ctx := range contextStack {
			if ctx == "stream" {
				inStream = true
			}
			if ctx == "upstream" {
				inUpstream = true
			}
		}

		if inStream && inUpstream {
			return "stream_upstream_server"
		} else if inStream {
			return "stream_server"
		}
		return "server"
	default:
		return directiveName
	}
}
