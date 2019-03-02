// Package model contains code for the model in the MVC pattern.
package model

import (
	"time"

	"github.com/golang/glog"

	"gitlab.com/btmura/ponzi2/internal/util"
)

// TODO(btmura): check arguments in functions and return errors

// TODO(btmura): add validation functions for symbol, quote, etc.

// now is a function to get the current time. Mocked out in tests to return a fixed time.
var now = time.Now

// Model models the app's state.
type Model struct {
	CurrentStock *Stock
	SavedStocks  []*Stock
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
	return &Model{}
}

// SetCurrentStock sets the current stock by symbol.
// It returns the corresponding Stock and true if the current stock changed.
func (m *Model) SetCurrentStock(symbol string) (st *Stock, changed bool) {
	if symbol == "" {
		glog.V(2).Info("cannot set current stock to empty symbol")
	}

	if m.CurrentStock != nil && m.CurrentStock.Symbol == symbol {
		return m.CurrentStock, false
	}

	if m.CurrentStock = m.Stock(symbol); m.CurrentStock == nil {
		m.CurrentStock = &Stock{Symbol: symbol}
	}
	return m.CurrentStock, true
}

// AddSavedStock adds the stock by symbol.
// It returns the corresponding Stock and true if the stock was newly added.
func (m *Model) AddSavedStock(symbol string) (st *Stock, added bool) {
	if symbol == "" {
		glog.V(2).Info("cannot add empty symbol")
	}

	for _, st := range m.SavedStocks {
		if st.Symbol == symbol {
			return st, false
		}
	}

	if st = m.Stock(symbol); st == nil {
		st = &Stock{Symbol: symbol}
	}
	m.SavedStocks = append(m.SavedStocks, st)
	return st, true
}

// RemoveSavedStock removes the stock by symbol and returns true if removed.
func (m *Model) RemoveSavedStock(symbol string) (removed bool) {
	if symbol == "" {
		glog.V(2).Info("cannot remove empty symbol")
	}

	for i, st := range m.SavedStocks {
		if st.Symbol == symbol {
			m.SavedStocks = append(m.SavedStocks[:i], m.SavedStocks[i+1:]...)
			return true
		}
	}
	return false
}

// UpdateChart inserts or updates the chart for a stock if it is in the model.
func (m *Model) UpdateChart(symbol string, chart *Chart) error {
	if err := validateSymbol(symbol); err != nil {
		return err
	}

	if err := validateChart(chart); err != nil {
		return err
	}

	st := m.Stock(symbol)
	if st == nil {
		return nil
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
	if m.CurrentStock != nil && m.CurrentStock.Symbol == symbol {
		return m.CurrentStock
	}

	for _, st := range m.SavedStocks {
		if st.Symbol == symbol {
			return st
		}
	}

	return nil
}

func validateSymbol(symbol string) error {
	if symbol == "" {
		return util.Error("empty symbol")
	}
	return nil
}

func validateChart(ch *Chart) error {
	if ch == nil {
		return util.Error("missing chart")
	}

	if ch.Range == RangeUnspecified {
		return util.Errorf("missing range")
	}

	return nil
}
