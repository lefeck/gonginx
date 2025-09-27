package config

// IBlock represents any directive block
// IBlock 用于表示 Nginx 配置文件中的块，它包含指令和子块的内容。
type IBlock interface {
	GetDirectives() []IDirective
	FindDirectives(directiveName string) []IDirective
	GetCodeBlock() string
	SetParent(IDirective)
	GetParent() IDirective
}

// IDirective represents any directive
// IDirective 用于表示 Nginx 配置文件中的指令，它定义了指令的名称、参数、块内容以及注释等信息。
type IDirective interface {
	GetName() string //the directive name.
	GetParameters() []Parameter
	GetBlock() IBlock
	GetComment() []string
	SetComment(comment []string)
	SetParent(IDirective)
	GetParent() IDirective
	GetLine() int
	SetLine(int)
	InlineCommenter
}

// InlineCommenter represents the inline comment holder
type InlineCommenter interface {
	GetInlineComment() []InlineComment
	SetInlineComment(comment InlineComment)
}

// DefaultInlineComment represents the default inline comment holder
type DefaultInlineComment struct {
	InlineComment []InlineComment
}

// GetInlineComment returns the inline comment
func (d *DefaultInlineComment) GetInlineComment() []InlineComment {
	return d.InlineComment
}

// SetInlineComment sets the inline comment
func (d *DefaultInlineComment) SetInlineComment(comment InlineComment) {
	d.InlineComment = append(d.InlineComment, comment)
}

// FileDirective a statement that saves its own file
type FileDirective interface {
	isFileDirective()
}

// IncludeDirective represents include statement in nginx
type IncludeDirective interface {
	FileDirective
}

// ParameterType represents the type of a parameter
type ParameterType int

const (
	// ParameterTypeString represents a regular string parameter
	ParameterTypeString ParameterType = iota
	// ParameterTypeVariable represents a variable parameter (starts with $)
	ParameterTypeVariable
	// ParameterTypeNumber represents a numeric parameter
	ParameterTypeNumber
	// ParameterTypeSize represents a size parameter (e.g., 1M, 512k)
	ParameterTypeSize
	// ParameterTypeTime represents a time parameter (e.g., 30s, 5m)
	ParameterTypeTime
	// ParameterTypePath represents a file/directory path
	ParameterTypePath
	// ParameterTypeURL represents a URL
	ParameterTypeURL
	// ParameterTypeRegex represents a regular expression
	ParameterTypeRegex
	// ParameterTypeBoolean represents a boolean value (on/off, yes/no, true/false)
	ParameterTypeBoolean
	// ParameterTypeQuoted represents a quoted string
	ParameterTypeQuoted
)

// String returns the string representation of the parameter type
func (pt ParameterType) String() string {
	switch pt {
	case ParameterTypeString:
		return "string"
	case ParameterTypeVariable:
		return "variable"
	case ParameterTypeNumber:
		return "number"
	case ParameterTypeSize:
		return "size"
	case ParameterTypeTime:
		return "time"
	case ParameterTypePath:
		return "path"
	case ParameterTypeURL:
		return "url"
	case ParameterTypeRegex:
		return "regex"
	case ParameterTypeBoolean:
		return "boolean"
	case ParameterTypeQuoted:
		return "quoted"
	default:
		return "unknown"
	}
}

// Parameter represents a parameter in a directive
type Parameter struct {
	Value             string
	Type              ParameterType // parameter type
	RelativeLineIndex int           // relative line index to the directive
}

// String returns the value of the parameter
func (p *Parameter) String() string {
	return p.Value
}

// SetValue sets the value of the parameter
func (p *Parameter) SetValue(v string) {
	p.Value = v
}

// GetValue returns the value of the parameter
func (p *Parameter) GetValue() string {
	return p.Value
}

// SetRelativeLineIndex sets the relative line index of the parameter
func (p *Parameter) SetRelativeLineIndex(i int) {
	p.RelativeLineIndex = i
}

// GetRelativeLineIndex returns the relative line index of the parameter
func (p *Parameter) GetRelativeLineIndex() int {
	return p.RelativeLineIndex
}

// GetType returns the type of the parameter
func (p *Parameter) GetType() ParameterType {
	return p.Type
}

// SetType sets the type of the parameter
func (p *Parameter) SetType(t ParameterType) {
	p.Type = t
}

// IsVariable returns true if the parameter is a variable
func (p *Parameter) IsVariable() bool {
	return p.Type == ParameterTypeVariable
}

// IsNumber returns true if the parameter is a number
func (p *Parameter) IsNumber() bool {
	return p.Type == ParameterTypeNumber
}

// IsSize returns true if the parameter is a size value
func (p *Parameter) IsSize() bool {
	return p.Type == ParameterTypeSize
}

// IsTime returns true if the parameter is a time value
func (p *Parameter) IsTime() bool {
	return p.Type == ParameterTypeTime
}

// IsPath returns true if the parameter is a path
func (p *Parameter) IsPath() bool {
	return p.Type == ParameterTypePath
}

// IsURL returns true if the parameter is a URL
func (p *Parameter) IsURL() bool {
	return p.Type == ParameterTypeURL
}

// IsRegex returns true if the parameter is a regex
func (p *Parameter) IsRegex() bool {
	return p.Type == ParameterTypeRegex
}

// IsBoolean returns true if the parameter is a boolean
func (p *Parameter) IsBoolean() bool {
	return p.Type == ParameterTypeBoolean
}

// IsQuoted returns true if the parameter is a quoted string
func (p *Parameter) IsQuoted() bool {
	return p.Type == ParameterTypeQuoted
}

// InlineComment represents an inline comment
type InlineComment Parameter
