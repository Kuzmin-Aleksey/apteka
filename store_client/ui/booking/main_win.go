package booking

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"log"
	"os/exec"
	"slices"
	"store_client/models"
	"store_client/ui/util"
	"sync"
	"time"
)

const WinTittle = "Booking"

type Window struct {
	fyne.Window
	s                Service
	selectedId       int
	bookingsList     *widget.List
	bookings         []models.Booking
	bookingContainer *fyne.Container
}

func (w *Window) Run() {
	w.pingServer()

	bookings, err := w.s.GetBookings()
	time.Sleep(time.Second)
	if err != nil {
		util.ShowError(w, err)
		w.Close()
	}
	w.bookings = bookings

	w.bookingsList = widget.NewList(
		func() int {
			return len(w.bookings)
		},
		w.createItem,
		w.updateItem,
	)

	w.bookingContainer = container.NewStack()

	w.bookingsList.OnSelected = func(i widget.ListItemID) {
		id := w.bookings[i].Id
		w.selectedId = id

		tab := w.createBookingTab(w.bookings[i], func(status string) {
			if err := w.s.SetBookingStatus(id, status); err != nil {
				go util.ShowError(w, err)
			}

			for i := range w.bookings {
				if w.bookings[i].Id == id {
					w.bookings[i].Status = status
				}
			}

		}, func() {
			w.deleteBooking(id)
		})

		w.bookingContainer.Objects = []fyne.CanvasObject{tab}
		w.bookingContainer.Refresh()
	}

	split := container.NewHSplit(w.bookingsList, w.bookingContainer)
	split.SetOffset(0.3)
	w.SetContent(split)

	w.loadingBookings()
}

func (w *Window) createItem() fyne.CanvasObject {
	return container.NewHBox(iconNull, widget.NewLabel("loading"), layout.NewSpacer(), widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {}))
}

var (
	iconCreated = func() *widget.Icon {
		res := theme.NewThemedResource(theme.RadioButtonFillIcon())
		res.ColorName = theme.ColorNameSuccess
		icon := widget.NewIcon(res)
		return icon
	}()

	iconConfirmed = widget.NewIcon(theme.RadioButtonFillIcon())

	iconNull = func() *widget.Icon {
		icon := widget.NewIcon(nil)
		return icon
	}()
)

func (w *Window) updateItem(i widget.ListItemID, item fyne.CanvasObject) {
	switch w.bookings[i].Status {
	case models.BookStatusCreated:
		item.(*fyne.Container).Objects[0] = iconCreated
	case models.BookStatusConfirmed:
		item.(*fyne.Container).Objects[0] = iconConfirmed
	default:
		item.(*fyne.Container).Objects[0] = iconNull
	}

	label := item.(*fyne.Container).Objects[1].(*widget.Label)
	label.SetText(fmt.Sprintf("№%d %s", w.bookings[i].Id, getStatusDisplayName(w.bookings[i].Status)))

	delBtn := item.(*fyne.Container).Objects[3].(*widget.Button)

	id := w.bookings[i].Id
	delBtn.OnTapped = func() {
		w.deleteBooking(id)
	}
}

func (w *Window) loadingBookings() {
	bookingsCh := w.s.GetBookingsChan()

	go func() {
		for {
			bookingResp := <-bookingsCh
			bookings, err := bookingResp.Bookings, bookingResp.Err
			if err != nil {
				log.Println(" <-bookingsCh err:", err)
				util.ShowError(w, err)
				continue
			}

			for _, booking := range bookings {
				if booking.Status == models.BookStatusCreated {
					if !slices.ContainsFunc(w.bookings, func(b models.Booking) bool { return b.Id == booking.Id }) {
						w.OnNewBooking(&booking)
					}
				}
			}

			w.bookings = bookings

			for i := range w.bookings {
				if w.bookings[i].Id == w.selectedId {
					w.bookingsList.Select(i)
				}
			}

			w.bookingsList.Refresh()
		}
	}()
}

func (w *Window) deleteBooking(id int) {
	d := dialog.NewConfirm(
		"Удаление заказа",
		fmt.Sprintf("Заказ №%d будет удален", id),
		func(ok bool) {
			if !ok {
				return
			}

			log.Println("delete: ", id)
			if err := w.s.DeleteBooking(id); err != nil {
				util.ShowError(w, err)
			}

			w.bookings = slices.DeleteFunc(w.bookings, func(booking models.Booking) bool {
				return booking.Id == id
			})

			w.bookingContainer.RemoveAll()
			w.bookingsList.Refresh()
		},
		w,
	)
	d.SetConfirmText("Удалить")
	d.SetDismissText("Отменить")

	d.Show()

}

func NewBookingWindow(a fyne.App, s Service) *Window {
	w := a.NewWindow(WinTittle)

	return &Window{
		Window: w,
		s:      s,
	}
}

func (w *Window) ShowAndRun() {
	loadLabel := widget.NewRichTextFromMarkdown("# Загрузка")

	progress := widget.NewProgressBarInfinite()
	progress.Start()

	vbox := container.NewVBox(loadLabel, progress)
	content := container.NewCenter(vbox)

	w.SetContent(content)

	go w.Run()

	w.Window.ShowAndRun()
}

func (w *Window) pingServer() {
	if err := w.s.Ping(); err != nil {
		log.Println(err)

		var wg sync.WaitGroup
		wg.Add(1)

		d := dialog.NewError(errors.New("Сервер недоступен"), w)
		d.Show()
		d.SetOnClosed(func() {
			wg.Done()
		})

		wg.Wait()
	}
}

func (w *Window) OnNewBooking(b *models.Booking) {
	if err := exec.Command("./notify.exe", "--title", "Новый заказ", "--content", fmt.Sprintf("Заказ №%d", b.Id)).Start(); err != nil {
		log.Println("show notify error", err)
	}
}
