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
			Symbol:            symbol,
			DailySessions:     convertTradingSessions(resp.TradingSessions),
			MovingAverage25:   makeMovingAverage(resp.TradingSessions, 25),
			MovingAverage50:   makeMovingAverage(resp.TradingSessions, 50),
			MovingAverage200:  makeMovingAverage(resp.TradingSessions, 200),
			DailyStochastics:  makeStochastics(resp.TradingSessions),
			WeeklyStochastics: makeStochastics(weeklySessions(resp.TradingSessions)),
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

func makeStochastics(ss []*iex.TradingSession) *model.Stochastics {
	const (
		k = 10
		d = 3
	)

	// Calculate fast %K for stochastics.
	fastK := make([]float32, len(ss))
	for i := range ss {
		if i+1 < k {
			continue
		}

		highestHigh, lowestLow := ss[i].High, ss[i].Low
		for j := 0; j < k; j++ {
			if highestHigh < ss[i-j].High {
				highestHigh = ss[i-j].High
			}
			if lowestLow > ss[i-j].Low {
				lowestLow = ss[i-j].Low
			}
		}
		fastK[i] = (ss[i].Close - lowestLow) / (highestHigh - lowestLow)
	}

	var vs []*model.StochasticValue
	for i := range ss {
		vs = append(vs, &model.StochasticValue{Date: ss[i].Date})
	}

	// Calculate fast %D (slow %K) for stochastics.
	for i := range ss {
		if i+1 < k+d {
			continue
		}
		vs[i].K = (fastK[i] + fastK[i-1] + fastK[i-2]) / 3
	}

	// Calculate slow %D for stochastics.
	for i := range ss {
		if i+1 < k+d+d {
			continue
		}
		vs[i].D = (vs[i].K + vs[i-1].K + vs[i-2].K) / 3
	}

	return &model.Stochastics{Values: vs}
}

func weeklySessions(ds []*iex.TradingSession) (ws []*iex.TradingSession) {
	for _, s := range ds {
		diffWeek := ws == nil
		if !diffWeek {
			_, week := s.Date.ISOWeek()
			_, prevWeek := ws[len(ws)-1].Date.ISOWeek()
			diffWeek = week != prevWeek
		}

		if diffWeek {
			sc := *s
			ws = append(ws, &sc)
		} else {
			ls := ws[len(ws)-1]
			if ls.High < s.High {
				ls.High = s.High
			}
			if ls.Low > s.Low {
				ls.Low = s.Low
			}
			ls.Close = s.Close
			ls.Volume += s.Volume
		}
	}
	return ws
}
