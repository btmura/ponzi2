package model

import (
	"time"

	"github.com/golang/glog"
)

// Model models the app's state.
type Model struct {
	CurrentStock *Stock
	SavedStocks  []*Stock
}

// Stock models a single stock's data.
type Stock struct {
	Symbol                      string
	DailyTradingSessionSeries   *TradingSessionSeries
	DailyMovingAverageSeries25  *MovingAverageSeries
	DailyMovingAverageSeries50  *MovingAverageSeries
	DailyMovingAverageSeries200 *MovingAverageSeries
	DailyStochasticSeries       *StochasticSeries
	WeeklyStochasticSeries      *StochasticSeries
	LastUpdateTime              time.Time
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

// StockUpdate is an update that can be applied to a single stock.
type StockUpdate struct {
	Symbol                      string
	DailyTradingSessionSeries   *TradingSessionSeries
	DailyMovingAverageSeries25  *MovingAverageSeries
	DailyMovingAverageSeries50  *MovingAverageSeries
	DailyMovingAverageSeries200 *MovingAverageSeries
	DailyStochasticSeries       *StochasticSeries
	WeeklyStochasticSeries      *StochasticSeries
}

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

	if m.CurrentStock = m.stock(symbol); m.CurrentStock == nil {
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

	if st = m.stock(symbol); st == nil {
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

// UpdateStock updates a stock with the update if it is in the model.
func (m *Model) UpdateStock(u *StockUpdate) (st *Stock, updated bool) {
	if st = m.stock(u.Symbol); st == nil {
		return nil, false
	}
	st.DailyTradingSessionSeries = u.DailyTradingSessionSeries
	st.DailyStochasticSeries = u.DailyStochasticSeries
	st.WeeklyStochasticSeries = u.WeeklyStochasticSeries
	st.DailyMovingAverageSeries25 = u.DailyMovingAverageSeries25
	st.DailyMovingAverageSeries50 = u.DailyMovingAverageSeries50
	st.DailyMovingAverageSeries200 = u.DailyMovingAverageSeries200
	st.LastUpdateTime = time.Now()
	return st, true
}

func (m *Model) stock(symbol string) *Stock {
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

// Price returns the most recent price or 0 if no data.
func (m *Stock) Price() float32 {
	if m.DailyTradingSessionSeries == nil || len(m.DailyTradingSessionSeries.TradingSessions) == 0 {
		return 0
	}
	return m.DailyTradingSessionSeries.TradingSessions[len(m.DailyTradingSessionSeries.TradingSessions)-1].Close
}

// Change returns the most recent change or 0 if no data.
func (m *Stock) Change() float32 {
	if ts := m.recentSession(); ts != nil {
		return ts.Change
	}
	return 0
}

// PercentChange returns the most recent percent change or 0 if no data.
func (m *Stock) PercentChange() float32 {
	if ts := m.recentSession(); ts != nil {
		return ts.PercentChange
	}
	return 0
}

// Date returns the most recent date or zero time if no data.
func (m *Stock) Date() time.Time {
	if ts := m.recentSession(); ts != nil {
		return ts.Date
	}
	return time.Time{}
}

func (m *Stock) recentSession() *TradingSession {
	sr := m.DailyTradingSessionSeries
	if sr == nil {
		return nil
	}

	ts := sr.TradingSessions
	if len(ts) == 0 {
		return nil
	}

	return ts[len(ts)-1]
}
