package booking

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"store_client/models"
	"strconv"
	"time"
)

func (w *Window) createBookingTab(booking models.Booking, onStatusChange func(newStatus string), onDelete func()) fyne.CanvasObject {
	idLabel := widget.NewLabel(fmt.Sprintf("Заказ №%d", booking.Id))
	idLabel.Alignment = fyne.TextAlignCenter
	idLabel.TextStyle.Bold = true

	dateLabel := widget.NewLabel("Дата создания: " + booking.CreatedAt.Format(time.DateTime))

	statusLabel := widget.NewLabel("Статус: " + getStatusDisplayName(booking.Status))
	statusLabel.TextStyle.Bold = true

	usernameLabel := widget.NewLabel("Имя клиента: " + booking.Username)

	phoneLabel := widget.NewLabel("Телефон: " + booking.Phone)

	messageLabel := widget.NewLabel("Комментарий к заказу: " + booking.Message)

	statusButtons := container.NewHBox()

	btnConfirm := widget.NewButton("Подтвердить", func() {
		onStatusChange(models.BookStatusConfirmed)
		updateStatusButtons(models.BookStatusConfirmed, statusButtons)
		statusLabel.SetText("Статус: " + getStatusDisplayName(models.BookStatusConfirmed))
	})
	btnReject := widget.NewButton("Отменить", func() {
		onStatusChange(models.BookStatusRejected)
		updateStatusButtons(models.BookStatusRejected, statusButtons)
		statusLabel.SetText("Статус: " + getStatusDisplayName(models.BookStatusRejected))
	})
	btnDone := widget.NewButton("Готов", func() {
		onStatusChange(models.BookStatusDone)
		updateStatusButtons(models.BookStatusDone, statusButtons)
		statusLabel.SetText("Статус: " + getStatusDisplayName(models.BookStatusDone))
	})
	btnReceive := widget.NewButton("Выдано", func() {
		onStatusChange(models.BookStatusReceive)
		updateStatusButtons(models.BookStatusReceive, statusButtons)
		statusLabel.SetText("Статус: " + getStatusDisplayName(models.BookStatusReceive))
	})

	btnDelete := widget.NewButtonWithIcon("", theme.DeleteIcon(), onDelete)

	statusButtons.Objects = []fyne.CanvasObject{btnConfirm, btnReject, btnDone, btnReceive, layout.NewSpacer(), btnDelete}
	updateStatusButtons(booking.Status, statusButtons)

	// Создаем таблицу для товаров
	productsHeader := widget.NewLabel("Заказанные товары:")
	productsHeader.TextStyle = fyne.TextStyle{Bold: true}

	// Создаем таблицу с товарами
	productsTable := widget.NewTable(
		func() (int, int) {
			return len(booking.Products) + 1, 3
		},
		func() fyne.CanvasObject {
			copyBtn := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {})
			copyBtn.Hide()
			return container.NewHScroll(container.NewHBox(widget.NewLabel("Ширина столбца"), layout.NewSpacer(), copyBtn))
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			scrollContainer := cell.(*container.Scroll)
			hBox := scrollContainer.Content.(*fyne.Container)
			label := hBox.Objects[0].(*widget.Label)

			if id.Row == 0 {
				switch id.Col {
				case 0:
					label.SetText("Код STU")
					label.TextStyle = fyne.TextStyle{Bold: true}
				case 1:
					label.SetText("Название")
					label.TextStyle = fyne.TextStyle{Bold: true}
				case 2:
					label.SetText("Количество")
					label.TextStyle = fyne.TextStyle{Bold: true}
				}
				scrollContainer.Content.Refresh()
			} else {
				product := booking.Products[id.Row-1]
				switch id.Col {
				case 0:
					label.SetText(strconv.Itoa(product.CodeSTU))
					copyBtn := hBox.Objects[2].(*widget.Button)
					copyBtn.OnTapped = func() {
						w.Clipboard().SetContent(strconv.Itoa(product.CodeSTU))
					}
					copyBtn.Show()

				case 1:
					label.SetText(product.Name)
				case 2:
					label.SetText(strconv.Itoa(product.Quantity))
				}
				scrollContainer.Content.Refresh()
			}
		},
	)

	productsTable.SetColumnWidth(0, 115)
	productsTable.SetColumnWidth(1, 350)
	productsTable.SetColumnWidth(2, 105)

	infoContainer := container.NewVBox(
		idLabel,
		dateLabel,
		statusLabel,
		usernameLabel,
		phoneLabel,
		messageLabel,
		layout.NewSpacer(),
		statusButtons,
		layout.NewSpacer(),
		productsHeader,
	)

	content := container.NewBorder(
		infoContainer,
		nil,
		nil,
		nil,
		productsTable,
	)

	return container.NewVScroll(content)
}

func updateStatusButtons(status string, cont *fyne.Container) {
	var (
		btnConfirm = cont.Objects[0].(*widget.Button)
		btnReject  = cont.Objects[1].(*widget.Button)
		btnDone    = cont.Objects[2].(*widget.Button)
		btnReceive = cont.Objects[3].(*widget.Button)
	)

	switch status {
	case models.BookStatusCreated:
		btnDone.Disable()
		btnReceive.Disable()
		btnConfirm.Enable()
		btnReject.Enable()
	case models.BookStatusConfirmed:
		btnConfirm.Disable()
		btnReceive.Disable()
		btnDone.Enable()
	case models.BookStatusRejected:
		btnConfirm.Disable()
		btnReject.Disable()
		btnDone.Disable()
		btnReceive.Disable()
	case models.BookStatusDone:
		btnConfirm.Disable()
		btnDone.Disable()
		btnReject.Enable()
		btnReceive.Enable()
	case models.BookStatusReceive:
		btnConfirm.Disable()
		btnReject.Disable()
		btnDone.Disable()
		btnReceive.Disable()
	}

	cont.Refresh()
}

func getStatusDisplayName(status string) string {
	switch status {
	case models.BookStatusCreated:
		return "Создан"
	case models.BookStatusConfirmed:
		return "Подтверждён"
	case models.BookStatusRejected:
		return "Отклонён"
	case models.BookStatusDone:
		return "Готов к выдаче"
	case models.BookStatusReceive:
		return "Выдано"
	default:
		return status
	}
}
