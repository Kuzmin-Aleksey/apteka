package ui

import (
	"apteka_booking/ui/booking"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

type UI struct {
	s booking.Service
	a fyne.App
}

func New(s booking.Service) *UI {
	a := app.New()

	return &UI{
		s: s,
		a: a,
	}
}

func (ui *UI) Run() {
	w := booking.NewBookingWindow(ui.a, ui.s)
	w.Resize(fyne.NewSize(900, 600))
	w.ShowAndRun()
}
