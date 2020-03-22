package model

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestSetCurrentSymbol(t *testing.T) {
	m := New()

	if diff := cmp.Diff("", m.CurrentSymbol()); diff != "" {
		t.Errorf("diff (-want, +got)\n%s", diff)
	}

	changed, err := m.SetCurrentSymbol("SPY")
	if !changed {
		t.Errorf("SetCurrentSymbol should return true if the input symbol is different.")
	}
	if err != nil {
		t.Errorf("SetCurrentSymbol should not return an error if given a valid symbol.")
	}

	if diff := cmp.Diff("SPY", m.CurrentSymbol()); diff != "" {
		t.Errorf("diff (-want, +got)\n%s", diff)
	}

	changed, err = m.SetCurrentSymbol("SPY")
	if changed {
		t.Errorf("SetCurrentSymbol should return false if the input symbol is the same.")
	}
	if err != nil {
		t.Errorf("SetCurrentSymbol should not return an error if given a valid symbol even if it is the same.")
	}

	if diff := cmp.Diff("SPY", m.CurrentSymbol()); diff != "" {
		t.Errorf("diff (-want, +got)\n%s", diff)
	}

	changed, err = m.SetCurrentSymbol("SPYSPY")
	if changed {
		t.Errorf("SetCurrentSymbol should return false if the given symbol is invalid.")
	}
	if err == nil {
		t.Errorf("SetCurrentSymbol should return an error if the given symbol is invalid.")
	}

	if diff := cmp.Diff("SPY", m.CurrentSymbol()); diff != "" {
		t.Errorf("diff (-want, +got)\n%s", diff)
	}
}

func TestAddSidebarSymbol(t *testing.T) {
	m := New()

	if diff := cmp.Diff([]string(nil), m.SidebarSymbols()); diff != "" {
		t.Errorf("diff (-want, +got)\n%s", diff)
	}

	added, err := m.AddSidebarSymbol("SPY")
	if !added {
		t.Errorf("AddSidebarSymbol should return true if the input symbol is new.")
	}
	if err != nil {
		t.Errorf("AddSidebarSymbol should not return an error if given a valid symbol.")
	}

	if diff := cmp.Diff([]string{"SPY"}, m.SidebarSymbols()); diff != "" {
		t.Errorf("diff (-want, +got)\n%s", diff)
	}

	added, err = m.AddSidebarSymbol("AAPL")
	if !added {
		t.Errorf("AddSidebarSymbol should return true if the input symbol is new.")
	}
	if err != nil {
		t.Errorf("AddSidebarSymbol should not return an error if given a valid symbol.")
	}

	if diff := cmp.Diff([]string{"SPY", "AAPL"}, m.SidebarSymbols()); diff != "" {
		t.Errorf("diff (-want, +got)\n%s", diff)
	}

	added, err = m.AddSidebarSymbol("AAPL")
	if added {
		t.Errorf("AddSidebarSymbol should return false if the input symbol exists.")
	}
	if err != nil {
		t.Errorf("AddSidebarSymbol should not return an error if given a valid symbol.")
	}

	if diff := cmp.Diff([]string{"SPY", "AAPL"}, m.SidebarSymbols()); diff != "" {
		t.Errorf("diff (-want, +got)\n%s", diff)
	}

	added, err = m.AddSidebarSymbol("AAPL AAPL")
	if added {
		t.Errorf("AddSidebarSymbol should return false if the input symbol is invalid.")
	}
	if err == nil {
		t.Errorf("AddSidebarSymbol should return an error if the given symbol is invalid.")
	}

	if diff := cmp.Diff([]string{"SPY", "AAPL"}, m.SidebarSymbols()); diff != "" {
		t.Errorf("diff (-want, +got)\n%s", diff)
	}
}

func TestRemoveSidebarSymbol(t *testing.T) {
	m := New()

	m.AddSidebarSymbol("SPY")
	m.AddSidebarSymbol("AAPL")
	m.AddSidebarSymbol("CEF")

	if diff := cmp.Diff([]string{"SPY", "AAPL", "CEF"}, m.SidebarSymbols()); diff != "" {
		t.Errorf("diff (-want, +got)\n%s", diff)
	}

	removed, err := m.RemoveSidebarSymbol("AAPL")
	if !removed {
		t.Errorf("RemoveSidebarSymbol should return true if the input symbol is in the sidebar.")
	}
	if err != nil {
		t.Errorf("RemoveSidebarSypmbol should not return an error if the given symbol is valid.")
	}

	if diff := cmp.Diff([]string{"SPY", "CEF"}, m.SidebarSymbols()); diff != "" {
		t.Errorf("diff (-want, +got)\n%s", diff)
	}

	removed, err = m.RemoveSidebarSymbol("FB")
	if removed {
		t.Errorf("RemoveSidebarSymbol should return false if the input symbol is not in the sidebar.")
	}
	if err != nil {
		t.Errorf("RemoveSidebarSymbol should not return an error if the given symbol is valid.")
	}

	if diff := cmp.Diff([]string{"SPY", "CEF"}, m.SidebarSymbols()); diff != "" {
		t.Errorf("diff (-want, +got)\n%s", diff)
	}
}

