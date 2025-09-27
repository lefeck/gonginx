package dumper

import (
	"fmt"
	"strings"

	"github.com/lefeck/gonginx/config"
)

// DumpStream dumps a stream block
func DumpStream(stream *config.Stream, style *Style) string {
	var result strings.Builder

	// Add indentation
	indent := strings.Repeat(" ", style.StartIndent)
	result.WriteString(indent)

	// Write stream directive
	result.WriteString("stream")

	// Add comment if present
	if len(stream.GetComment()) > 0 {
		result.WriteString(" #")
		result.WriteString(stream.GetComment()[0])
	}

	// Add opening brace
	result.WriteString(" {")
	result.WriteString("\n")

	// Dump block contents
	if stream.GetBlock() != nil {
		blockStyle := style.Iterate()
		result.WriteString(DumpBlock(stream.GetBlock(), blockStyle))
	}

	// Add closing brace
	result.WriteString(indent)
	result.WriteString("}")
	result.WriteString("\n")

	return result.String()
}

// DumpStreamUpstream dumps a stream upstream block
func DumpStreamUpstream(upstream *config.StreamUpstream, style *Style) string {
	var result strings.Builder

	// Add indentation
	indent := strings.Repeat(" ", style.StartIndent)
	result.WriteString(indent)

	// Write upstream directive with name
	result.WriteString("upstream")
	if upstream.UpstreamName != "" {
		result.WriteString(" ")
		result.WriteString(upstream.UpstreamName)
	}

	// Add comment if present
	if len(upstream.GetComment()) > 0 {
		result.WriteString(" #")
		result.WriteString(upstream.GetComment()[0])
	}

	// Add opening brace
	result.WriteString(" {")
	result.WriteString("\n")

	// Dump block contents
	if upstream.GetBlock() != nil {
		blockStyle := style.Iterate()
		result.WriteString(DumpBlock(upstream.GetBlock(), blockStyle))
	}

	// Add closing brace
	result.WriteString(indent)
	result.WriteString("}")
	result.WriteString("\n")

	return result.String()
}

// DumpStreamServer dumps a stream server block
func DumpStreamServer(server *config.StreamServer, style *Style) string {
	var result strings.Builder

	// Add indentation
	indent := strings.Repeat(" ", style.StartIndent)
	result.WriteString(indent)

	// Write server directive
	result.WriteString("server")

	// Add comment if present
	if len(server.GetComment()) > 0 {
		result.WriteString(" #")
		result.WriteString(server.GetComment()[0])
	}

	// Add opening brace
	result.WriteString(" {")
	result.WriteString("\n")

	// Dump block contents
	if server.GetBlock() != nil {
		blockStyle := style.Iterate()
		result.WriteString(DumpBlock(server.GetBlock(), blockStyle))
	}

	// Add closing brace
	result.WriteString(indent)
	result.WriteString("}")
	result.WriteString("\n")

	return result.String()
}

// DumpStreamUpstreamServer dumps a stream upstream server directive
func DumpStreamUpstreamServer(server *config.StreamUpstreamServer, style *Style) string {
	var result strings.Builder

	// Add indentation
	indent := strings.Repeat(" ", style.StartIndent)
	result.WriteString(indent)

	// Write server directive with address
	result.WriteString("server")
	if server.Address != "" {
		result.WriteString(" ")
		result.WriteString(server.Address)
	}

	// Add parameters
	for key, value := range server.Parameters {
		result.WriteString(" ")
		if value == "true" {
			// Boolean parameter
			result.WriteString(key)
		} else {
			// Key=value parameter
			result.WriteString(fmt.Sprintf("%s=%s", key, value))
		}
	}

	// Add comment if present
	if len(server.GetComment()) > 0 {
		result.WriteString(" #")
		result.WriteString(server.GetComment()[0])
	}

	// Add semicolon
	result.WriteString(";")
	result.WriteString("\n")

	return result.String()
}
