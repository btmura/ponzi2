package controller

import (
	"context"

	"github.com/golang/glog"

	"gitlab.com/btmura/ponzi2/internal/app/model"
	"gitlab.com/btmura/ponzi2/internal/status"
	"gitlab.com/btmura/ponzi2/internal/stock/iex"
)

type stockUpdate struct {
	symbol    string
	chart     *model.Chart
	updateErr error
}

// addPendingStockUpdatesLocked locks the pendingStockUpdates slice
// and adds the new stock updates to the existing slice.
func (c *Controller) addPendingStockUpdatesLocked(us []stockUpdate) {
	c.pendingMutex.Lock()
	defer c.pendingMutex.Unlock()
	c.pendingStockUpdates = append(c.pendingStockUpdates, us...)
}

// takePendingStockUpdatesLocked locks the pendingStockUpdates slice,
// returns a copy of the updates, and empties the existing updates.
func (c *Controller) takePendingStockUpdatesLocked() []stockUpdate {
	c.pendingMutex.Lock()
	defer c.pendingMutex.Unlock()

	var us []stockUpdate
	for _, u := range c.pendingStockUpdates {
		us = append(us, u)
	}
	c.pendingStockUpdates = nil
	return us
}

type stockRefreshRequest struct {
	symbols   []string
	dataRange model.Range
}

func (c *Controller) currentStockRefreshRequests() []stockRefreshRequest {
	st := c.model.CurrentStock
	if st == nil {
		return nil
	}

	return []stockRefreshRequest{
		{
			symbols:   []string{st.Symbol},
			dataRange: c.chartRange,
		},
	}
}

func (c *Controller) sidebarStockRefreshRequests() []stockRefreshRequest {
	var symbols []string
	for _, st := range c.model.SavedStocks {
		symbols = append(symbols, st.Symbol)
	}
	if len(symbols) == 0 {
		return nil
	}

	return []stockRefreshRequest{
		{
			symbols:   symbols,
			dataRange: c.chartThumbRange,
		},
	}
}

func (c *Controller) allStockRefreshRequests() []stockRefreshRequest {
	range2Symbols := map[model.Range][]string{}

	for _, req := range c.currentStockRefreshRequests() {
		for _, s := range req.symbols {
			range2Symbols[req.dataRange] = append(range2Symbols[req.dataRange], s)
		}
	}

	for _, req := range c.sidebarStockRefreshRequests() {
		for _, s := range req.symbols {
			range2Symbols[req.dataRange] = append(range2Symbols[req.dataRange], s)
		}
	}

	var reqs []stockRefreshRequest
	for r, ss := range range2Symbols {
		symbolSet := map[string]bool{}
		for _, s := range ss {
			symbolSet[s] = true
		}

		var symbols []string
		for s := range symbolSet {
			symbols = append(symbols, s)
		}

		reqs = append(reqs, stockRefreshRequest{
			symbols:   symbols,
			dataRange: r,
		})
	}
	return reqs
}

func (c *Controller) refreshStocks(ctx context.Context, reqs []stockRefreshRequest) error {
	if !c.enableRefreshingStocks {
		glog.V(2).Infof("ignoring stock refresh request, refreshing disabled")
		return nil
	}

	if len(reqs) == 0 {
		return nil
	}

	for _, req := range reqs {
		if req.dataRange == c.chartRange {
			for _, s := range req.symbols {
				if ch, ok := c.symbolToChartMap[s]; ok {
					ch.SetLoading(true)
					ch.SetError(false)
				}
			}
		}

		if req.dataRange == c.chartThumbRange {
			for _, s := range req.symbols {
				if th, ok := c.symbolToChartThumbMap[s]; ok {
					th.SetLoading(true)
					th.SetError(false)
				}
			}
		}

		go func(symbols []string, dataRange model.Range) {
			handleErr := func(err error) {
				var us []stockUpdate
				for _, s := range symbols {
					us = append(us, stockUpdate{
						symbol:    s,
						updateErr: err,
					})
				}
				c.addPendingStockUpdatesLocked(us)
				c.view.WakeLoop()
			}

			var r iex.Range

			switch dataRange {
			case model.OneDay:
				r = iex.OneDay
			case model.OneYear:
				r = iex.TwoYears // Need additional data for weekly stochastics.
			default:
				handleErr(status.Errorf("bad range: %v", dataRange))
				return
			}

			req := &iex.GetStocksRequest{
				Symbols: symbols,
				Range:   r,
			}
			stocks, err := c.iexClient.GetStocks(ctx, req)
			if err != nil {
				handleErr(err)
				return
			}

			var us []stockUpdate

			found := map[string]bool{}
			for _, st := range stocks {
				found[st.Symbol] = true

				switch dataRange {
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

			for _, s := range symbols {
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
		}(req.symbols, req.dataRange)
	}

	return nil
}

func (c *Controller) processStockUpdates(ctx context.Context) error {
	for _, u := range c.takePendingStockUpdatesLocked() {
		switch {
		case u.updateErr != nil:
			if ch, ok := c.symbolToChartMap[u.symbol]; ok {
				ch.SetLoading(false)
				ch.SetError(true)
			}
			if th, ok := c.symbolToChartThumbMap[u.symbol]; ok {
				th.SetLoading(false)
				th.SetError(true)
			}

		case u.chart != nil:
			if err := c.model.UpdateChart(u.symbol, u.chart); err != nil {
				return err
			}

			if ch, ok := c.symbolToChartMap[u.symbol]; ok {
				ch.SetLoading(false)

				data, err := c.chartData(u.symbol, c.chartRange)
				if err != nil {
					return err
				}

				if err := c.title.SetData(data); err != nil {
					return err
				}

				if err := ch.SetData(data); err != nil {
					return err
				}
			}

			if th, ok := c.symbolToChartThumbMap[u.symbol]; ok {
				th.SetLoading(false)

				data, err := c.chartData(u.symbol, c.chartThumbRange)
				if err != nil {
					return err
				}

				if err := th.SetData(data); err != nil {
					return err
				}
			}

		default:
			return status.Errorf("bad update: %v", u)
		}
	}

	return nil
}
