// Package controller contains code for the controller in the MVC pattern.
package controller

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/golang/glog"

	"gitlab.com/btmura/ponzi2/internal/app/config"
	"gitlab.com/btmura/ponzi2/internal/app/model"
	"gitlab.com/btmura/ponzi2/internal/app/view"
	"gitlab.com/btmura/ponzi2/internal/stock/iex"
	"gitlab.com/btmura/ponzi2/internal/util"
)

// loc is the timezone to use when parsing dates.
var loc = mustLoadLocation("America/New_York")

// zoomRanges are the ranges from most zoomed out to most zoomed in.
var zoomRanges = []model.Range{
	model.OneYear,
	model.OneDay,
}

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

	// chartRange is the current data range to use for Charts.
	chartRange model.Range

	// chartThumbRange is the current data range to use for ChartThumbnails.
	chartThumbRange model.Range

	// enableSavingConfigs enables saving config changes.
	enableSavingConfigs bool

	// pendingConfigSaves is a channel with configs to save.
	pendingConfigSaves chan *config.Config

	// doneSavingConfigs indicates saving is done and the program may quit.
	doneSavingConfigs chan bool
}

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
		chartRange:            model.OneYear,
		chartThumbRange:       model.OneYear,
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

	c.view.SetChartZoomChangeCallback(func(zoomChange view.ZoomChange) {
		// Find the current zoom range.
		i := 0
		for j := range zoomRanges {
			if zoomRanges[j] == c.chartRange {
				i = j
			}
		}

		// Adjust the zoom one increment.
		switch zoomChange {
		case view.ZoomIn:
			if i+1 < len(zoomRanges) {
				i++
			}
		case view.ZoomOut:
			if i-1 >= 0 {
				i--
			}
		}

		// Ignore if no change in zoom.
		if c.chartRange == zoomRanges[i] {
			return
		}

		// Set zoom and refresh the current stock.
		c.chartRange = zoomRanges[i]
		c.refreshStock(ctx, c.currentSymbol(), c.chartRange)
	})

	defer func() {
		ticker.Stop()

		// Disable config changes to start shutting down save processor.
		c.enableSavingConfigs = false
		close(c.pendingConfigSaves)
		<-c.doneSavingConfigs
	}()

	return c.view.RunLoop(ctx, c.update)
}

func (c *Controller) update(ctx context.Context) error {
	if err := c.processStockUpdates(ctx); err != nil {
		return err
	}

	if err := c.processPendingSignals(ctx); err != nil {
		return err
	}

	return nil
}

func (c *Controller) setChart(ctx context.Context, symbol string) error {
	if symbol == "" {
		return util.Error("missing symbol")
	}

	_, changed := c.model.SetCurrentStock(symbol)
	if !changed {
		c.refreshStock(ctx, []string{symbol}, c.chartRange)
		return nil
	}

	for symbol, ch := range c.symbolToChartMap {
		delete(c.symbolToChartMap, symbol)
		ch.Close()
	}

	ch := view.NewChart()
	c.symbolToChartMap[symbol] = ch

	data, err := c.chartData(symbol, c.chartRange)
	if err != nil {
		return err
	}

	if err := c.title.SetData(data); err != nil {
		return err
	}

	if err := ch.SetData(data); err != nil {
		return err
	}

	ch.SetRefreshButtonClickCallback(func() {
		c.refreshStock(ctx, c.allSymbols(), c.chartRange)
	})
	ch.SetAddButtonClickCallback(func() {
		c.addChartThumb(ctx, symbol)
	})

	c.view.SetChart(ch)
	c.refreshStock(ctx, []string{symbol}, c.chartRange)
	c.saveConfig()

	return nil
}

func (c *Controller) addChartThumb(ctx context.Context, symbol string) error {
	if symbol == "" {
		return util.Error("missing symbol")
	}

	_, added := c.model.AddSavedStock(symbol)
	if !added {
		c.refreshStock(ctx, []string{symbol}, c.chartThumbRange)
		return nil
	}

	th := view.NewChartThumb()
	c.symbolToChartThumbMap[symbol] = th

	data, err := c.chartData(symbol, c.chartThumbRange)
	if err != nil {
		return err
	}

	if err := th.SetData(data); err != nil {
		return err
	}

	th.SetRemoveButtonClickCallback(func() {
		c.removeChartThumb(symbol)
	})
	th.SetThumbClickCallback(func() {
		c.setChart(ctx, symbol)
	})

	c.view.AddChartThumb(th)
	c.refreshStock(ctx, []string{symbol}, c.chartThumbRange)
	c.saveConfig()

	return nil
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

func (c *Controller) chartData(symbol string, dataRange model.Range) (*view.ChartData, error) {
	if symbol == "" {
		return nil, util.Error("missing symbol")
	}

	data := &view.ChartData{Symbol: symbol}

	st := c.model.Stock(symbol)
	if st == nil {
		return data, nil
	}

	for _, ch := range st.Charts {
		if ch.Range == dataRange {
			data.Quote = ch.Quote
			data.Chart = ch
			return data, nil
		}
	}

	return data, nil
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
