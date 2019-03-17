package controller

import (
	"context"

	"github.com/golang/glog"

	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/status"
	"github.com/btmura/ponzi2/internal/stock/iex"
)

func (c *Controller) refreshStock(ctx context.Context, symbol string, dataRange model.Range) error {
	// TODO(btmura): validate symbols
	d := new(dataRequestBuilder)
	if err := d.add([]string{symbol}, dataRange); err != nil {
		return err
	}
	return c.refreshStockInternal(ctx, d)
}

func (c *Controller) refreshCurrentStock(ctx context.Context) error {
	d := new(dataRequestBuilder)
	if s := c.model.CurrentSymbol(); s != "" {
		if err := d.add([]string{s}, c.chartRange); err != nil {
			return err
		}
	}
	return c.refreshStockInternal(ctx, d)
}

func (c *Controller) refreshAllStocks(ctx context.Context) error {
	d := new(dataRequestBuilder)

	if s := c.model.CurrentSymbol(); s != "" {
		if err := d.add([]string{s}, c.chartRange); err != nil {
			return err
		}
	}

	if err := d.add(c.model.SidebarSymbols(), c.chartThumbRange); err != nil {
		return err
	}

	return c.refreshStockInternal(ctx, d)
}

func (c *Controller) refreshStockInternal(ctx context.Context, d *dataRequestBuilder) error {
	if !c.enableRefreshingStocks {
		glog.V(2).Infof("ignoring stock refresh request, refreshing disabled")
		return nil
	}

	for s, ch := range c.symbolToChartMap {
		ok, err := d.contains(s, c.chartRange)
		if err != nil {
			return err
		}
		if ok {
			ch.SetLoading(true)
			ch.SetError(false)
		}
	}

	for s, th := range c.symbolToChartThumbMap {
		ok, err := d.contains(s, c.chartThumbRange)
		if err != nil {
			return err
		}
		if ok {
			th.SetLoading(true)
			th.SetError(false)
		}
	}

	reqs, err := d.dataRequests()
	if err != nil {
		return err
	}

	for _, req := range reqs {
		go func(req *dataRequest) {
			handleErr := func(err error) {
				var us []stockUpdate
				for _, s := range req.symbols {
					us = append(us, stockUpdate{
						symbol:    s,
						updateErr: err,
					})
				}
				c.addPendingStockUpdatesLocked(us)
				c.view.WakeLoop()
			}

			stocks, err := c.iexClient.GetStocks(ctx, req.iexRequest)
			if err != nil {
				handleErr(err)
				return
			}

			var us []stockUpdate

			found := map[string]bool{}
			for _, st := range stocks {
				found[st.Symbol] = true

				switch req.dataRange {
				case model.OneDay:
					ch, err := modelOneDayChart(st)
					us = append(us, stockUpdate{
						symbol:    st.Symbol,
						chart:     ch,
						updateErr: err,
					})

				case model.OneYear:
					ch, err := modelOneYearChart(st)
					us = append(us, stockUpdate{
						symbol:    st.Symbol,
						chart:     ch,
						updateErr: err,
					})
				}
			}

			for _, s := range req.symbols {
				if found[s] {
					continue
				}
				us = append(us, stockUpdate{
					symbol:    s,
					updateErr: status.Errorf("no stock data for %q", s),
				})
			}

			c.addPendingStockUpdatesLocked(us)
			c.view.WakeLoop()

		}(req)
	}

	return nil
}

// dataRequestBuilder accumulates symbols and data ranges and builds a slice of data requests.
type dataRequestBuilder struct {
	range2Symbols map[model.Range][]string
}

func (d *dataRequestBuilder) add(symbols []string, dataRange model.Range) error {
	// TODO(btmura): validate symbols

	if dataRange == model.RangeUnspecified {
		return status.Error("range not set")
	}

	if len(symbols) == 0 {
		return nil
	}

	sset := make(map[string]bool)
	for _, s := range d.range2Symbols[dataRange] {
		sset[s] = true
	}
	for _, s := range symbols {
		sset[s] = true
	}

	var ss []string
	for s := range sset {
		ss = append(ss, s)
	}

	if d.range2Symbols == nil {
		d.range2Symbols = make(map[model.Range][]string)
	}
	d.range2Symbols[dataRange] = ss

	return nil
}

func (d *dataRequestBuilder) contains(symbol string, dataRange model.Range) (bool, error) {
	// TODO(btmura): validate symbols
	for _, s := range d.range2Symbols[dataRange] {
		if s == symbol {
			return true, nil
		}
	}
	return false, nil
}

type dataRequest struct {
	symbols    []string
	dataRange  model.Range
	iexRequest *iex.GetStocksRequest
}

func (d *dataRequestBuilder) dataRequests() ([]*dataRequest, error) {
	var reqs []*dataRequest
	for r, ss := range d.range2Symbols {
		var ir iex.Range

		switch r {
		case model.OneDay:
			ir = iex.OneDay
		case model.OneYear:
			ir = iex.TwoYears // Need additional data for weekly stochastics.
		default:
			return nil, status.Errorf("bad range: %v", r)
		}

		reqs = append(reqs, &dataRequest{
			symbols:   ss,
			dataRange: r,
			iexRequest: &iex.GetStocksRequest{
				Symbols: ss,
				Range:   ir,
			},
		})
	}
	return reqs, nil
}
