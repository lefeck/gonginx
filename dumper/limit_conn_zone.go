package dumper

import (
	"fmt"
	"strings"

	"github.com/lefeck/gonginx/config"
)

// DumpLimitConnZone converts a LimitConnZone to a string representation
func DumpLimitConnZone(lcz *config.LimitConnZone, style *Style) string {
	if lcz == nil {
		return ""
	}

	result := ""

	// Add comments before the limit_conn_zone directive
	if len(lcz.GetComment()) > 0 {
		for _, comment := range lcz.GetComment() {
			result += fmt.Sprintf("%s%s\n", strings.Repeat(" ", style.StartIndent), comment)
		}
	}

	// Add the limit_conn_zone directive with parameters
	result += strings.Repeat(" ", style.StartIndent) + "limit_conn_zone " + lcz.Key

	// Add zone parameter
	result += " zone=" + lcz.ZoneName + ":" + lcz.ZoneSize

	// Add sync parameter if enabled
	if lcz.Sync {
		result += " sync"
	}

	result += ";"

	// Add inline comments
	if len(lcz.GetInlineComment()) > 0 {
		for _, inlineComment := range lcz.GetInlineComment() {
			result += " " + inlineComment.Value
		}
	}

	return result
}
