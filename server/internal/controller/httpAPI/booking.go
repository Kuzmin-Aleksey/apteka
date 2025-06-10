package httpAPI

import (
	"encoding/json"
	"math"
	"net/http"
	"server/internal/domain/service/booking"
	"server/pkg/failure"
	"strconv"
	"time"
)

type BookResponse struct {
	BookId int `json:"book_id"`
}

func (h *Handler) ApiToBook(w http.ResponseWriter, r *http.Request) {
	dto := new(booking.CreateBookDTO)
	if err := json.NewDecoder(r.Body).Decode(dto); err != nil {
		h.writeError(w, failure.NewInvalidRequestError("invalid json"+": "+err.Error()))
		return
	}

	if err := r.ParseForm(); err != nil {
		h.writeError(w, failure.NewInvalidRequestError("invalid form"+": "+err.Error()))
		return
	}
	storeId, err := strconv.Atoi(r.Form.Get("store_id"))
	if err != nil {
		h.writeError(w, failure.NewInvalidRequestError("invalid store id"+": "+err.Error()))
		return
	}

	bookId, err := h.booking.ToBook(r.Context(), storeId, dto)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, BookResponse{BookId: bookId})
}

func (h *Handler) ApiBookUpdateStatus(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.writeError(w, failure.NewInvalidRequestError("invalid form"+": "+err.Error()))
		return
	}

	bookId, err := strconv.Atoi(r.Form.Get("book_id"))
	if err != nil {
		h.writeError(w, failure.NewInvalidRequestError("invalid book id"+": "+err.Error()))
		return
	}

	status := r.Form.Get("status")

	if err := h.booking.UpdateStatus(r.Context(), bookId, status); err != nil {
		h.writeError(w, err)
		return
	}
}

func (h *Handler) ApiGetBook(w http.ResponseWriter, r *http.Request) {
	bookId, err := strconv.Atoi(r.FormValue("book_id"))
	if err != nil {
		h.writeError(w, failure.NewInvalidRequestError("invalid book id"+": "+err.Error()))
		return
	}

	book, err := h.booking.Get(r.Context(), bookId)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, book)
}

func (h *Handler) ApiGetBooks(w http.ResponseWriter, r *http.Request) {
	var ids []int

	if err := json.Unmarshal([]byte(r.FormValue("ids")), &ids); err != nil {
		h.writeError(w, failure.NewInvalidRequestError("invalid book ids"+": "+err.Error()))
		return
	}

	books, err := h.booking.GetByIds(r.Context(), ids)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, books)
}

func (h *Handler) ApiGetStoreBookings(w http.ResponseWriter, r *http.Request) {
	storeId, err := strconv.Atoi(r.FormValue("store_id"))
	if err != nil {
		h.writeError(w, failure.NewInvalidRequestError("invalid store id"+": "+err.Error()))
		return
	}

	bookings, err := h.booking.GetByStore(r.Context(), storeId)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, bookings)
}

func (h *Handler) ApiDeleteBooking(w http.ResponseWriter, r *http.Request) {
	bookId, err := strconv.Atoi(r.FormValue("book_id"))
	if err != nil {
		h.writeError(w, failure.NewInvalidRequestError("invalid book id"+": "+err.Error()))
		return
	}

	if err := h.booking.Delete(r.Context(), bookId); err != nil {
		h.writeError(w, err)
		return
	}
}

func (h *Handler) ApiSetBookingDelay(w http.ResponseWriter, r *http.Request) {
	delay, err := strconv.Atoi(r.FormValue("delay"))
	if err != nil {
		h.writeError(w, failure.NewInvalidRequestError("invalid delay id"+": "+err.Error()))
		return
	}

	if err := h.booking.StoreBookingDelay(time.Duration(delay) * time.Hour); err != nil {
		h.writeError(w, err)
		return
	}
}

func (h *Handler) ApiGetBookingDelay(w http.ResponseWriter, r *http.Request) {
	delay := h.booking.GetBookingDelay()

	w.Write([]byte(strconv.Itoa(int(math.Round(delay.Hours())))))
}
