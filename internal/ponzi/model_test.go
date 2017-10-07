package ponzi

import "testing"

func TestSetCurrentStock(t *testing.T) {
	m := NewModel()

	check := func(t *testing.T, got *ModelStock, gotChanged bool, wantSymbol string, wantChanged bool) {
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

	t.Run("initial stock", func(t *testing.T) {
		st, changed := m.SetCurrentStock("SPY")
		check(t, st, changed, "SPY", true)
	})

	t.Run("same stock", func(t *testing.T) {
		st, changed := m.SetCurrentStock("SPY")
		check(t, st, changed, "SPY", false)
	})

	t.Run("change stock", func(t *testing.T) {
		st, changed := m.SetCurrentStock("MO")
		check(t, st, changed, "MO", true)
	})
}
