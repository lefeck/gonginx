package dumper

import (
	"fmt"
	"strings"

	"github.com/lefeck/gonginx/config"
)

// DumpMap converts a Map to a string representation
func DumpMap(m *config.Map, style *Style) string {
	if m == nil {
		return ""
	}

	result := ""

	// Add comments before the map block
	if len(m.GetComment()) > 0 {
		for _, comment := range m.GetComment() {
			result += fmt.Sprintf("%s%s\n", strings.Repeat(" ", style.StartIndent), comment)
		}
	}

	// Add the map directive with parameters
	result += strings.Repeat(" ", style.StartIndent) + "map " + m.Variable + " " + m.MappedVariable

	// Add inline comments
	if len(m.GetInlineComment()) > 0 {
		for _, inlineComment := range m.GetInlineComment() {
			result += " " + inlineComment.Value
		}
	}

	if style.SpaceBeforeBlocks {
		result += " "
	}
	result += " {\n"

	// Increase indentation for map entries
	style.StartIndent++

	// Add map entries
	for _, mapping := range m.Mappings {
		result += DumpMapEntry(mapping, style)
	}

	// Decrease indentation
	style.StartIndent--

	// Close the map block
	result += strings.Repeat(" ", style.StartIndent) + "}"

	return result
}

// DumpMapEntry converts a MapEntry to a string representation
func DumpMapEntry(entry *config.MapEntry, style *Style) string {
	if entry == nil {
		return ""
	}

	result := ""

	// Add comments before the entry
	if len(entry.GetComment()) > 0 {
		for _, comment := range entry.GetComment() {
			result += fmt.Sprintf("%s%s\n", strings.Repeat(" ", style.StartIndent), comment)
		}
	}

	// Add the mapping entry
	result += strings.Repeat(" ", style.StartIndent) + entry.Pattern + " " + entry.Value + ";"

	// Add inline comments
	if len(entry.GetInlineComment()) > 0 {
		for _, inlineComment := range entry.GetInlineComment() {
			result += " " + inlineComment.Value
		}
	}

	result += "\n"

	return result
}
