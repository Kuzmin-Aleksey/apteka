package main

import (
	"flag"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	var (
		flagTitle   = new(string)
		flagContent = new(string)
	)

	flag.StringVar(flagTitle, "title", "", "")
	flag.StringVar(flagContent, "content", "", "")

	flag.Parse()

	showNotify(*flagTitle, *flagContent)
}

func showNotify(title string, content string) {
	a := app.New()
	a.Settings().SetTheme(greenTheme{})

	w := a.NewWindow("Уведомление")
	w.Resize(fyne.NewSize(250, 110))

	w.SetContent(container.NewVBox(
		container.NewCenter(widget.NewRichTextFromMarkdown("## "+title)),
		container.NewCenter(widget.NewLabel(content)),
	))

	w.ShowAndRun()
}
