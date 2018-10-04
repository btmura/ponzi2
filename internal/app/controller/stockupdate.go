package controller

import (
	"sort"

	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/stock/iex"
)

// maxDataWeeks is maximum number of weeks of data to retain.
const maxDataWeeks = 12 /* months */ * 4 /* weeks = 1 year */

// Stochastic parameters.
const (
	k = 10
	d = 3
)

func modelStockUpdate(st *iex.Stock) *model.StockUpdate {
	ds := modelTradingSessions(st.Chart)
	ws := modelTradingSessions(weeklyChartPoints(st.Chart))

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
		Symbol: st.Symbol,
		Quote:  modelQuote(st.Quote),
		DailyTradingSessionSeries:   &model.TradingSessionSeries{TradingSessions: ds},
		DailyMovingAverageSeries25:  &model.MovingAverageSeries{MovingAverages: m25},
		DailyMovingAverageSeries50:  &model.MovingAverageSeries{MovingAverages: m50},
		DailyMovingAverageSeries200: &model.MovingAverageSeries{MovingAverages: m200},
		DailyStochasticSeries:       &model.StochasticSeries{Stochastics: dsto},
		WeeklyStochasticSeries:      &model.StochasticSeries{Stochastics: wsto},
	}
}

func modelQuote(q *iex.Quote) *model.Quote {
	return &model.Quote{
		CompanyName:   q.CompanyName,
		LatestPrice:   q.LatestPrice,
		LatestSource:  q.LatestSource,
		LatestTime:    q.LatestTime,
		LatestUpdate:  q.LatestUpdate,
		LatestVolume:  q.LatestVolume,
		Open:          q.Open,
		High:          q.High,
		Low:           q.Low,
		Close:         q.Close,
		Change:        q.Change,
		ChangePercent: q.ChangePercent,
	}
}

func modelTradingSessions(ps []*iex.ChartPoint) []*model.TradingSession {
	var ts []*model.TradingSession
	for _, p := range ps {
		ts = append(ts, &model.TradingSession{
			Date:          p.Date,
			Open:          p.Open,
			High:          p.High,
			Low:           p.Low,
			Close:         p.Close,
			Volume:        p.Volume,
			Change:        p.Change,
			PercentChange: p.ChangePercent,
		})
	}
	sort.Slice(ts, func(i, j int) bool {
		return ts[i].Date.Before(ts[j].Date)
	})
	return ts
}

func weeklyChartPoints(ps []*iex.ChartPoint) (ws []*iex.ChartPoint) {
	for _, p := range ps {
		diffWeek := ws == nil
		if !diffWeek {
			_, week := p.Date.ISOWeek()
			_, prevWeek := ws[len(ws)-1].Date.ISOWeek()
			diffWeek = week != prevWeek
		}

		if diffWeek {
			pcopy := *p
			ws = append(ws, &pcopy)
		} else {
			ls := ws[len(ws)-1]
			if ls.High < p.High {
				ls.High = p.High
			}
			if ls.Low > p.Low {
				ls.Low = p.Low
			}
			ls.Close = p.Close
			ls.Volume += p.Volume
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
