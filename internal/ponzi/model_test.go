package ponzi

import "testing"

func TestSetCurrentStock(t *testing.T) {
	m := NewModel()

	check := func(t *testing.T, got *ModelStock, gotChanged, wantChanged bool, wantMappings int) {
		if got == nil {
			t.Error("SetCurrentStock should NEVER return a nil stock.")
		}
		if gotChanged != wantChanged {
			t.Errorf("SetCurrentStock returned %t, wanted %t", gotChanged, wantChanged)
		}
		if len(m.symbolToStockMap) != wantMappings {
			t.Errorf("Model has %d mappings, want %d: %v", len(m.symbolToStockMap), wantMappings, keys(m))
		}
	}

	t.Run("initial stock", func(t *testing.T) {
		st, changed := m.SetCurrentStock("SPY")
		check(t, st, changed, true, 1)
	})

	t.Run("same stock", func(t *testing.T) {
		st, changed := m.SetCurrentStock("SPY")
		check(t, st, changed, false, 1)
	})

	t.Run("change stock", func(t *testing.T) {
		st, changed := m.SetCurrentStock("MO")
		check(t, st, changed, true, 1)
	})
}

func keys(m *Model) []string {
	var keys []string
	for s := range m.symbolToStockMap {
		keys = append(keys, s)
	}
	return keys
}
