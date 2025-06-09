package httpAPI

import (
	"net/http"
	"strconv"
)

func (h *Handler) HandleNotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (h *Handler) HandleMainPage(w http.ResponseWriter, r *http.Request) {
	_, err := strconv.Atoi(r.FormValue("store"))
	if err != nil {
		return
	}

	h.handleTemplate(
		"web/templates/main_page.html",
		"web/templates/navbar.html",
		"web/templates/booking-cart.html",
		"web/templates/footer.html")(w, r)
}

func (h *Handler) HandleBookingsPage(w http.ResponseWriter, r *http.Request) {
	h.handleTemplate("web/templates/bookings.html",
		"web/templates/navbar.html",
		"web/templates/booking-cart.html",
		"web/templates/footer.html")(w, r)
}

func (h *Handler) HandleStoresPage(w http.ResponseWriter, r *http.Request) {
	h.handleTemplate(
		"web/templates/stores_page.html",
		"web/templates/navbar.html",
		"web/templates/booking-cart.html",
		"web/templates/footer.html")(w, r)
}
