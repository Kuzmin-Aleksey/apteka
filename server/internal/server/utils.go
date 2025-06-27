package server

import (
	"log/slog"
	"net/http"
	"os"
	"server/pkg/contextx"
	"server/pkg/logx"
)

func (s *Server) handleFile(path string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l := contextx.GetLoggerOrDefault(r.Context())

		f, err := os.Open(path)
		if err != nil {
			if os.IsNotExist(err) {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			l.Error("handle file error", slog.String("file", path), logx.Error(err))
			return
		}
		defer f.Close()

		fi, err := f.Stat()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			l.Error("handle file error", slog.String("file", path), logx.Error(err))
			return
		}

		http.ServeContent(w, r, fi.Name(), fi.ModTime(), f)
	})
}
