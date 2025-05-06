package report

import (
	"fmt"

	"fyne.io/fyne/v2/widget"
	"go.ddosify.com/ddosify/core/types"
)

var AvailableOutputServices = make(map[string]func(*widget.TextGrid) ReportService)

type ReportService interface {
	DoneChan() <-chan struct{}
	Init(debug bool) error
	Start(input chan *types.ScenarioResult)
	Stop() // Добавляем метод Stop
}

func NewReportService(s string, resultGrid *widget.TextGrid) (ReportService, error) {
	if constructor, ok := AvailableOutputServices[s]; ok {
		return constructor(resultGrid), nil // Передаем GUI
	}
	return nil, fmt.Errorf("unsupported output type: %s", s)
}
