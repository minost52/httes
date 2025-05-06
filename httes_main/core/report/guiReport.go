package report

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"go.ddosify.com/ddosify/core/types"
)

const OutputTypeGui = "gui"

func init() {
	AvailableOutputServices[OutputTypeGui] = func(resultGrid *widget.TextGrid) ReportService {
		return NewGuiReportService(resultGrid)
	}
}

type guiReport struct {
	resultGrid *widget.TextGrid
	doneChan   chan struct{}
	result     *Result
	mu         sync.Mutex
	debug      bool
	closed     bool // Флаг для отслеживания состояния doneChan
}

type duration struct {
	name     string
	duration float32
	order    int
}

var keyToStr = map[string]duration{
	"dnsDuration":           {name: "DNS", order: 1},
	"connDuration":          {name: "Connection", order: 2},
	"tlsDuration":           {name: "TLS", order: 3},
	"reqDuration":           {name: "Request Write", order: 4},
	"serverProcessDuration": {name: "Server Processing", order: 5},
	"resDuration":           {name: "Response Read", order: 6},
	"duration":              {name: "Total", order: 7},
}

func NewGuiReportService(resultGrid *widget.TextGrid) ReportService {
	if resultGrid == nil {
		panic("resultGrid cannot be nil")
	}
	return &guiReport{
		resultGrid: resultGrid,
		doneChan:   make(chan struct{}),
		result: &Result{
			StepResults:    make(map[uint16]*ScenarioStepResultSummary),
			Durations:      make(map[string]float32),
			StatusCodeDist: make(map[int]int),
			ProgressPoints: make(map[int]float32),
		},
	}
}

func (r *guiReport) Init(debug bool) error {
	r.debug = debug
	return nil
}

func (r *guiReport) Start(input chan *types.ScenarioResult) {
	if input == nil {
		return
	}

	if r.debug {
		r.printInDebugMode(input)
		return
	}

	// Запускаем тикер для промежуточных уведомлений (каждую секунду)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Выводим заголовок TEST
	fmt.Println("TEST")
	fmt.Println("-------------------------------------")

	// Канал для завершения обработки результатов
	done := make(chan struct{})

	// Обрабатываем результаты в отдельной горутине
	go func() {
		count := 0
		for scr := range input {
			r.mu.Lock()
			aggregate(r.result, scr)
			r.printLiveResult()
			r.mu.Unlock()
			count++
		}
		close(done)
	}()

	// Основной цикл: вывод промежуточных уведомлений и ожидание завершения
	for {
		select {
		case <-ticker.C:
			// Промежуточное уведомление
			r.mu.Lock()
			if r.result.SuccessCount+r.result.FailedCount > 0 {
				fmt.Printf("✔️  Successful Run: %-6d %d%%       ❌ Failed Run: %-6d %d%%       ⏱️  Avg. Duration: %.5fs\n",
					r.result.SuccessCount, r.result.successPercentage(),
					r.result.FailedCount, r.result.failedPercentage(),
					r.result.AvgDuration)
			}
			r.mu.Unlock()

		case <-done:
			// Результаты обработаны, выводим итоговый отчёт
			r.mu.Lock()
			r.printDetails()
			r.mu.Unlock()
			select {
			case <-r.doneChan:
			default:
				r.doneChan <- struct{}{}
			}
			return
		}
	}
}

func (r *guiReport) DoneChan() <-chan struct{} {
	return r.doneChan
}

func (r *guiReport) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.closed {
		close(r.doneChan)
		r.closed = true
	}
}

func (r *guiReport) printLiveResult() {
	if r.resultGrid == nil || r.result == nil {
		return
	}

	b := strings.Builder{}
	b.WriteString("🚀 Live Test Results\n")
	b.WriteString("----------------------------------------------------\n")
	b.WriteString(fmt.Sprintf("✅ Successful Run: %-6d (%d%%)\n", r.result.SuccessCount, r.result.successPercentage()))
	b.WriteString(fmt.Sprintf("❌ Failed Run: %-6d (%d%%)\n", r.result.FailedCount, r.result.failedPercentage()))
	b.WriteString(fmt.Sprintf("⏳ Avg. Duration: %.5fs\n", r.result.AvgDuration))
	b.WriteString("----------------------------------------------------\n")

	fyne.Do(func() {
		if r.debug {
			fmt.Println("Updating TextGrid (live) with content:", b.String())
		}
		r.resultGrid.SetText(b.String())
	})
}

