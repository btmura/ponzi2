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
	sync.Mutex                     // Mutex guards the model.
	currentStock   *ModelStock     // currentStock is the stock currently being viewed.
	sideBarStocks  []*ModelStock   // sideBarStocks are the ordered stocks in the sidebar.
	sideBarSymbols map[string]bool // sideBarSymbols is a set of the sidebar's symbols.
}

type ModelQuote struct {
	symbol        string
	price         float32
	change        float32
	percentChange float32
}

type ModelStock struct {
	symbol         string
	quote          *ModelQuote
	dailySessions  []*ModelTradingSession
	weeklySessions []*ModelTradingSession
	lastUpdateTime time.Time
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

func NewModel(symbol string) *Model {
	return &Model{
		currentStock:   NewModelStock("SPY"),
		sideBarSymbols: map[string]bool{},
	}
}

func (m *Model) AddSideBarStock(symbol string) *ModelStock {
	if m.sideBarSymbols[symbol] {
		return nil
	}
	st := NewModelStock(symbol)
	m.sideBarStocks = append(m.sideBarStocks, st)
	m.sideBarSymbols[symbol] = true
	return st
}

func (m *Model) RemoveSideBarStock(symbol string) {
	if !m.sideBarSymbols[symbol] {
		return
	}
	var newStocks []*ModelStock
	for _, st := range m.sideBarStocks {
		if st.symbol == symbol {
			continue // Don't keep it.
		}
		newStocks = append(newStocks, st)
	}
	m.sideBarStocks = newStocks
	delete(m.sideBarSymbols, symbol)
}

func (m *Model) Refresh() error {
	if err := m.currentStock.Refresh(); err != nil {
		return err
	}
	for _, st := range m.sideBarStocks {
		if err := st.Refresh(); err != nil {
			return err
		}
	}
	return nil
}

func NewModelStock(symbol string) *ModelStock {
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
