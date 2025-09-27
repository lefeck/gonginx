package config

import (
	"errors"
	"net"
	"strings"
)

// Geo represents a geo block in nginx configuration
// geo $variable { ... } or geo $remote_addr $variable { ... }
type Geo struct {
	SourceAddress  string // Source address variable (optional, default is $remote_addr)
	Variable       string // Target variable to set
	Entries        []*GeoEntry
	DefaultValue   string   // Default value when no match
	Ranges         bool     // Whether ranges are used instead of CIDR
	Delete         []string // IP addresses to delete from inherited geo
	Proxy          []string // Trusted proxy addresses
	ProxyRecursive bool     // Whether to use recursive proxy lookup
	Comment        []string
	DefaultInlineComment
	Parent IDirective
	Line   int
}

// GeoEntry represents a single entry in a geo block
type GeoEntry struct {
	Network string // IP network (CIDR) or IP address or range
	Value   string // Value to set for this network
	Comment []string
	DefaultInlineComment
	Parent IDirective
	Line   int
}

// SetLine sets the line number
func (g *Geo) SetLine(line int) {
	g.Line = line
}

// GetLine returns the line number
func (g *Geo) GetLine() int {
	return g.Line
}

// SetParent sets the parent directive
func (g *Geo) SetParent(parent IDirective) {
	g.Parent = parent
}

// GetParent returns the parent directive
func (g *Geo) GetParent() IDirective {
	return g.Parent
}

// SetComment sets the directive comment
func (g *Geo) SetComment(comment []string) {
	g.Comment = comment
}

// GetName implements the IDirective interface
func (g *Geo) GetName() string {
	return "geo"
}

// GetParameters returns the geo parameters
func (g *Geo) GetParameters() []Parameter {
	if g.SourceAddress != "" && g.SourceAddress != "$remote_addr" {
		return []Parameter{
			{Value: g.SourceAddress},
			{Value: g.Variable},
		}
	}
	return []Parameter{
		{Value: g.Variable},
	}
}

// GetBlock returns the geo itself, which implements IBlock
func (g *Geo) GetBlock() IBlock {
	return g
}

// GetComment returns the directive comment
func (g *Geo) GetComment() []string {
	return g.Comment
}

// GetDirectives returns the geo entries as directives
func (g *Geo) GetDirectives() []IDirective {
	directives := make([]IDirective, 0)

	// Add special directives first
	if g.DefaultValue != "" {
		defaultEntry := &GeoEntry{
			Network: "default",
			Value:   g.DefaultValue,
			Parent:  g,
		}
		directives = append(directives, defaultEntry)
	}

	if g.Ranges {
		rangesEntry := &GeoEntry{
			Network: "ranges",
			Value:   "",
			Parent:  g,
		}
		directives = append(directives, rangesEntry)
	}

	if g.ProxyRecursive {
		proxyRecursiveEntry := &GeoEntry{
			Network: "proxy_recursive",
			Value:   "",
			Parent:  g,
		}
		directives = append(directives, proxyRecursiveEntry)
	}

	// Add delete entries
	for _, deleteIP := range g.Delete {
		deleteEntry := &GeoEntry{
			Network: "delete",
			Value:   deleteIP,
			Parent:  g,
		}
		directives = append(directives, deleteEntry)
	}

	// Add proxy entries
	for _, proxyIP := range g.Proxy {
		proxyEntry := &GeoEntry{
			Network: "proxy",
			Value:   proxyIP,
			Parent:  g,
		}
		directives = append(directives, proxyEntry)
	}

	// Add regular entries
	for _, entry := range g.Entries {
		directives = append(directives, entry)
	}

	return directives
}

// GetCodeBlock returns empty string (not a literal code block)
func (g *Geo) GetCodeBlock() string {
	return ""
}

// FindDirectives finds directives in the geo block
func (g *Geo) FindDirectives(directiveName string) []IDirective {
	var directives []IDirective
	for _, entry := range g.Entries {
		if entry.GetName() == directiveName {
			directives = append(directives, entry)
		}
	}
	return directives
}

// AddEntry adds a new IP network mapping to the geo block
func (g *Geo) AddEntry(network, value string) error {
	// Validate network format for regular entries
	if network != "default" && network != "ranges" && network != "proxy_recursive" {
		if !strings.HasPrefix(network, "delete ") && !strings.HasPrefix(network, "proxy ") {
			// Check if it's a valid CIDR or IP address or range
			if !g.isValidNetworkFormat(network) {
				return errors.New("invalid network format: " + network)
			}
		}
	}

	entry := &GeoEntry{
		Network: network,
		Value:   value,
		Parent:  g,
	}
	g.Entries = append(g.Entries, entry)
	return nil
}

// isValidNetworkFormat checks if the network format is valid
func (g *Geo) isValidNetworkFormat(network string) bool {
	// Check for IP range format (start-end)
	if strings.Contains(network, "-") {
		parts := strings.Split(network, "-")
		if len(parts) == 2 {
			startIP := net.ParseIP(strings.TrimSpace(parts[0]))
			endIP := net.ParseIP(strings.TrimSpace(parts[1]))
			return startIP != nil && endIP != nil
		}
		return false
	}

	// Check for CIDR format
	if strings.Contains(network, "/") {
		_, _, err := net.ParseCIDR(network)
		return err == nil
	}

	// Check for single IP address
	ip := net.ParseIP(network)
	return ip != nil
}

