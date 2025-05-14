package util

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"log"
	"sync"
)

func ShowError(w fyne.Window, err error) {
	if err != nil {
		var wg sync.WaitGroup
		wg.Add(1)

		d := dialog.NewError(err, w)
		d.Show()
		d.SetOnClosed(func() {
			wg.Done()
			log.Println("closed")
		})

		wg.Wait()
	}
}
