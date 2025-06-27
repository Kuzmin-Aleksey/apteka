package httpAPI

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"time"
)

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

func (h *Handler) handleSitemap(w http.ResponseWriter, r *http.Request) {
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
		f, err := os.Stat(templatePart)
		if err != nil {
			h.l.Printf("reading file (%s) : %s", templatePart, err.Error())
			continue
		}

		if f.ModTime().After(templatePartsLastMod) {
			templatePartsLastMod = f.ModTime()
		}
	}

	stores, err := h.store.GetAll(r.Context())
	if err != nil {
		h.l.Println(err.Error())
		return
	}

	for _, store := range stores {
		url := fmt.Sprintf("/?store=%d", store.Id)
		urlFiles[url] = "web/templates/main_page.html"
		priorityMap[url] = 1
	}

	for url, file := range urlFiles {
		f, err := os.Stat(file)
		if err != nil {
			h.l.Printf("reading file (%s) : %s", file, err.Error())
			continue
		}

		lastMod := f.ModTime()
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
		h.l.Println(err.Error())
	}

}
