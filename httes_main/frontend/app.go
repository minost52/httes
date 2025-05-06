package frontend

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
)

func Run() {
	a := app.New()
	a.Settings().SetTheme(theme.DarkTheme())
	w := a.NewWindow("Ddosify UI")
	w.Resize(fyne.NewSize(700, 500))

	// Инициализация страниц
	pages := initPages()
	contentContainer := container.NewVBox(pages[0].Content)

	// Создание бокового меню
	sidebar := createSidebar(pages, contentContainer)

	// Основной контейнер
	content := container.NewHSplit(sidebar, contentContainer)
	content.SetOffset(0.25)

	w.SetContent(content)
	w.ShowAndRun()
}
