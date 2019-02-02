package model

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
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

func TestUpdateStock(t *testing.T) {
	old := now
	defer func() { now = old }()
	n := time.Date(2018, time.October, 11, 0, 0, 0, 0, time.UTC)
	now = func() time.Time { return n }

	m := New()

	st, changed := m.SetCurrentStock("SPY")
	if !changed {
		t.Errorf("SetCurrentStock should return changed for new symbols.")
	}

	st2, updated := m.UpdateStock(&StockUpdate{
		Symbol: "SPY",
		Range:  OneDay,
	})
	if !updated {
		t.Errorf("UpdateStock should return updated for existing symbols.")
	}

	if st != st2 {
		t.Errorf("UpdateStock should return the same pointer to the existing stock.\n%v\n%v", st, st2)
	}

	// TODO(btmura): check more fields in want

	want := &Stock{
		Symbol:         "SPY",
		Range:          OneDay,
		LastUpdateTime: n,
	}

	if diff := cmp.Diff(want, st2); diff != "" {
		t.Errorf("differs: (-want, +got)\n%s", diff)
	}
}
