package middlewarex

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"strings"

	"server/pkg/logx"
)

func RequestLogging(
	sensitiveDataMasker logx.SensitiveDataMaskerInterface,
	logFieldMaxLen int,
) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if strings.Contains(r.RequestURI, "image") ||
				strings.Contains(r.RequestURI, "static") ||
				strings.Contains(r.RequestURI, "ico") {
				next.ServeHTTP(w, r)
				return
			}

			ctx := r.Context()
			dumpBody := true

			if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
				dumpBody = false
			}

			dump, err := httputil.DumpRequest(r, dumpBody)

			if len(dump) > logFieldMaxLen {
				dump = dump[:logFieldMaxLen]
			}

			logger(ctx).Info(
				logx.FieldHTTPRequest,
				slog.String(logx.FieldRequestBody, string(sensitiveDataMasker.Mask(dump))),
				logx.Error(err),
			)

			next.ServeHTTP(w, r)
		})
	}
}
