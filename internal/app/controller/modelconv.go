package controller

import (
	"sort"
	"time"

	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/errs"
	"github.com/btmura/ponzi2/internal/logger"
	"github.com/btmura/ponzi2/internal/stock/iex"
)

// maxDataWeeks is maximum number of weeks of data to retain.
const maxDataWeeks = 12 /* months */ * 4 /* weeks = 1 year */

func modelIntradayChart(chart *iex.Chart) *model.Chart {
	var ts []*model.TradingSession
	for _, p := range chart.ChartPoints {
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

	return &model.Chart{
		Interval: model.Intraday,
		TradingSessionSeries: &model.TradingSessionSeries{
			TradingSessions: ts,
		},
	}
}

func modelDailyChart(quote *iex.Quote, chart *iex.Chart) *model.Chart {
	ds := modelTradingSessions(quote, chart)
	ws := weeklyModelTradingSessions(ds)
	m8 := modelExponentialMovingAverages(ds, 8)
	m21 := modelExponentialMovingAverages(ds, 21)
	m50 := modelSimpleMovingAverages(ds, 50)
	m200 := modelSimpleMovingAverages(ds, 200)
	v50 := modelAverageVolumes(ds, 50)

	if len(ws) > maxDataWeeks {
		start := ws[len(ws)-maxDataWeeks:][0].Date
		ds = trimmedTradingSessions(ds, start)
		m8 = trimmedMovingAverages(m8, start)
		m21 = trimmedMovingAverages(m21, start)
		m50 = trimmedMovingAverages(m50, start)
		m200 = trimmedMovingAverages(m200, start)
		v50 = trimmedAverageVolumes(v50, start)
	}

	return &model.Chart{
		Interval:             model.Daily,
		TradingSessionSeries: &model.TradingSessionSeries{TradingSessions: ds},
		MovingAverageSeriesSet: []*model.MovingAverageSeries{
			{Type: model.Exponential, Intervals: 8, Values: m8},
			{Type: model.Exponential, Intervals: 21, Values: m21},
			{Type: model.Simple, Intervals: 50, Values: m50},
			{Type: model.Simple, Intervals: 200, Values: m200},
		},
		AverageVolumeSeries: &model.AverageVolumeSeries{Values: v50},
	}
}

func modelWeeklyChart(quote *iex.Quote, chart *iex.Chart) *model.Chart {
	ds := modelTradingSessions(quote, chart)
	ws := weeklyModelTradingSessions(ds)

	m10 := modelSimpleMovingAverages(ws, 10)
	m40 := modelSimpleMovingAverages(ws, 40)

	v10 := modelAverageVolumes(ws, 10)

	return &model.Chart{
		Interval:             model.Weekly,
		TradingSessionSeries: &model.TradingSessionSeries{TradingSessions: ws},
		MovingAverageSeriesSet: []*model.MovingAverageSeries{
			{Type: model.Simple, Intervals: 10, Values: m10},
			{Type: model.Simple, Intervals: 40, Values: m40},
		},
		AverageVolumeSeries: &model.AverageVolumeSeries{Values: v10},
	}
}

func modelQuote(q *iex.Quote) (*model.Quote, error) {
	if q == nil {
		return nil, errs.Errorf("missing quote")
	}

	return &model.Quote{
		CompanyName:   q.CompanyName,
		LatestPrice:   q.LatestPrice,
		LatestSource:  modelSource(q.LatestSource),
		LatestTime:    q.LatestTime,
		LatestUpdate:  q.LatestUpdate,
		LatestVolume:  q.LatestVolume,
		Open:          q.Open,
		High:          q.High,
		Low:           q.Low,
		Close:         q.Close,
		Change:        q.Change,
		ChangePercent: q.ChangePercent,
	}, nil
}

func modelSource(src iex.Source) model.Source {
	switch src {
	case iex.SourceUnspecified:
		return model.SourceUnspecified
	case iex.RealTimePrice:
		return model.RealTimePrice
	case iex.FifteenMinuteDelayedPrice:
		return model.FifteenMinuteDelayedPrice
	case iex.Close:
		return model.Close
	case iex.PreviousClose:
		return model.PreviousClose
	case iex.Price:
		return model.Price
	case iex.LastTrade:
		return model.LastTrade
	default:
		logger.Errorf("unrecognized iex source: %v", src)
		return model.SourceUnspecified
	}
}

func modelTradingSessions(quote *iex.Quote, chart *iex.Chart) []*model.TradingSession {
	var ts []*model.TradingSession

	for _, p := range chart.ChartPoints {
		if p.Open <= 0 || p.High <= 0 || p.Low <= 0 || p.Close <= 0 {
			logger.Errorf("skipping bad data for %s: %v", chart.Symbol, p)
			continue
		}

		ts = append(ts, &model.TradingSession{
			Date:          p.Date,
			Source:        model.Close,
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

	// Add a trading session for the current quote if we do not have data
	// for today's trading session, so that the chart includes the latest quote.

	if quote == nil {
		return ts
	}

	q := quote

	// Real-time quotes won't have OHLC set, but they will have a latest price.
	// Fake OHLC so something shows up on the chart by using the latest price.
	o, h, l, c := q.Open, q.High, q.Low, q.Close
	if o == 0 && h == 0 && l == 0 && c == 0 {
		o = q.LatestPrice - q.Change
		c = q.LatestPrice

		l = o
		if l > c {
			l = c
		}

		h = o
		if h < c {
			h = c
		}
	}

	t := &model.TradingSession{
		Date:          q.LatestTime,
		Source:        modelSource(q.LatestSource),
		Open:          o,
		High:          h,
		Low:           l,
		Close:         c,
		Volume:        q.LatestVolume,
		Change:        q.Change,
		PercentChange: q.ChangePercent,
	}

	if len(ts) == 0 {
		return []*model.TradingSession{t}
	}

	sameDate := func(t1, t2 time.Time) bool {
		return t1.Year() == t2.Year() && t1.Month() == t2.Month() && t1.Day() == t2.Day()
	}

	if !sameDate(t.Date, ts[len(ts)-1].Date) {
		ts = append(ts, t)
	}

	// Run through the final trading sessions and recalculate price and volume changes.
	for i, t := range ts {
		if i == 0 {
			continue
		}

		// Calculate the change and percent change as IEX sometimes has incorrect data!
		pc := ts[i-1].Close
		t.Change = t.Close - pc
		t.PercentChange = t.Change / pc * 100.0

		// Calculate volume percent change, since IEX does not have this field.
		pv := ts[i-1].Volume
		t.VolumePercentChange = float32(t.Volume-pv) / float32(pv) * 100.0
	}

	return ts
}

func weeklyModelTradingSessions(ds []*model.TradingSession) (ws []*model.TradingSession) {
	for _, p := range ds {
		// Append if empty series.
		if len(ws) == 0 {
			pCopy := *p
			ws = append(ws, &pCopy)
			continue
		}

		// Append if different week as previous.
		_, week := p.Date.ISOWeek()
		_, prevWeek := ws[len(ws)-1].Date.ISOWeek()
		if week != prevWeek {
			pCopy := *p
			ws = append(ws, &pCopy)
			continue
		}

		// Combine if same week as previous.
		ls := ws[len(ws)-1]
		if ls.High < p.High {
			ls.High = p.High
		}
		if ls.Low > p.Low {
			ls.Low = p.Low
		}
		ls.Source = p.Source
		ls.Close = p.Close
		ls.Volume += p.Volume
		ls.Change = ls.Close - ls.Open
		if len(ws)-2 >= 0 {
			prev := ws[len(ws)-2]
			ls.PercentChange = (ls.Close - prev.Close) / prev.Close
		} else {
			ls.PercentChange = 0
		}
	}
	return ws
}

func modelExponentialMovingAverages(ts []*model.TradingSession, n int) []*model.MovingAverageValue {
	var values []*model.MovingAverageValue

	smoothing := 2.0 / (float32(n) + 1.0)

	value := func(i int) (avg float32) {
		var prevEMA float32
		switch {
		case i < n:
			// Not enough points to calculate SMA.
			return 0

		case i == n:
			// Use yesterday's SMA for today's previous EMA.
			var sum float32
			for j := 0; j < n; j++ {
				sum += ts[i-1-j].Close
			}
			prevEMA = sum / float32(n)

		default:
			// Use prev EMA.
			prevEMA = values[i-1].Value
		}
		return ts[i].Close*smoothing + prevEMA*(1-smoothing)
	}

	for i := range ts {
		values = append(values, &model.MovingAverageValue{
			Date:  ts[i].Date,
			Value: value(i),
		})
	}
	return values
}

func modelSimpleMovingAverages(ts []*model.TradingSession, n int) []*model.MovingAverageValue {
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

	var ms []*model.MovingAverageValue
	for i := range ts {
		ms = append(ms, &model.MovingAverageValue{
			Date:  ts[i].Date,
			Value: average(i, n),
		})
	}
	return ms
}

func modelAverageVolumes(ts []*model.TradingSession, n int) []*model.AverageVolumeValue {
	average := func(i, n int) (avg float32) {
		if i+1-n < 0 {
			return 0 // Not enough data
		}
		var sum float32
		for j := 0; j < n; j++ {
			sum += float32(ts[i-j].Volume)
		}
		return sum / float32(n)
	}

	var vs []*model.AverageVolumeValue
	for i := range ts {
		vs = append(vs, &model.AverageVolumeValue{
			Date:  ts[i].Date,
			Value: average(i, n),
		})
	}
	return vs
}

func trimmedTradingSessions(vs []*model.TradingSession, start time.Time) []*model.TradingSession {
	for i, v := range vs {
		if v.Date == start {
			return vs[i:]
		}
	}
	return vs
}

func trimmedMovingAverages(vs []*model.MovingAverageValue, start time.Time) []*model.MovingAverageValue {
	for i, v := range vs {
		if v.Date == start {
			return vs[i:]
		}
	}
	return vs
}

func trimmedAverageVolumes(vs []*model.AverageVolumeValue, start time.Time) []*model.AverageVolumeValue {
	for i, v := range vs {
		if v.Date == start {
			return vs[i:]
		}
	}
	return vs
}
