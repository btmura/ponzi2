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
	"gitlab.com/btmura/ponzi2/internal/status"
	"gitlab.com/btmura/ponzi2/internal/stock/iex"
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
	pendingStockUpdates []stockUpdate

	// pendingSignals are the signals to be processed by the main thread.
	pendingSignals []signal

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

	// enableRefreshingStocks enables refreshing stocks.
	enableRefreshingStocks bool

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
			if err := c.addChartThumb(ctx, s); err != nil {
				return err
			}
		}
	}

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

		if err := c.refreshStocks(ctx, c.currentStockRefreshRequests()); err != nil {
			glog.Fatalf("TODO(btmura): remove log fatal, refreshStocks: %v", err)
		}
	})

	// Process config changes in the background until the program ends.
	go func() {
		for cfg := range c.pendingConfigSaves {
			if err := config.Save(cfg); err != nil {
				glog.V(2).Infof("failed to save config: %v", err)
			}
		}
		c.doneSavingConfigs <- true
	}()

	// Refresh stocks during market hours.
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

			c.addPendingSignalsLocked([]signal{refreshAllStocks})
			c.view.WakeLoop()
		}
	}()

	defer func() {
		ticker.Stop()

		// Disable refreshing stocks to avoid unnecessary work.
		c.enableRefreshingStocks = false

		// Disable config changes to start shutting down save processor.
		c.enableSavingConfigs = false
		close(c.pendingConfigSaves)
		<-c.doneSavingConfigs
	}()

	// Enable refreshing stocks and saving configs after the UI is setup and go routines launched.
	c.enableRefreshingStocks = true
	c.enableSavingConfigs = true

	// Fire requests to get data for the entire UI.
	if err := c.refreshStocks(ctx, c.allStockRefreshRequests()); err != nil {
		return err
	}

	return c.view.RunLoop(ctx, func(ctx context.Context) error {
		if err := c.processStockUpdates(ctx); err != nil {
			return err
		}

		if err := c.processPendingSignals(ctx); err != nil {
			return err
		}

		return nil
	})
}

func (c *Controller) setChart(ctx context.Context, symbol string) error {
	if symbol == "" {
		return status.Error("missing symbol")
	}

	changed, err := c.model.SetCurrentSymbol(symbol)
	if err != nil {
		return err
	}

	// If the stock is already the current one, just refresh it.
	if !changed {
		return c.refreshStocks(ctx, c.currentStockRefreshRequests())
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
		if err := c.refreshStocks(ctx, c.allStockRefreshRequests()); err != nil {
			glog.Fatalf("TODO(btmura): remove log fatal, refreshStocks: %v", err)
		}
	})
	ch.SetAddButtonClickCallback(func() {
		if err := c.addChartThumb(ctx, symbol); err != nil {
			glog.Fatalf("TODO(btmura): remove log fatal, addChartThumb: %v", err)
		}
	})

	c.view.SetChart(ch)

	if err := c.refreshStocks(ctx, c.currentStockRefreshRequests()); err != nil {
		return err
	}

	c.saveConfig()

	return nil
}

func (c *Controller) addChartThumb(ctx context.Context, symbol string) error {
	if symbol == "" {
		return status.Error("missing symbol")
	}

	added, err := c.model.AddSidebarSymbol(symbol)
	if err != nil {
		return err
	}

	// If the stock is already added, just refresh it.
	if !added {
		return c.refreshStocks(ctx, []stockRefreshRequest{{
			symbols:   []string{symbol},
			dataRange: c.chartThumbRange,
		}})
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
		if err := c.removeChartThumb(symbol); err != nil {
			glog.Fatalf("TODO(btmura): remove log fatal, removeChartThumb: %v", err)
		}
	})
	th.SetThumbClickCallback(func() {
		if err := c.setChart(ctx, symbol); err != nil {
			glog.Fatalf("TODO(btmura): remove log fatal, setChart: %v", err)
		}
	})

	c.view.AddChartThumb(th)

	if err := c.refreshStocks(ctx, []stockRefreshRequest{{
		symbols:   []string{symbol},
		dataRange: c.chartThumbRange,
	}}); err != nil {
		return err
	}

	c.saveConfig()

	return nil
}

func (c *Controller) removeChartThumb(symbol string) error {
	if symbol == "" {
		return nil
	}

	removed, err := c.model.RemoveSidebarSymbol(symbol)
	if err != nil {
		return err
	}

	if !removed {
		return nil
	}

	th := c.symbolToChartThumbMap[symbol]
	delete(c.symbolToChartThumbMap, symbol)
	th.Close()

	c.view.RemoveChartThumb(th)
	c.saveConfig()

	return nil
}

func (c *Controller) chartData(symbol string, dataRange model.Range) (*view.ChartData, error) {
	if symbol == "" {
		return nil, status.Error("missing symbol")
	}

	data := &view.ChartData{Symbol: symbol}

	st, err := c.model.Stock(symbol)
	if err != nil {
		return nil, err
	}

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
	if s := c.model.CurrentSymbol(); s != "" {
		cfg.CurrentStock = &config.Stock{Symbol: s}
	}
	for _, s := range c.model.SidebarSymbols() {
		cfg.Stocks = append(cfg.Stocks, &config.Stock{Symbol: s})
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