func (r *guiReport) printDetails() {
	if r.resultGrid == nil || r.result == nil {
		return
	}

	// Формируем консольный вывод
	bConsole := strings.Builder{}
	bConsole.WriteString("\nRESULT\n")
	bConsole.WriteString("-------------------------------------\n")
	bConsole.WriteString(fmt.Sprintf("Success Count:    %-6d (%d%%)\n", r.result.SuccessCount, r.result.successPercentage()))
	bConsole.WriteString(fmt.Sprintf("Failed Count:     %-6d (%d%%)\n", r.result.FailedCount, r.result.failedPercentage()))

	// Всегда выводим Durations (Avg), даже если данные отсутствуют
	bConsole.WriteString("\nDurations (Avg):\n")
	var durationList = make([]duration, 0)
	for d, s := range r.result.Durations {
		dur, ok := keyToStr[d]
		if !ok {
			dur = duration{name: d, order: 999}
		}
		dur.duration = s
		durationList = append(durationList, dur)
	}
	// Добавляем все ожидаемые длительности с нулевыми значениями, если их нет
	for _, dur := range keyToStr {
		found := false
		for _, d := range durationList {
			if d.name == dur.name {
				found = true
				break
			}
		}
		if !found {
			durationList = append(durationList, duration{name: dur.name, duration: 0, order: dur.order})
		}
	}
	sort.Slice(durationList, func(i, j int) bool {
		return durationList[i].order < durationList[j].order
	})
	for _, v := range durationList {
		bConsole.WriteString(fmt.Sprintf("  %-20s:%.4fs\n", v.name, v.duration))
	}

	if len(r.result.StatusCodeDist) > 0 {
		bConsole.WriteString("\nStatus Code (Message) :Count\n")
		keys := make([]int, 0, len(r.result.StatusCodeDist))
		for k := range r.result.StatusCodeDist {
			keys = append(keys, k)
		}
		sort.Ints(keys)
		for _, s := range keys {
			c := r.result.StatusCodeDist[s]
			bConsole.WriteString(fmt.Sprintf("  %-20s:%d\n", fmt.Sprintf("%d (%s)", s, http.StatusText(s)), c))
		}
	}

	// Выводим в консоль
	fmt.Print(bConsole.String())

	// Формируем вывод для GUI (с Avg. Parameter Count)
	bGui := strings.Builder{}
	// Промежуточные уведомления (для GUI)
	progressKeys := make([]int, 0, len(r.result.ProgressPoints))
	for k := range r.result.ProgressPoints {
		progressKeys = append(progressKeys, k)
	}
	sort.Ints(progressKeys)
	for _, milestone := range progressKeys {
		if milestone <= r.result.SuccessCount {
			avgDuration := r.result.ProgressPoints[milestone]
			bGui.WriteString(fmt.Sprintf("✔️  Successful Run: %-6d %d%%       ❌ Failed Run: %-6d %d%%       ⏱️  Avg. Duration: %.5fs\n",
				milestone, r.result.successPercentage(), r.result.FailedCount, r.result.failedPercentage(), avgDuration))
		}
	}

	bGui.WriteString(bConsole.String())

	// Добавляем Avg. Parameter Count для GUI
	avgParamCount := float32(0)
	if r.result.TotalRequests > 0 {
		avgParamCount = float32(r.result.TotalParamCount) / float32(r.result.TotalRequests)
	}
	bGui.WriteString(fmt.Sprintf("\nAvg. Parameter Count: %.2f\n", avgParamCount))

	fyne.Do(func() {
		if r.debug {
			fmt.Println("Updating TextGrid (details) with content:", bGui.String())
		}
		r.resultGrid.SetText(bGui.String())
	})
}

