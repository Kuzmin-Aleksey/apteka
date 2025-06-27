package middlewarex

import (
	"server/pkg/contextx"
)

var logger = contextx.GetLoggerOrDefault //nolint:gochecknoglobals
