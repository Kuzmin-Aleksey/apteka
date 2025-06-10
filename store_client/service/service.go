package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"slices"
	"store_client/config"
	"store_client/models"
	"store_client/pkg/failure"
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
		if errors.As(err, new(net.Error)) {
			return nil, failure.NewNetworkError(err.Error())
		}
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return nil, failure.NewUnauthorizedError(readError(resp.Body), resp.StatusCode)
		}
		return nil, failure.NewServerError(readError(resp.Body), resp.StatusCode)
	}

	var bookings []models.Booking

	if err := json.NewDecoder(resp.Body).Decode(&bookings); err != nil {
		return nil, err
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
		if errors.As(err, new(net.Error)) {
			return failure.NewNetworkError(err.Error())
		}
		return err
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return failure.NewUnauthorizedError(readError(resp.Body), resp.StatusCode)
		}
		return failure.NewServerError(readError(resp.Body), resp.StatusCode)
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

func (s *Service) DeleteBooking(id int) error {
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
		if errors.As(err, new(net.Error)) {
			return failure.NewNetworkError(err.Error())
		}
		return err
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return failure.NewUnauthorizedError(readError(resp.Body), resp.StatusCode)
		}
		return failure.NewServerError(readError(resp.Body), resp.StatusCode)
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

type errorResponse struct {
	Error string `json:"error"`
}

func readError(r io.Reader) string {
	var e errorResponse

	if err := json.NewDecoder(r).Decode(&e); err != nil {
		log.Println("decode error:", err)
	}

	return e.Error
}
