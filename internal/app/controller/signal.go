package controller

import "context"

//go:generate stringer -type=signal
type signal int

const (
	signalUnspecified signal = iota
	refreshCurrentStock
)

// addPendingSignalsLocked locks the pendingSignals slice
// and adds the new signals to the existing slice.
func (c *Controller) addPendingSignalsLocked(signals []signal) {
	c.pendingMutex.Lock()
	defer c.pendingMutex.Unlock()
	c.pendingSignals = append(c.pendingSignals, signals...)
}

// takePendingSignalsLocked locks the pendingSignals slice,
// returns a copy of the current signals, and empties the existing signals.
func (c *Controller) takePendingSignalsLocked() []signal {
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
	for _, s := range c.takePendingSignalsLocked() {
		switch s {
		case refreshCurrentStock:
			if err := c.refreshStocks(ctx, c.currentStockRefreshRequests()); err != nil {
				return err
			}
		}
	}

	return nil
}
