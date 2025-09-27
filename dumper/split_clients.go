package dumper

import (
	"fmt"
	"strings"

	"github.com/lefeck/gonginx/config"
)

// DumpSplitClients converts a SplitClients to a string representation
func DumpSplitClients(sc *config.SplitClients, style *Style) string {
	if sc == nil {
		return ""
	}

	result := ""

	// Add comments before the split_clients block
	if len(sc.GetComment()) > 0 {
		for _, comment := range sc.GetComment() {
			result += fmt.Sprintf("%s%s\n", strings.Repeat(" ", style.StartIndent), comment)
		}
	}

	// Add the split_clients directive with parameters
	result += strings.Repeat(" ", style.StartIndent) + "split_clients " + sc.Variable + " " + sc.MappedVariable

	// Add inline comments
	if len(sc.GetInlineComment()) > 0 {
		for _, inlineComment := range sc.GetInlineComment() {
			result += " " + inlineComment.Value
		}
	}

	if style.SpaceBeforeBlocks {
		result += " "
	}
	result += " {\n"

	// Increase indentation for split_clients entries
	style.StartIndent++

	// Add split_clients entries
	for _, entry := range sc.Entries {
		result += DumpSplitClientsEntry(entry, style)
	}

	// Decrease indentation
	style.StartIndent--

	// Close the split_clients block
	result += strings.Repeat(" ", style.StartIndent) + "}"

	return result
}

// DumpSplitClientsEntry converts a SplitClientsEntry to a string representation
func DumpSplitClientsEntry(entry *config.SplitClientsEntry, style *Style) string {
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

	// Add the percentage entry
	result += strings.Repeat(" ", style.StartIndent) + entry.Percentage + " " + entry.Value + ";"

	// Add inline comments
	if len(entry.GetInlineComment()) > 0 {
		for _, inlineComment := range entry.GetInlineComment() {
			result += " " + inlineComment.Value
		}
	}

	result += "\n"

	return result
}
