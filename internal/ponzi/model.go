package ponzi

import (
	"sort"
	"sync"
	"time"

	"github.com/btmura/ponzi2/internal/stock"
	t2 "github.com/btmura/ponzi2/internal/time"
)

// Model is the state of the program separate from the view.
type Model struct {
	// Mutex guards the model.
	sync.Mutex

	// CurrentStock is the stock currently being viewed.
	CurrentStock *ModelStock

	// Sidebar models the sidebar with the user's stocks.
	Sidebar *ModelSidebar
}

func NewModel(currentSymbol string, sidebarSymbols []string) *Model {
	return &Model{
		CurrentStock: newModelStock(currentSymbol),
		Sidebar:      newModelSidebar(sidebarSymbols),
	}
}

func (m *Model) Refresh() error {
	if err := m.CurrentStock.Refresh(); err != nil {
		return err
	}
	for _, st := range m.Sidebar.Stocks {
		if err := st.Refresh(); err != nil {
			return err
		}
	}
	return nil
}

// ModelSidebar models the sidebar with the user's stocks.
type ModelSidebar struct {
	// Stocks are the ordered stocks from top to bottom.
	Stocks []*ModelStock
}

func newModelSidebar(symbols []string) *ModelSidebar {
	var sts []*ModelStock
	for _, s := range symbols {
		sts = append(sts, newModelStock(s))
	}
	return &ModelSidebar{
		Stocks: sts,
	}
}

// AddStock adds a stock to the sidebar.
func (m *ModelSidebar) AddStock(symbol string) bool {
	if m.Stock(symbol) != nil {
		return false // Already have it.
	}
	m.Stocks = append(m.Stocks, newModelStock(symbol))
	return true
}

// RemoveStock removes a stock from the sidebar.
func (m *ModelSidebar) RemoveStock(symbol string) bool {
	if m.Stock(symbol) == nil {
		return false // Don't have it.
	}

	var ss []*ModelStock
	for _, st := range m.Stocks {
		if st.symbol == symbol {
			continue
		}
		ss = append(ss, st)
	}
	m.Stocks = ss
	return true
}

// Stock returns the stock with the symbol or nil if the sidebar doesn't have it.
func (m *ModelSidebar) Stock(symbol string) *ModelStock {
	for _, st := range m.Stocks {
		if st.symbol == symbol {
			return st
		}
	}
	return nil
}

type ModelStock struct {
	symbol         string
	quote          *ModelQuote
	dailySessions  []*ModelTradingSession
	weeklySessions []*ModelTradingSession
	lastUpdateTime time.Time
}

type ModelQuote struct {
	symbol        string
	price         float32
	change        float32
	percentChange float32
}

type ModelTradingSession struct {
	date          time.Time
	open          float32
	high          float32
	low           float32
	close         float32
	volume        int
	change        float32
	percentChange float32
	k             float32
	d             float32
}

func newModelStock(symbol string) *ModelStock {
	return &ModelStock{
		symbol: symbol,
		quote:  &ModelQuote{symbol: symbol},
	}
}

func (m *ModelStock) Refresh() error {
	// Get the trading history for the current stock.
	end := t2.Midnight(time.Now().In(t2.NewYorkLoc))
	start := end.Add(-6 * 30 * 24 * time.Hour)
	hist, err := stock.GetTradingHistory(&stock.GetTradingHistoryRequest{
		Symbol:    m.symbol,
		StartDate: start,
		EndDate:   end,
	})
	if err != nil {
		return err
	}

	m.dailySessions, m.weeklySessions = convertSessions(hist.Sessions)
	if len(m.dailySessions) > 0 {
		last := m.dailySessions[len(m.dailySessions)-1]
		m.quote.price = last.close
		m.quote.change = last.change
		m.quote.percentChange = last.percentChange
	}
	m.lastUpdateTime = time.Now()

	return nil
}

func convertSessions(sessions []*stock.TradingSession) (dailySessions, weeklySessions []*ModelTradingSession) {
	// Convert the trading sessions into daily sessions.
	var ds []*ModelTradingSession
	for _, s := range sessions {
		ds = append(ds, &ModelTradingSession{
			date:   s.Date,
			open:   s.Open,
			high:   s.High,
			low:    s.Low,
			close:  s.Close,
			volume: s.Volume,
		})
	}
	sortByModelTradingSessionDate(ds)

	// Convert the daily sessions into weekly sessions.
	var ws []*ModelTradingSession
	for _, s := range ds {
		diffWeek := ws == nil
		if !diffWeek {
			_, week := s.date.ISOWeek()
			_, prevWeek := ws[len(ws)-1].date.ISOWeek()
			diffWeek = week != prevWeek
		}

		if diffWeek {
			sc := *s
			ws = append(ws, &sc)
		} else {
			ls := ws[len(ws)-1]
			if ls.high < s.high {
				ls.high = s.high
			}
			if ls.low > s.low {
				ls.low = s.low
			}
			ls.close = s.close
			ls.volume += s.volume
		}
	}

	// Fill in the change and percent change fields.
	addChanges := func(ss []*ModelTradingSession) {
		for i := range ss {
			if i > 0 {
				ss[i].change = ss[i].close - ss[i-1].close
				ss[i].percentChange = ss[i].change / ss[i-1].close
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

			highestHigh, lowestLow := ss[i].high, ss[i].low
			for j := 0; j < k; j++ {
				if highestHigh < ss[i-j].high {
					highestHigh = ss[i-j].high
				}
				if lowestLow > ss[i-j].low {
					lowestLow = ss[i-j].low
				}
			}
			fastK[i] = (ss[i].close - lowestLow) / (highestHigh - lowestLow)
		}

		// Calculate fast %D (slow %K) for stochastics.
		for i := range ss {
			if i+1 < k+d {
				continue
			}
			ss[i].k = (fastK[i] + fastK[i-1] + fastK[i-2]) / 3
		}

		// Calculate slow %D for stochastics.
		for i := range ss {
			if i+1 < k+d+d {
				continue
			}
			ss[i].d = (ss[i].k + ss[i-1].k + ss[i-2].k) / 3
		}
	}
	addStochastics(ds)
	addStochastics(ws)

	return ds, ws
}

func sortByModelTradingSessionDate(ss []*ModelTradingSession) {
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].date.Before(ss[j].date)
	})
}
