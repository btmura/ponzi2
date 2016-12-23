package ponzi

import "sync"

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
	symbol string
	quote  *modelQuote
}

type modelQuote struct {
	price         float32
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
	const (
		dowSymbol    = ".DJI"
		sapSymbol    = ".INX"
		nasdaqSymbol = ".IXIC"
	)

	symbols := []string{
		dowSymbol,
		sapSymbol,
		nasdaqSymbol,
	}

	m.RLock()
	s := m.currentStock.symbol
	m.RUnlock()

	if s != "" {
		symbols = append(symbols, s)
	}

	resp, err := listQuotes(&listQuotesRequest{symbols})
	if err != nil {
		return err
	}

	getQuote := func(resp *listQuotesResponse, symbol string) *modelQuote {
		if q := resp.quotes[symbol]; q != nil {
			return &modelQuote{
				price:         q.price,
				change:        q.change,
				percentChange: q.percentChange,
			}
		}
		return nil
	}

	m.Lock()
	m.dow = getQuote(resp, dowSymbol)
	m.sap = getQuote(resp, sapSymbol)
	m.nasdaq = getQuote(resp, nasdaqSymbol)
	if s != "" && s == m.currentStock.symbol {
		m.currentStock.quote = getQuote(resp, s)
	}
	m.Unlock()

	return nil
}
