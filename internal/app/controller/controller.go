// Package controller contains code for the controller in the MVC pattern.
package controller

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/golang/glog"

	"gitlab.com/btmura/ponzi2/internal/app/config"
	"gitlab.com/btmura/ponzi2/internal/app/model"
	"gitlab.com/btmura/ponzi2/internal/app/view"
	"gitlab.com/btmura/ponzi2/internal/stock/iex"
)

// loc is the timezone to use when parsing dates.
var loc = mustLoadLocation("America/New_York")

// Controller runs the program in a "game loop".
type Controller struct {
	// model is the data that the Controller connects to the View.
	model *model.Model

	// iexClient fetches stock data to update the model.
	iexClient *iex.Client

	// pendingStockUpdates are the updates to be processed by the main thread.
	pendingStockUpdates []controllerStockUpdate

	// pendingSignals are the signals to be processed by the main thread.
	pendingSignals []controllerSignal

	// pendingMutex guards pendingUpdates and pendingSignals.
	pendingMutex *sync.Mutex

	// view is the UI that the Controller updates.
	view *view.View

	// title controls the title bar.
	title *view.Title

	// symbolToChartMap maps symbol to Chart. Only one entry right now.
	symbolToChartMap map[string]*view.Chart

	// symbolToChartThumbMap maps symbol to ChartThumbnail.
	symbolToChartThumbMap map[string]*view.ChartThumb

	// enableSavingConfigs enables saving config changes.
	enableSavingConfigs bool

	// pendingConfigSaves is a channel with configs to save.
	pendingConfigSaves chan *config.Config

	// doneSavingConfigs indicates saving is done and the program may quit.
	doneSavingConfigs chan bool
}

type controllerStockUpdate struct {
	symbol    string
	update    *model.StockUpdate
	updateErr error
}

//go:generate stringer -type=controllerSignal
type controllerSignal int

const (
	signalUnspecified controllerSignal = iota
	signalRefreshCurrentStock
)

// New creates a new Controller.
func New(iexClient *iex.Client) *Controller {
	return &Controller{
		model:                 model.New(),
		iexClient:             iexClient,
		pendingMutex:          &sync.Mutex{},
		view:                  view.New(),
		title:                 view.NewTitle(),
		symbolToChartMap:      map[string]*view.Chart{},
		symbolToChartThumbMap: map[string]*view.ChartThumb{},
		pendingConfigSaves:    make(chan *config.Config),
		doneSavingConfigs:     make(chan bool),
	}
}

// RunLoop runs the loop until the user exits the app.
func (c *Controller) RunLoop() error {
	ctx := context.Background()

	cleanup, err := c.view.Init(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	// Load the config and setup the initial UI.
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if s := cfg.GetCurrentStock().GetSymbol(); s != "" {
		c.setChart(ctx, s)
	}

	for _, cs := range cfg.GetStocks() {
		if s := cs.GetSymbol(); s != "" {
			c.addChartThumb(ctx, s)
		}
	}

	// Process config changes in the background until the program ends.
	go func() {
		for cfg := range c.pendingConfigSaves {
			if err := config.Save(cfg); err != nil {
				glog.V(2).Infof("failed to save config: %v", err)
			}
		}
		c.doneSavingConfigs <- true
	}()

	// Enable saving configs after UI is setup and change processor started.
	c.enableSavingConfigs = true

	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for t := range ticker.C {
			n := time.Now()
			open := time.Date(n.Year(), n.Month(), n.Day(), 9, 30, 0, 0, loc)
			close := time.Date(n.Year(), n.Month(), n.Day(), 16, 0, 0, 0, loc)

			if t.Before(open) || t.After(close) {
				glog.V(2).Infof("ignoring refresh ticker at %v", t.Format("1/2/2006 3:04:05 PM"))
				continue
			}

			c.addPendingSignalsLocked([]controllerSignal{signalRefreshCurrentStock})
			c.view.WakeLoop()
		}
	}()

	c.view.SetTitle(c.title)

	c.view.SetInputSymbolSubmittedCallback(func(symbol string) {
		c.setChart(ctx, symbol)
	})

	c.view.RunLoop(ctx, c.update)

	ticker.Stop()

	// Disable config changes to start shutting down save processor.
	c.enableSavingConfigs = false
	close(c.pendingConfigSaves)
	<-c.doneSavingConfigs

	return nil
}

