package httpAPI

import (
	"github.com/gorilla/mux"
)

type AdminPagesHandler struct {
	*mux.Router
	baseHandler *Handler
}

func NewAdminPagesHandler(base *Handler) (*AdminPagesHandler, error) {
	h := &AdminPagesHandler{
		Router:      mux.NewRouter(),
		baseHandler: base,
	}

	templateFiles := map[string][]string{
		"login":     {"web/templates/login.html"},
		"admin":     {"web/templates/admin.html", "web/templates/admin_root.html"},
		"stores":    {"web/templates/admin.html", "web/templates/admin_store.html"},
		"promotion": {"web/templates/admin.html", "web/templates/admin_promotion.html"},
		"images":    {"web/templates/admin.html", "web/templates/admin_images.html"},
	}

	h.HandleFunc("/", h.baseHandler.handleTemplate(templateFiles["admin"]...))
	h.HandleFunc("/login", h.baseHandler.handleTemplate(templateFiles["login"]...))
	h.HandleFunc("/stores", h.baseHandler.handleTemplate(templateFiles["stores"]...))
	h.HandleFunc("/promotion", h.baseHandler.handleTemplate(templateFiles["promotion"]...))
	h.HandleFunc("/images", h.baseHandler.handleTemplate(templateFiles["images"]...))

	h.NotFoundHandler = h.baseHandler.MwLogging(h.NotFoundHandler)

	return h, nil
}
