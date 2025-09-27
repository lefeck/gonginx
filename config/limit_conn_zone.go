package config

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// LimitConnZone represents a limit_conn_zone directive in nginx configuration
// limit_conn_zone $variable zone=name:size [sync];
type LimitConnZone struct {
	Key      string // Key variable (e.g., $binary_remote_addr)
	ZoneName string // Zone name (e.g., "addr")
	ZoneSize string // Zone size (e.g., "10m")
	Sync     bool   // Whether sync is enabled
	Comment  []string
	DefaultInlineComment
	Parent IDirective
	Line   int
}

// SetLine sets the line number
func (lcz *LimitConnZone) SetLine(line int) {
	lcz.Line = line
}

// GetLine returns the line number
func (lcz *LimitConnZone) GetLine() int {
	return lcz.Line
}

// SetParent sets the parent directive
func (lcz *LimitConnZone) SetParent(parent IDirective) {
	lcz.Parent = parent
}

// GetParent returns the parent directive
func (lcz *LimitConnZone) GetParent() IDirective {
	return lcz.Parent
}

// SetComment sets the directive comment
func (lcz *LimitConnZone) SetComment(comment []string) {
	lcz.Comment = comment
}

// GetName implements the IDirective interface
func (lcz *LimitConnZone) GetName() string {
	return "limit_conn_zone"
}

// GetParameters returns the limit_conn_zone parameters
func (lcz *LimitConnZone) GetParameters() []Parameter {
	params := []Parameter{
		{Value: lcz.Key},
		{Value: fmt.Sprintf("zone=%s:%s", lcz.ZoneName, lcz.ZoneSize)},
	}

	if lcz.Sync {
		params = append(params, Parameter{Value: "sync"})
	}

	return params
}

// GetBlock returns nil as limit_conn_zone doesn't have a block
func (lcz *LimitConnZone) GetBlock() IBlock {
	return nil
}

// GetComment returns the directive comment
func (lcz *LimitConnZone) GetComment() []string {
	return lcz.Comment
}

// GetZoneSizeBytes returns the zone size in bytes
func (lcz *LimitConnZone) GetZoneSizeBytes() (int64, error) {
	return parseSizeToBytes(lcz.ZoneSize)
}

// SetZoneSize sets the zone size
func (lcz *LimitConnZone) SetZoneSize(size string) error {
	if err := validateSize(size); err != nil {
		return err
	}
	lcz.ZoneSize = size
	return nil
}

// EstimateMaxConnections estimates the maximum number of connections that can be tracked
// Each connection entry takes approximately 64 bytes
func (lcz *LimitConnZone) EstimateMaxConnections() (int64, error) {
	sizeBytes, err := lcz.GetZoneSizeBytes()
	if err != nil {
		return 0, err
	}

	// Each connection entry takes approximately 64 bytes
	const bytesPerConnection = 64
	return sizeBytes / bytesPerConnection, nil
}

// validateConnZoneName validates the zone name for connection limiting
func validateConnZoneName(name string) error {
	if name == "" {
		return errors.New("zone name cannot be empty")
	}

	// Zone name should be alphanumeric with underscores and hyphens
	re := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !re.MatchString(name) {
		return fmt.Errorf("invalid zone name: %s (only alphanumeric, underscore, and hyphen allowed)", name)
	}

	return nil
}

// NewLimitConnZone creates a new LimitConnZone from a directive
func NewLimitConnZone(directive IDirective) (*LimitConnZone, error) {
	parameters := directive.GetParameters()
	if len(parameters) < 2 {
		return nil, errors.New("limit_conn_zone directive requires at least 2 parameters: key and zone")
	}

	limitConnZone := &LimitConnZone{
		Comment:              directive.GetComment(),
		DefaultInlineComment: DefaultInlineComment{InlineComment: directive.GetInlineComment()},
	}

	// Parse parameters
	for i, param := range parameters {
		value := param.GetValue()

		switch i {
		case 0:
			// First parameter is the key variable
			limitConnZone.Key = value
		default:
			// Remaining parameters are name=value pairs or flags
			if strings.HasPrefix(value, "zone=") {
				zoneSpec := strings.TrimPrefix(value, "zone=")
				parts := strings.Split(zoneSpec, ":")
				if len(parts) != 2 {
					return nil, fmt.Errorf("invalid zone specification: %s (expected format: zone=name:size)", value)
				}

				zoneName := parts[0]
				zoneSize := parts[1]

				if err := validateConnZoneName(zoneName); err != nil {
					return nil, err
				}

				if err := validateSize(zoneSize); err != nil {
					return nil, err
				}

				limitConnZone.ZoneName = zoneName
				limitConnZone.ZoneSize = zoneSize

			} else if value == "sync" {
				limitConnZone.Sync = true

			} else {
				return nil, fmt.Errorf("unknown parameter: %s", value)
			}
		}
	}

	// Validate required parameters
	if limitConnZone.Key == "" {
		return nil, errors.New("key parameter is required")
	}

	if limitConnZone.ZoneName == "" {
		return nil, errors.New("zone name is required")
	}

	if limitConnZone.ZoneSize == "" {
		return nil, errors.New("zone size is required")
	}

	return limitConnZone, nil
}

// SetSync enables or disables sync for the zone
func (lcz *LimitConnZone) SetSync(enable bool) {
	lcz.Sync = enable
}

// IsCompatibleWith checks if this zone is compatible with another zone for merging
func (lcz *LimitConnZone) IsCompatibleWith(other *LimitConnZone) bool {
	if other == nil {
		return false
	}

	// Zones are compatible if they use the same key variable
	return lcz.Key == other.Key
}

// GetMemoryUsageEstimate returns an estimate of memory usage per connection
func (lcz *LimitConnZone) GetMemoryUsageEstimate() string {
	return "~64 bytes per connection"
}

// GetRecommendedLimits returns recommended connection limits based on zone size
func (lcz *LimitConnZone) GetRecommendedLimits() (map[string]int, error) {
	maxConnections, err := lcz.EstimateMaxConnections()
	if err != nil {
		return nil, err
	}

	recommendations := map[string]int{
		"conservative": int(maxConnections * 70 / 100), // 70% of max
		"moderate":     int(maxConnections * 85 / 100), // 85% of max
		"aggressive":   int(maxConnections * 95 / 100), // 95% of max
	}

	return recommendations, nil
}
