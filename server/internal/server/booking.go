package server

import (
	"encoding/json"
	"math"
	"net/http"
	"server/internal/domain/service/booking"
	"server/pkg/failure"
	"strconv"
	"time"
)

type BookingServer struct {
	booking *booking.BookingService
}

func NewBookingServer(booking *booking.BookingService) *BookingServer {
	return &BookingServer{booking: booking}
}

type BookResponse struct {
	BookId int `json:"book_id"`
}

func (s *BookingServer) ApiToBook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	dto := new(booking.CreateBookDTO)
	if err := json.NewDecoder(r.Body).Decode(dto); err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError("invalid json"+": "+err.Error()))
		return
	}

	if err := r.ParseForm(); err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError("invalid form"+": "+err.Error()))
		return
	}
	storeId, err := strconv.Atoi(r.Form.Get("store_id"))
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError("invalid store id"+": "+err.Error()))
		return
	}

	bookId, err := s.booking.ToBook(ctx, storeId, dto)
	if err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}

	writeJson(ctx, w, BookResponse{BookId: bookId}, http.StatusOK)
}

func (s *BookingServer) ApiBookUpdateStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError("invalid form"+": "+err.Error()))
		return
	}

	bookId, err := strconv.Atoi(r.Form.Get("book_id"))
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError("invalid book id"+": "+err.Error()))
		return
	}

	status := r.Form.Get("status")

	if err := s.booking.UpdateStatus(ctx, bookId, status); err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}
}

func (s *BookingServer) ApiGetBook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	bookId, err := strconv.Atoi(r.FormValue("book_id"))
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError("invalid book id"+": "+err.Error()))
		return
	}

	book, err := s.booking.Get(ctx, bookId)
	if err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}

	writeJson(ctx, w, book, http.StatusOK)
}

func (s *BookingServer) ApiGetBooks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var ids []int

	if err := json.Unmarshal([]byte(r.FormValue("ids")), &ids); err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError("invalid book ids"+": "+err.Error()))
		return
	}

	books, err := s.booking.GetByIds(ctx, ids)
	if err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}

	writeJson(ctx, w, books, http.StatusOK)
}

func (s *BookingServer) ApiGetStoreBookings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	storeId, err := strconv.Atoi(r.FormValue("store_id"))
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError("invalid store id"+": "+err.Error()))
		return
	}

	bookings, err := s.booking.GetByStore(ctx, storeId)
	if err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}

	writeJson(ctx, w, bookings, http.StatusOK)
}

func (s *BookingServer) ApiDeleteBooking(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	bookId, err := strconv.Atoi(r.FormValue("book_id"))
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError("invalid book id"+": "+err.Error()))
		return
	}

	if err := s.booking.Delete(ctx, bookId); err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}
}

func (s *BookingServer) ApiSetBookingDelay(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	delay, err := strconv.Atoi(r.FormValue("delay"))
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError("invalid delay id"+": "+err.Error()))
		return
	}

	if err := s.booking.StoreBookingDelay(time.Duration(delay) * time.Hour); err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}
}

func (s *BookingServer) ApiGetBookingDelay(w http.ResponseWriter, r *http.Request) {
	delay := s.booking.GetBookingDelay()

	w.Write([]byte(strconv.Itoa(int(math.Round(delay.Hours())))))
}
