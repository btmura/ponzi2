// Package controller contains code for the controller in the MVC pattern.
package controller

import (
	"context"

	"github.com/btmura/ponzi2/internal/app/config"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view/chart"
	"github.com/btmura/ponzi2/internal/app/view/ui"
	"github.com/btmura/ponzi2/internal/errors"
	"github.com/btmura/ponzi2/internal/logger"
	"github.com/btmura/ponzi2/internal/stock/iex"
)

// Controller runs the program in a "game loop".
type Controller struct {
	// model is the data that the Controller connects to the View.
	model *model.Model

	// ui is the UI that the Controller updates.
	ui *ui.UI

	// chartInterval is the current interval to use for charts and thumbnails.
	chartInterval model.Interval

	// chartPriceStyle is the current price style for charts and thumbnails.
	chartPriceStyle chart.PriceStyle

	// stockRefresher offers methods to refresh one or many stocks.
	stockRefresher *stockRefresher

	// configSaver offers methods to save configs in the background.
	configSaver *configSaver

	// eventController offers methods to queue and process events in the main loop.
	eventController *eventController
}

// iexClientInterface is implemented by clients in the iex package to get stock data.
type iexClientInterface interface {
	GetQuotes(ctx context.Context, req *iex.GetQuotesRequest) ([]*iex.Quote, error)
	GetCharts(ctx context.Context, req *iex.GetChartsRequest) ([]*iex.Chart, error)
}

// New creates a new Controller.
func New(iexClient iexClientInterface, token string) *Controller {
	c := &Controller{
		model:           model.New(),
		ui:              ui.New(),
		chartInterval:   model.Daily,
		chartPriceStyle: chart.Bar,
		configSaver:     newConfigSaver(),
	}
	c.eventController = newEventController(c)
	c.stockRefresher = newStockRefresher(iexClient, token, c.eventController)
	return c
}

