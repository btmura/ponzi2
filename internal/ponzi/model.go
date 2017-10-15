package ponzi

import (
	"sort"
	"time"

	"github.com/golang/glog"

	"github.com/btmura/ponzi2/internal/stock"
)

// Model keeps track of the user's stocks.
type Model struct {
	// CurrentStock is the stock currently being viewed.
	CurrentStock *ModelStock

	// SavedStocks is the user's ordered set of saved stocks.
	SavedStocks []*ModelStock
}

// NewModel creates a new Model.
func NewModel() *Model {
	return &Model{}
}

// SetCurrentStock sets the current stock by symbol. It returns the
// corresponding ModelStock and true if the current stock changed.
func (m *Model) SetCurrentStock(symbol string) (st *ModelStock, changed bool) {
	if symbol == "" {
		glog.Fatal("SetCurrentStock: cannot set current stock to empty symbol")
	}

	if m.CurrentStock != nil && m.CurrentStock.Symbol == symbol {
		return m.CurrentStock, false
	}

	if m.CurrentStock = m.stock(symbol); m.CurrentStock == nil {
		m.CurrentStock = &ModelStock{Symbol: symbol}
	}
	return m.CurrentStock, true
}

// AddSavedStock adds the stock by symbol. It returns the corresponding
// ModelStock and true if the stock was newly added.
func (m *Model) AddSavedStock(symbol string) (st *ModelStock, added bool) {
	if symbol == "" {
		glog.Fatal("AddSavedStock: cannot add empty symbol")
	}

	for _, st := range m.SavedStocks {
		if st.Symbol == symbol {
			return st, false
		}
	}

	if st = m.stock(symbol); st == nil {
		st = &ModelStock{Symbol: symbol}
	}
	m.SavedStocks = append(m.SavedStocks, st)
	return st, true
}

// RemoveSavedStock removes the stock by symbol and returns true if removed.
func (m *Model) RemoveSavedStock(symbol string) (removed bool) {
	if symbol == "" {
		glog.Fatal("RemovedSavedStock: cannot remove empty symbol")
	}

	for i, st := range m.SavedStocks {
		if st.Symbol == symbol {
			m.SavedStocks = append(m.SavedStocks[:i], m.SavedStocks[i+1:]...)
			return true
		}
	}
	return false
}

// UpdateStock updates the stock with the TradingHistory if it is in the model.
func (m *Model) UpdateStock(symbol string, hist *stock.TradingHistory) (st *ModelStock, updated bool) {
	if st = m.stock(symbol); st == nil {
		return nil, false
	}
	st.DailySessions, st.WeeklySessions = convertSessions(hist.Sessions)
	st.LastUpdateTime = time.Now()
	return st, true
}

func (m *Model) stock(symbol string) *ModelStock {
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

// ModelStock models a single stock.
type ModelStock struct {
	Symbol         string
	DailySessions  []*ModelTradingSession
	WeeklySessions []*ModelTradingSession
	LastUpdateTime time.Time
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

// Date returns the most recent date or zero time if no data.
func (m *ModelStock) Date() time.Time {
	if len(m.DailySessions) == 0 {
		return time.Time{}
	}
	return m.DailySessions[len(m.DailySessions)-1].Date
}

func convertSessions(ts []*stock.TradingSession) (ds, ws []*ModelTradingSession) {
	ds = dailySessions(ts)
	ws = weeklySessions(ds)

	fillChangeValues(ds)
	fillChangeValues(ws)

	fillStochastics(ds)
	fillStochastics(ws)

	return ds, ws
}

func dailySessions(ts []*stock.TradingSession) (ds []*ModelTradingSession) {
	for _, s := range ts {
		ds = append(ds, &ModelTradingSession{
			Date:   s.Date,
			Open:   s.Open,
			High:   s.High,
			Low:    s.Low,
			Close:  s.Close,
			Volume: s.Volume,
		})
	}
	sort.Slice(ds, func(i, j int) bool {
		return ds[i].Date.Before(ds[j].Date)
	})
	return ds
}

func weeklySessions(ds []*ModelTradingSession) (ws []*ModelTradingSession) {
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
	return ws
}

func fillChangeValues(ss []*ModelTradingSession) {
	for i := range ss {
		if i > 0 {
			ss[i].Change = ss[i].Close - ss[i-1].Close
			ss[i].PercentChange = ss[i].Change / ss[i-1].Close
		}
	}
}

func fillStochastics(ss []*ModelTradingSession) {
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
