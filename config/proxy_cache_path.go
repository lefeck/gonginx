package config

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ProxyCachePath represents a proxy_cache_path directive in nginx configuration
// proxy_cache_path path [levels=levels] [use_temp_path=on|off] keys_zone=name:size
// [inactive=time] [max_size=size] [min_free=size] [manager_files=number]
// [manager_sleep=time] [manager_threshold=time] [loader_files=number]
// [loader_sleep=time] [loader_threshold=time] [purger=on|off]
// [purger_files=number] [purger_sleep=time] [purger_threshold=time];
type ProxyCachePath struct {
	Path             string // Cache directory path
	Levels           string // Directory hierarchy levels (e.g., "1:2")
	KeysZoneName     string // Shared memory zone name
	KeysZoneSize     string // Shared memory zone size
	UseTemPath       *bool  // Whether to use temp path (nil = not set)
	Inactive         string // Inactive time (e.g., "60m")
	MaxSize          string // Maximum cache size (e.g., "10g")
	MinFree          string // Minimum free space
	ManagerFiles     *int   // Manager files parameter
	ManagerSleep     string // Manager sleep time
	ManagerThreshold string // Manager threshold time
	LoaderFiles      *int   // Loader files parameter
	LoaderSleep      string // Loader sleep time
	LoaderThreshold  string // Loader threshold time
	Purger           *bool  // Whether purger is enabled
	PurgerFiles      *int   // Purger files parameter
	PurgerSleep      string // Purger sleep time
	PurgerThreshold  string // Purger threshold time
	Comment          []string
	DefaultInlineComment
	Parent IDirective
	Line   int
}

// SetLine sets the line number
func (pcp *ProxyCachePath) SetLine(line int) {
	pcp.Line = line
}

// GetLine returns the line number
func (pcp *ProxyCachePath) GetLine() int {
	return pcp.Line
}

// SetParent sets the parent directive
func (pcp *ProxyCachePath) SetParent(parent IDirective) {
	pcp.Parent = parent
}

// GetParent returns the parent directive
func (pcp *ProxyCachePath) GetParent() IDirective {
	return pcp.Parent
}

// SetComment sets the directive comment
func (pcp *ProxyCachePath) SetComment(comment []string) {
	pcp.Comment = comment
}

// GetName implements the IDirective interface
func (pcp *ProxyCachePath) GetName() string {
	return "proxy_cache_path"
}

// GetParameters returns the proxy_cache_path parameters
func (pcp *ProxyCachePath) GetParameters() []Parameter {
	params := []Parameter{
		{Value: pcp.Path},
	}

	if pcp.Levels != "" {
		params = append(params, Parameter{Value: fmt.Sprintf("levels=%s", pcp.Levels)})
	}

	if pcp.UseTemPath != nil {
		value := "off"
		if *pcp.UseTemPath {
			value = "on"
		}
		params = append(params, Parameter{Value: fmt.Sprintf("use_temp_path=%s", value)})
	}

	// keys_zone is required
	params = append(params, Parameter{Value: fmt.Sprintf("keys_zone=%s:%s", pcp.KeysZoneName, pcp.KeysZoneSize)})

	if pcp.Inactive != "" {
		params = append(params, Parameter{Value: fmt.Sprintf("inactive=%s", pcp.Inactive)})
	}

	if pcp.MaxSize != "" {
		params = append(params, Parameter{Value: fmt.Sprintf("max_size=%s", pcp.MaxSize)})
	}

	if pcp.MinFree != "" {
		params = append(params, Parameter{Value: fmt.Sprintf("min_free=%s", pcp.MinFree)})
	}

	if pcp.ManagerFiles != nil {
		params = append(params, Parameter{Value: fmt.Sprintf("manager_files=%d", *pcp.ManagerFiles)})
	}

	if pcp.ManagerSleep != "" {
		params = append(params, Parameter{Value: fmt.Sprintf("manager_sleep=%s", pcp.ManagerSleep)})
	}

	if pcp.ManagerThreshold != "" {
		params = append(params, Parameter{Value: fmt.Sprintf("manager_threshold=%s", pcp.ManagerThreshold)})
	}

	if pcp.LoaderFiles != nil {
		params = append(params, Parameter{Value: fmt.Sprintf("loader_files=%d", *pcp.LoaderFiles)})
	}

	if pcp.LoaderSleep != "" {
		params = append(params, Parameter{Value: fmt.Sprintf("loader_sleep=%s", pcp.LoaderSleep)})
	}

	if pcp.LoaderThreshold != "" {
		params = append(params, Parameter{Value: fmt.Sprintf("loader_threshold=%s", pcp.LoaderThreshold)})
	}

	if pcp.Purger != nil {
		value := "off"
		if *pcp.Purger {
			value = "on"
		}
		params = append(params, Parameter{Value: fmt.Sprintf("purger=%s", value)})
	}

	if pcp.PurgerFiles != nil {
		params = append(params, Parameter{Value: fmt.Sprintf("purger_files=%d", *pcp.PurgerFiles)})
	}

	if pcp.PurgerSleep != "" {
		params = append(params, Parameter{Value: fmt.Sprintf("purger_sleep=%s", pcp.PurgerSleep)})
	}

	if pcp.PurgerThreshold != "" {
		params = append(params, Parameter{Value: fmt.Sprintf("purger_threshold=%s", pcp.PurgerThreshold)})
	}

	return params
}

