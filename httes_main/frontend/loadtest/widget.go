package loadtest

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func CreateLoadTestContent() fyne.CanvasObject {
	title := widget.NewLabel("Httes")
	title.TextStyle = fyne.TextStyle{Bold: true}

	description := widget.NewLabel("Модуль нагрузочного тестирования для определения таймингов HTTP, HTTPS запросов.")

	// Создание секций UI
	urlSection := createURLSection()
	proxySection := createProxySection()
	paramsSection := createParamsSection()
	buttons := createButtons()
	resultOutput = widget.NewTextGrid()

	// Контейнер формы
	form := container.NewVBox(
		urlSection,
		proxySection,
		widget.NewSeparator(),
		paramsSection,
		buttons,
	)

	return container.NewVBox(
		title,
		description,
		form,
		resultOutput,
	)
}
