// Package controller contains code for the controller in the MVC pattern.
package controller

import (
	"context"

	"github.com/btmura/ponzi2/internal/app/config"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view"
	"github.com/btmura/ponzi2/internal/app/view/chart"
	"github.com/btmura/ponzi2/internal/app/view/ui"
	"github.com/btmura/ponzi2/internal/errors"
	"github.com/btmura/ponzi2/internal/log"
	"github.com/btmura/ponzi2/internal/stock/iex"
)

// Controller runs the program in a "game loop".
type Controller struct {
	// model is the data that the Controller connects to the View.
	model *model.Model

	// ui is the UI that the Controller updates.
	ui *ui.UI

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
func New(iexClient iexClientInterface) *Controller {
	c := &Controller{
		model:       model.New(),
		ui:          ui.New(),
		configSaver: newConfigSaver(),
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

	c.ui.SetInputSymbolSubmittedCallback(func(symbol string) {
		if err := c.setChart(ctx, symbol); err != nil {
			log.Errorf("setChart: %v", err)
		}
	})

	c.ui.SetChartZoomChangeCallback(func(zoomChange view.ZoomChange) {
		if err := c.refreshCurrentStock(ctx); err != nil {
			log.Errorf("refreshCurrentStock: %v", err)
		}
	})

	c.ui.SetChartRefreshButtonClickCallback(func(symbol string) {
		if err := c.refreshAllStocks(ctx); err != nil {
			log.Errorf("refreshStocks: %v", err)
		}
	})

	c.ui.SetChartAddButtonClickCallback(func(symbol string) {
		if err := c.addChartThumb(ctx, symbol); err != nil {
			log.Errorf("addChartThumb: %v", err)
		}
	})

	c.ui.SetThumbRemoveButtonClickCallback(func(symbol string) {
		if err := c.removeChartThumb(symbol); err != nil {
			log.Errorf("removeChartThumb: %v", err)
		}
	})

	c.ui.SetThumbClickCallback(func(symbol string) {
		if err := c.setChart(ctx, symbol); err != nil {
			log.Errorf("setChart: %v", err)
		}
	})

	// Process config changes in the background until the program ends.
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

	data, err := c.chartData(symbol, c.ui.ChartRange())
	if err != nil {
		return err
	}

	if err := c.ui.SetChart(symbol, data); err != nil {
		return err
	}

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
		return c.stockRefresher.refreshOne(ctx, symbol, c.ui.ChartThumbRange())
	}

	data, err := c.chartData(symbol, c.ui.ChartThumbRange())
	if err != nil {
		return err
	}

	if err := c.ui.AddChartThumb(symbol, data); err != nil {
		return err
	}

	if err := c.stockRefresher.refreshOne(ctx, symbol, c.ui.ChartThumbRange()); err != nil {
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

	c.ui.RemoveChartThumb(symbol)

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
			data.Quote = st.Quote
			data.Chart = ch
			return data, nil
		}
	}

	return data, nil
}

func (c *Controller) refreshCurrentStock(ctx context.Context) error {
	d := new(dataRequestBuilder)
	if s := c.model.CurrentSymbol(); s != "" {
		if err := d.add([]string{s}, c.ui.ChartRange()); err != nil {
			return err
		}
	}
	return c.stockRefresher.refresh(ctx, d)
}

func (c *Controller) refreshAllStocks(ctx context.Context) error {
	d := new(dataRequestBuilder)

	if s := c.model.CurrentSymbol(); s != "" {
		if err := d.add([]string{s}, c.ui.ChartRange()); err != nil {
			return err
		}
	}

	if err := d.add(c.model.SidebarSymbols(), c.ui.ChartThumbRange()); err != nil {
		return err
	}

	return c.stockRefresher.refresh(ctx, d)
}

// onStockRefreshStarted implements the eventHandler interface.
func (c *Controller) onStockRefreshStarted(symbol string, dataRange model.Range) error {
	return c.ui.SetLoading(symbol, dataRange)
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
		data, err := c.chartData(symbol, c.ui.ChartRange())
		if err != nil {
			return err
		}

		return c.ui.SetData(symbol, data)
	}

	return nil
}

// onStockUpdateError implements the eventHandler interface.
func (c *Controller) onStockUpdateError(symbol string, updateErr error) error {
	log.Errorf("stock update for %s failed: %v\n", symbol, updateErr)
	return c.ui.SetError(symbol, updateErr)
}

// onEventAdded implements the eventHandler interface.
func (c *Controller) onEventAdded() {
	c.ui.WakeLoop()
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
