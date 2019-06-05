// Package controller contains code for the controller in the MVC pattern.
package controller

import (
	"context"

	"github.com/golang/glog"

	"github.com/btmura/ponzi2/internal/app/config"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view"
	"github.com/btmura/ponzi2/internal/app/view/chart"
	"github.com/btmura/ponzi2/internal/app/view/title"
	"github.com/btmura/ponzi2/internal/app/view/ui"
	"github.com/btmura/ponzi2/internal/errors"
	"github.com/btmura/ponzi2/internal/stock/iex"
)

// Controller runs the program in a "game loop".
type Controller struct {
	// model is the data that the Controller connects to the View.
	model *model.Model

	// ui is the UI that the Controller updates.
	ui *ui.UI

	// title controls the title bar.
	title *title.Title

	// symbolToChartMap maps symbol to Chart. Only one entry right now.
	symbolToChartMap map[string]*chart.Chart

	// symbolToChartThumbMap maps symbol to ChartThumbnail.
	symbolToChartThumbMap map[string]*chart.Thumb

	// chartRange is the current data range to use for Charts.
	chartRange model.Range

	// chartThumbRange is the current data range to use for ChartThumbnails.
	chartThumbRange model.Range

	// stockRefresher offers methods to refresh one or many stocks.
	stockRefresher *stockRefresher

	// configSaver offers methods to save configs in the background.
	configSaver *configSaver

	// eventController offers methods to queue and process events in the main loop.
	eventController *eventController
}

// New creates a new Controller.
func New(iexClient *iex.Client) *Controller {
	c := &Controller{
		model:                 model.New(),
		ui:                    ui.New(),
		title:                 title.New(),
		symbolToChartMap:      map[string]*chart.Chart{},
		symbolToChartThumbMap: map[string]*chart.Thumb{},
		chartRange:            model.OneYear,
		chartThumbRange:       model.OneYear,
		configSaver:           newConfigSaver(),
	}
	c.eventController = newEventController(c)
	c.stockRefresher = newStockRefresher(iexClient, c.eventController)
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

	c.ui.SetTitle(c.title)

	c.ui.SetInputSymbolSubmittedCallback(func(symbol string) {
		c.setChart(ctx, symbol)
	})

	c.ui.SetChartZoomChangeCallback(func(zoomChange view.ZoomChange) {
		r := nextRange(c.chartRange, zoomChange)

		if c.chartRange == r {
			return
		}

		c.chartRange = r

		if err := c.refreshCurrentStock(ctx); err != nil {
			glog.Fatalf("TODO(btmura): remove log fatal, refreshStocks: %v", err)
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

	for symbol, ch := range c.symbolToChartMap {
		delete(c.symbolToChartMap, symbol)
		ch.Close()
	}

	ch := c.ui.NewChart()
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
		if err := c.refreshAllStocks(ctx); err != nil {
			glog.Fatalf("TODO(btmura): remove log fatal, refreshStocks: %v", err)
		}
	})
	ch.SetAddButtonClickCallback(func() {
		if err := c.addChartThumb(ctx, symbol); err != nil {
			glog.Fatalf("TODO(btmura): remove log fatal, addChartThumb: %v", err)
		}
	})

	c.ui.SetChart(ch)

	if err := c.refreshCurrentStock(ctx); err != nil {
		return err
	}

	c.configSaver.save(toConfig(c.model))

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
		return c.stockRefresher.refreshOne(ctx, symbol, c.chartThumbRange)
	}

	th := c.ui.NewChartThumb()
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

	c.ui.AddChartThumb(th)

	if err := c.stockRefresher.refreshOne(ctx, symbol, c.chartThumbRange); err != nil {
		return err
	}

	c.configSaver.save(toConfig(c.model))

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

	c.ui.RemoveChartThumb(th)
	c.configSaver.save(toConfig(c.model))

	return nil
}

func (c *Controller) chartData(symbol string, dataRange model.Range) (*chart.Data, error) {
	if symbol == "" {
		return nil, errors.Errorf("missing symbol")
	}

	data := &chart.Data{Symbol: symbol}

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

func (c *Controller) refreshCurrentStock(ctx context.Context) error {
	d := new(dataRequestBuilder)
	if s := c.model.CurrentSymbol(); s != "" {
		if err := d.add([]string{s}, c.chartRange); err != nil {
			return err
		}
	}
	return c.stockRefresher.refresh(ctx, d)
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

	return c.stockRefresher.refresh(ctx, d)
}

// onStockRefreshStarted implements the eventHandler interface.
func (c *Controller) onStockRefreshStarted(symbol string, dataRange model.Range) error {
	if err := model.ValidateSymbol(symbol); err != nil {
		return err
	}

	if dataRange == model.RangeUnspecified {
		return errors.Errorf("range not set")
	}

	for s, ch := range c.symbolToChartMap {
		if s == symbol && c.chartRange == dataRange {
			ch.SetLoading(true)
			ch.SetError(false)
		}
	}

	for s, th := range c.symbolToChartThumbMap {
		if s == symbol && c.chartThumbRange == dataRange {
			th.SetLoading(true)
			th.SetError(false)
		}
	}

	return nil
}

// onStockChartUpdate implements the eventHandler interface.
func (c *Controller) onStockChartUpdate(symbol string, ch *model.Chart) error {
	if err := model.ValidateSymbol(symbol); err != nil {
		return err
	}

	if err := model.ValidateChart(ch); err != nil {
		return err
	}

	if err := c.model.UpdateStockChart(symbol, ch); err != nil {
		return err
	}

	if ch, ok := c.symbolToChartMap[symbol]; ok {
		ch.SetLoading(false)

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
	}

	if th, ok := c.symbolToChartThumbMap[symbol]; ok {
		th.SetLoading(false)

		data, err := c.chartData(symbol, c.chartThumbRange)
		if err != nil {
			return err
		}

		if err := th.SetData(data); err != nil {
			return err
		}
	}

	return nil
}

// onStockChartUpdateError implements the eventHandler interface.
func (c *Controller) onStockChartUpdateError(symbol string, updateErr error) error {
	if err := model.ValidateSymbol(symbol); err != nil {
		return err
	}

	if ch, ok := c.symbolToChartMap[symbol]; ok {
		ch.SetLoading(false)
		ch.SetError(true)
	}

	if th, ok := c.symbolToChartThumbMap[symbol]; ok {
		th.SetLoading(false)
		th.SetError(true)
	}

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

func nextRange(r model.Range, zoomChange view.ZoomChange) model.Range {
	// zoomRanges are the ranges from most zoomed out to most zoomed in.
	var zoomRanges = []model.Range{
		model.OneYear,
		model.OneDay,
	}

	// Find the current zoom range.
	i := 0
	for j := range zoomRanges {
		if zoomRanges[j] == r {
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

	return zoomRanges[i]
}

func toConfig(model *model.Model) *config.Config {
	cfg := &config.Config{}
	if s := model.CurrentSymbol(); s != "" {
		cfg.CurrentStock = &config.Stock{Symbol: s}
	}
	for _, s := range model.SidebarSymbols() {
		cfg.Stocks = append(cfg.Stocks, &config.Stock{Symbol: s})
	}
	return cfg
}
