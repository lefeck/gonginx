package config

import (
	"errors"
)

// Map represents a map block in nginx configuration
// map $variable $mapped_variable { ... }
type Map struct {
	Variable       string // Source variable (e.g., $http_host)
	MappedVariable string // Target variable (e.g., $backend)
	Mappings       []*MapEntry
	Comment        []string
	DefaultInlineComment
	Parent IDirective
	Line   int
}

// MapEntry represents a single mapping entry in a map block
type MapEntry struct {
	Pattern string // Pattern to match (can be literal, regex, or "default")
	Value   string // Value to map to
	Comment []string
	DefaultInlineComment
	Parent IDirective
	Line   int
}

// SetLine sets the line number
func (m *Map) SetLine(line int) {
	m.Line = line
}

// GetLine returns the line number
func (m *Map) GetLine() int {
	return m.Line
}

// SetParent sets the parent directive
func (m *Map) SetParent(parent IDirective) {
	m.Parent = parent
}

// GetParent returns the parent directive
func (m *Map) GetParent() IDirective {
	return m.Parent
}

// SetComment sets the directive comment
func (m *Map) SetComment(comment []string) {
	m.Comment = comment
}

// GetName implements the IDirective interface
func (m *Map) GetName() string {
	return "map"
}

// GetParameters returns the map parameters (source and target variables)
func (m *Map) GetParameters() []Parameter {
	return []Parameter{
		{Value: m.Variable},
		{Value: m.MappedVariable},
	}
}

// GetBlock returns the map itself, which implements IBlock
func (m *Map) GetBlock() IBlock {
	return m
}

// GetComment returns the directive comment
func (m *Map) GetComment() []string {
	return m.Comment
}

// GetDirectives returns the map entries as directives
func (m *Map) GetDirectives() []IDirective {
	directives := make([]IDirective, len(m.Mappings))
	for i, mapping := range m.Mappings {
		directives[i] = mapping
	}
	return directives
}

// GetCodeBlock returns empty string (not a literal code block)
func (m *Map) GetCodeBlock() string {
	return ""
}

// FindDirectives finds directives in the map block
func (m *Map) FindDirectives(directiveName string) []IDirective {
	// Map blocks typically only contain mapping entries
	// but we implement this for consistency
	var directives []IDirective
	for _, mapping := range m.Mappings {
		if mapping.GetName() == directiveName {
			directives = append(directives, mapping)
		}
	}
	return directives
}

// AddMapping adds a new mapping to the map block
func (m *Map) AddMapping(pattern, value string) {
	mapping := &MapEntry{
		Pattern: pattern,
		Value:   value,
		Parent:  m,
	}
	m.Mappings = append(m.Mappings, mapping)
}

// GetDefaultValue returns the default value for the map
func (m *Map) GetDefaultValue() string {
	for _, mapping := range m.Mappings {
		if mapping.Pattern == "default" {
			return mapping.Value
		}
	}
	return ""
}

// SetDefaultValue sets or updates the default value for the map
func (m *Map) SetDefaultValue(value string) {
	// Check if default already exists
	for _, mapping := range m.Mappings {
		if mapping.Pattern == "default" {
			mapping.Value = value
			return
		}
	}
	// Add new default mapping
	m.AddMapping("default", value)
}

// NewMap creates a new Map from a directive
func NewMap(directive IDirective) (*Map, error) {
	parameters := directive.GetParameters()
	if len(parameters) < 2 {
		return nil, errors.New("map directive requires at least 2 parameters: source and target variables")
	}

	mapBlock := &Map{
		Variable:             parameters[0].GetValue(),
		MappedVariable:       parameters[1].GetValue(),
		Mappings:             make([]*MapEntry, 0),
		Comment:              directive.GetComment(),
		DefaultInlineComment: DefaultInlineComment{InlineComment: directive.GetInlineComment()},
	}

	if directive.GetBlock() == nil {
		return nil, errors.New("map directive must have a block")
	}

	// Parse map entries from the block
	for _, d := range directive.GetBlock().GetDirectives() {
		entry, err := NewMapEntry(d)
		if err != nil {
			return nil, err
		}
		entry.SetParent(mapBlock)
		entry.SetLine(d.GetLine())
		mapBlock.Mappings = append(mapBlock.Mappings, entry)
	}

	return mapBlock, nil
}

// MapEntry methods implementing IDirective interface

// SetLine sets the line number for MapEntry
func (me *MapEntry) SetLine(line int) {
	me.Line = line
}

// GetLine returns the line number for MapEntry
func (me *MapEntry) GetLine() int {
	return me.Line
}

// SetParent sets the parent directive for MapEntry
func (me *MapEntry) SetParent(parent IDirective) {
	me.Parent = parent
}

// GetParent returns the parent directive for MapEntry
func (me *MapEntry) GetParent() IDirective {
	return me.Parent
}

// SetComment sets the comment for MapEntry
func (me *MapEntry) SetComment(comment []string) {
	me.Comment = comment
}

// GetName returns the pattern as the "directive name" for MapEntry
func (me *MapEntry) GetName() string {
	return me.Pattern
}

// GetParameters returns the value as parameter for MapEntry
func (me *MapEntry) GetParameters() []Parameter {
	return []Parameter{{Value: me.Value}}
}

// GetBlock returns nil as MapEntry doesn't have a block
func (me *MapEntry) GetBlock() IBlock {
	return nil
}

// GetComment returns the comment for MapEntry
func (me *MapEntry) GetComment() []string {
	return me.Comment
}

// NewMapEntry creates a new MapEntry from a directive
func NewMapEntry(directive IDirective) (*MapEntry, error) {
	parameters := directive.GetParameters()
	if len(parameters) < 1 {
		return nil, errors.New("map entry must have at least one parameter (value)")
	}

	// The pattern is the directive name, value is the first parameter
	entry := &MapEntry{
		Pattern:              directive.GetName(),
		Value:                parameters[0].GetValue(),
		Comment:              directive.GetComment(),
		DefaultInlineComment: DefaultInlineComment{InlineComment: directive.GetInlineComment()},
	}

	return entry, nil
}
