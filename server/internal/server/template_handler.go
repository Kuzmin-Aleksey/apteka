package server

import (
	"html/template"
	"net/http"
	"server/internal/config"
	"server/pkg/failure"
	"strings"
)

type TemplateData struct {
	Title   string `json:"title"`
	Logo    string `json:"logo" `
	LogoMin string `json:"logo_min"`
}

type templateHandler struct {
	cfg            *config.WebConfig
	templatesCache map[string]*template.Template
}

func newTemplateHandler(cfg *config.WebConfig) *templateHandler {
	return &templateHandler{
		cfg:            cfg,
		templatesCache: make(map[string]*template.Template),
	}
}

func (th *templateHandler) handleTemplate(tmpPath ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		templateCacheKey := strings.Join(tmpPath, "&")

		var tmp *template.Template
		if th.cfg.CacheTemplate {
			tmp = th.templatesCache[templateCacheKey]
		}
		if tmp == nil {
			var err error
			tmp, err = template.ParseFiles(tmpPath...)
			if err != nil {
				writeAndLogErr(ctx, w, failure.NewInternalError("parse template: "+err.Error()))
				return
			}
			if th.cfg.CacheTemplate {
				th.templatesCache[templateCacheKey] = tmp
			}
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmp.Execute(w, &TemplateData{
			Title:   th.cfg.Title,
			Logo:    th.cfg.Logo,
			LogoMin: th.cfg.LogoMin,
		}); err != nil {
			writeAndLogErr(ctx, w, failure.NewInternalError("execute template: "+err.Error()))
			return
		}
	}
}
