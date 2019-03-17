package controller

import (
	"context"
	"sync"
)

//go:generate stringer -type=signal
type signal int

const (
	signalUnspecified signal = iota
	refreshAllStocks
)

type signalController struct {
	// pendingSignals are the signals to be processed by the main thread.
	pendingSignals []signal

	// pendingMutex guards pendingSignals.
	pendingMutex *sync.Mutex
}

func newSignalController() *signalController {
	return &signalController{pendingMutex: new(sync.Mutex)}
}

// addPendingSignalsLocked locks the pendingSignals slice
// and adds the new signals to the existing slice.
func (c *signalController) addPendingSignalsLocked(signals []signal) {
	c.pendingMutex.Lock()
	defer c.pendingMutex.Unlock()
	c.pendingSignals = append(c.pendingSignals, signals...)
}

// takePendingSignalsLocked locks the pendingSignals slice,
// returns a copy of the current signals, and empties the existing signals.
func (c *signalController) takePendingSignalsLocked() []signal {
	c.pendingMutex.Lock()
	defer c.pendingMutex.Unlock()

	var ss []signal
	for _, s := range c.pendingSignals {
		ss = append(ss, s)
	}
	c.pendingSignals = nil
	return ss
}

func (c *Controller) processPendingSignals(ctx context.Context) error {
	for _, s := range c.signalController.takePendingSignalsLocked() {
		switch s {
		case refreshAllStocks:
			if err := c.refreshAllStocks(ctx); err != nil {
				return err
			}
		}
	}

	return nil
}