func (r *guiReport) printInDebugMode(input chan *types.ScenarioResult) {
	if input == nil {
		fmt.Println("printInDebugMode: ERROR: input channel is nil")
		return
	}

	b := strings.Builder{}
	b.WriteString("🐞 Debug Mode\n")
	b.WriteString("----------------------------------------------------\n")

	count := 0
	for scr := range input {
		fmt.Printf("Debug: Processing ScenarioResult #%d: %+v\n", count, scr)
		r.mu.Lock()
		aggregate(r.result, scr)

		for _, sr := range scr.StepResults {
			fmt.Printf("Debug: StepResult: StepID=%d, StatusCode=%d, Duration=%v, Err=%+v\n",
				sr.StepID, sr.StatusCode, sr.Duration, sr.Err)
			verboseInfo := ScenarioStepResultToVerboseHttpRequestInfo(sr)
			b.WriteString(fmt.Sprintf("\n\nSTEP (%d) %s\n", verboseInfo.StepId, verboseInfo.StepName))
			b.WriteString("------------------------------------\n")
			b.WriteString("- Environment Variables\n")
			for eKey, eVal := range verboseInfo.Envs {
				switch eVal.(type) {
				case map[string]interface{}, []string, []float64, []bool:
					valPretty, _ := json.Marshal(eVal)
					b.WriteString(fmt.Sprintf("  %s: %s\n", eKey, valPretty))
				default:
					b.WriteString(fmt.Sprintf("  %s: %v\n", eKey, eVal))
				}
			}

			if verboseInfo.Error != "" && isVerboseInfoRequestEmpty(verboseInfo.Request) {
				b.WriteString(fmt.Sprintf("\n⚠️ Error: %s\n", verboseInfo.Error))
				continue
			}

			b.WriteString("\n- Request\n")
			b.WriteString(fmt.Sprintf("  Target: %s\n", verboseInfo.Request.Url))
			b.WriteString(fmt.Sprintf("  Method: %s\n", verboseInfo.Request.Method))
			b.WriteString("  Headers:\n")
			for hKey, hVal := range verboseInfo.Request.Headers {
				b.WriteString(fmt.Sprintf("    %s: %s\n", hKey, hVal))
			}

			contentType := sr.DebugInfo["requestHeaders"].(http.Header).Get("content-type")
			b.WriteString("  Body: ")
			if verboseInfo.Request.Body == nil {
				b.WriteString("null\n")
			} else if strings.Contains(contentType, "application/json") {
				valPretty, _ := json.MarshalIndent(verboseInfo.Request.Body, "    ", "  ")
				b.WriteString(fmt.Sprintf("\n    %s\n", valPretty))
			} else {
				b.WriteString(fmt.Sprintf("%v\n", verboseInfo.Request.Body))
			}

			if verboseInfo.Error != "" {
				if len(verboseInfo.FailedCaptures) > 0 {
					b.WriteString("\n- Failed Captures\n")
					for wKey, wVal := range verboseInfo.FailedCaptures {
						b.WriteString(fmt.Sprintf("    %s: %s\n", wKey, wVal))
					}
				}
				b.WriteString(fmt.Sprintf("\n⚠️ Error: %s\n", verboseInfo.Error))
			} else {
				b.WriteString("\n- Response\n")
				b.WriteString(fmt.Sprintf("  StatusCode: %d\n", verboseInfo.Response.StatusCode))
				b.WriteString("  Headers:\n")
				for hKey, hVal := range verboseInfo.Response.Headers {
					b.WriteString(fmt.Sprintf("    %s: %s\n", hKey, hVal))
				}

				contentType := sr.DebugInfo["responseHeaders"].(http.Header).Get("content-type")
				b.WriteString("  Body: ")
				if verboseInfo.Response.Body == nil {
					b.WriteString("null\n")
				} else if strings.Contains(contentType, "application/json") {
					valPretty, _ := json.MarshalIndent(verboseInfo.Response.Body, "    ", "  ")
					b.WriteString(fmt.Sprintf("\n    %s\n", valPretty))
				} else {
					b.WriteString(fmt.Sprintf("%v\n", verboseInfo.Response.Body))
				}

				if len(verboseInfo.FailedCaptures) > 0 {
					b.WriteString("\n- Failed Captures\n")
					for wKey, wVal := range verboseInfo.FailedCaptures {
						b.WriteString(fmt.Sprintf("    %s: %s\n", wKey, wVal))
					}
				}
			}
		}

		fyne.Do(func() {
			fmt.Println("Debug: Updating TextGrid with content:", b.String())
			r.resultGrid.SetText(b.String())
		})
		r.mu.Unlock()
		count++
	}

	fmt.Println("Debug: Finished processing input channel")
	r.mu.Lock()
	r.printDetails()
	r.mu.Unlock()
	select {
	case <-r.doneChan:
		fmt.Println("doneChan already closed, skipping signal")
	default:
		fmt.Println("Debug: Sending done signal")
		r.doneChan <- struct{}{}
	}
}
