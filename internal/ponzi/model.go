package ponzi

import "sync"

// model is the state of the program separate from the view.
type model struct {
	// mutex guards the entire model. Only use read lock outside of this file.
	sync.RWMutex

	dow    *quote
	sap    *quote
	nasdaq *quote

	// inputSymbol is the symbol being entered by the user.
	inputSymbol string

	currentSymbol string
	currentQuote  *quote
}

func (m *model) pushSymbolLetter(s string) {
	m.Lock()
	m.inputSymbol += s
	m.Unlock()
}

func (m *model) popSymbolLetter() {
	m.Lock()
	if l := len(m.inputSymbol); l > 0 {
		m.inputSymbol = m.inputSymbol[:l-1]
	}
	m.Unlock()
}

func (m *model) submitSymbol() {
	m.Lock()
	m.currentSymbol = m.inputSymbol
	m.currentQuote = nil
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
	s := m.currentSymbol
	m.RUnlock()

	if s != "" {
		symbols = append(symbols, s)
	}

	resp, err := listQuotes(&listQuotesRequest{symbols})
	if err != nil {
		return err
	}

	m.Lock()
	m.dow = resp.quotes[dowSymbol]
	m.sap = resp.quotes[sapSymbol]
	m.nasdaq = resp.quotes[nasdaqSymbol]
	if s != "" && s == m.currentSymbol {
		m.currentQuote = resp.quotes[m.currentSymbol]
	} else {
		m.currentQuote = nil
	}
	m.Unlock()

	return nil
}
