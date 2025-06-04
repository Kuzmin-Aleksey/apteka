package util

import (
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"log"
	"store_client/pkg/failure"
	"sync"
)

var (
	errServer       = errors.New("Ошибка сервера")
	errUnauthorized = errors.New("Ошибка авторизации")
	errNetwork      = errors.New("Ошибка подключения")
	errUnknown      = errors.New("Ошибка")
)

func ShowError(w fyne.Window, err error) {
	if err != nil {
		log.Println(err)

		var clientErr error

		switch {
		case failure.IsServerError(err):
			clientErr = errServer
		case failure.IsNetworkError(err):
			clientErr = errNetwork
		case failure.IsUnauthorizedError(err):
			clientErr = errUnauthorized
		default:
			clientErr = errUnknown
		}

		var wg sync.WaitGroup
		wg.Add(1)

		d := dialog.NewError(clientErr, w)
		d.Show()
		d.SetOnClosed(func() {
			wg.Done()
			log.Println("closed")
		})

		wg.Wait()
	}
}
