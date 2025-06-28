package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"regexp"
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
	getBookingsWsUrl       = "/api/booking/by-store/ws"
	updateBookingStatusUrl = "/api/booking/update-status"
	deleteBookingUrl       = "/api/booking/delete"
)

type Service struct {
	cfg    *config.ServiceConfig
	wsAddr string
}

func NewService(cfg *config.ServiceConfig) *Service {
	return &Service{
		cfg:    cfg,
		wsAddr: regexp.MustCompile(`\Ahttps?`).ReplaceAllString(cfg.ServerAddr, "ws"),
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

func (s *Service) GetBookingsChan() <-chan models.BookingsWithError {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	d := websocket.Dialer{}
	ch := make(chan models.BookingsWithError)

	go func() {
		defer close(ch)

		for {
			values := urlValues{
				"store_id": strconv.Itoa(s.cfg.StoreId),
			}

			conn, resp, err := d.Dial(s.wsAddr+getBookingsWsUrl+values.String(), http.Header{
				"Authorization": []string{"ApiKey " + s.cfg.ApiKey},
			})

			if err != nil {
				if errors.As(err, new(net.Error)) {
					ch <- models.BookingsWithError{
						Err: failure.NewNetworkError(err.Error()),
					}
					continue
				}
				ch <- models.BookingsWithError{
					Err: failure.NewNetworkError(err.Error()),
				}
				continue
			}

			if resp.StatusCode != http.StatusSwitchingProtocols {
				if resp.StatusCode == http.StatusUnauthorized {
					ch <- models.BookingsWithError{
						Err: failure.NewUnauthorizedError(readError(resp.Body), resp.StatusCode),
					}
					continue
				}
				ch <- models.BookingsWithError{
					Err: failure.NewServerError(readError(resp.Body), resp.StatusCode),
				}
				continue
			}

			log.Println("connected to ws")

			func() {
				defer func() {
					log.Println("close conn...")
					if err := conn.Close(); err != nil {
						log.Println("close conn:", err)
					}
					log.Println("close conn - ok")
				}()

				errorsCount := 0

				for {
					var bookings []models.Booking

					if err := conn.ReadJSON(&bookings); err != nil {
						log.Println("read bookings error:", err)
						errorsCount++
						if errorsCount < 10 {
							continue
						}

						ch <- models.BookingsWithError{
							Err: failure.NewNetworkError(err.Error()),
						}
						log.Println("reconnect...")
						return
					}

					SortBookings(bookings)
					ch <- models.BookingsWithError{
						Bookings: bookings,
					}
				}
			}()
		}
	}()

	return ch
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
		k := statusesSortMap[i.Status] - statusesSortMap[j.Status]
		if k == 0 {
			return -i.CreatedAt.Compare(j.CreatedAt)
		}
		return k

	})
}

var statusesSortMap = map[string]int{
	models.BookStatusCreated:   1,
	models.BookStatusConfirmed: 2,
	models.BookStatusDone:      3,
	models.BookStatusRejected:  4,
	models.BookStatusReceive:   4,
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
