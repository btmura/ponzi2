package controller

import (
	"context"
	"sync"

	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/status"
)

// event is a single event that the Controller should process on the main thread.
type event struct {
	symbol           string
	chart            *model.Chart
	updateErr        error
	refreshAllStocks bool
}

// eventController collects events in a queue. It is thread-safe.
type eventController struct {
	// queue is the queue where all events are stored.
	queue []event

	// queueMutex guards the queue.
	queueMutex *sync.Mutex
}

func newEventController() *eventController {
	return &eventController{queueMutex: new(sync.Mutex)}
}

// addEventLocked locks the queue and adds the new event to the queue.
func (c *eventController) addEventLocked(es ...event) {
	if len(es) == 0 {
		return
	}

	c.queueMutex.Lock()
	defer c.queueMutex.Unlock()
	c.queue = append(c.queue, es...)
}

// takeEventLocked locks the queue, takes an event from the queue, an returns it.
func (c *eventController) takeEventLocked() []event {
	c.queueMutex.Lock()
	defer c.queueMutex.Unlock()

	var es []event
	for _, u := range c.queue {
		es = append(es, u)
	}
	c.queue = nil
	return es
}

func (c *Controller) processStockUpdates(ctx context.Context) error {
	for _, e := range c.eventController.takeEventLocked() {
		switch {
		case e.updateErr != nil:
			if ch, ok := c.symbolToChartMap[e.symbol]; ok {
				ch.SetLoading(false)
				ch.SetError(true)
			}
			if th, ok := c.symbolToChartThumbMap[e.symbol]; ok {
				th.SetLoading(false)
				th.SetError(true)
			}

		case e.refreshAllStocks:
			if err := c.refreshAllStocks(ctx); err != nil {
				return err
			}

		case e.chart != nil:
			if err := c.model.UpdateStockChart(e.symbol, e.chart); err != nil {
				return err
			}

			if ch, ok := c.symbolToChartMap[e.symbol]; ok {
				ch.SetLoading(false)

				data, err := c.chartData(e.symbol, c.chartRange)
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

			if th, ok := c.symbolToChartThumbMap[e.symbol]; ok {
				th.SetLoading(false)

				data, err := c.chartData(e.symbol, c.chartThumbRange)
				if err != nil {
					return err
				}

				if err := th.SetData(data); err != nil {
					return err
				}
			}

		default:
			return status.Errorf("bad event: %v", e)
		}
	}

	return nil
}
