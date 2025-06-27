package server

import (
	"encoding/xml"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"server/internal/domain/service/store"
	"server/pkg/contextx"
	"server/pkg/logx"
	"time"
)

type SitemapGenerator struct {
	store *store.StoreService
}

func NewSitemapGenerator(store *store.StoreService) *SitemapGenerator {
	return &SitemapGenerator{store: store}
}

type sitemap struct {
	XMLName xml.Name     `xml:"urlset"`
	Xmlns   string       `xml:"xmlns,attr"`
	Url     []siteMapUrl `xml:"url"`
}

type siteMapUrl struct {
	Loc      string    `xml:"loc"`
	LastMod  time.Time `xml:"lastmod"`
	Priority float32   `xml:"priority"`
}

func (g *SitemapGenerator) handleSitemap(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := contextx.GetLoggerOrDefault(ctx)

	w.Header().Set("Content-Type", "text/xml")

	proto := "http"
	if r.TLS != nil {
		proto = "https"
	}

	var sm sitemap
	sm.Xmlns = "http://www.sitemaps.org/schemas/sitemap/0.9"

	urlFiles := map[string]string{
		"/stores": "web/templates/stores_page.html",
	}

	priorityMap := map[string]float32{
		"/stores": 0.9,
	}

	templateParts := []string{
		"web/templates/navbar.html",
		"web/templates/booking-cart.html",
		"web/templates/footer.html",
		"web/static/js/init.js",
		"web/static/js/util.js",
	}

	var templatePartsLastMod time.Time

	for _, templatePart := range templateParts {
		fileInfo, err := os.Stat(templatePart)
		if err != nil {
			l.Error("generate sitemap", logx.Error(fmt.Errorf("get file info: %w", err)), slog.String("file", templatePart))
			continue
		}

		if fileInfo.ModTime().After(templatePartsLastMod) {
			templatePartsLastMod = fileInfo.ModTime()
		}
	}

	stores, err := g.store.GetAll(ctx)
	if err != nil {
		l.Error("generate sitemap", logx.Error(fmt.Errorf("get stores: %w", err)))
		return
	}

	for _, st := range stores {
		url := fmt.Sprintf("/?store=%d", st.Id)
		urlFiles[url] = "web/templates/main_page.html"
		priorityMap[url] = 1
	}

	for url, file := range urlFiles {
		fileInfo, err := os.Stat(file)
		if err != nil {
			l.Error("generate sitemap", logx.Error(fmt.Errorf("get file info: %w", err)), slog.String("file", file))
			continue
		}

		lastMod := fileInfo.ModTime()
		if lastMod.Before(templatePartsLastMod) {
			lastMod = templatePartsLastMod
		}

		sm.Url = append(sm.Url, siteMapUrl{
			Loc:      fmt.Sprintf("%s://%s%s", proto, r.Host, url),
			LastMod:  lastMod,
			Priority: priorityMap[url],
		})
	}

	w.Write([]byte(xml.Header))
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	if err := enc.Encode(sm); err != nil {
		l.Error("generate sitemap", logx.Error(fmt.Errorf("encode xml: %w", err)), slog.Any("sitemap", sm))
	}

}
