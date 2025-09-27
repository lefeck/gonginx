package dumper

import (
	"fmt"
	"strings"

	"github.com/lefeck/gonginx/config"
)

// DumpGeo converts a Geo to a string representation
func DumpGeo(g *config.Geo, style *Style) string {
	if g == nil {
		return ""
	}

	result := ""

	// Add comments before the geo block
	if len(g.GetComment()) > 0 {
		for _, comment := range g.GetComment() {
			result += fmt.Sprintf("%s%s\n", strings.Repeat(" ", style.StartIndent), comment)
		}
	}

	// Add the geo directive with parameters
	result += strings.Repeat(" ", style.StartIndent) + "geo"

	// Add parameters based on configuration
	if g.SourceAddress != "" && g.SourceAddress != "$remote_addr" {
		// geo $source_addr $variable
		result += " " + g.SourceAddress + " " + g.Variable
	} else {
		// geo $variable
		result += " " + g.Variable
	}

	// Add inline comments
	if len(g.GetInlineComment()) > 0 {
		for _, inlineComment := range g.GetInlineComment() {
			result += " " + inlineComment.Value
		}
	}

	if style.SpaceBeforeBlocks {
		result += " "
	}
	result += " {\n"

	// Increase indentation for geo entries
	style.StartIndent++

	// Add special directives first
	if g.Ranges {
		result += strings.Repeat(" ", style.StartIndent) + "ranges;\n"
	}

	if g.ProxyRecursive {
		result += strings.Repeat(" ", style.StartIndent) + "proxy_recursive;\n"
	}

	// Add proxy entries
	for _, proxy := range g.Proxy {
		result += strings.Repeat(" ", style.StartIndent) + "proxy " + proxy + ";\n"
	}

	// Add delete entries
	for _, deleteIP := range g.Delete {
		result += strings.Repeat(" ", style.StartIndent) + "delete " + deleteIP + ";\n"
	}

	// Add default value if set
	if g.DefaultValue != "" {
		result += strings.Repeat(" ", style.StartIndent) + "default " + g.DefaultValue + ";\n"
	}

	// Add regular geo entries
	for _, entry := range g.Entries {
		result += DumpGeoEntry(entry, style)
	}

	// Decrease indentation
	style.StartIndent--

	// Close the geo block
	result += strings.Repeat(" ", style.StartIndent) + "}"

	return result
}

// DumpGeoEntry converts a GeoEntry to a string representation
func DumpGeoEntry(entry *config.GeoEntry, style *Style) string {
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

	// Handle special entries
	switch entry.Network {
	case "default":
		result += strings.Repeat(" ", style.StartIndent) + "default " + entry.Value + ";"
	case "ranges":
		result += strings.Repeat(" ", style.StartIndent) + "ranges;"
	case "proxy_recursive":
		result += strings.Repeat(" ", style.StartIndent) + "proxy_recursive;"
	case "delete":
		result += strings.Repeat(" ", style.StartIndent) + "delete " + entry.Value + ";"
	case "proxy":
		result += strings.Repeat(" ", style.StartIndent) + "proxy " + entry.Value + ";"
	default:
		// Regular network entry
		result += strings.Repeat(" ", style.StartIndent) + entry.Network + " " + entry.Value + ";"
	}

	// Add inline comments
	if len(entry.GetInlineComment()) > 0 {
		for _, inlineComment := range entry.GetInlineComment() {
			result += " " + inlineComment.Value
		}
	}

	result += "\n"

	return result
}
