package dumper

import (
	"fmt"
	"strings"

	"github.com/lefeck/gonginx/config"
)

// DumpProxyCachePath converts a ProxyCachePath to a string representation
func DumpProxyCachePath(pcp *config.ProxyCachePath, style *Style) string {
	if pcp == nil {
		return ""
	}

	result := ""

	// Add comments before the proxy_cache_path directive
	if len(pcp.GetComment()) > 0 {
		for _, comment := range pcp.GetComment() {
			result += fmt.Sprintf("%s%s\n", strings.Repeat(" ", style.StartIndent), comment)
		}
	}

	// Add the proxy_cache_path directive with parameters
	result += strings.Repeat(" ", style.StartIndent) + "proxy_cache_path " + pcp.Path

	// Add levels parameter
	if pcp.Levels != "" {
		result += " levels=" + pcp.Levels
	}

	// Add use_temp_path parameter
	if pcp.UseTemPath != nil {
		value := "off"
		if *pcp.UseTemPath {
			value = "on"
		}
		result += " use_temp_path=" + value
	}

	// Add keys_zone parameter (required)
	result += " keys_zone=" + pcp.KeysZoneName + ":" + pcp.KeysZoneSize

	// Add inactive parameter
	if pcp.Inactive != "" {
		result += " inactive=" + pcp.Inactive
	}

	// Add max_size parameter
	if pcp.MaxSize != "" {
		result += " max_size=" + pcp.MaxSize
	}

	// Add min_free parameter
	if pcp.MinFree != "" {
		result += " min_free=" + pcp.MinFree
	}

	// Add manager parameters
	if pcp.ManagerFiles != nil {
		result += fmt.Sprintf(" manager_files=%d", *pcp.ManagerFiles)
	}

	if pcp.ManagerSleep != "" {
		result += " manager_sleep=" + pcp.ManagerSleep
	}

	if pcp.ManagerThreshold != "" {
		result += " manager_threshold=" + pcp.ManagerThreshold
	}

	// Add loader parameters
	if pcp.LoaderFiles != nil {
		result += fmt.Sprintf(" loader_files=%d", *pcp.LoaderFiles)
	}

	if pcp.LoaderSleep != "" {
		result += " loader_sleep=" + pcp.LoaderSleep
	}

	if pcp.LoaderThreshold != "" {
		result += " loader_threshold=" + pcp.LoaderThreshold
	}

	// Add purger parameters
	if pcp.Purger != nil {
		value := "off"
		if *pcp.Purger {
			value = "on"
		}
		result += " purger=" + value
	}

	if pcp.PurgerFiles != nil {
		result += fmt.Sprintf(" purger_files=%d", *pcp.PurgerFiles)
	}

	if pcp.PurgerSleep != "" {
		result += " purger_sleep=" + pcp.PurgerSleep
	}

	if pcp.PurgerThreshold != "" {
		result += " purger_threshold=" + pcp.PurgerThreshold
	}

	result += ";"

	// Add inline comments
	if len(pcp.GetInlineComment()) > 0 {
		for _, inlineComment := range pcp.GetInlineComment() {
			result += " " + inlineComment.Value
		}
	}

	return result
}
