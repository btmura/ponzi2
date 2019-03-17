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
	dataRange        model.Range
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

	// proc is the eventProcessor that processes events.
	proc eventProcessor
}

// eventProcessor is an interface for handling all event types.
type eventProcessor interface {
	processStockRefreshStarted(symbol string, dataRange model.Range) error
	processStockChartUpdate(symbol string, ch *model.Chart) error
	processStockChartUpdateError(symbol string, updateErr error) error
	processRefreshAllStocks(ctx context.Context) error
	notifyProcessor()
}

func newEventController(proc eventProcessor) *eventController {
	return &eventController{
		queueMutex: new(sync.Mutex),
		proc:       proc,
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
	c.proc.notifyProcessor()
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

func (c *eventController) process(ctx context.Context) error {
	for _, e := range c.takeEventLocked() {
		switch {
		case e.updateErr != nil:
			if err := c.proc.processStockChartUpdateError(e.symbol, e.updateErr); err != nil {
				return err
			}

		case e.chart != nil:
			if err := c.proc.processStockChartUpdate(e.symbol, e.chart); err != nil {
				return err
			}

		case e.refreshAllStocks:
			if err := c.proc.processRefreshAllStocks(ctx); err != nil {
				return err
			}

		case e.refreshStarted:
			if err := c.proc.processStockRefreshStarted(e.symbol, e.dataRange); err != nil {
				return err
			}

		default:
			return status.Errorf("bad event: %v", e)
		}
	}

	return nil
}
