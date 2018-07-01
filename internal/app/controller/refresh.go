package controller

import (
	"context"
	"sort"

	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/stock"

	"golang.org/x/sync/errgroup"
)

type stockUpdateData struct {
	hist  *stock.History
	ds    *stock.Stochastics
	ws    *stock.Stochastics
	ma25  *stock.MovingAverage
	ma50  *stock.MovingAverage
	ma250 *stock.MovingAverage
}

func (c *Controller) stockUpdate(ctx context.Context, symbol string) controllerStockUpdate {
	g, gCtx := errgroup.WithContext(ctx)

	var data stockUpdateData

	g.Go(func() error {
		req := &stock.GetHistoryRequest{Symbol: symbol}
		h, err := c.stockDataFetcher.GetHistory(gCtx, req)
		if err != nil {
			return err
		}
		data.hist = h
		return nil
	})

	g.Go(func() error {
		req := &stock.GetStochasticsRequest{
			Symbol:   symbol,
			Interval: stock.Daily,
		}
		s, err := c.stockDataFetcher.GetStochastics(gCtx, req)
		if err != nil {
			return err
		}
		data.ds = s
		return nil
	})

	g.Go(func() error {
		req := &stock.GetStochasticsRequest{
			Symbol:   symbol,
			Interval: stock.Weekly,
		}
		s, err := c.stockDataFetcher.GetStochastics(gCtx, req)
		if err != nil {
			return err
		}
		data.ws = s
		return nil
	})

	g.Go(func() error {
		req := &stock.GetMovingAverageRequest{
			Symbol:     symbol,
			TimePeriod: 25,
		}
		m, err := c.stockDataFetcher.GetMovingAverage(gCtx, req)
		if err != nil {
			return err
		}
		data.ma25 = m
		return nil
	})

	g.Go(func() error {
		req := &stock.GetMovingAverageRequest{
			Symbol:     symbol,
			TimePeriod: 50,
		}
		m, err := c.stockDataFetcher.GetMovingAverage(gCtx, req)
		if err != nil {
			return err
		}
		data.ma50 = m
		return nil
	})

	g.Go(func() error {
		req := &stock.GetMovingAverageRequest{
			Symbol:     symbol,
			TimePeriod: 250,
		}
		m, err := c.stockDataFetcher.GetMovingAverage(gCtx, req)
		if err != nil {
			return err
		}
		data.ma250 = m
		return nil
	})

	if err := g.Wait(); err != nil {
		logger.Printf("getting data failed: %v", err)
		return controllerStockUpdate{
			symbol:    symbol,
			updateErr: err,
		}
	}

	return controllerStockUpdate{
		symbol: symbol,
		update: makeStockUpdate(symbol, data),
	}
}

func makeStockUpdate(symbol string, data stockUpdateData) *model.StockUpdate {
	ds := dailySessions(data.hist.TradingSessions)
	ws := weeklySessions(ds)

	fillChangeValues(ds)
	fillChangeValues(ws)

	dsto := &model.Stochastics{}
	for _, v := range data.ds.Values {
		sv := &model.StochasticValue{
			Date: v.Date,
			K:    v.K / 100,
			D:    v.D / 100,
		}
		dsto.Values = append(dsto.Values, sv)
	}

	wsto := &model.Stochastics{}
	for _, v := range data.ws.Values {
		sv := &model.StochasticValue{
			Date: v.Date,
			K:    v.K / 100,
			D:    v.D / 100,
		}
		wsto.Values = append(wsto.Values, sv)
	}

	fillMovingAverages(ds)
	fillMovingAverages(ws)

	ds, ws = trimSessions(ds, ws)

	return &model.StockUpdate{
		Symbol:            symbol,
		DailySessions:     ds,
		WeeklySessions:    ws,
		DailyStochastics:  dsto,
		WeeklyStochastics: wsto,
	}
}

func dailySessions(ts []*stock.TradingSession) (ds []*model.TradingSession) {
	for _, s := range ts {
		ds = append(ds, &model.TradingSession{
			Date:   s.Date,
			Open:   s.Open,
			High:   s.High,
			Low:    s.Low,
			Close:  s.Close,
			Volume: s.Volume,
		})
	}
	sort.Slice(ds, func(i, j int) bool {
		return ds[i].Date.Before(ds[j].Date)
	})
	return ds
}

func weeklySessions(ds []*model.TradingSession) (ws []*model.TradingSession) {
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

func fillChangeValues(ss []*model.TradingSession) {
	for i := range ss {
		if i > 0 {
			ss[i].Change = ss[i].Close - ss[i-1].Close
			ss[i].PercentChange = ss[i].Change / ss[i-1].Close
		}
	}
}

func fillMovingAverages(ss []*model.TradingSession) {
	average := func(i, n int) (avg float32) {
		if i+1-n < 0 {
			return 0 // Not enough data
		}
		var sum float32
		for j := 0; j < n; j++ {
			sum += ss[i-j].Close
		}
		return sum / float32(n)
	}

	for i := range ss {
		ss[i].MovingAverage25 = average(i, 25)
		ss[i].MovingAverage50 = average(i, 50)
		ss[i].MovingAverage200 = average(i, 200)
	}
}

func trimSessions(ds, ws []*model.TradingSession) (trimDs, trimWs []*model.TradingSession) {
	const sixMonthWeeks = 4 /* weeks */ * 6 /* months */
	if len(ws) >= sixMonthWeeks {
		ws = ws[len(ws)-sixMonthWeeks:]
		for i := range ds {
			if ds[i].Date == ws[0].Date {
				ds = ds[i:]
				return ds, ws
			}
		}
	}
	return ds, ws
}
