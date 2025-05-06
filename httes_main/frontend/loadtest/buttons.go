package loadtest

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"go.ddosify.com/ddosify/core/proxy"
	"go.ddosify.com/ddosify/core/report"
	"go.ddosify.com/ddosify/core/types"
)

var isRunning bool
var testCtx context.Context
var testCancel context.CancelFunc
var reportService report.ReportService // Сохраняем reportService для вызова Stop()

// showErrorDialog отображает диалоговое окно с ошибкой
func showErrorDialog(msg string) {
	fyne.Do(func() {
		dialog := widget.NewLabel(msg)
		w := fyne.CurrentApp().NewWindow("Error")
		w.SetContent(container.NewVBox(dialog))
		w.Show()
	})
}

// setupStartButton настраивает обработчик для кнопки "Start"
func setupStartButton(startBtn, stopBtn *widget.Button, debugCheck *widget.Check) {
	startBtn.OnTapped = func() {
		if isRunning {
			showErrorDialog("Тест уже запущен!")
			return
		}
		isRunning = true
		defer func() { isRunning = false }()

		// Создаём контекст для теста
		testCtx, testCancel = context.WithCancel(context.Background())

		// Формируем URL
		targetURL := strings.ToLower(protocolSelect.Selected) + "://" + urlEntry.Text
		parsedURL, err := url.ParseRequestURI(targetURL)
		if err != nil || parsedURL.Host == "" {
			showErrorDialog("Target URL is invalid: " + targetURL)
			return
		}

		// Формируем шаг сценария
		step := types.ScenarioStep{
			ID:      1,
			Method:  strings.ToUpper(methodSelect.Selected),
			URL:     targetURL,
			Timeout: 30,
			Headers: map[string]string{},
		}
		scenario := types.Scenario{Steps: []types.ScenarioStep{step}}

		// Настройка прокси
		var proxyAddr *url.URL
		if proxyEntry.Text != "" {
			proxyAddr, err = url.Parse(proxyEntry.Text)
			if err != nil {
				showErrorDialog("Invalid proxy URL")
				return
			}
		}
		p := proxy.Proxy{
			Strategy: proxy.ProxyTypeSingle,
			Addr:     proxyAddr,
		}

		// Создаём объект Heart
		h := types.Heart{
			IterationCount:    parseInt(reqCount.Text),
			LoadType:          strings.ToLower(loadType.Selected),
			TestDuration:      parseInt(duration.Text),
			Scenario:          scenario,
			Proxy:             p,
			ReportDestination: "gui",
			Debug:             debugCheck.Checked,
		}

		// Обновляем GUI
		fyne.Do(func() {
			startBtn.Disable()
			stopBtn.Enable()
			resultOutput.SetText("🚀 Тест запущен...")
		})

		// Создаём resultGrid для отчета
		resultGrid := widget.NewTextGrid()
		resultGrid.SetText("TEST: GUI initialization")

		// Запуск теста в фоне
		go func() {
			var err error
			reportService, err = report.NewReportService(h.ReportDestination, resultGrid)
			if err != nil {
				showErrorDialog("Ошибка создания сервиса отчета: " + err.Error())
				fyne.Do(func() {
					startBtn.Enable()
					stopBtn.Disable()
				})
				return
			}
			err = RunLoadTest(testCtx, testCancel, h, reportService)
			if err != nil {
				showErrorDialog("Ошибка тестирования, RunLoadTest: " + err.Error())
			}

			fyne.Do(func() {
				startBtn.Enable()
				stopBtn.Disable()
				resultOutput.SetText("✅ Тест завершен!")
			})
		}()
	}
}

// setupStopButton настраивает обработчик для кнопки "Stop"
func setupStopButton(startBtn, stopBtn *widget.Button) {
	stopBtn.OnTapped = func() {
		if testCancel != nil {
			testCancel()
		}
		if reportService != nil {
			reportService.Stop()
		}

		fyne.Do(func() {
			startBtn.Enable()
			stopBtn.Disable()
			resultOutput.SetText("🛑 Тест остановлен пользователем.")
		})
	}
}

// createButtons создаёт контейнер с кнопками и чекбоксом Debug
func createButtons() *fyne.Container {
	// Создание кнопок
	startBtn := widget.NewButton("Start Load Test", nil)
	stopBtn := widget.NewButton("Stop", nil)

	// Создание чекбокса для Debug
	debugCheck := widget.NewCheck("Debug Mode", nil)
	debugCheck.Checked = false // Отключаем дебаг для проверки

	// Отключаем кнопку Stop в начале
	stopBtn.Disable()

	// Настраиваем обработчики
	setupStartButton(startBtn, stopBtn, debugCheck)
	setupStopButton(startBtn, stopBtn)

	// Возвращаем контейнер с кнопками и чекбоксом
	return container.NewHBox(startBtn, stopBtn, debugCheck)
}

func parseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		fmt.Printf("Invalid number %s: %v\n", s, err)
		return 0 // Или выбросить ошибку
	}
	return i
}
