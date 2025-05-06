package loadtest

import (
	"context"
	"fmt"

	"go.ddosify.com/ddosify/core"
	"go.ddosify.com/ddosify/core/report"
	"go.ddosify.com/ddosify/core/types"
)

func RunLoadTest(ctx context.Context, cancel context.CancelFunc, h types.Heart, rs report.ReportService) error {
	e, err := core.NewEngine(ctx, h, rs)
	if err != nil {
		return err
	}
	if err = e.Init(); err != nil {
		return err
	}
	resultChan := e.GetResultChan()
	if resultChan == nil {
		return fmt.Errorf("resultChan is nil")
	}
	// Запуск reportService в отдельной горутине
	go rs.Start(resultChan)
	// Проверка контекста перед запуском
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	// Запуск engine в текущей горутине
	e.Start()
	return nil
}
