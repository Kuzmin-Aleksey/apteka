package service

import (
	"apteka_booking/config"
	"apteka_booking/models"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
)

const (
	pingUrl                = "/ping"
	getBookingsUrl         = "/api/booking/by-store"
	updateBookingStatusUrl = "/api/booking/update-status"
	deleteBookingUrl       = "/api/booking/delete"
)

type Service struct {
	cfg *config.ServiceConfig
}

func NewService(cfg *config.ServiceConfig) *Service {
	return &Service{
		cfg: cfg,
	}
}

func (s *Service) GetBookings() ([]models.Booking, error) {
	values := urlValues{
		"store_id": strconv.Itoa(s.cfg.StoreId),
	}

	r, err := http.NewRequest(http.MethodGet, s.cfg.ServerAddr+getBookingsUrl+values.String(), nil)
	if err != nil {
		return nil, err
	}
	r.Header.Set("Authorization", "ApiKey "+s.cfg.ApiKey)

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	var books []models.Booking

	if err := json.NewDecoder(resp.Body).Decode(&books); err != nil {
		return nil, err
	}

	var bookings []models.Booking

	for _, booking := range books {
		if _, ok := deletingBookings[booking.Id]; !ok {
			bookings = append(bookings, booking)
		}
	}

	SortBookings(bookings)
	log.Println("loaded booking")

	return bookings, nil
}

func (s *Service) SetBookingStatus(id int, status string) error {
	log.Println(id, "set booking status", status)

	values := urlValues{
		"book_id": strconv.Itoa(id),
		"status":  status,
	}

	r, err := http.NewRequest(http.MethodPost, s.cfg.ServerAddr+updateBookingStatusUrl+values.String(), nil)
	if err != nil {
		return err
	}
	r.Header.Set("Authorization", "ApiKey "+s.cfg.ApiKey)

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	return nil
}

func (s *Service) Ping() error {
	r, err := http.NewRequest(http.MethodGet, s.cfg.ServerAddr+pingUrl, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	pong, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if string(pong) != "pong" {
		return errors.New("invalid server response")
	}

	return nil
}

var deletingBookings = make(map[int]struct{})

func (s *Service) DeleteBooking(id int) error {
	deletingBookings[id] = struct{}{}
	return nil
}

func (s *Service) deleteBooking(id int) error {
	values := urlValues{
		"book_id": strconv.Itoa(id),
	}
	r, err := http.NewRequest(http.MethodPost, s.cfg.ServerAddr+deleteBookingUrl+values.String(), nil)
	if err != nil {
		return err
	}
	r.Header.Set("Authorization", "ApiKey "+s.cfg.ApiKey)

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	return nil

}

type urlValues map[string]string

func (v urlValues) String() string {
	var s string

	for k, v := range v {
		s += k + "=" + url.QueryEscape(v) + "&"
	}

	s = strings.TrimRight(s, "&")

	return "?" + s
}

func SortBookings(bookings []models.Booking) {
	slices.SortFunc(bookings, func(i, j models.Booking) int {
		if i.Status == j.Status {
			return -i.CreatedAt.Compare(j.CreatedAt)
		}
		return statusesSortMap[i.Status] - statusesSortMap[j.Status]

	})
}

var statusesSortMap = map[string]int{
	models.BookStatusCreated:   1,
	models.BookStatusConfirmed: 2,
	models.BookStatusRejected:  3,
	models.BookStatusDone:      4,
}
