package httpAPI

import (
	"fmt"
	"golang.org/x/net/context"
	"net/http"
	"server/pkg/failure"
	"strings"
	"time"
)

type CustomWriter struct {
	http.ResponseWriter
	http.Hijacker // if you use websocket
	StatusCode    int
	ContentLength int
}

func (r *CustomWriter) WriteHeader(statusCode int) {
	r.StatusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *CustomWriter) Write(p []byte) (int, error) {
	n, err := r.ResponseWriter.Write(p)
	r.ContentLength += n
	return n, err
}

func (h *Handler) MwLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		customWriter := &CustomWriter{
			ResponseWriter: w,
			Hijacker:       w.(http.Hijacker),
			StatusCode:     http.StatusOK,
		}

		start := time.Now()
		next.ServeHTTP(customWriter, r)
		end := time.Now()

		h.info.Log(r.Method, r.URL.Path, customWriter.StatusCode, r.RemoteAddr, uint64(r.ContentLength), uint64(customWriter.ContentLength), end.Sub(start))
	})

}

func (h *Handler) MwWithApiKey(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		typeAndKey := strings.Split(r.Header.Get("Authorization"), " ")
		if len(typeAndKey) != 2 {
			h.writeError(w, failure.NewUnauthorizedError("invalid authorization header"))
			return
		}
		if typeAndKey[0] != "ApiKey" {
			h.writeError(w, failure.NewUnauthorizedError("invalid auth type"))
			return
		}
		key := typeAndKey[1]
		if key != h.ApiKey {
			h.writeError(w, failure.NewUnauthorizedError("invalid api key"))
			return
		}

		next(w, r)
	}
}

func (h *Handler) MwAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		str := r.Header.Get("Authorization")
		if len(str) == 0 {
			str = r.FormValue("authorization")
		}

		typeAndToken := strings.Split(str, " ")
		if len(typeAndToken) != 2 {
			h.writeError(w, failure.NewUnauthorizedError("invalid authorization header"+fmt.Sprint(typeAndToken)))
			return
		}
		if typeAndToken[0] != "Bearer" {
			h.writeError(w, failure.NewUnauthorizedError("invalid auth type"))
			return
		}
		token := typeAndToken[1]
		ok, err := h.auth.CheckSession(r.Context(), token)
		if err != nil {
			h.writeError(w, failure.NewUnauthorizedError(err.Error()))
			return
		}
		if !ok {
			h.writeError(w, failure.NewUnauthorizedError("invalid token"))
			return
		}
		next(w, r)
	}
}

func (h *Handler) MwWithTimeout(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), h.Timeout)
		defer cancel()

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handler) MwNoCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		next.ServeHTTP(w, r)
	})
}
