package ponzi

import (
	"sync"
	"time"
)

// model is the state of the program separate from the view.
type model struct {
	// mutex guards the entire model. Only use read lock outside of this file.
	sync.RWMutex

	dow    *modelQuote
	sap    *modelQuote
	nasdaq *modelQuote

	// inputSymbol is the symbol being entered by the user.
	inputSymbol string

	currentStock modelStock
}

type modelStock struct {
	symbol   string
	quote    *modelQuote
	sessions []*modelTradingSession
}

type modelQuote struct {
	price         float32
	change        float32
	percentChange float32
}

type modelTradingSession struct {
	date          time.Time
	open          float32
	close         float32
	change        float32
	percentChange float32
	volume        int
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
	m.currentStock = modelStock{
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

	// Get live quotes for the major indices.

	const (
		dowSymbol    = ".DJI"
		sapSymbol    = ".INX"
		nasdaqSymbol = ".IXIC"
	)

	resp, err := listQuotes(&listQuotesRequest{[]string{
		dowSymbol,
		sapSymbol,
		nasdaqSymbol,
	}})
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

	m.RLock()
	s := m.currentStock.symbol
	m.RUnlock()

	var hist *tradingHistory
	if s != "" {
		end := midnight(time.Now().In(newYorkLoc))
		start := end.Add(-30 * 24 * time.Hour)
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
	if s != "" && s == m.currentStock.symbol {
		m.currentStock.sessions = convertTradingSessions(hist.sessions)
		if len(m.currentStock.sessions) > 0 {
			m.currentStock.quote = &modelQuote{
				price:         m.currentStock.sessions[0].close,
				change:        m.currentStock.sessions[0].change,
				percentChange: m.currentStock.sessions[0].percentChange,
			}
		}
	}
	m.Unlock()

	return nil
}

func convertTradingSessions(sessions []*tradingSession) []*modelTradingSession {
	// Copy the trading sessions into a slice of structs.
	var ms []*modelTradingSession
	for _, s := range sessions {
		ms = append(ms, &modelTradingSession{
			date:   s.date,
			open:   s.open,
			close:  s.close,
			volume: s.volume,
		})
	}

	// Calculate the price change which is today's close minus yesterday's close.
	for i := range ms {
		if i+1 < len(ms) {
			ms[i].change = ms[i].close - ms[i+1].close
			ms[i].percentChange = ms[i].change / ms[i+1].close
		}
	}

	return ms
}
