// Package model contains code for the model in the MVC pattern.
package model

import (
	"regexp"
	"time"

	"github.com/btmura/ponzi2/internal/errs"
	"github.com/btmura/ponzi2/internal/logger"
)

// now is a function to get the current time. Mocked out in tests to return a fixed time.
var now = time.Now

// validSymbolRegexp is a regexp that accepts valid stock symbols. Examples: X, FB, SPY, AAPL
var validSymbolRegexp = regexp.MustCompile("^[A-Z]{1,5}$")

// Model models the app's state.
type Model struct {
	// currentSymbol is the symbol of the stock shown in the main area.
	currentSymbol string

	// sidebar contains the user's saved symbols.
	sidebar *Sidebar

	// symbol2Stock is map from symbol to Stock data.
	symbol2Stock map[string]*Stock
}

// Sidebar has slots which each have stock symbols.
type Sidebar struct {
	// Slots have one or more stock symbols.
	Slots []*Slot
}

// Slots have one or more stock symbols.
type Slot struct {
	Symbols []string
}

// Stock has a stock's symbol and charts.
type Stock struct {
	// Symbol is the stock's non-empty symbol.
	Symbol string

	// Quote is the stock's quote.
	Quote *Quote

	// Charts are the stock's unsorted charts. Nil initially.
	Charts []*Chart
}

// Chart has multiple series of data to be graphed.
type Chart struct {
	Interval               Interval
	TradingSessionSeries   *TradingSessionSeries
	MovingAverageSeriesSet []*AverageSeries
	AverageVolumeSeries    *AverageSeries
	LastUpdateTime         time.Time
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
	RealTimePrice
	FifteenMinuteDelayedPrice
	Close
	PreviousClose
	Price
	LastTrade
)

// Interval is the interval spanned by each trading session.
type Interval int

// Interval values.
//go:generate stringer -type=Interval
const (
	IntervalUnspecified Interval = iota
	Intraday
	Daily
	Weekly
)

// TradingSessionSeries is a time series of trading sessions.
type TradingSessionSeries struct {
	// TradingSessions are sorted by date in ascending order.
	TradingSessions []*TradingSession
}

// DeepCopy returns a deep copy of the series.
func (t *TradingSessionSeries) DeepCopy() *TradingSessionSeries {
	if t == nil {
		return nil
	}
	deep := *t
	if len(deep.TradingSessions) != 0 {
		deep.TradingSessions = make([]*TradingSession, len(t.TradingSessions))
		for i, ts := range t.TradingSessions {
			deep.TradingSessions[i] = ts.DeepCopy()
		}
	}
	return &deep
}

// TradingSession models a single trading session.
type TradingSession struct {
	Date                time.Time
	Source              Source
	Open                float32
	High                float32
	Low                 float32
	Close               float32
	Volume              int
	Change              float32
	PercentChange       float32
	VolumePercentChange float32
}

// DeepCopy returns a deep copy of the session.
func (s *TradingSession) DeepCopy() *TradingSession {
	if s == nil {
		return nil
	}
	deep := *s
	return &deep
}

// AverageSeries is a time series of average values like a moving average or average volume.
type AverageSeries struct {
	// Type is the average type like simple or exponential.
	Type AverageType

	// Intervals is how many days or weeks a moving average value spans.
	Intervals int

	// Values are sorted by date in ascending order.
	Values []*AverageValue
}

// DeepCopy returns a deep copy of the series.
func (a *AverageSeries) DeepCopy() *AverageSeries {
	if a == nil {
		return nil
	}
	deep := *a
	if len(deep.Values) != 0 {
		deep.Values = make([]*AverageValue, len(a.Values))
		for i, ma := range a.Values {
			deep.Values[i] = ma.DeepCopy()
		}
	}
	return &deep
}

// AverageType is the type of the average.
type AverageType int

// AverageType values.
//go:generate stringer -type=AverageType
const (
	AverageTypeUnspecified AverageType = iota
	Simple
	Exponential
)

// AverageValue is a single data point in an AverageSeries.
type AverageValue struct {
	// Date is the start date of the data point.
	Date time.Time

	// Value is the moving average value.
	Value float32
}

// DeepCopy returns a deep copy of the value.
func (a *AverageValue) DeepCopy() *AverageValue {
	if a == nil {
		return nil
	}
	deep := *a
	return &deep
}

// New creates a new Model.
func New() *Model {
	return &Model{
		sidebar:      new(Sidebar),
		symbol2Stock: make(map[string]*Stock),
	}
}

// CurrentSymbol returns the current symbol or empty string if no stock is set.
func (m *Model) CurrentSymbol() string {
	return m.currentSymbol
}

// SetCurrentSymbol sets the current stock symbol and returns true if the current symbol changed.
func (m *Model) SetCurrentSymbol(symbol string) (changed bool, err error) {
	if err := ValidateSymbol(symbol); err != nil {
		return false, err
	}

	if m.currentSymbol == symbol {
		return false, nil
	}

	old := m.currentSymbol

	m.currentSymbol = symbol

	// Remove the old stock if it's not in the model.
	if !m.containsSymbol(old) {
		delete(m.symbol2Stock, old)
	}

	// Add a stock placeholder for the new symbol if it doesn't exist.
	if m.symbol2Stock[symbol] == nil {
		m.symbol2Stock[symbol] = &Stock{Symbol: symbol}
	}

	return true, nil
}

