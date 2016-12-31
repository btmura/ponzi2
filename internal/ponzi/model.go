package ponzi

import (
	"sort"
	"sync"
	"time"
)

// model is the state of the program separate from the view.
type model struct {
	// mutex guards the entire model. Only use read lock outside of this file.
	sync.RWMutex

	// dow is a quote for the Dow Jones index.
	dow *modelQuote

	// sap is a quote for the S&P 500 index.
	sap *modelQuote

	// nasdaq is a quote for the NASDAQ index.
	nasdaq *modelQuote

	// inputSymbol is the symbol being entered by the user.
	inputSymbol string

	// currentStock is the stock data currently being viewed.
	currentStock *modelStock
}

type modelStock struct {
	symbol         string
	quote          *modelQuote
	dailySessions  []*modelTradingSession
	weeklySessions []*modelTradingSession
}

type modelQuote struct {
	price         float32
	change        float32
	percentChange float32
}

type modelTradingSession struct {
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

func (m *model) currentSymbol() string {
	// TODO(btmura): fix locking issues
	if m.currentStock == nil {
		return ""
	}
	return m.currentStock.symbol
}

func (m *model) pushSymbolChar(ch rune) {
	m.Lock()
	m.inputSymbol += string(ch)
	m.Unlock()
}

func (m *model) popSymbolChar() {
	m.Lock()
	if l := len(m.inputSymbol); l > 0 {
		m.inputSymbol = m.inputSymbol[:l-1]
	}
	m.Unlock()
}

func (m *model) submitSymbol() {
	m.Lock()
	m.currentStock = &modelStock{
		symbol: m.inputSymbol,
	}
	m.inputSymbol = ""
	m.startRefresh()
	m.Unlock()
}

func (m *model) startRefresh() error {
	go func() {
		m.refresh()
	}()
	return nil
}

func (m *model) refresh() error {
	// Get the current symbol being viewed.
	m.RLock()
	s := m.currentSymbol()
	m.RUnlock()

	// Get live quotes for the major indices and the current symbol.
	const (
		dowSymbol    = ".DJI"
		sapSymbol    = ".INX"
		nasdaqSymbol = ".IXIC"
	)

	symbols := []string{dowSymbol, sapSymbol, nasdaqSymbol}
	if s != "" {
		symbols = append(symbols, s)
	}

	resp, err := listQuotes(&listQuotesRequest{symbols})
	if err != nil {
		return err
	}

	getQuote := func(symbol string) *modelQuote {
		if q := resp.quotes[symbol]; q != nil {
			return &modelQuote{
				price:         q.price,
				change:        q.change,
				percentChange: q.percentChange,
			}
		}
		return nil
	}

	// Get the trading history for the current stock.

	var hist *tradingHistory
	if s != "" {
		end := midnight(time.Now().In(newYorkLoc))
		start := end.Add(-6 * 30 * 24 * time.Hour)
		hist, err = getTradingHistory(&getTradingHistoryRequest{
			symbol:    s,
			startDate: start,
			endDate:   end,
		})
		if err != nil {
			return err
		}
	}

	// Acquire lock and update the model once.

	m.Lock()
	m.dow = getQuote(dowSymbol)
	m.sap = getQuote(sapSymbol)
	m.nasdaq = getQuote(nasdaqSymbol)
	if s != "" && s == m.currentSymbol() {
		m.currentStock = createModelStock(s, getQuote(s), hist.sessions)
	} else {
		m.currentStock = nil
	}
	m.Unlock()

	return nil
}

func createModelStock(symbol string, quote *modelQuote, sessions []*tradingSession) *modelStock {
	// Convert the trading sessions into daily sessions.
	var ds []*modelTradingSession
	for _, s := range sessions {
		ds = append(ds, &modelTradingSession{
			date:   s.date,
			open:   s.open,
			high:   s.high,
			low:    s.low,
			close:  s.close,
			volume: s.volume,
		})
	}
	sort.Sort(byModelTradingSessionDate(ds))

	// Convert the daily sessions into weekly sessions.
	var ws []*modelTradingSession
	for _, s := range ds {
		diffWeek := ws == nil
		if !diffWeek {
			_, week := s.date.ISOWeek()
			_, prevWeek := ws[len(ws)-1].date.ISOWeek()
			diffWeek = week != prevWeek
		}

		if diffWeek {
			ws = append(ws, &modelTradingSession{
				date:   s.date,
				open:   s.open,
				high:   s.high,
				low:    s.low,
				close:  s.close,
				volume: s.volume,
			})
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
	addChanges := func(ss []*modelTradingSession) {
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
	addStochastics := func(ss []*modelTradingSession) {
		const (
			kDays = 10
			dDays = 3
		)

		// Calculate fast %K for stochastics.
		fastK := make([]float32, len(ss))
		for i := range ss {
			if i+1 < kDays {
				continue
			}

			highestHigh, lowestLow := ss[i].high, ss[i].low
			for j := 0; j < kDays; j++ {
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
			if i+1 < kDays+dDays {
				continue
			}
			ss[i].k = (fastK[i] + fastK[i-1] + fastK[i-2]) / 3
		}

		// Calculate slow %D for stochastics.
		for i := range ss {
			if i+1 < kDays+dDays+dDays {
				continue
			}
			ss[i].d = (ss[i].k + ss[i-1].k + ss[i-2].k) / 3
		}
	}
	addStochastics(ds)
	addStochastics(ws)

	return &modelStock{
		symbol:         symbol,
		quote:          quote,
		dailySessions:  ds,
		weeklySessions: ws,
	}
}

// byModelTradingSessionDate is a sortable modelTradingSession slice.
type byModelTradingSessionDate []*modelTradingSession

// Len implements sort.Interface.
func (ss byModelTradingSessionDate) Len() int {
	return len(ss)
}

// Less implements sort.Interface.
func (ss byModelTradingSessionDate) Less(i, j int) bool {
	return ss[i].date.Before(ss[j].date)
}

// Swap implements sort.Interface.
func (ss byModelTradingSessionDate) Swap(i, j int) {
	ss[i], ss[j] = ss[j], ss[i]
}
