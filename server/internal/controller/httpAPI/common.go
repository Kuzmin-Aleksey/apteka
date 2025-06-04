package httpAPI

import (
	"net/http"
	"strconv"
)

func (h *Handler) HandleNotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (h *Handler) HandleRootPage(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) HandleMainPage(w http.ResponseWriter, r *http.Request) {
	storeId, err := strconv.Atoi(r.FormValue("store"))
	if err != nil {
		h.HandleRootPage(w, r)
		return
	}
	h.l.Println(storeId)

	h.handleTemplate("web/templates/main_page.html")(w, r)
}

func (h *Handler) HandleStoresPage(w http.ResponseWriter, r *http.Request) {
	h.handleTemplate("web/templates/stores_page.html")(w, r)
}