// GetBlock returns nil as proxy_cache_path doesn't have a block
func (pcp *ProxyCachePath) GetBlock() IBlock {
	return nil
}

// GetComment returns the directive comment
func (pcp *ProxyCachePath) GetComment() []string {
	return pcp.Comment
}

// GetMaxSizeBytes returns the max_size in bytes
func (pcp *ProxyCachePath) GetMaxSizeBytes() (int64, error) {
	if pcp.MaxSize == "" {
		return 0, nil
	}
	return parseSizeToBytes(pcp.MaxSize)
}

// GetMinFreeBytes returns the min_free in bytes
func (pcp *ProxyCachePath) GetMinFreeBytes() (int64, error) {
	if pcp.MinFree == "" {
		return 0, nil
	}
	return parseSizeToBytes(pcp.MinFree)
}

// GetKeysZoneSizeBytes returns the keys_zone size in bytes
func (pcp *ProxyCachePath) GetKeysZoneSizeBytes() (int64, error) {
	return parseSizeToBytes(pcp.KeysZoneSize)
}

// GetInactiveDuration returns the inactive time as Duration
func (pcp *ProxyCachePath) GetInactiveDuration() (time.Duration, error) {
	if pcp.Inactive == "" {
		return 0, nil
	}
	return parseTimeDuration(pcp.Inactive)
}

// SetMaxSize sets the max_size parameter
func (pcp *ProxyCachePath) SetMaxSize(size string) error {
	if size != "" {
		if err := validateSize(size); err != nil {
			return err
		}
	}
	pcp.MaxSize = size
	return nil
}

// SetInactive sets the inactive time parameter
func (pcp *ProxyCachePath) SetInactive(inactive string) error {
	if inactive != "" {
		if err := validateTime(inactive); err != nil {
			return err
		}
	}
	pcp.Inactive = inactive
	return nil
}

// SetKeysZoneSize sets the keys_zone size
func (pcp *ProxyCachePath) SetKeysZoneSize(size string) error {
	if err := validateSize(size); err != nil {
		return err
	}
	pcp.KeysZoneSize = size
	return nil
}

// SetUseTemPath sets the use_temp_path parameter
func (pcp *ProxyCachePath) SetUseTemPath(enabled bool) {
	pcp.UseTemPath = &enabled
}

// SetPurger sets the purger parameter
func (pcp *ProxyCachePath) SetPurger(enabled bool) {
	pcp.Purger = &enabled
}

// GetLevelsDepth returns the depth of directory levels
func (pcp *ProxyCachePath) GetLevelsDepth() ([]int, error) {
	if pcp.Levels == "" {
		return nil, nil
	}

	parts := strings.Split(pcp.Levels, ":")
	depths := make([]int, len(parts))

	for i, part := range parts {
		depth, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid levels format: %s", pcp.Levels)
		}
		if depth < 1 || depth > 2 {
			return nil, fmt.Errorf("level depth must be 1 or 2, got %d", depth)
		}
		depths[i] = depth
	}

	return depths, nil
}

// EstimateKeyCapacity estimates the number of cache keys that can be stored
func (pcp *ProxyCachePath) EstimateKeyCapacity() (int64, error) {
	sizeBytes, err := pcp.GetKeysZoneSizeBytes()
	if err != nil {
		return 0, err
	}

	// Each cache key takes approximately 256 bytes in memory
	const bytesPerKey = 256
	return sizeBytes / bytesPerKey, nil
}

// validateTime validates nginx time format (e.g., 60m, 1h, 30s, 50ms)
func validateTime(timeStr string) error {
	re := regexp.MustCompile(`^(\d+)(ms|[smhd])$`)
	if !re.MatchString(timeStr) {
		return fmt.Errorf("invalid time format: %s (expected format: 60m, 1h, 30s, 50ms, 7d)", timeStr)
	}
	return nil
}

// parseTimeDuration parses nginx time format to Go Duration
func parseTimeDuration(timeStr string) (time.Duration, error) {
	if timeStr == "" {
		return 0, nil
	}

	re := regexp.MustCompile(`^(\d+)(ms|[smhd])$`)
	matches := re.FindStringSubmatch(timeStr)
	if len(matches) < 3 {
		return 0, fmt.Errorf("invalid time format: %s", timeStr)
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, err
	}

	unit := matches[2]
	switch unit {
	case "ms":
		return time.Duration(value) * time.Millisecond, nil
	case "s":
		return time.Duration(value) * time.Second, nil
	case "m":
		return time.Duration(value) * time.Minute, nil
	case "h":
		return time.Duration(value) * time.Hour, nil
	case "d":
		return time.Duration(value) * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("unknown time unit: %s", unit)
	}
}