// SetDefaultValue sets the default value for the geo block
func (g *Geo) SetDefaultValue(value string) {
	g.DefaultValue = value
}

// GetDefaultValue returns the default value for the geo block
func (g *Geo) GetDefaultValue() string {
	return g.DefaultValue
}

// AddProxy adds a trusted proxy address
func (g *Geo) AddProxy(proxyAddr string) {
	g.Proxy = append(g.Proxy, proxyAddr)
}

// AddDelete adds an IP address to delete from inherited geo
func (g *Geo) AddDelete(ipAddr string) {
	g.Delete = append(g.Delete, ipAddr)
}

// SetRanges enables or disables ranges mode
func (g *Geo) SetRanges(enable bool) {
	g.Ranges = enable
}

// SetProxyRecursive enables or disables recursive proxy lookup
func (g *Geo) SetProxyRecursive(enable bool) {
	g.ProxyRecursive = enable
}

// NewGeo creates a new Geo from a directive
func NewGeo(directive IDirective) (*Geo, error) {
	parameters := directive.GetParameters()
	if len(parameters) < 1 {
		return nil, errors.New("geo directive requires at least 1 parameter: target variable")
	}

	var sourceAddress, variable string

	// Determine if source address is specified
	if len(parameters) == 1 {
		// geo $variable { ... }
		sourceAddress = "$remote_addr" // default
		variable = parameters[0].GetValue()
	} else if len(parameters) == 2 {
		// geo $source_addr $variable { ... }
		sourceAddress = parameters[0].GetValue()
		variable = parameters[1].GetValue()
	} else {
		return nil, errors.New("geo directive accepts 1 or 2 parameters only")
	}

	geoBlock := &Geo{
		SourceAddress:        sourceAddress,
		Variable:             variable,
		Entries:              make([]*GeoEntry, 0),
		Comment:              directive.GetComment(),
		DefaultInlineComment: DefaultInlineComment{InlineComment: directive.GetInlineComment()},
	}

	if directive.GetBlock() == nil {
		return nil, errors.New("geo directive must have a block")
	}

	// Parse geo entries from the block
	for _, d := range directive.GetBlock().GetDirectives() {
		entry, err := NewGeoEntry(d, geoBlock)
		if err != nil {
			return nil, err
		}
		entry.SetParent(geoBlock)
		entry.SetLine(d.GetLine())

		// Handle special directives
		switch entry.Network {
		case "default":
			geoBlock.DefaultValue = entry.Value
		case "ranges":
			geoBlock.Ranges = true
		case "proxy_recursive":
			geoBlock.ProxyRecursive = true
		default:
			if strings.HasPrefix(entry.Network, "delete") {
				// delete 127.0.0.1;
				if entry.Value != "" {
					geoBlock.Delete = append(geoBlock.Delete, entry.Value)
				}
			} else if strings.HasPrefix(entry.Network, "proxy") {
				// proxy 192.168.1.0/24;
				if entry.Value != "" {
					geoBlock.Proxy = append(geoBlock.Proxy, entry.Value)
				}
			} else {
				// Regular network entry
				geoBlock.Entries = append(geoBlock.Entries, entry)
			}
		}
	}

	return geoBlock, nil
}

// GeoEntry methods implementing IDirective interface

// SetLine sets the line number for GeoEntry
func (ge *GeoEntry) SetLine(line int) {
	ge.Line = line
}

// GetLine returns the line number for GeoEntry
func (ge *GeoEntry) GetLine() int {
	return ge.Line
}

// SetParent sets the parent directive for GeoEntry
func (ge *GeoEntry) SetParent(parent IDirective) {
	ge.Parent = parent
}

// GetParent returns the parent directive for GeoEntry
func (ge *GeoEntry) GetParent() IDirective {
	return ge.Parent
}

// SetComment sets the comment for GeoEntry
func (ge *GeoEntry) SetComment(comment []string) {
	ge.Comment = comment
}

// GetName returns the network as the "directive name" for GeoEntry
func (ge *GeoEntry) GetName() string {
	return ge.Network
}

// GetParameters returns the value as parameter for GeoEntry
func (ge *GeoEntry) GetParameters() []Parameter {
	if ge.Value == "" {
		return []Parameter{}
	}
	return []Parameter{{Value: ge.Value}}
}

// GetBlock returns nil as GeoEntry doesn't have a block
func (ge *GeoEntry) GetBlock() IBlock {
	return nil
}

// GetComment returns the comment for GeoEntry
func (ge *GeoEntry) GetComment() []string {
	return ge.Comment
}

// NewGeoEntry creates a new GeoEntry from a directive
func NewGeoEntry(directive IDirective, parent *Geo) (*GeoEntry, error) {
	parameters := directive.GetParameters()

	network := directive.GetName()
	var value string

	// Handle different geo entry formats
	if len(parameters) >= 1 {
		value = parameters[0].GetValue()
	}

	// Handle special cases where the directive name contains the value
	if network == "delete" || network == "proxy" {
		if len(parameters) >= 1 {
			value = parameters[0].GetValue()
		}
	}

	entry := &GeoEntry{
		Network:              network,
		Value:                value,
		Comment:              directive.GetComment(),
		DefaultInlineComment: DefaultInlineComment{InlineComment: directive.GetInlineComment()},
	}

	return entry, nil
}
