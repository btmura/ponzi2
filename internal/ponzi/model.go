package ponzi

import (
	"sort"
	"time"

	"github.com/btmura/ponzi2/internal/stock"
)

// Model keeps track of the user's stocks.
type Model struct {
	// CurrentStock is the stock currently being viewed.
	CurrentStock *ModelStock

	// Stocks are the user's other stocks.
	Stocks []*ModelStock
}

// NewModel creates a new Model.
func NewModel(currentSymbol string, symbols []string) *Model {
	cs := NewModelStock(currentSymbol)
	var sts []*ModelStock
	for _, s := range symbols {
		if s == currentSymbol {
			sts = append(sts, cs)
		} else {
			sts = append(sts, NewModelStock(s))
		}
	}
	return &Model{
		CurrentStock: cs,
		Stocks:       sts,
	}
}

// AddStock adds a stock to the model.
func (m *Model) AddStock(st *ModelStock) bool {
	if m.Stock(st.Symbol) != nil {
		return false // Already have it.
	}
	m.Stocks = append(m.Stocks, st)
	return true
}

// RemoveStock removes a stock from the model.
func (m *Model) RemoveStock(st *ModelStock) bool {
	if m.Stock(st.Symbol) == nil {
		return false // Don't have it.
	}

	for i, stock := range m.Stocks {
		if stock.Symbol == st.Symbol {
			m.Stocks = append(m.Stocks[:i], m.Stocks[i+1:]...)
			break
		}
	}

	return true
}

// Stock returns the stock with the symbol or nil if the sidebar doesn't have it.
func (m *Model) Stock(symbol string) *ModelStock {
	for _, st := range m.Stocks {
		if st.Symbol == symbol {
			return st
		}
	}
	return nil
}

// ModelStock models a single stock.
type ModelStock struct {
	Symbol          string
	DailySessions   []*ModelTradingSession
	WeeklySessions  []*ModelTradingSession
	LastUpdateTime  time.Time
	changeCallbacks []func()
}

// ModelTradingSession models a single trading session.
type ModelTradingSession struct {
	Date          time.Time
	Open          float32
	High          float32
	Low           float32
	Close         float32
	Volume        int
	Change        float32
	PercentChange float32
	K             float32
	D             float32
}

// NewModelStock creates a new ModelStock.
func NewModelStock(symbol string) *ModelStock {
	return &ModelStock{Symbol: symbol}
}

// AddChangeCallback adds a callback to be fired when the stock changes.
func (m *ModelStock) AddChangeCallback(cb func()) {
	m.changeCallbacks = append(m.changeCallbacks, cb)
}

// Price returns the most recent price or 0 if no data.
func (m *ModelStock) Price() float32 {
	if len(m.DailySessions) == 0 {
		return 0
	}
	return m.DailySessions[len(m.DailySessions)-1].Close
}

// Change returns the most recent change or 0 if no data.
func (m *ModelStock) Change() float32 {
	if len(m.DailySessions) == 0 {
		return 0
	}
	return m.DailySessions[len(m.DailySessions)-1].Change
}

// PercentChange returns the most recent percent change or 0 if no data.
func (m *ModelStock) PercentChange() float32 {
	if len(m.DailySessions) == 0 {
		return 0
	}
	return m.DailySessions[len(m.DailySessions)-1].PercentChange
}

// Update updates the stock with the new data.
func (m *ModelStock) Update(hist *stock.TradingHistory) error {
	m.DailySessions, m.WeeklySessions = convertSessions(hist.Sessions)
	m.LastUpdateTime = time.Now()
	for _, cb := range m.changeCallbacks {
		cb()
	}
	return nil
}

func convertSessions(sessions []*stock.TradingSession) (dailySessions, weeklySessions []*ModelTradingSession) {
	// Convert the trading sessions into daily sessions.
	var ds []*ModelTradingSession
	for _, s := range sessions {
		ds = append(ds, &ModelTradingSession{
			Date:   s.Date,
			Open:   s.Open,
			High:   s.High,
			Low:    s.Low,
			Close:  s.Close,
			Volume: s.Volume,
		})
	}
	sortByModelTradingSessionDate(ds)

	// Convert the daily sessions into weekly sessions.
	var ws []*ModelTradingSession
	for _, s := range ds {
		diffWeek := ws == nil
		if !diffWeek {
			_, week := s.Date.ISOWeek()
			_, prevWeek := ws[len(ws)-1].Date.ISOWeek()
			diffWeek = week != prevWeek
		}

		if diffWeek {
			sc := *s
			ws = append(ws, &sc)
		} else {
			ls := ws[len(ws)-1]
			if ls.High < s.High {
				ls.High = s.High
			}
			if ls.Low > s.Low {
				ls.Low = s.Low
			}
			ls.Close = s.Close
			ls.Volume += s.Volume
		}
	}

	// Fill in the change and percent change fields.
	addChanges := func(ss []*ModelTradingSession) {
		for i := range ss {
			if i > 0 {
				ss[i].Change = ss[i].Close - ss[i-1].Close
				ss[i].PercentChange = ss[i].Change / ss[i-1].Close
			}
		}
	}
	addChanges(ds)
	addChanges(ws)

	// Fill in the stochastics.
	addStochastics := func(ss []*ModelTradingSession) {
		const (
			k = 10
			d = 3
		)

		// Calculate fast %K for stochastics.
		fastK := make([]float32, len(ss))
		for i := range ss {
			if i+1 < k {
				continue
			}

			highestHigh, lowestLow := ss[i].High, ss[i].Low
			for j := 0; j < k; j++ {
				if highestHigh < ss[i-j].High {
					highestHigh = ss[i-j].High
				}
				if lowestLow > ss[i-j].Low {
					lowestLow = ss[i-j].Low
				}
			}
			fastK[i] = (ss[i].Close - lowestLow) / (highestHigh - lowestLow)
		}

		// Calculate fast %D (slow %K) for stochastics.
		for i := range ss {
			if i+1 < k+d {
				continue
			}
			ss[i].K = (fastK[i] + fastK[i-1] + fastK[i-2]) / 3
		}

		// Calculate slow %D for stochastics.
		for i := range ss {
			if i+1 < k+d+d {
				continue
			}
			ss[i].D = (ss[i].K + ss[i-1].K + ss[i-2].K) / 3
		}
	}
	addStochastics(ds)
	addStochastics(ws)

	return ds, ws
}

func sortByModelTradingSessionDate(ss []*ModelTradingSession) {
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Date.Before(ss[j].Date)
	})
}
