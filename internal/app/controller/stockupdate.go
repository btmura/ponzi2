package controller

import (
	"context"
	"sync"

	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/status"
)

type stockUpdate struct {
	symbol    string
	chart     *model.Chart
	updateErr error
}

type stockUpdateController struct {
	// pendingStockUpdates are the updates to be processed by the main thread.
	pendingStockUpdates []stockUpdate

	// pendingMutex guards pendingUpdates.
	pendingMutex *sync.Mutex
}

func newStockUpdateController() *stockUpdateController {
	return &stockUpdateController{pendingMutex: new(sync.Mutex)}
}

// addPendingStockUpdatesLocked locks the pendingStockUpdates slice
// and adds the new stock updates to the existing slice.
func (c *stockUpdateController) addPendingStockUpdatesLocked(us []stockUpdate) {
	c.pendingMutex.Lock()
	defer c.pendingMutex.Unlock()
	c.pendingStockUpdates = append(c.pendingStockUpdates, us...)
}

// takePendingStockUpdatesLocked locks the pendingStockUpdates slice,
// returns a copy of the updates, and empties the existing updates.
func (c *stockUpdateController) takePendingStockUpdatesLocked() []stockUpdate {
	c.pendingMutex.Lock()
	defer c.pendingMutex.Unlock()

	var us []stockUpdate
	for _, u := range c.pendingStockUpdates {
		us = append(us, u)
	}
	c.pendingStockUpdates = nil
	return us
}

func (c *Controller) processStockUpdates(ctx context.Context) error {
	for _, u := range c.stockUpdateController.takePendingStockUpdatesLocked() {
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
			if err := c.model.UpdateStockChart(u.symbol, u.chart); err != nil {
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
