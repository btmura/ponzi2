// Package controller contains code for the controller in the MVC pattern.
package controller

import (
	"context"
	"fmt"

	"github.com/golang/glog"

	"github.com/btmura/ponzi2/internal/app/config"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view"
	"github.com/btmura/ponzi2/internal/stock/iex"
)

// Controller runs the program in a "game loop".
type Controller struct {
	// iexClient fetches stock data.
	iexClient *iex.Client

	// model is the data that the Controller connects to the View.
	model *model.Model

	// view is the UI that the Controller updates.
	view *view.View

	// symbolToChartMap maps symbol to Chart. Only one entry right now.
	symbolToChartMap map[string]*view.Chart

	// symbolToChartThumbMap maps symbol to ChartThumbnail.
	symbolToChartThumbMap map[string]*view.ChartThumb

	// pendingStockUpdates has stock updates ready to apply to the model.
	pendingStockUpdates chan controllerStockUpdate

	// enableSavingConfigs enables saving config changes.
	enableSavingConfigs bool

	// pendingConfigSaves is a channel with configs to save.
	pendingConfigSaves chan *config.Config

	// doneSavingConfigs indicates saving is done and the program may quit.
	doneSavingConfigs chan bool
}

// controllerStockUpdate bundles a stock and new data for that stock.
type controllerStockUpdate struct {
	// symbol is the stock's symbol.
	symbol string

	// update is the new data for the stock. Nil if an error happened.
	update *model.StockUpdate

	// updateErr is the error getting the update. Nil if no error happened.
	updateErr error
}

// New creates a new Controller.
func New(iexClient *iex.Client) *Controller {
	return &Controller{
		iexClient:             iexClient,
		model:                 model.New(),
		view:                  view.New(),
		symbolToChartMap:      map[string]*view.Chart{},
		symbolToChartThumbMap: map[string]*view.ChartThumb{},
		pendingStockUpdates:   make(chan controllerStockUpdate),
		pendingConfigSaves:    make(chan *config.Config),
		doneSavingConfigs:     make(chan bool),
	}
}

// Run initializes and runs the "game loop".
func (c *Controller) Run() error {
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

	c.view.SetInputSymbolSubmittedCallback(func(symbol string) {
		c.setChart(ctx, symbol)
	})

	c.view.Run(c.update)

	// Disable config changes to start shutting down save processor.
	c.enableSavingConfigs = false
	close(c.pendingConfigSaves)
	<-c.doneSavingConfigs

	return nil
}

func (c *Controller) update() {
	// Process any stock updates.
loop:
	for {
		select {
		case u := <-c.pendingStockUpdates:
			switch {
			case u.update != nil:
				st, updated := c.model.UpdateStock(u.update)
				if !updated {
					break loop
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
		default:
			break loop
		}
	}
}

func (c *Controller) setChart(ctx context.Context, symbol string) {
	if symbol == "" {
		return
	}

	st, changed := c.model.SetCurrentStock(symbol)
	if !changed {
		return
	}

	for symbol, ch := range c.symbolToChartMap {
		delete(c.symbolToChartMap, symbol)
		ch.Close()
	}

	ch := view.NewChart()
	c.symbolToChartMap[symbol] = ch

	ch.SetData(st)
	ch.SetRefreshButtonClickCallback(func() {
		c.refreshStock(ctx, []string{symbol})
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
			for _, s := range symbols {
				c.pendingStockUpdates <- controllerStockUpdate{
					symbol:    s,
					updateErr: err,
				}
			}
			return
		}

		found := map[string]bool{}
		for _, st := range stocks {
			found[st.Symbol] = true
			c.pendingStockUpdates <- controllerStockUpdate{
				symbol: st.Symbol,
				update: modelStockUpdate(st),
			}
		}

		for _, s := range symbols {
			if found[s] {
				continue
			}
			c.pendingStockUpdates <- controllerStockUpdate{
				symbol:    s,
				updateErr: fmt.Errorf("no stock data for %q", s),
			}
		}

		c.view.PostEmptyEvent()
	}()
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