func (c *Controller) update(ctx context.Context) {
	for _, u := range c.takePendingStockUpdatesLocked() {
		switch {
		case u.update != nil:
			st, updated := c.model.UpdateStock(u.update)
			if !updated {
				continue
			}
			if st == c.model.CurrentStock {
				c.title.SetData(st)
			}
			if ch, ok := c.symbolToChartMap[u.symbol]; ok {
				ch.SetLoading(false)
				ch.SetData(st)
			}
			if th, ok := c.symbolToChartThumbMap[u.symbol]; ok {
				th.SetLoading(false)
				th.SetData(st)
			}

		case u.updateErr != nil:
			if ch, ok := c.symbolToChartMap[u.symbol]; ok {
				ch.SetLoading(false)
				ch.SetError(true)
			}
			if th, ok := c.symbolToChartThumbMap[u.symbol]; ok {
				th.SetLoading(false)
				th.SetError(true)
			}
		}
	}

	for _, s := range c.takePendingSignalsLocked() {
		switch s {
		case signalRefreshCurrentStock:
			c.refreshStock(ctx, c.allSymbols())
		}
	}
}

func (c *Controller) setChart(ctx context.Context, symbol string) {
	if symbol == "" {
		return
	}

	st, changed := c.model.SetCurrentStock(symbol)
	if !changed {
		c.refreshStock(ctx, []string{symbol})
		return
	}

	for symbol, ch := range c.symbolToChartMap {
		delete(c.symbolToChartMap, symbol)
		ch.Close()
	}

	c.title.SetData(st)

	ch := view.NewChart()
	c.symbolToChartMap[symbol] = ch

	ch.SetData(st)
	ch.SetRefreshButtonClickCallback(func() {
		c.refreshStock(ctx, c.allSymbols())
	})
	ch.SetAddButtonClickCallback(func() {
		c.addChartThumb(ctx, symbol)
	})

	c.view.SetChart(ch)
	c.refreshStock(ctx, []string{symbol})
	c.saveConfig()
}

func (c *Controller) addChartThumb(ctx context.Context, symbol string) {
	if symbol == "" {
		return
	}

	st, added := c.model.AddSavedStock(symbol)
	if !added {
		c.refreshStock(ctx, []string{symbol})
		return
	}

	th := view.NewChartThumb()
	c.symbolToChartThumbMap[symbol] = th

	th.SetData(st)
	th.SetRemoveButtonClickCallback(func() {
		c.removeChartThumb(symbol)
	})
	th.SetThumbClickCallback(func() {
		c.setChart(ctx, symbol)
	})

	c.view.AddChartThumb(th)
	c.refreshStock(ctx, []string{symbol})
	c.saveConfig()
}

func (c *Controller) removeChartThumb(symbol string) {
	if symbol == "" {
		return
	}

	if !c.model.RemoveSavedStock(symbol) {
		return
	}

	th := c.symbolToChartThumbMap[symbol]
	delete(c.symbolToChartThumbMap, symbol)
	th.Close()

	c.view.RemoveChartThumb(th)
	c.saveConfig()
}

func (c *Controller) allSymbols() []string {
	var symbols []string
	if st := c.model.CurrentStock; st != nil {
		symbols = append(symbols, st.Symbol)
	}
	for _, st := range c.model.SavedStocks {
		symbols = append(symbols, st.Symbol)
	}
	return symbols
}