func TestSetSidebarSymbols(t *testing.T) {
	m := New()

	m.AddSidebarSymbol("SPY")
	m.AddSidebarSymbol("AAPL")
	m.AddSidebarSymbol("CEF")

	if diff := cmp.Diff([]string{"SPY", "AAPL", "CEF"}, m.SidebarSymbols()); diff != "" {
		t.Errorf("diff (-want, +got)\n%s", diff)
	}

	m.SetSidebarSymbols([]string{"SPY", "CEF", "AAPL"})

	if diff := cmp.Diff([]string{"SPY", "CEF", "AAPL"}, m.SidebarSymbols()); diff != "" {
		t.Errorf("diff (-want, +got)\n%s", diff)
	}

	m.SetSidebarSymbols([]string{"SPY"})

	if diff := cmp.Diff([]string{"SPY"}, m.SidebarSymbols()); diff != "" {
		t.Errorf("diff (-want, +got)\n%s", diff)
	}
}

func TestUpdateStockQuote(t *testing.T) {
	old := now
	defer func() { now = old }()
	now = func() time.Time { return time.Date(2019, time.March, 10, 30, 0, 0, 0, time.UTC) }

	m := New()

	// Add SPY to the model so UpdateStockQuote works.
	m.SetCurrentSymbol("SPY")

	if err := m.UpdateStockQuote("", &Quote{}); err == nil {
		t.Errorf("UpdateStockQuote should return an error when the input symbol is invalid.")
	}

	if err := m.UpdateStockQuote("SPY", nil /* quote can't be nil */); err == nil {
		t.Errorf("UpdateStockQuote should return an error when the input quote is invalid.")
	}

	q1 := &Quote{Close: 1337}
	if err := m.UpdateStockQuote("SPY", q1); err != nil {
		t.Errorf("UpdateStockQuote should not return an error if the inputs are valid.")
	}

	want := &Stock{
		Symbol: "SPY",
		Quote:  &Quote{Close: 1337},
	}

	st, err := m.Stock("SPY")
	if err != nil {
		t.Errorf("Stock should not return an error if the given symbol is valid.")
	}
	if diff := cmp.Diff(want, st); diff != "" {
		t.Errorf("diff (-want, +got)\n%s", diff)
	}

	q2 := &Quote{Open: 42}
	if err := m.UpdateStockQuote("SPY", q2); err != nil {
		t.Errorf("UpdateStockQuote should not return an error if the inputs are valid.")
	}

	want = &Stock{
		Symbol: "SPY",
		Quote:  &Quote{Open: 42},
	}

	st, err = m.Stock("SPY")
	if err != nil {
		t.Errorf("Stock should not return an error if the given symbol is valid.")
	}
	if diff := cmp.Diff(want, st); diff != "" {
		t.Errorf("diff (-want, +got)\n%s", diff)
	}
}

func TestUpdateStockChart(t *testing.T) {
	old := now
	defer func() { now = old }()
	now = func() time.Time { return time.Date(2019, time.March, 10, 30, 0, 0, 0, time.UTC) }

	m := New()

	// Add SPY to the model so UpdateStockChart works.
	m.SetCurrentSymbol("SPY")

	if err := m.UpdateStockChart("", &Chart{Range: OneDay}); err == nil {
		t.Errorf("UpdateStockChart should return an error when the input symbol is invalid.")
	}

	if err := m.UpdateStockChart("SPY", nil /* chart can't be nil */); err == nil {
		t.Errorf("UpdateStockChart should return an error when the input chart is invalid.")
	}

	c1 := &Chart{Range: OneDay}
	if err := m.UpdateStockChart("SPY", c1); err != nil {
		t.Errorf("UpdateStcokChart should not return an error if the inputs are valid.")
	}

	want := &Stock{
		Symbol: "SPY",
		Charts: []*Chart{
			{
				Range:          OneDay,
				LastUpdateTime: time.Date(2019, time.March, 10, 30, 0, 0, 0, time.UTC),
			},
		},
	}

	st, err := m.Stock("SPY")
	if err != nil {
		t.Errorf("Stock should not return an error if the given symbol is valid.")
	}
	if diff := cmp.Diff(want, st); diff != "" {
		t.Errorf("diff (-want, +got)\n%s", diff)
	}

	c2 := &Chart{Range: OneYear}
	if err := m.UpdateStockChart("SPY", c2); err != nil {
		t.Errorf("UpdateStockChart should not return an error if the inputs are valid.")
	}

	want = &Stock{
		Symbol: "SPY",
		Charts: []*Chart{
			{
				Range:          OneDay,
				LastUpdateTime: time.Date(2019, time.March, 10, 30, 0, 0, 0, time.UTC),
			},
			{
				Range:          OneYear,
				LastUpdateTime: time.Date(2019, time.March, 10, 30, 0, 0, 0, time.UTC),
			},
		},
	}

	st, err = m.Stock("SPY")
	if err != nil {
		t.Errorf("Stock should not return an error if the given symbol is valid.")
	}
	if diff := cmp.Diff(want, st); diff != "" {
		t.Errorf("diff (-want, +got)\n%s", diff)
	}
}

