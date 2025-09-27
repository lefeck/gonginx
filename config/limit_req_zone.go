package config

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// LimitReqZone represents a limit_req_zone directive in nginx configuration
// limit_req_zone $variable zone=name:size rate=rate [sync];
type LimitReqZone struct {
	Key      string // Key variable (e.g., $binary_remote_addr)
	ZoneName string // Zone name (e.g., "one")
	ZoneSize string // Zone size (e.g., "10m")
	Rate     string // Rate limit (e.g., "1r/s")
	Sync     bool   // Whether sync is enabled
	Comment  []string
	DefaultInlineComment
	Parent IDirective
	Line   int
}

// SetLine sets the line number
func (lrz *LimitReqZone) SetLine(line int) {
	lrz.Line = line
}

// GetLine returns the line number
func (lrz *LimitReqZone) GetLine() int {
	return lrz.Line
}

// SetParent sets the parent directive
func (lrz *LimitReqZone) SetParent(parent IDirective) {
	lrz.Parent = parent
}

// GetParent returns the parent directive
func (lrz *LimitReqZone) GetParent() IDirective {
	return lrz.Parent
}

// SetComment sets the directive comment
func (lrz *LimitReqZone) SetComment(comment []string) {
	lrz.Comment = comment
}

// GetName implements the IDirective interface
func (lrz *LimitReqZone) GetName() string {
	return "limit_req_zone"
}

// GetParameters returns the limit_req_zone parameters
func (lrz *LimitReqZone) GetParameters() []Parameter {
	params := []Parameter{
		{Value: lrz.Key},
		{Value: fmt.Sprintf("zone=%s:%s", lrz.ZoneName, lrz.ZoneSize)},
		{Value: fmt.Sprintf("rate=%s", lrz.Rate)},
	}

	if lrz.Sync {
		params = append(params, Parameter{Value: "sync"})
	}

	return params
}

// GetBlock returns nil as limit_req_zone doesn't have a block
func (lrz *LimitReqZone) GetBlock() IBlock {
	return nil
}

// GetComment returns the directive comment
func (lrz *LimitReqZone) GetComment() []string {
	return lrz.Comment
}

// GetRateNumber returns the numeric part of the rate (e.g., 10 from "10r/s")
func (lrz *LimitReqZone) GetRateNumber() (float64, error) {
	re := regexp.MustCompile(`^(\d+(?:\.\d+)?)r/[sm]$`)
	matches := re.FindStringSubmatch(lrz.Rate)
	if len(matches) < 2 {
		return 0, fmt.Errorf("invalid rate format: %s", lrz.Rate)
	}

	return strconv.ParseFloat(matches[1], 64)
}

// GetRateUnit returns the unit part of the rate (e.g., "s" from "10r/s")
func (lrz *LimitReqZone) GetRateUnit() (string, error) {
	re := regexp.MustCompile(`^(\d+(?:\.\d+)?)r/([sm])$`)
	matches := re.FindStringSubmatch(lrz.Rate)
	if len(matches) < 3 {
		return "", fmt.Errorf("invalid rate format: %s", lrz.Rate)
	}

	return matches[2], nil
}

// GetZoneSizeBytes returns the zone size in bytes
func (lrz *LimitReqZone) GetZoneSizeBytes() (int64, error) {
	return parseSizeToBytes(lrz.ZoneSize)
}

// SetRate sets the rate limit
func (lrz *LimitReqZone) SetRate(rate string) error {
	if err := validateRate(rate); err != nil {
		return err
	}
	lrz.Rate = rate
	return nil
}

// SetZoneSize sets the zone size
func (lrz *LimitReqZone) SetZoneSize(size string) error {
	if err := validateSize(size); err != nil {
		return err
	}
	lrz.ZoneSize = size
	return nil
}

// parseSizeToBytes converts size string to bytes (e.g., "10m" -> 10485760)
func parseSizeToBytes(size string) (int64, error) {
	if size == "" {
		return 0, errors.New("empty size")
	}

	size = strings.ToLower(size)
	unit := size[len(size)-1:]

	var multiplier int64 = 1
	var numStr string

	switch unit {
	case "k":
		multiplier = 1024
		numStr = size[:len(size)-1]
	case "m":
		multiplier = 1024 * 1024
		numStr = size[:len(size)-1]
	case "g":
		multiplier = 1024 * 1024 * 1024
		numStr = size[:len(size)-1]
	default:
		// No unit, assume bytes
		numStr = size
	}

	num, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size format: %s", size)
	}

	return num * multiplier, nil
}

// validateRate validates the rate format
func validateRate(rate string) error {
	re := regexp.MustCompile(`^(\d+(?:\.\d+)?)r/([sm])$`)
	if !re.MatchString(rate) {
		return fmt.Errorf("invalid rate format: %s (expected format: 10r/s or 5r/m)", rate)
	}
	return nil
}

// validateSize validates the size format
func validateSize(size string) error {
	re := regexp.MustCompile(`^(\d+)([kmg]?)$`)
	if !re.MatchString(strings.ToLower(size)) {
		return fmt.Errorf("invalid size format: %s (expected format: 10m, 1g, 512k, etc.)", size)
	}
	return nil
}

// validateZoneName validates the zone name
func validateZoneName(name string) error {
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

// NewLimitReqZone creates a new LimitReqZone from a directive
func NewLimitReqZone(directive IDirective) (*LimitReqZone, error) {
	parameters := directive.GetParameters()
	if len(parameters) < 3 {
		return nil, errors.New("limit_req_zone directive requires at least 3 parameters: key, zone, and rate")
	}

	limitReqZone := &LimitReqZone{
		Comment:              directive.GetComment(),
		DefaultInlineComment: DefaultInlineComment{InlineComment: directive.GetInlineComment()},
	}

	// Parse parameters
	for i, param := range parameters {
		value := param.GetValue()

		switch i {
		case 0:
			// First parameter is the key variable
			limitReqZone.Key = value
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

				if err := validateZoneName(zoneName); err != nil {
					return nil, err
				}

				if err := validateSize(zoneSize); err != nil {
					return nil, err
				}

				limitReqZone.ZoneName = zoneName
				limitReqZone.ZoneSize = zoneSize

			} else if strings.HasPrefix(value, "rate=") {
				rate := strings.TrimPrefix(value, "rate=")

				if err := validateRate(rate); err != nil {
					return nil, err
				}

				limitReqZone.Rate = rate

			} else if value == "sync" {
				limitReqZone.Sync = true

			} else {
				return nil, fmt.Errorf("unknown parameter: %s", value)
			}
		}
	}

	// Validate required parameters
	if limitReqZone.Key == "" {
		return nil, errors.New("key parameter is required")
	}

	if limitReqZone.ZoneName == "" {
		return nil, errors.New("zone name is required")
	}

	if limitReqZone.ZoneSize == "" {
		return nil, errors.New("zone size is required")
	}

	if limitReqZone.Rate == "" {
		return nil, errors.New("rate parameter is required")
	}

	return limitReqZone, nil
}