func (c *Controller) refreshStock(ctx context.Context, symbols []string) {
	if len(symbols) == 0 {
		return
	}

	for _, s := range symbols {
		if ch, ok := c.symbolToChartMap[s]; ok {
			ch.SetLoading(true)
			ch.SetError(false)
		}
		if th, ok := c.symbolToChartThumbMap[s]; ok {
			th.SetLoading(true)
			th.SetError(false)
		}
	}

	go func() {
		req := &iex.GetStocksRequest{
			Symbols: symbols,
			Range:   iex.RangeTwoYears,
		}
		stocks, err := c.iexClient.GetStocks(ctx, req)
		if err != nil {
			var us []controllerStockUpdate
			for _, s := range symbols {
				us = append(us, controllerStockUpdate{
					symbol:    s,
					updateErr: err,
				})
			}
			c.addPendingStockUpdatesLocked(us)
			c.view.WakeLoop()
			return
		}

		var us []controllerStockUpdate

		found := map[string]bool{}
		for _, st := range stocks {
			found[st.Symbol] = true
			u, err := modelStockUpdate(st)
			us = append(us, controllerStockUpdate{
				symbol:    st.Symbol,
				update:    u,
				updateErr: err,
			})
		}

		for _, s := range symbols {
			if found[s] {
				continue
			}
			us = append(us, controllerStockUpdate{
				symbol:    s,
				updateErr: fmt.Errorf("no stock data for %q", s),
			})
		}

		c.addPendingStockUpdatesLocked(us)
		c.view.WakeLoop()
	}()
}

// addPendingStockUpdatesLocked locks the pendingStockUpdates slice
// and adds the new stock updates to the existing slice.
func (c *Controller) addPendingStockUpdatesLocked(us []controllerStockUpdate) {
	c.pendingMutex.Lock()
	defer c.pendingMutex.Unlock()
	c.pendingStockUpdates = append(c.pendingStockUpdates, us...)
}

// takePendingStockUpdatesLocked locks the pendingStockUpdates slice,
// returns a copy of the updates, and empties the existing updates.
func (c *Controller) takePendingStockUpdatesLocked() []controllerStockUpdate {
	c.pendingMutex.Lock()
	defer c.pendingMutex.Unlock()

	var us []controllerStockUpdate
	for _, u := range c.pendingStockUpdates {
		us = append(us, u)
	}
	c.pendingStockUpdates = nil
	return us
}

// addPendingSignalsLocked locks the pendingSignals slice
// and adds the new signals to the existing slice.
func (c *Controller) addPendingSignalsLocked(signals []controllerSignal) {
	c.pendingMutex.Lock()
	defer c.pendingMutex.Unlock()
	c.pendingSignals = append(c.pendingSignals, signals...)
}

// takePendingSignalsLocked locks the pendingSignals slice,
// returns a copy of the current signals, and empties the existing signals.
func (c *Controller) takePendingSignalsLocked() []controllerSignal {
	c.pendingMutex.Lock()
	defer c.pendingMutex.Unlock()

	var ss []controllerSignal
	for _, s := range c.pendingSignals {
		ss = append(ss, s)
	}
	c.pendingSignals = nil
	return ss
}

func (c *Controller) saveConfig() {
	if !c.enableSavingConfigs {
		glog.V(2).Infof("ignoring save request, saving disabled")
		return
	}

	// Make the config on the main thread to save the exact config at the time.
	cfg := &config.Config{}
	if st := c.model.CurrentStock; st != nil {
		cfg.CurrentStock = &config.Stock{Symbol: st.Symbol}
	}
	for _, st := range c.model.SavedStocks {
		cfg.Stocks = append(cfg.Stocks, &config.Stock{Symbol: st.Symbol})
	}

	// Queue the config for saving.
	go func() {
		c.pendingConfigSaves <- cfg
	}()
}

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		log.Fatalf("time.LoadLocation(%s) failed: %v", name, err)
	}
	return loc
}
