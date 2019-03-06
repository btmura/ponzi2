// Package model contains code for the model in the MVC pattern.
package model

import (
	"time"

	"gitlab.com/btmura/ponzi2/internal/status"
)

// TODO(btmura): add validation functions for symbol, quote, etc.

// now is a function to get the current time. Mocked out in tests to return a fixed time.
var now = time.Now

// Model models the app's state.
type Model struct {
	// currentSymbol is the symbol of the stock shown in the main area.
	currentSymbol string

	// sidebarSymbols is an ordered list of symbols shown in the sidebar.
	sidebarSymbols []string

	// symbol2Stock is map from symbol to Stock data.
	symbol2Stock map[string]*Stock
}

// Stock has a stock's symbol and charts.
type Stock struct {
	// Symbol is the stock's non-empty symbol.
	Symbol string

	// Charts are the stock's unsorted charts. Nil initially.
	Charts []*Chart
}

// Chart has multiple series of data to be graphed.
type Chart struct {
	Quote                  *Quote
	Range                  Range
	TradingSessionSeries   *TradingSessionSeries
	MovingAverageSeries25  *MovingAverageSeries
	MovingAverageSeries50  *MovingAverageSeries
	MovingAverageSeries200 *MovingAverageSeries
	DailyStochasticSeries  *StochasticSeries
	WeeklyStochasticSeries *StochasticSeries
	LastUpdateTime         time.Time
}

// TradingSessionSeries is a time series of trading sessions.
type TradingSessionSeries struct {
	// TradingSessions are sorted by date in ascending order.
	TradingSessions []*TradingSession
}

// TradingSession models a single trading session.
type TradingSession struct {
	Date          time.Time
	Open          float32
	High          float32
	Low           float32
	Close         float32
	Volume        int
	Change        float32
	PercentChange float32
}

// Empty returns whether the TradingSession has any data to report.
func (s *TradingSession) Empty() bool {
	// IEX sets the low and high to -1 when it has no data to report.
	return s.Low < 0 || s.High < 0
}

// MovingAverageSeries is a time series of moving average values.
type MovingAverageSeries struct {
	// MovingAverages are sorted by date in ascending order.
	MovingAverages []*MovingAverage
}

// MovingAverage is a single data point in a MovingAverageSeries.
type MovingAverage struct {
	// Date is the start date of the data point.
	Date time.Time

	// Value is the moving average value.
	Value float32
}

// StochasticSeries is a time series of stochastic values.
type StochasticSeries struct {
	// Stochastics are sorted by date in ascending order.
	Stochastics []*Stochastic
}

// Stochastic is a single data point in a StochasticSeries.
type Stochastic struct {
	// Date is the start date of the data point.
	Date time.Time

	// K measures the stock's momentum.
	K float32

	// D is some moving average of K.
	D float32
}

// Quote is the latest quote for the stock.
type Quote struct {
	CompanyName   string
	LatestPrice   float32
	LatestSource  Source
	LatestTime    time.Time
	LatestUpdate  time.Time
	LatestVolume  int
	Open          float32
	High          float32
	Low           float32
	Close         float32
	Change        float32
	ChangePercent float32
}

// Source is the quote data source.
type Source int

// Source values.
//go:generate stringer -type=Source
const (
	SourceUnspecified Source = iota
	IEXRealTimePrice
	FifteenMinuteDelayedPrice
	Close
	PreviousClose
)

// Range is the range to specify in the request.
type Range int

// Range values.
//go:generate stringer -type=Range
const (
	RangeUnspecified Range = iota
	OneDay
	OneYear
)

// New creates a new Model.
func New() *Model {
	return &Model{
		symbol2Stock: map[string]*Stock{},
	}
}

// CurrentSymbol returns the current symbol or empty string if no stock is set.
func (m *Model) CurrentSymbol() string {
	return m.currentSymbol
}

// SetCurrentSymbol sets the current stock symbol and returns true if the current symbol changed.
func (m *Model) SetCurrentSymbol(symbol string) (changed bool, err error) {
	if err := validateSymbol(symbol); err != nil {
		return false, err
	}

	if m.currentSymbol == symbol {
		return false, nil
	}

	m.currentSymbol = symbol
	return true, nil
}

// SidebarSymbols returns the sidebar's symbols.
func (m *Model) SidebarSymbols() []string {
	var symbols []string
	for _, s := range m.sidebarSymbols {
		symbols = append(symbols, s)
	}
	return symbols
}

// AddSidebarSymbol adds a symbol to the sidebar and returns true if the stock was newly added.
func (m *Model) AddSidebarSymbol(symbol string) (added bool, err error) {
	if err := validateSymbol(symbol); err != nil {
		return false, err
	}

	for _, s := range m.sidebarSymbols {
		if s == symbol {
			return false, nil
		}
	}

	m.sidebarSymbols = append(m.sidebarSymbols, symbol)
	return true, nil
}

// RemoveSidebarSymbol removes a symbol from the sidebar and returns true if removed.
func (m *Model) RemoveSidebarSymbol(symbol string) (removed bool, err error) {
	if err := validateSymbol(symbol); err != nil {
		return false, err
	}

	for i, s := range m.sidebarSymbols {
		if s == symbol {
			m.sidebarSymbols = append(m.sidebarSymbols[:i], m.sidebarSymbols[i+1:]...)
			return true, nil
		}
	}
	return false, nil
}

// UpdateChart inserts or updates the chart for a stock if it is in the model.
func (m *Model) UpdateChart(symbol string, chart *Chart) error {
	if err := validateSymbol(symbol); err != nil {
		return err
	}

	if err := validateChart(chart); err != nil {
		return err
	}

	st := m.symbol2Stock[symbol]
	if st == nil {
		st = &Stock{Symbol: symbol}
		m.symbol2Stock[symbol] = st
	}

	ch := &Chart{}
	*ch = *chart
	ch.LastUpdateTime = now()

	for i := range st.Charts {
		if st.Charts[i].Range == ch.Range {
			st.Charts[i] = ch
			return nil
		}
	}

	st.Charts = append(st.Charts, ch)
	return nil
}

// Stock returns the stock for the symbol if it is in the model. Nil otherwise.
func (m *Model) Stock(symbol string) *Stock {
	return m.symbol2Stock[symbol]
}

func validateSymbol(symbol string) error {
	if symbol == "" {
		return status.Error("empty symbol")
	}
	return nil
}

func validateChart(ch *Chart) error {
	if ch == nil {
		return status.Error("missing chart")
	}

	if ch.Range == RangeUnspecified {
		return status.Errorf("missing range")
	}

	return nil
}
