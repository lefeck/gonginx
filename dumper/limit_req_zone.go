package dumper

import (
	"fmt"
	"strings"

	"github.com/lefeck/gonginx/config"
)

// DumpLimitReqZone converts a LimitReqZone to a string representation
func DumpLimitReqZone(lrz *config.LimitReqZone, style *Style) string {
	if lrz == nil {
		return ""
	}

	result := ""

	// Add comments before the limit_req_zone directive
	if len(lrz.GetComment()) > 0 {
		for _, comment := range lrz.GetComment() {
			result += fmt.Sprintf("%s%s\n", strings.Repeat(" ", style.StartIndent), comment)
		}
	}

	// Add the limit_req_zone directive with parameters
	result += strings.Repeat(" ", style.StartIndent) + "limit_req_zone " + lrz.Key

	// Add zone parameter
	result += " zone=" + lrz.ZoneName + ":" + lrz.ZoneSize

	// Add rate parameter
	result += " rate=" + lrz.Rate

	// Add sync parameter if enabled
	if lrz.Sync {
		result += " sync"
	}

	result += ";"

	// Add inline comments
	if len(lrz.GetInlineComment()) > 0 {
		for _, inlineComment := range lrz.GetInlineComment() {
			result += " " + inlineComment.Value
		}
	}

	return result
}
