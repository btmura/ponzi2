package model

import (
	"testing"

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
		t.Errorf("AddSidebarSymbols should return true if the input symbol is new.")
	}
	if err != nil {
		t.Errorf("AddSidebarSymbols should not return an error if given a valid symbol.")
	}

	if diff := cmp.Diff([]string{"SPY"}, m.SidebarSymbols()); diff != "" {
		t.Errorf("diff (-want, +got)\n%s", diff)
	}

	added, err = m.AddSidebarSymbol("AAPL")
	if !added {
		t.Errorf("AddSidebarSymbols should return true if the input symbol is new.")
	}
	if err != nil {
		t.Errorf("AddSidebarSymbols should not return an error if given a valid symbol.")
	}

	if diff := cmp.Diff([]string{"SPY", "AAPL"}, m.SidebarSymbols()); diff != "" {
		t.Errorf("diff (-want, +got)\n%s", diff)
	}

	added, err = m.AddSidebarSymbol("AAPL")
	if added {
		t.Errorf("AddSidebarSymbols should return false if the input symbol exists.")
	}
	if err != nil {
		t.Errorf("AddSidebarSymbols should not return an error if given a valid symbol.")
	}

	if diff := cmp.Diff([]string{"SPY", "AAPL"}, m.SidebarSymbols()); diff != "" {
		t.Errorf("diff (-want, +got)\n%s", diff)
	}

	added, err = m.AddSidebarSymbol("AAPL AAPL")
	if added {
		t.Errorf("AddSidebarSymbols should return false if the input symbol is invalid.")
	}
	if err == nil {
		t.Errorf("AddSidebarSymbols should return an error if the given symbol is invalid.")
	}

	if diff := cmp.Diff([]string{"SPY", "AAPL"}, m.SidebarSymbols()); diff != "" {
		t.Errorf("diff (-want, +got)\n%s", diff)
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
			desc:    "too short",
			input:   "SP",
			wantErr: true,
		},
		{
			desc:    "empty string not allowed",
			input:   "",
			wantErr: true,
		},
	} {
		t.Run(tt.desc, func(t *testing.T) {
			gotErr := validateSymbol(tt.input) != nil
			if gotErr != tt.wantErr {
				t.Errorf("gotErr: %t, wantErr: %t", gotErr, tt.wantErr)
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
			gotErr := validateChart(tt.input) != nil
			if gotErr != tt.wantErr {
				t.Errorf("gotErr: %t, wantErr: %t", gotErr, tt.wantErr)
			}
		})
	}
}