// SidebarSymbols returns the sidebar's symbols.
func (m *Model) SidebarSymbols() []string {
	var symbols []string
	for _, slot := range m.sidebar.Slots {
		for _, s := range slot.Symbols {
			symbols = append(symbols, s)
		}
	}
	return symbols
}

// AddSidebarSymbol adds a symbol to the sidebar and returns true if the stock was newly added.
func (m *Model) AddSidebarSymbol(symbol string) (added bool, err error) {
	if err := ValidateSymbol(symbol); err != nil {
		return false, err
	}

	for _, slot := range m.sidebar.Slots {
		for _, s := range slot.Symbols {
			if s == symbol {
				return false, nil
			}
		}
	}

	m.sidebar.Slots = append(m.sidebar.Slots, &Slot{Symbols: []string{symbol}})

	// Add a stock placeholder for the new symbol if it doesn't exist.
	if m.symbol2Stock[symbol] == nil {
		m.symbol2Stock[symbol] = &Stock{Symbol: symbol}
	}

	return true, nil
}

// RemoveSidebarSymbol removes a symbol from the sidebar and returns true if removed.
func (m *Model) RemoveSidebarSymbol(symbol string) (removed bool, err error) {
	if err := ValidateSymbol(symbol); err != nil {
		return false, err
	}

	for _, slot := range m.sidebar.Slots {
		for i, s := range slot.Symbols {
			if s == symbol {
				slot.Symbols = append(slot.Symbols[:i], slot.Symbols[i+1:]...)
				if !m.containsSymbol(symbol) {
					delete(m.symbol2Stock, symbol)
				}
				return true, nil
			}
		}
	}

	return false, nil
}

// SwapSidebarSlots swaps two sidebar slots.
func (m *Model) SwapSidebarSlots(i, j int) (swapped bool) {
	if i < 0 || i >= len(m.sidebar.Slots) {
		logger.Errorf("slot index i (%d) is out of bounds (%d)", i, len(m.sidebar.Slots))
		return false
	}

	if j < 0 || j >= len(m.sidebar.Slots) {
		logger.Errorf("slot index j (%d) is out of bounds (%d)", j, len(m.sidebar.Slots))
		return false
	}

	if i == j {
		return false
	}

	m.sidebar.Slots[i], m.sidebar.Slots[j] = m.sidebar.Slots[j], m.sidebar.Slots[i]
	return true
}

// Stock returns the stock for the symbol if it is in the model. Nil otherwise.
func (m *Model) Stock(symbol string) (*Stock, error) {
	if err := ValidateSymbol(symbol); err != nil {
		return nil, err
	}
	return m.symbol2Stock[symbol], nil
}

// UpdateStockQuote updates the quote for a stock if the stock is in the model.
func (m *Model) UpdateStockQuote(symbol string, quote *Quote) error {
	if err := ValidateSymbol(symbol); err != nil {
		return err
	}

	if err := ValidateQuote(quote); err != nil {
		return err
	}

	st := m.symbol2Stock[symbol]

	// Don't do anything if the stock isn't in the model.
	if st == nil {
		return nil
	}

	st.Quote = quote

	return nil
}

// UpdateStockChart inserts or updates the chart for a stock if it is in the model.
func (m *Model) UpdateStockChart(symbol string, chart *Chart) error {
	if err := ValidateSymbol(symbol); err != nil {
		return err
	}

	if err := ValidateChart(chart); err != nil {
		return err
	}

	st := m.symbol2Stock[symbol]

	// Don't do anything if the stock isn't in the model.
	if st == nil {
		return nil
	}

	ch := &Chart{}
	*ch = *chart
	ch.LastUpdateTime = now()

	for i := range st.Charts {
		if st.Charts[i].Interval == ch.Interval {
			st.Charts[i] = ch
			return nil
		}
	}

	st.Charts = append(st.Charts, ch)

	return nil
}

// containsSymbol return true if the symbol is either the current symbol or in the sidebar.
func (m *Model) containsSymbol(symbol string) bool {
	if m.currentSymbol == symbol {
		return true
	}

	for _, slot := range m.sidebar.Slots {
		for _, s := range slot.Symbols {
			if s == symbol {
				return true
			}
		}
	}

	return false
}

// ValidateSymbol validates a symbol and returns an error if it's invalid.
func ValidateSymbol(symbol string) error {
	if !validSymbolRegexp.MatchString(symbol) {
		return errs.Errorf("bad symbol: got %s, want: %v", symbol, validSymbolRegexp)
	}
	return nil
}

// ValidateQuote validates a Quote and returns an error if it's invalid.
func ValidateQuote(q *Quote) error {
	if q == nil {
		return errs.Errorf("missing quote")
	}

	return nil
}

// ValidateChart validates a Chart and returns an error if it's invalid.
func ValidateChart(ch *Chart) error {
	if ch == nil {
		return errs.Errorf("missing chart")
	}

	if ch.Interval == IntervalUnspecified {
		return errs.Errorf("missing interval")
	}

	return nil
}
