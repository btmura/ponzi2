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

	dow    *modelQuote
	sap    *modelQuote
	nasdaq *modelQuote

	// inputSymbol is the symbol being entered by the user.
	inputSymbol string

	currentSymbol          string
	currentQuote           *modelQuote
	currentTradingSessions []*modelTradingSession
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
	m.currentSymbol, m.inputSymbol = m.inputSymbol, ""
	m.currentQuote = nil
	m.currentTradingSessions = nil
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
	s := m.currentSymbol
	m.RUnlock()

	var hist *tradingHistory
	if s != "" {
		end := midnight(time.Now().In(newYorkLoc))
		start := end.Add(-30 * 24 * time.Hour * 3)
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
	if s != "" && s == m.currentSymbol {
		m.currentQuote, m.currentTradingSessions = convertTradingSessions(hist.sessions)
	} else {
		m.currentQuote = nil
		m.currentTradingSessions = nil
	}
	m.Unlock()

	return nil
}

func convertTradingSessions(sessions []*tradingSession) (*modelQuote, []*modelTradingSession) {
	// Copy the trading sessions into a slice of structs.
	var ms []*modelTradingSession
	for _, s := range sessions {
		ms = append(ms, &modelTradingSession{
			date:   s.date,
			open:   s.open,
			high:   s.high,
			low:    s.low,
			close:  s.close,
			volume: s.volume,
		})
	}

	// Most recent trading sessions at the back.
	sort.Sort(byModelTradingSessionDate(ms))

	// Calculate the price change which is today's close minus yesterday's close.
	for i := range ms {
		if i > 0 {
			ms[i].change = ms[i].close - ms[i-1].close
			ms[i].percentChange = ms[i].change / ms[i-1].close
		}
	}

	// Use the trading history to create the current quote.
	var quote *modelQuote
	if len(ms) != 0 {
		quote = &modelQuote{
			price:         ms[len(ms)-1].close,
			change:        ms[len(ms)-1].change,
			percentChange: ms[len(ms)-1].percentChange,
		}
	}

	return quote, ms
}

// byModelTradingSessionDate is a sortable modelTradingSession slice.
type byModelTradingSessionDate []*modelTradingSession

// Len implements sort.Interface.
func (sessions byModelTradingSessionDate) Len() int {
	return len(sessions)
}

// Less implements sort.Interface.
func (sessions byModelTradingSessionDate) Less(i, j int) bool {
	return sessions[i].date.Before(sessions[j].date)
}

// Swap implements sort.Interface.
func (sessions byModelTradingSessionDate) Swap(i, j int) {
	sessions[i], sessions[j] = sessions[j], sessions[i]
}
