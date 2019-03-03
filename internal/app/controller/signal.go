package controller

import "context"

//go:generate stringer -type=controllerSignal
type controllerSignal int

const (
	signalUnspecified controllerSignal = iota
	signalRefreshCurrentStock
)

// addPendingSignalsLocked locks the pendingSignals slice
// and adds the new signals to the existing slice.
func (c *Controller) addPendingSignalsLocked(signals []controllerSignal) {
	c.pendingMutex.Lock()
	defer c.pendingMutex.Unlock()
	c.pendingSignals = append(c.pendingSignals, signals...)
}

// takePendingSignalsLocked locks the pendingSignals slice,
// returns a copy of the current signals, and empties the existing signals.
func (c *Controller) takePendingSignalsLocked() []controllerSignal {
	c.pendingMutex.Lock()
	defer c.pendingMutex.Unlock()

	var ss []controllerSignal
	for _, s := range c.pendingSignals {
		ss = append(ss, s)
	}
	c.pendingSignals = nil
	return ss
}

func (c *Controller) processPendingSignals(ctx context.Context) error {
	for _, s := range c.takePendingSignalsLocked() {
		switch s {
		case signalRefreshCurrentStock:
			if err := c.refreshStocks(ctx, c.currentStockRefreshRequests()); err != nil {
				return err
			}
		}
	}

	return nil
}
