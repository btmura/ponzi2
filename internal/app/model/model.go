package model

import (
	"log"
	"os"
	"time"
)

var logger = log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)

// Model keeps track of the user's stocks.
type Model struct {
	// CurrentStock is the stock currently being viewed.
	CurrentStock *Stock

	// SavedStocks is the user's saved stocks.
	SavedStocks []*Stock
}

// Stock models a single stock.
type Stock struct {
	// Symbol is the symbol of the stock.
	Symbol string

	// DailySessions are trading sessions that span a single day.
	DailySessions []*TradingSession

	// WeeklySessions are trading sessions that span a single week.
	WeeklySessions []*TradingSession

	// DailyStochastics is the daily stochastics series.
	DailyStochastics *Stochastics

	// WeeklyStochastics is the weekly stochastic series.
	WeeklyStochastics *Stochastics

	// LastUpdateTime is when the ModelStock was last updated.
	LastUpdateTime time.Time
}

// TradingSession models a single trading session.
type TradingSession struct {
	Date   time.Time
	Open   float32
	High   float32
	Low    float32
	Close  float32
	Volume int

	Change        float32
	PercentChange float32

	MovingAverage25  float32
	MovingAverage50  float32
	MovingAverage200 float32
}

// Stochastics is a time series of stochastic values.
type Stochastics struct {
	// Values are the stochastic values with earlier values in front.
	Values []*StochasticValue
}

// StochasticValue are the stochastic values for some date.
type StochasticValue struct {
	// Date is the start date of the time span covered by this value.
	Date time.Time

	// K measures the stock's momentum.
	K float32

	// D is some moving average of K.
	D float32
}

// StockUpdate is an update that can be applied to the model.
type StockUpdate struct {
	// Symbol is the symbol of the stock.
	Symbol string

	// DailySessions are trading sessions that span a single day.
	DailySessions []*TradingSession

	// WeeklySessions are trading sessions that span a single week.
	WeeklySessions []*TradingSession

	// DailyStochastics is the daily stochastics series.
	DailyStochastics *Stochastics

	// WeeklyStochastics is the weekly stochastic series.
	WeeklyStochastics *Stochastics
}

// NewModel creates a new Model.
func NewModel() *Model {
	return &Model{}
}

// SetCurrentStock sets the current stock by symbol. It returns the
// corresponding ModelStock and true if the current stock changed.
func (m *Model) SetCurrentStock(symbol string) (st *Stock, changed bool) {
	if symbol == "" {
		logger.Print("SetCurrentStock: cannot set current stock to empty symbol")
	}

	if m.CurrentStock != nil && m.CurrentStock.Symbol == symbol {
		return m.CurrentStock, false
	}

	if m.CurrentStock = m.stock(symbol); m.CurrentStock == nil {
		m.CurrentStock = &Stock{Symbol: symbol}
	}
	return m.CurrentStock, true
}

// AddSavedStock adds the stock by symbol. It returns the corresponding
// ModelStock and true if the stock was newly added.
func (m *Model) AddSavedStock(symbol string) (st *Stock, added bool) {
	if symbol == "" {
		logger.Print("AddSavedStock: cannot add empty symbol")
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
		logger.Print("RemovedSavedStock: cannot remove empty symbol")
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
func (m *Model) UpdateStock(update *StockUpdate) (st *Stock, updated bool) {
	if st = m.stock(update.Symbol); st == nil {
		return nil, false
	}
	st.DailySessions = update.DailySessions
	st.WeeklySessions = update.WeeklySessions
	st.DailyStochastics = update.DailyStochastics
	st.WeeklyStochastics = update.WeeklyStochastics
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
	if len(m.DailySessions) == 0 {
		return 0
	}
	return m.DailySessions[len(m.DailySessions)-1].Close
}

// Change returns the most recent change or 0 if no data.
func (m *Stock) Change() float32 {
	if len(m.DailySessions) == 0 {
		return 0
	}
	return m.DailySessions[len(m.DailySessions)-1].Change
}

// PercentChange returns the most recent percent change or 0 if no data.
func (m *Stock) PercentChange() float32 {
	if len(m.DailySessions) == 0 {
		return 0
	}
	return m.DailySessions[len(m.DailySessions)-1].PercentChange
}

// Date returns the most recent date or zero time if no data.
func (m *Stock) Date() time.Time {
	if len(m.DailySessions) == 0 {
		return time.Time{}
	}
	return m.DailySessions[len(m.DailySessions)-1].Date
}
