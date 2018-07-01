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
	ds := dailySessions(data.hist.TradingSessions)

	fillChangeValues(ds)

	convertMovingAverage := func(src *stock.MovingAverage) *model.MovingAverage {
		dst := &model.MovingAverage{}
		for _, v := range src.Values {
			mv := &model.MovingAverageValue{
				Date:    v.Date,
				Average: v.Average,
			}
			dst.Values = append(dst.Values, mv)
		}
		return dst
	}

	ma25 := convertMovingAverage(data.ma25)
	ma50 := convertMovingAverage(data.ma50)
	ma200 := convertMovingAverage(data.ma200)

	convertStochastics := func(src *stock.Stochastics) *model.Stochastics {
		dst := &model.Stochastics{}
		for _, v := range src.Values {
			sv := &model.StochasticValue{
				Date: v.Date,
				K:    v.K / 100,
				D:    v.D / 100,
			}
			dst.Values = append(dst.Values, sv)
		}
		return dst
	}

	dsto := convertStochastics(data.dsto)
	wsto := convertStochastics(data.wsto)

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

func fillChangeValues(ss []*model.TradingSession) {
	for i := range ss {
		if i > 0 {
			ss[i].Change = ss[i].Close - ss[i-1].Close
			ss[i].PercentChange = ss[i].Change / ss[i-1].Close
		}
	}
}