// RunLoop runs the loop until the user exits the app.
func (c *Controller) RunLoop() error {
	ctx := context.Background()

	cleanup, err := c.ui.Init(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	// Load the config and setup the initial UI.
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if cfg.CurrentStock != nil {
		if s := cfg.CurrentStock.Symbol; s != "" {
			if err := c.setChart(ctx, s); err != nil {
				return err
			}
		}
	}

	for _, cs := range cfg.Stocks {
		if s := cs.Symbol; s != "" {
			if err := c.addChartThumb(ctx, s); err != nil {
				return err
			}
		}
	}

	if priceStyle := cfg.Settings.ChartSettings.PriceStyle; priceStyle != chart.PriceStyleUnspecified {
		c.setChartPriceStyle(priceStyle)
	}

	c.ui.SetInputSymbolSubmittedCallback(func(symbol string) {
		if err := c.setChart(ctx, symbol); err != nil {
			logger.Errorf("setChart: %v", err)
		}
	})

	c.ui.SetSidebarChangeCallback(func(symbols []string) {
		if err := c.setSidebarSymbols(symbols); err != nil {
			logger.Errorf("setSidebarSymbols: %v", err)
		}
	})

	c.ui.SetChartPriceStyleButtonClickCallback(func(newPriceStyle chart.PriceStyle) {
		if newPriceStyle == chart.PriceStyleUnspecified {
			logger.Error("unspecified price style")
			return
		}
		c.setChartPriceStyle(newPriceStyle)
	})

	c.ui.SetChartZoomChangeCallback(func(zoomChange chart.ZoomChange) {
		if zoomChange == chart.ZoomChangeUnspecified {
			logger.Error("unspecified zoom change")
			return
		}
		c.setChartInterval(nextInterval(c.chartInterval, zoomChange))
	})

	c.ui.SetChartRefreshButtonClickCallback(func(symbol string) {
		if err := c.refreshAllStocks(ctx); err != nil {
			logger.Errorf("refreshAllStocks: %v", err)
		}
	})

	c.ui.SetChartAddButtonClickCallback(func(symbol string) {
		if err := c.addChartThumb(ctx, symbol); err != nil {
			logger.Errorf("addChartThumb: %v", err)
		}
	})

	c.ui.SetThumbRemoveButtonClickCallback(func(symbol string) {
		if err := c.removeChartThumb(symbol); err != nil {
			logger.Errorf("removeChartThumb: %v", err)
		}
	})

	c.ui.SetThumbClickCallback(func(symbol string) {
		if err := c.setChart(ctx, symbol); err != nil {
			logger.Errorf("setChart: %v", err)
		}
	})

	// Process stock refreshes and config changes in the background until the program ends.
	go c.stockRefresher.refreshLoop()
	go c.configSaver.saveLoop()

	defer func() {
		c.stockRefresher.stop()
		c.configSaver.stop()
	}()

	c.stockRefresher.start()
	c.configSaver.start()

	// Fire requests to get data for the entire UI.
	if err := c.refreshAllStocks(ctx); err != nil {
		return err
	}

	return c.ui.RunLoop(ctx, c.eventController.process)
}

func (c *Controller) setChart(ctx context.Context, symbol string) error {
	if symbol == "" {
		return errors.Errorf("missing symbol")
	}

	changed, err := c.model.SetCurrentSymbol(symbol)
	if err != nil {
		return err
	}

	// If the stock is already the current one, just refresh it.
	if !changed {
		return c.refreshCurrentStock(ctx)
	}

	data := c.chartData(symbol, c.chartInterval)

	if err := c.ui.SetChart(symbol, data, c.chartPriceStyle); err != nil {
		return err
	}

	if err := c.refreshCurrentStock(ctx); err != nil {
		return err
	}

	c.configSaver.save(c.makeConfig())

	return nil
}

func (c *Controller) addChartThumb(ctx context.Context, symbol string) error {
	if symbol == "" {
		return errors.Errorf("missing symbol")
	}

	added, err := c.model.AddSidebarSymbol(symbol)
	if err != nil {
		return err
	}

	// If the stock is already added, just refresh it.
	if !added {
		return c.stockRefresher.refreshOne(ctx, symbol, c.chartInterval)
	}

	data := c.chartData(symbol, c.chartInterval)

	if err := c.ui.AddChartThumb(symbol, data, c.chartPriceStyle); err != nil {
		return err
	}

	if err := c.stockRefresher.refreshOne(ctx, symbol, c.chartInterval); err != nil {
		return err
	}

	c.configSaver.save(c.makeConfig())

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

	if err := c.ui.RemoveChartThumb(symbol); err != nil {
		return err
	}

	c.configSaver.save(c.makeConfig())

	return nil
}

func (c *Controller) setSidebarSymbols(symbols []string) error {
	for _, s := range symbols {
		if err := model.ValidateSymbol(s); err != nil {
			return err
		}
	}

	if err := c.model.SetSidebarSymbols(symbols); err != nil {
		return err
	}

	c.configSaver.save(c.makeConfig())

	return nil
}

func (c *Controller) setChartPriceStyle(newPriceStyle chart.PriceStyle) {
	if newPriceStyle == chart.PriceStyleUnspecified {
		logger.Error("unspecified price style")
		return
	}

	if newPriceStyle == c.chartPriceStyle {
		return
	}

	c.chartPriceStyle = newPriceStyle
	c.ui.SetChartPriceStyle(newPriceStyle)
	c.configSaver.save(c.makeConfig())
}

func (c *Controller) setChartInterval(newInterval model.Interval) {
	if newInterval == model.IntervalUnspecified {
		logger.Error("unspecified interval")
		return
	}

	if newInterval == c.chartInterval {
		return
	}

	c.chartInterval = newInterval

	if s := c.model.CurrentSymbol(); s != "" {
		data := c.chartData(s, c.chartInterval)
		c.ui.SetData(s, data)
	}

	for _, s := range c.model.SidebarSymbols() {
		data := c.chartData(s, c.chartInterval)
		c.ui.SetData(s, data)
	}
}

func (c *Controller) chartData(symbol string, interval model.Interval) chart.Data {
	if symbol == "" {
		logger.Error("missing symbol")
		return chart.Data{}
	}

	data := chart.Data{Symbol: symbol}

	st, err := c.model.Stock(symbol)
	if err != nil {
		return data
	}

	if st == nil {
		return data
	}

	for _, ch := range st.Charts {
		if ch.Interval == interval {
			data.Quote = st.Quote
			data.Chart = ch
			return data
		}
	}

	return data
}

func (c *Controller) refreshCurrentStock(ctx context.Context) error {
	d := new(dataRequestBuilder)
	if s := c.model.CurrentSymbol(); s != "" {
		if err := d.add([]string{s}, c.chartInterval); err != nil {
			return err
		}
	}
	return c.stockRefresher.refresh(ctx, d)
}

func (c *Controller) refreshAllStocks(ctx context.Context) error {
	d := new(dataRequestBuilder)

	if s := c.model.CurrentSymbol(); s != "" {
		if err := d.add([]string{s}, c.chartInterval); err != nil {
			return err
		}
	}

	if err := d.add(c.model.SidebarSymbols(), c.chartInterval); err != nil {
		return err
	}

	return c.stockRefresher.refresh(ctx, d)
}

// onStockRefreshStarted implements the eventHandler interface.
func (c *Controller) onStockRefreshStarted(symbol string, interval model.Interval) error {
	return c.ui.SetLoading(symbol, interval)
}

// onStockUpdate implements the eventHandler interface.
func (c *Controller) onStockUpdate(symbol string, q *model.Quote, ch *model.Chart) error {
	if err := model.ValidateSymbol(symbol); err != nil {
		return err
	}

	if q != nil {
		if err := model.ValidateQuote(q); err != nil {
			return err
		}

		if err := c.model.UpdateStockQuote(symbol, q); err != nil {
			return err
		}
	}

	if ch != nil {
		if err := model.ValidateChart(ch); err != nil {
			return err
		}

		if err := c.model.UpdateStockChart(symbol, ch); err != nil {
			return err
		}
	}

	if q != nil || ch != nil {
		data := c.chartData(symbol, c.chartInterval)
		c.ui.SetData(symbol, data)
	}

	return nil
}

// onStockUpdateError implements the eventHandler interface.
func (c *Controller) onStockUpdateError(symbol string, updateErr error) error {
	logger.Errorf("stock update for %s failed: %v\n", symbol, updateErr)
	c.ui.SetError(symbol, updateErr)
	return nil
}

// onRefreshAllStocksRequest implements the eventHandler interface.
func (c *Controller) onRefreshAllStocksRequest(ctx context.Context) error {
	return c.refreshAllStocks(ctx)
}

// onEventAdded implements the eventHandler interface.
func (c *Controller) onEventAdded() {
	c.ui.WakeLoop()
}

func nextInterval(interval model.Interval, zoomChange chart.ZoomChange) model.Interval {
	// zoomIntervals are the ranges from most zoomed out to most zoomed in.
	var zoomIntervals = []model.Interval{
		model.Weekly,
		model.Daily,
	}

	// Find the current zoom range.
	i := 0
	for j := range zoomIntervals {
		if zoomIntervals[j] == interval {
			i = j
		}
	}

	// Adjust the zoom one increment.
	switch zoomChange {
	case chart.ZoomIn:
		if i+1 < len(zoomIntervals) {
			i++
		}
	case chart.ZoomOut:
		if i-1 >= 0 {
			i--
		}
	}

	return zoomIntervals[i]
}

func (c *Controller) makeConfig() *config.Config {
	cfg := &config.Config{}
	if s := c.model.CurrentSymbol(); s != "" {
		cfg.CurrentStock = &config.Stock{Symbol: s}
	}
	for _, s := range c.model.SidebarSymbols() {
		cfg.Stocks = append(cfg.Stocks, &config.Stock{Symbol: s})
	}
	cfg.Settings.ChartSettings.PriceStyle = c.chartPriceStyle
	return cfg
}