func TestStockBookkeeping(t *testing.T) {
	m := New()
	if st, _ := m.Stock("SPY"); st != nil {
		t.Errorf("Stock should return nil if the symbol is not in the model.")
	}

	m.SetCurrentSymbol("SPY")
	if st, _ := m.Stock("SPY"); st == nil {
		t.Errorf("SetCurrentSymbol should insert the new Stock.")
	}

	m.SetCurrentSymbol("AAPL")
	if st, _ := m.Stock("SPY"); st != nil {
		t.Error("SetCurrentSymbol should remove the unused Stock.")
	}

	m.AddSidebarSymbol("FB")
	if st, _ := m.Stock("FB"); st == nil {
		t.Errorf("AddSidebarSymbol should insert the new Stock.")
	}

	m.RemoveSidebarSymbol("FB")
	if st, _ := m.Stock("FB"); st != nil {
		t.Error("RemoveSidebarSymbol should remove the unused Stock.")
	}
}

func TestContainsSymbol(t *testing.T) {
	m := New()

	if m.containsSymbol("SPY") {
		t.Errorf("containsSymbol should return false, since nothing was added yet.")
	}

	m.SetCurrentSymbol("SPY")

	if !m.containsSymbol("SPY") {
		t.Errorf("containsSymbol should return true, since the current symbol was set to SPY.")
	}

	m.SetCurrentSymbol("AAPL")

	if m.containsSymbol("SPY") {
		t.Errorf("containsSymbol should return false, since the current symbol changed to AAPL.")
	}

	m.AddSidebarSymbol("SPY")

	if !m.containsSymbol("SPY") {
		t.Errorf("containsSymbol should return true, since the sidebar now has SPY.")
	}

	m.RemoveSidebarSymbol("SPY")

	if m.containsSymbol("SPY") {
		t.Errorf("containsSymbol should return false, since the sidebar no longer has SPY.")
	}
}

func TestValidateSymbol(t *testing.T) {
	for _, tt := range []struct {
		desc    string
		input   string
		wantErr bool
	}{
		{
			desc:  "valid three letter symbol",
			input: "SPY",
		},
		{
			desc:  "valid four letter symbol",
			input: "QQQQ",
		},
		{
			desc:    "lowercase not allowed",
			input:   "spy",
			wantErr: true,
		},
		{
			desc:    "spaces not allowed",
			input:   "S P Y",
			wantErr: true,
		},
		{
			desc:    "too long",
			input:   "SPYSPY",
			wantErr: true,
		},
		{
			desc:    "empty string not allowed",
			input:   "",
			wantErr: true,
		},
	} {
		t.Run(tt.desc, func(t *testing.T) {
			gotErr := ValidateSymbol(tt.input)

			if (gotErr != nil) != tt.wantErr {
				t.Errorf("got error: %v, wanted err: %t", gotErr, tt.wantErr)
			}
		})
	}
}

func TestValidateChart(t *testing.T) {
	for _, tt := range []struct {
		desc    string
		input   *Chart
		wantErr bool
	}{
		{
			desc:  "valid",
			input: &Chart{Range: OneDay},
		},
		{
			desc:    "missing range",
			input:   &Chart{},
			wantErr: true,
		},
		{
			desc:    "nil chart",
			input:   nil,
			wantErr: true,
		},
	} {
		t.Run(tt.desc, func(t *testing.T) {
			gotErr := ValidateChart(tt.input)

			if (gotErr != nil) != tt.wantErr {
				t.Errorf("got error: %v, wanted err: %t", gotErr, tt.wantErr)
			}
		})
	}
}
