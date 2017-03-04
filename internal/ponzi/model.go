package ponzi

import (
	"sort"
	"time"
)

// model is the state of the program separate from the view.
type model struct {
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

func newModel(symbol string) *model {
	return &model{
		dow:          newModelQuote(".DJI"),
		sap:          newModelQuote(".INX"),
		nasdaq:       newModelQuote(".IXIC"),
		currentStock: newModelStock("SPY"),
	}
}

func (m *model) currentSymbol() string {
	if m.currentStock == nil {
		return ""
	}
	return m.currentStock.symbol
}

func (m *model) pushSymbolChar(ch rune) {
	m.inputSymbol += string(ch)
}

func (m *model) popSymbolChar() {
	if l := len(m.inputSymbol); l > 0 {
		m.inputSymbol = m.inputSymbol[:l-1]
	}
}

func (m *model) submitSymbol() {
	m.currentStock = newModelStock(m.inputSymbol)
	m.inputSymbol = ""
}

func (m *model) refresh() error {
	if err := m.dow.refresh(); err != nil {
		return err
	}
	if err := m.sap.refresh(); err != nil {
		return err
	}
	if err := m.nasdaq.refresh(); err != nil {
		return err
	}
	if err := m.currentStock.refresh(); err != nil {
		return err
	}
	return nil
}

type modelQuote struct {
	symbol        string
	price         float32
	change        float32
	percentChange float32
}

func newModelQuote(symbol string) *modelQuote {
	return &modelQuote{symbol: symbol}
}

func (m *modelQuote) refresh() error {
	if m == nil {
		return nil
	}

	resp, err := listQuotes(&listQuotesRequest{[]string{m.symbol}})
	if err != nil {
		return err
	}

	if q := resp.quotes[m.symbol]; q != nil {
		*m = modelQuote{
			symbol:        q.symbol,
			price:         q.price,
			change:        q.change,
			percentChange: q.percentChange,
		}
	}
	return nil
}

type modelStock struct {
	symbol         string
	quote          *modelQuote
	dailySessions  []*modelTradingSession
	weeklySessions []*modelTradingSession
}

func newModelStock(symbol string) *modelStock {
	return &modelStock{
		symbol: symbol,
		quote:  &modelQuote{symbol: symbol},
	}
}

func (m *modelStock) refresh() error {
	if m == nil {
		return nil
	}

	if err := m.quote.refresh(); err != nil {
		return err
	}

	// Get the trading history for the current stock.
	end := midnight(time.Now().In(newYorkLoc))
	start := end.Add(-6 * 30 * 24 * time.Hour)
	hist, err := getTradingHistory(&getTradingHistoryRequest{
		symbol:    m.quote.symbol,
		startDate: start,
		endDate:   end,
	})
	if err != nil {
		return err
	}

	m.dailySessions, m.weeklySessions = convertSessions(hist.sessions)

	return nil
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

func convertSessions(sessions []*tradingSession) (dailySessions, weeklySessions []*modelTradingSession) {
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
