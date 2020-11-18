package controller

import (
	"context"
	"sync"

	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/errs"
)

// event is a single event that the Controller should process on the main thread.
type event struct {
	symbol           string
	interval         model.Interval
	quote            *model.Quote
	chart            *model.Chart
	updateErr        error
	refreshAllStocks bool
	refreshStarted   bool
}

// eventController collects events in a queue. It is thread-safe.
type eventController struct {
	// queue is the queue where all events are stored.
	queue []event

	// queueMutex guards the queue.
	queueMutex *sync.Mutex

	// handler is the eventHandler that handles events.
	handler eventHandler
}

// eventHandler is an interface for handling all event types.
type eventHandler interface {
	onStockRefreshStarted(symbol string) error
	onStockUpdate(symbol string, q *model.Quote, ch *model.Chart) error
	onStockUpdateError(symbol string, updateErr error) error
	onRefreshAllStocksRequest(ctx context.Context) error
	onEventAdded()
}

func newEventController(handler eventHandler) *eventController {
	return &eventController{
		queueMutex: new(sync.Mutex),
		handler:    handler,
	}
}

// addEventLocked locks the queue and adds the new event to the queue.
func (c *eventController) addEventLocked(es ...event) {
	if len(es) == 0 {
		return
	}

	c.queueMutex.Lock()
	defer c.queueMutex.Unlock()

	c.queue = append(c.queue, es...)
	c.handler.onEventAdded()
}

// takeEventLocked locks the queue, takes an event from the queue, an returns it.
func (c *eventController) takeEventLocked() []event {
	c.queueMutex.Lock()
	defer c.queueMutex.Unlock()

	if len(c.queue) == 0 {
		return nil
	}

	var es []event
	es = append(es, c.queue[0])
	c.queue = c.queue[1:]

	return es
}

// process takes an event from the queue and processes it.
func (c *eventController) process(ctx context.Context) error {
	for _, e := range c.takeEventLocked() {
		switch {
		case e.updateErr != nil:
			if err := c.handler.onStockUpdateError(e.symbol, e.updateErr); err != nil {
				return err
			}

		case e.quote != nil, e.chart != nil:
			if err := c.handler.onStockUpdate(e.symbol, e.quote, e.chart); err != nil {
				return err
			}

		case e.refreshAllStocks:
			if err := c.handler.onRefreshAllStocksRequest(ctx); err != nil {
				return err
			}

		case e.refreshStarted:
			if err := c.handler.onStockRefreshStarted(e.symbol); err != nil {
				return err
			}

		default:
			return errs.Errorf("bad event: %v", e)
		}
	}

	return nil
}
