package controller

import (
	"context"
	"sort"

	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/stock/iex"
)

const (
	maxDataWeeks = 12 /* months */ * 4 /* weeks = 1 year */
	k            = 10
	d            = 3
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
		update: modelStockUpdate(symbol, resp),
	}
}

func modelStockUpdate(symbol string, resp *iex.ListTradingSessionsResponse) *model.StockUpdate {
	ds := modelTradingSessions(resp.TradingSessions)
	ws := modelTradingSessions(weeklyTradingSessions(resp.TradingSessions))

	m25 := modelMovingAverages(ds, 25)
	m50 := modelMovingAverages(ds, 50)
	m200 := modelMovingAverages(ds, 200)

	dsto := modelStochastics(ds)
	wsto := modelStochastics(ws)

	if len(ws) > maxDataWeeks {
		start := ws[len(ws)-maxDataWeeks:][0].Date

		trimmedTradingSessions := func(vs []*model.TradingSession) []*model.TradingSession {
			for i, v := range vs {
				if v.Date == start {
					return vs[i:]
				}
			}
			return vs
		}

		trimmedMovingAverages := func(vs []*model.MovingAverage) []*model.MovingAverage {
			for i, v := range vs {
				if v.Date == start {
					return vs[i:]
				}
			}
			return vs
		}

		trimmedStochastics := func(vs []*model.Stochastic) []*model.Stochastic {
			for i, v := range vs {
				if v.Date == start {
					return vs[i:]
				}
			}
			return vs
		}

		ds = trimmedTradingSessions(ds)
		m25 = trimmedMovingAverages(m25)
		m50 = trimmedMovingAverages(m50)
		m200 = trimmedMovingAverages(m200)
		dsto = trimmedStochastics(dsto)
		wsto = trimmedStochastics(wsto)
	}

	return &model.StockUpdate{
		Symbol: symbol,
		DailyTradingSessionSeries:   &model.TradingSessionSeries{TradingSessions: ds},
		DailyMovingAverageSeries25:  &model.MovingAverageSeries{MovingAverages: m25},
		DailyMovingAverageSeries50:  &model.MovingAverageSeries{MovingAverages: m50},
		DailyMovingAverageSeries200: &model.MovingAverageSeries{MovingAverages: m200},
		DailyStochasticSeries:       &model.StochasticSeries{Stochastics: dsto},
		WeeklyStochasticSeries:      &model.StochasticSeries{Stochastics: wsto},
	}
}

func modelTradingSessions(ts []*iex.TradingSession) []*model.TradingSession {
	var ms []*model.TradingSession
	for _, s := range ts {
		ms = append(ms, &model.TradingSession{
			Date:          s.Date,
			Open:          s.Open,
			High:          s.High,
			Low:           s.Low,
			Close:         s.Close,
			Volume:        s.Volume,
			Change:        s.Change,
			PercentChange: s.ChangePercent,
		})
	}
	sort.Slice(ms, func(i, j int) bool {
		return ms[i].Date.Before(ms[j].Date)
	})
	return ms
}

func weeklyTradingSessions(ds []*iex.TradingSession) (ws []*iex.TradingSession) {
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

func modelMovingAverages(ts []*model.TradingSession, n int) []*model.MovingAverage {
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

	var ms []*model.MovingAverage
	for i := range ts {
		ms = append(ms, &model.MovingAverage{
			Date:  ts[i].Date,
			Value: average(i, n),
		})
	}
	return ms
}

func modelStochastics(ts []*model.TradingSession) []*model.Stochastic {
	// Calculate fast %K for stochastics.
	fastK := make([]float32, len(ts))
	for i := range ts {
		if i+1 < k {
			continue
		}

		highestHigh, lowestLow := ts[i].High, ts[i].Low
		for j := 0; j < k; j++ {
			if highestHigh < ts[i-j].High {
				highestHigh = ts[i-j].High
			}
			if lowestLow > ts[i-j].Low {
				lowestLow = ts[i-j].Low
			}
		}
		fastK[i] = (ts[i].Close - lowestLow) / (highestHigh - lowestLow)
	}

	// Setup slice to hold stochastics.
	var ms []*model.Stochastic
	for i := range ts {
		ms = append(ms, &model.Stochastic{Date: ts[i].Date})
	}

	// Calculate fast %D (slow %K) for stochastics.
	for i := range ts {
		if i+1 < k+d {
			continue
		}
		ms[i].K = (fastK[i] + fastK[i-1] + fastK[i-2]) / 3
	}

	// Calculate slow %D for stochastics.
	for i := range ts {
		if i+1 < k+d+d {
			continue
		}
		ms[i].D = (ms[i].K + ms[i-1].K + ms[i-2].K) / 3
	}

	return ms
}
