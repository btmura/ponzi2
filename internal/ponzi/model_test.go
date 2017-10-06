package ponzi

import "testing"

func TestSetCurrentStock(t *testing.T) {
	m := NewModel()

	st, changed := m.SetCurrentStock("SPY")
	if st == nil {
		t.Error("SetCurrentStock should return some stock.")
	}
	if !changed {
		t.Error("SetCurrentStock should report TRUE since the stock changed.")
	}
	if len(m.symbolToStockMap) != 1 {
		t.Errorf("SetCurrentStock should have stored a single mapping.")
	}

	st, changed = m.SetCurrentStock("SPY")
	if st == nil {
		t.Error("SetCurrentStock should STILL return some stock.")
	}
	if changed {
		t.Error("SetCurrentStock should report FALSE since the stock did NOT change.")
	}
	if len(m.symbolToStockMap) != 1 {
		t.Errorf("SetCurrentStock should STILL have a single mapping.")
	}
}