// validateLevels validates the levels format
func validateLevels(levels string) error {
	if levels == "" {
		return nil
	}

	parts := strings.Split(levels, ":")
	for _, part := range parts {
		depth, err := strconv.Atoi(part)
		if err != nil {
			return fmt.Errorf("invalid levels format: %s", levels)
		}
		if depth < 1 || depth > 2 {
			return fmt.Errorf("level depth must be 1 or 2, got %d", depth)
		}
	}

	return nil
}

// NewProxyCachePath creates a new ProxyCachePath from a directive
func NewProxyCachePath(directive IDirective) (*ProxyCachePath, error) {
	parameters := directive.GetParameters()
	if len(parameters) < 2 {
		return nil, errors.New("proxy_cache_path directive requires at least 2 parameters: path and keys_zone")
	}

	proxyCachePath := &ProxyCachePath{
		Comment:              directive.GetComment(),
		DefaultInlineComment: DefaultInlineComment{InlineComment: directive.GetInlineComment()},
	}

	// Parse parameters
	for i, param := range parameters {
		value := param.GetValue()

		if i == 0 {
			// First parameter is the path
			proxyCachePath.Path = value
			continue
		}

		// Parse key=value parameters
		if strings.Contains(value, "=") {
			parts := strings.SplitN(value, "=", 2)
			if len(parts) != 2 {
				continue
			}

			key := parts[0]
			val := parts[1]

			switch key {
			case "levels":
				if err := validateLevels(val); err != nil {
					return nil, err
				}
				proxyCachePath.Levels = val

			case "keys_zone":
				zoneParts := strings.Split(val, ":")
				if len(zoneParts) != 2 {
					return nil, fmt.Errorf("invalid keys_zone format: %s (expected format: name:size)", val)
				}
				proxyCachePath.KeysZoneName = zoneParts[0]
				proxyCachePath.KeysZoneSize = zoneParts[1]

				if err := validateSize(proxyCachePath.KeysZoneSize); err != nil {
					return nil, err
				}

			case "use_temp_path":
				enabled := val == "on"
				proxyCachePath.UseTemPath = &enabled

			case "inactive":
				if err := validateTime(val); err != nil {
					return nil, err
				}
				proxyCachePath.Inactive = val

			case "max_size":
				if err := validateSize(val); err != nil {
					return nil, err
				}
				proxyCachePath.MaxSize = val

			case "min_free":
				if err := validateSize(val); err != nil {
					return nil, err
				}
				proxyCachePath.MinFree = val

			case "manager_files":
				files, err := strconv.Atoi(val)
				if err != nil {
					return nil, fmt.Errorf("invalid manager_files value: %s", val)
				}
				proxyCachePath.ManagerFiles = &files

			case "manager_sleep":
				if err := validateTime(val); err != nil {
					return nil, err
				}
				proxyCachePath.ManagerSleep = val

			case "manager_threshold":
				if err := validateTime(val); err != nil {
					return nil, err
				}
				proxyCachePath.ManagerThreshold = val

			case "loader_files":
				files, err := strconv.Atoi(val)
				if err != nil {
					return nil, fmt.Errorf("invalid loader_files value: %s", val)
				}
				proxyCachePath.LoaderFiles = &files

			case "loader_sleep":
				if err := validateTime(val); err != nil {
					return nil, err
				}
				proxyCachePath.LoaderSleep = val

			case "loader_threshold":
				if err := validateTime(val); err != nil {
					return nil, err
				}
				proxyCachePath.LoaderThreshold = val

			case "purger":
				enabled := val == "on"
				proxyCachePath.Purger = &enabled

			case "purger_files":
				files, err := strconv.Atoi(val)
				if err != nil {
					return nil, fmt.Errorf("invalid purger_files value: %s", val)
				}
				proxyCachePath.PurgerFiles = &files

			case "purger_sleep":
				if err := validateTime(val); err != nil {
					return nil, err
				}
				proxyCachePath.PurgerSleep = val

			case "purger_threshold":
				if err := validateTime(val); err != nil {
					return nil, err
				}
				proxyCachePath.PurgerThreshold = val

			default:
				return nil, fmt.Errorf("unknown parameter: %s", value)
			}
		} else {
			return nil, fmt.Errorf("invalid parameter format: %s", value)
		}
	}

	// Validate required parameters
	if proxyCachePath.Path == "" {
		return nil, errors.New("cache path is required")
	}

	if proxyCachePath.KeysZoneName == "" {
		return nil, errors.New("keys_zone name is required")
	}

	if proxyCachePath.KeysZoneSize == "" {
		return nil, errors.New("keys_zone size is required")
	}

	return proxyCachePath, nil
}
