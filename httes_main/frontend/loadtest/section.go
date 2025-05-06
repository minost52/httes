package loadtest

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Глобальные переменные для UI элементов
var methodSelect *widget.Select
var protocolSelect *widget.Select
var urlEntry *widget.Entry
var proxyEntry *widget.Entry
var reqCount *widget.Entry
var duration *widget.Entry
var loadType *widget.RadioGroup
var resultOutput *widget.TextGrid

func createURLSection() *fyne.Container {
	methodSelect = widget.NewSelect([]string{"GET", "POST"}, nil)
	methodSelect.SetSelected("GET")

	protocolSelect = widget.NewSelect([]string{"HTTP", "HTTPS"}, nil)
	protocolSelect.SetSelected("HTTPS")

	urlEntry = widget.NewEntry()
	urlEntry.SetText("example.com")

	urlSection := container.NewHBox(
		container.NewGridWrap(fyne.NewSize(80, methodSelect.MinSize().Height), methodSelect),
		container.NewGridWrap(fyne.NewSize(100, protocolSelect.MinSize().Height), protocolSelect),
		container.NewGridWrap(fyne.NewSize(250, urlEntry.MinSize().Height), urlEntry),
	)

	return urlSection
}

func createProxySection() *fyne.Container {
	proxyEntry = widget.NewEntry()
	proxyEntry.SetPlaceHolder("http://127.0.0.1:8080")
	return container.NewVBox(
		widget.NewLabel("Прокси (необязательно)"),
		container.NewGridWrap(fyne.NewSize(250, proxyEntry.MinSize().Height), proxyEntry),
	)
}

func createParamsSection() *fyne.Container {
	reqCount = widget.NewEntry()
	reqCount.SetText("100")

	duration = widget.NewEntry()
	duration.SetText("5")

	loadType = widget.NewRadioGroup([]string{"Linear", "Incremental", "Waved"}, nil)
	loadType.SetSelected("Linear")

	return container.NewHBox(
		container.NewGridWrap(
			fyne.NewSize(120, widget.NewLabel("Request Count*").MinSize().Height+reqCount.MinSize().Height),
			container.NewVBox(
				widget.NewLabel("Request Count*"),
				reqCount,
			),
		),
		container.NewGridWrap(
			fyne.NewSize(120, widget.NewLabel("Duration (s)*").MinSize().Height+duration.MinSize().Height),
			container.NewVBox(
				widget.NewLabel("Duration (s)*"),
				duration,
			),
		),
		container.NewGridWrap(
			fyne.NewSize(120, widget.NewLabel("Load Type*").MinSize().Height+loadType.MinSize().Height),
			container.NewVBox(
				widget.NewLabel("Load Type*"),
				loadType,
			),
		),
	)
}
