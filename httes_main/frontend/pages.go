package frontend

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"go.ddosify.com/ddosify/frontend/loadtest"
)

type Page struct {
	Name    string
	Button  *widget.Button
	Content fyne.CanvasObject
}

func initPages() []*Page {
	loadTestContent := loadtest.CreateLoadTestContent()
	emptyContent := widget.NewLabel("Пустая страница")

	return []*Page{
		{Name: "LoadTest", Content: loadTestContent},
		{Name: "Images", Content: emptyContent},
		{Name: "Volumes", Content: emptyContent},
	}
}

func createSidebar(pages []*Page, contentContainer *fyne.Container) fyne.CanvasObject {
	var currentPage *Page
	sidebarButtons := []fyne.CanvasObject{}

	updateButtonStyles := func(activePage *Page) {
		for _, page := range pages {
			if page == activePage {
				page.Button.Importance = widget.HighImportance
			} else {
				page.Button.Importance = widget.MediumImportance
			}
			page.Button.Refresh()
		}
	}

	for _, page := range pages {
		page := page // захват переменной
		btn := widget.NewButton(page.Name, func() {
			contentContainer.Objects = []fyne.CanvasObject{page.Content}
			contentContainer.Refresh()
			currentPage = page
			updateButtonStyles(page)
		})
		page.Button = btn
		sidebarButtons = append(sidebarButtons, btn)
	}

	// Установка начального состояния
	currentPage = pages[0]
	updateButtonStyles(currentPage)

	return container.NewVBox(sidebarButtons...)
}
