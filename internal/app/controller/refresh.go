package controller

import (
	"context"
	"sort"

	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/stock"

	"golang.org/x/sync/errgroup"
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

type stockUpdateData struct {
	hist  *stock.History
	ma25  *stock.MovingAverage
	ma50  *stock.MovingAverage
	ma200 *stock.MovingAverage
	dsto  *stock.Stochastics
	wsto  *stock.Stochastics
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
		data.dsto = s
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
		data.wsto = s
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
			TimePeriod: 200,
		}
		m, err := c.stockDataFetcher.GetMovingAverage(gCtx, req)
		if err != nil {
			return err
		}
		data.ma200 = m
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
	return trim(&model.StockUpdate{
		Symbol:            symbol,
		DailySessions:     convertHistory(data.hist),
		MovingAverage25:   convertMovingAverage(data.ma25),
		MovingAverage50:   convertMovingAverage(data.ma50),
		MovingAverage200:  convertMovingAverage(data.ma200),
		DailyStochastics:  convertStochastics(data.dsto),
		WeeklyStochastics: convertStochastics(data.wsto),
	})
}

func convertHistory(hist *stock.History) (ts []*model.TradingSession) {
	for _, s := range hist.TradingSessions {
		ts = append(ts, &model.TradingSession{
			Date:   s.Date,
			Open:   s.Open,
			High:   s.High,
			Low:    s.Low,
			Close:  s.Close,
			Volume: s.Volume,
		})
	}
	sort.Slice(ts, func(i, j int) bool {
		return ts[i].Date.Before(ts[j].Date)
	})

	for i := range ts {
		if i > 0 {
			ts[i].Change = ts[i].Close - ts[i-1].Close
			ts[i].PercentChange = ts[i].Change / ts[i-1].Close
		}
	}

	return ts
}

func convertMovingAverage(src *stock.MovingAverage) *model.MovingAverage {
	dst := &model.MovingAverage{}
	for _, v := range src.Values {
		mv := &model.MovingAverageValue{
			Date:    v.Date,
			Average: v.Average,
		}
		dst.Values = append(dst.Values, mv)
	}
	sort.Slice(dst.Values, func(i, j int) bool {
		return dst.Values[i].Date.Before(dst.Values[j].Date)
	})
	return dst
}

func convertStochastics(src *stock.Stochastics) *model.Stochastics {
	dst := &model.Stochastics{}
	for _, v := range src.Values {
		sv := &model.StochasticValue{
			Date: v.Date,
			K:    v.K / 100,
			D:    v.D / 100,
		}
		dst.Values = append(dst.Values, sv)
	}
	sort.Slice(dst.Values, func(i, j int) bool {
		return dst.Values[i].Date.Before(dst.Values[j].Date)
	})
	return dst
}

func trim(u *model.StockUpdate) *model.StockUpdate {
	if len(u.DailySessions) == 0 {
		return &model.StockUpdate{
			Symbol:            u.Symbol,
			MovingAverage25:   &model.MovingAverage{},
			MovingAverage50:   &model.MovingAverage{},
			MovingAverage200:  &model.MovingAverage{},
			DailyStochastics:  &model.Stochastics{},
			WeeklyStochastics: &model.Stochastics{},
		}
	}

	start := u.DailySessions[0].Date

	trimMovingAverage := func(ma *model.MovingAverage) {
		i := 0
		for ; i < len(ma.Values); i++ {
			if d := ma.Values[i].Date; d.Equal(start) || d.After(start) {
				break
			}
		}
		ma.Values = ma.Values[i:]
	}

	trimMovingAverage(u.MovingAverage25)
	trimMovingAverage(u.MovingAverage50)
	trimMovingAverage(u.MovingAverage200)

	trimStochastics := func(sto *model.Stochastics) {
		i := 0
		for ; i < len(sto.Values); i++ {
			if d := sto.Values[i].Date; d.Equal(start) || d.After(start) {
				break
			}
		}
		sto.Values = sto.Values[i:]
	}

	trimStochastics(u.DailyStochastics)
	trimStochastics(u.WeeklyStochastics)

	return u
}
