package controller

import (
	"context"
	"sort"
	"time"

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

	ma25 := modelMovingAverages(ds, 25)
	ma50 := modelMovingAverages(ds, 50)
	ma200 := modelMovingAverages(ds, 200)

	dsto := modelStochastics(ds)
	wsto := modelStochastics(ws)

	if len(ws) > maxDataWeeks {
		start := ws[len(ws)-maxDataWeeks:][0].Date

		trimTradingSessions := func(vs []*model.TradingSession) []*model.TradingSession {
			for i, v := range vs {
				if v.Date == start {
					return vs[i:]
				}
			}
			return vs
		}

		trimMovingAverageValues := func(vs []*model.MovingAverageValue) []*model.MovingAverageValue {
			for i, v := range vs {
				if v.Date == start {
					return vs[i:]
				}
			}
			return vs
		}

		trimStochasticValues := func(vs []*model.StochasticValue) []*model.StochasticValue {
			for i, v := range vs {
				if v.Date == start {
					return vs[i:]
				}
			}
			return vs
		}

		ds = trimTradingSessions(ds)
		ma25.Values = trimMovingAverageValues(ma25.Values)
		ma50.Values = trimMovingAverageValues(ma50.Values)
		ma200.Values = trimMovingAverageValues(ma200.Values)
		dsto.Values = trimStochasticValues(dsto.Values)
		wsto.Values = trimStochasticValues(wsto.Values)
	}

	return &model.StockUpdate{
		Symbol:            symbol,
		DailySessions:     ds,
		MovingAverage25:   ma25,
		MovingAverage50:   ma50,
		MovingAverage200:  ma200,
		DailyStochastics:  dsto,
		WeeklyStochastics: wsto,
	}
}

func trimmedMovingAverages(vs []*model.MovingAverageValue, start time.Time) []*model.MovingAverageValue {
	for i, v := range vs {
		if v.Date == start {
			return vs[i:]
		}
	}
	return vs
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

func modelMovingAverages(ts []*model.TradingSession, n int) *model.MovingAverages {
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
	for i, s := range ts {
		vs = append(vs, &model.MovingAverageValue{
			Date:    s.Date,
			Average: average(i, n),
		})
	}

	return &model.MovingAverages{Values: vs}
}

func modelStochastics(ts []*model.TradingSession) *model.Stochastics {
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
	var vs []*model.StochasticValue
	for i := range ts {
		vs = append(vs, &model.StochasticValue{Date: ts[i].Date})
	}

	// Calculate fast %D (slow %K) for stochastics.
	for i := range ts {
		if i+1 < k+d {
			continue
		}
		vs[i].K = (fastK[i] + fastK[i-1] + fastK[i-2]) / 3
	}

	// Calculate slow %D for stochastics.
	for i := range ts {
		if i+1 < k+d+d {
			continue
		}
		vs[i].D = (vs[i].K + vs[i-1].K + vs[i-2].K) / 3
	}

	return &model.Stochastics{Values: vs}
}
