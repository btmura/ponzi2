package model

import (
	"testing"
)

func TestSetCurrentStock(t *testing.T) {
	check := func(t *testing.T, got *Stock, gotChanged bool, wantSymbol string, wantChanged bool) {
		if got == nil {
			t.Error("SetCurrentStock should NEVER return a nil stock.")
		}
		if got.Symbol != wantSymbol {
			t.Errorf("SetCurrentStock returned %s, wanted %s", got.Symbol, wantSymbol)
		}
		if gotChanged != wantChanged {
			t.Errorf("SetCurrentStock returned %t, wanted %t", gotChanged, wantChanged)
		}
	}

	m := New()

	// Set initial stock.
	st, changed := m.SetCurrentStock("SPY")
	check(t, st, changed, "SPY", true)

	// Set the same stock.
	st, changed = m.SetCurrentStock("SPY")
	check(t, st, changed, "SPY", false)

	// Change to a different stock.
	st, changed = m.SetCurrentStock("MO")
	check(t, st, changed, "MO", true)
}
