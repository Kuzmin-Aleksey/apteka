package httpAPI

import (
	"encoding/json"
	"net/http"
	"server/domain/models"
	"server/domain/service/booking"
	"strconv"
	"time"
)

type BookResponse struct {
	BookId int `json:"book_id"`
}

func (h *Handler) ApiToBook(w http.ResponseWriter, r *http.Request) {
	dto := new(booking.CreateBookDTO)
	if err := json.NewDecoder(r.Body).Decode(dto); err != nil {
		h.writeError(w, models.NewError(models.ErrInvalidRequest, "invalid json", err))
		return
	}

	if err := r.ParseForm(); err != nil {
		h.writeError(w, models.NewError(models.ErrInvalidRequest, "invalid form", err))
		return
	}
	storeId, err := strconv.Atoi(r.Form.Get("store_id"))
	if err != nil {
		h.writeError(w, models.NewError(models.ErrInvalidRequest, "invalid store id", err))
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
		h.writeError(w, models.NewError(models.ErrInvalidRequest, "invalid form", err))
		return
	}

	bookId, err := strconv.Atoi(r.Form.Get("book_id"))
	if err != nil {
		h.writeError(w, models.NewError(models.ErrInvalidRequest, "invalid book id", err))
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
		h.writeError(w, models.NewError(models.ErrInvalidRequest, "invalid book id", err))
		return
	}

	book, err := h.booking.Get(r.Context(), bookId)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, book)
}

func (h *Handler) ApiGetStoreBookings(w http.ResponseWriter, r *http.Request) {
	storeId, err := strconv.Atoi(r.FormValue("store_id"))
	if err != nil {
		h.writeError(w, models.NewError(models.ErrInvalidRequest, "invalid store id", err))
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
		h.writeError(w, models.NewError(models.ErrInvalidRequest, "invalid book id", err))
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
		h.writeError(w, models.NewError(models.ErrInvalidRequest, "invalid delay", err))
		return
	}

	if err := h.booking.StoreBookingDelay(time.Duration(delay)); err != nil {
		h.writeError(w, err)
		return
	}
}
