// Package model contains code for the model in the MVC pattern.
package model

import (
	"regexp"
	"time"

	"github.com/btmura/ponzi2/internal/errors"
)

// now is a function to get the current time. Mocked out in tests to return a fixed time.
var now = time.Now

// validSymbolRegexp is a regexp that accepts valid stock symbols. Examples: X, FB, SPY, AAPL
var validSymbolRegexp = regexp.MustCompile("^[A-Z]{1,5}$")

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

	// Quote is the stock's quote.
	Quote *Quote

	// Charts are the stock's unsorted charts. Nil initially.
	Charts []*Chart
}

// Chart has multiple series of data to be graphed.
type Chart struct {
	Interval               Interval
	TradingSessionSeries   *TradingSessionSeries
	MovingAverageSeries20  *MovingAverageSeries
	MovingAverageSeries50  *MovingAverageSeries
	MovingAverageSeries200 *MovingAverageSeries
	AverageVolumeSeries    *AverageVolumeSeries
	DailyStochasticSeries  *StochasticSeries
	WeeklyStochasticSeries *StochasticSeries
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
	Date          time.Time
	Source        Source
	Open          float32
	High          float32
	Low           float32
	Close         float32
	Volume        int
	Change        float32
	PercentChange float32
}

// Skip returns true if the TradingSession should be skipped.
func (s *TradingSession) Skip() bool {
	// IEX sets the low and high to -1 when it has no data to report.
	return s.Open <= 0 || s.High <= 0 || s.Low <= 0 || s.Close <= 0
}

// DeepCopy returns a deep copy of the session.
func (s *TradingSession) DeepCopy() *TradingSession {
	if s == nil {
		return nil
	}
	deep := *s
	return &deep
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

// AverageVolumeSeries is a time series of average volume values.
type AverageVolumeSeries struct {
	// AverageVolumes are sorted by date in ascending order.
	AverageVolumes []*AverageVolume
}

// DeepCopy returns a deep copy of the series.
func (a *AverageVolumeSeries) DeepCopy() *AverageVolumeSeries {
	if a == nil {
		return nil
	}
	deep := *a
	if len(deep.AverageVolumes) != 0 {
		deep.AverageVolumes = make([]*AverageVolume, len(a.AverageVolumes))
		for i, av := range a.AverageVolumes {
			deep.AverageVolumes[i] = av.DeepCopy()
		}
	}
	return &deep
}

// AverageVolume is a single data point in a AverageVolumeSeries.
type AverageVolume struct {
	// Date is the start date of the data point.
	Date time.Time

	// Value is the average volume value.
	Value float32
}

// DeepCopy returns a deep copy of the volume.
func (a *AverageVolume) DeepCopy() *AverageVolume {
	if a == nil {
		return nil
	}
	deep := *a
	return &deep
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
	for _, s := range m.sidebarSymbols {
		symbols = append(symbols, s)
	}
	return symbols
}

// AddSidebarSymbol adds a symbol to the sidebar and returns true if the stock was newly added.
func (m *Model) AddSidebarSymbol(symbol string) (added bool, err error) {
	if err := ValidateSymbol(symbol); err != nil {
		return false, err
	}

	for _, s := range m.sidebarSymbols {
		if s == symbol {
			return false, nil
		}
	}

	m.sidebarSymbols = append(m.sidebarSymbols, symbol)

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

	for i, s := range m.sidebarSymbols {
		if s == symbol {
			m.sidebarSymbols = append(m.sidebarSymbols[:i], m.sidebarSymbols[i+1:]...)
			if !m.containsSymbol(symbol) {
				delete(m.symbol2Stock, symbol)
			}
			return true, nil
		}
	}

	return false, nil
}

// SetSidebarSymbol replaces the sidebar symbols with the given symbols.
func (m *Model) SetSidebarSymbols(symbols []string) error {
	for _, s := range symbols {
		if err := ValidateSymbol(s); err != nil {
			return err
		}
	}

	leftOver := map[string]bool{}
	for _, s := range m.sidebarSymbols {
		leftOver[s] = true
	}

	var newSidebarSymbols []string
	for _, s := range symbols {
		newSidebarSymbols = append(newSidebarSymbols, s)
		if !m.containsSymbol(s) {
			m.symbol2Stock[s] = &Stock{Symbol: s}
		}
		delete(leftOver, s)
	}
	m.sidebarSymbols = newSidebarSymbols

	for s := range leftOver {
		if !m.containsSymbol(s) {
			delete(m.symbol2Stock, s)
		}
	}

	return nil
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

	for _, s := range m.sidebarSymbols {
		if s == symbol {
			return true
		}
	}

	return false
}

// ValidateSymbol validates a symbol and returns an error if it's invalid.
func ValidateSymbol(symbol string) error {
	if !validSymbolRegexp.MatchString(symbol) {
		return errors.Errorf("bad symbol: got %s, want: %v", symbol, validSymbolRegexp)
	}
	return nil
}

// ValidateQuote validates a Quote and returns an error if it's invalid.
func ValidateQuote(q *Quote) error {
	if q == nil {
		return errors.Errorf("missing quote")
	}

	return nil
}

// ValidateChart validates a Chart and returns an error if it's invalid.
func ValidateChart(ch *Chart) error {
	if ch == nil {
		return errors.Errorf("missing chart")
	}

	if ch.Interval == IntervalUnspecified {
		return errors.Errorf("missing interval")
	}

	return nil
}
