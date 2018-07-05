package controller

import (
	"context"
	"sort"

	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/stock/iex"
)

// controllerStockUpdate bundles a stock and new data for that stock.
type controllerStockUpdate struct {
	// symbol is the stock's symbol.
	symbol string

	// update is the new data for the stock. Nil if an error happened.
	update *model.StockUpdate

	// updateErr is the error getting the update. Nil if no error happened.
	updateErr error
}

func (c *Controller) stockUpdate(ctx context.Context, symbol string) controllerStockUpdate {
	req := &iex.ListTradingSessionsRequest{Symbol: symbol}
	resp, err := c.stockDataFetcher.ListTradingSessions(ctx, req)
	if err != nil {
		return controllerStockUpdate{
			symbol:    symbol,
			updateErr: err,
		}
	}
	return controllerStockUpdate{
		symbol: symbol,
		update: &model.StockUpdate{
			Symbol:           symbol,
			DailySessions:    convertTradingSessions(resp.TradingSessions),
			MovingAverage25:  makeMovingAverage(resp.TradingSessions, 25),
			MovingAverage50:  makeMovingAverage(resp.TradingSessions, 50),
			MovingAverage200: makeMovingAverage(resp.TradingSessions, 200),
		},
	}
}

func convertTradingSessions(ts []*iex.TradingSession) []*model.TradingSession {
	var ms []*model.TradingSession
	for _, t := range ts {
		ms = append(ms, &model.TradingSession{
			Date:          t.Date,
			Open:          t.Open,
			High:          t.High,
			Low:           t.Low,
			Close:         t.Close,
			Volume:        t.Volume,
			Change:        t.Change,
			PercentChange: t.ChangePercent,
		})
	}
	sort.Slice(ms, func(i, j int) bool {
		return ms[i].Date.Before(ms[j].Date)
	})
	return ms
}

func makeMovingAverage(ts []*iex.TradingSession, n int) *model.MovingAverage {
	average := func(i, n int) (avg float32) {
		if i+1-n < 0 {
			return 0 // Not enough data
		}
		var sum float32
		for j := 0; j < n; j++ {
			sum += ts[i-j].Close
		}
		return sum / float32(n)
	}

	var vs []*model.MovingAverageValue
	for i, t := range ts {
		vs = append(vs, &model.MovingAverageValue{
			Date:    t.Date,
			Average: average(i, n),
		})
	}

	return &model.MovingAverage{Values: vs}
}
