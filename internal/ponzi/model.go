package ponzi

import (
	"sort"
	"sync"
	"time"

	"github.com/btmura/ponzi2/internal/stock"
)

// model is the state of the program separate from the view.
type model struct {
	// Mutex guards the model.
	sync.Mutex

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

type modelQuote struct {
	symbol        string
	price         float32
	change        float32
	percentChange float32
}

type modelStock struct {
	symbol         string
	quote          *modelQuote
	dailySessions  []*modelTradingSession
	weeklySessions []*modelTradingSession
	lastUpdateTime time.Time
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

func newModel(symbol string) *model {
	return &model{
		dow:          &modelQuote{symbol: ".DJI"},
		sap:          &modelQuote{symbol: ".INX"},
		nasdaq:       &modelQuote{symbol: ".IXIC"},
		currentStock: newModelStock("SPY"),
	}
}

func newModelStock(symbol string) *modelStock {
	return &modelStock{
		symbol: symbol,
		quote:  &modelQuote{symbol: symbol},
	}
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
	m.currentStock = newModelStock(m.inputSymbol)
	m.inputSymbol = ""
	m.Unlock()
}

func (m *model) refresh() error {
	m.Lock()
	s := m.currentStock.symbol
	m.Unlock()

	symbols := []string{m.dow.symbol, m.sap.symbol, m.nasdaq.symbol}
	if s != "" {
		symbols = append(symbols, s)
	}

	resp, err := stock.ListQuotes(&stock.ListQuotesRequest{Symbols: symbols})
	if err != nil {
		return err
	}

	// Get the trading history for the current stock.
	var hist *stock.TradingHistory
	if s != "" {
		end := midnight(time.Now().In(newYorkLoc))
		start := end.Add(-6 * 30 * 24 * time.Hour)
		hist, err = stock.GetTradingHistory(&stock.GetTradingHistoryRequest{
			Symbol:    s,
			StartDate: start,
			EndDate:   end,
		})
		if err != nil {
			return err
		}
	}

	updateQuote := func(mq *modelQuote) {
		if q := resp.Quotes[mq.symbol]; q != nil {
			mq.price = q.Price
			mq.change = q.Change
			mq.percentChange = q.PercentChange
		}
	}

	m.Lock()
	updateQuote(m.dow)
	updateQuote(m.sap)
	updateQuote(m.nasdaq)
	if s != "" && s == m.currentStock.symbol {
		updateQuote(m.currentStock.quote)
		m.currentStock.dailySessions, m.currentStock.weeklySessions = convertSessions(hist.Sessions)
		m.currentStock.lastUpdateTime = time.Now()
	}
	m.Unlock()

	return nil
}

func convertSessions(sessions []*stock.TradingSession) (dailySessions, weeklySessions []*modelTradingSession) {
	// Convert the trading sessions into daily sessions.
	var ds []*modelTradingSession
	for _, s := range sessions {
		ds = append(ds, &modelTradingSession{
			date:   s.Date,
			open:   s.Open,
			high:   s.High,
			low:    s.Low,
			close:  s.Close,
			volume: s.Volume,
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
