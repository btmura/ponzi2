package ponzi

// ModelSidebar represents the sidebar of ordered stocks on the left of the UI.
type ModelSidebar struct {
	Stocks  []*ModelStock   // Stocks is the ordered stocks that can be iterated over.
	symbols map[string]bool // symbols is a set to track symbols.
}

// HasStock returns true if the sidebar has a stock with the given symbol.
func (m *ModelSidebar) HasStock(symbol string) bool {
	return m.symbols[symbol]
}

// AddStock adds a stock to the sidebar with the given symbol.
func (m *ModelSidebar) AddStock(symbol string) {
	if m.symbols[symbol] {
		return
	}
	if m.symbols == nil {
		m.symbols = map[string]bool{}
	}
	m.symbols[symbol] = true

	m.Stocks = append(m.Stocks, NewModelStock(symbol))
}

// RemoveStock removes a stock from the sidebar with the given symbol.
func (m *ModelSidebar) RemoveStock(symbol string) {
	if !m.symbols[symbol] {
		return
	}
	delete(m.symbols, symbol)

	var stocks []*ModelStock
	for _, st := range m.Stocks {
		if symbol == st.symbol {
			continue
		}
		stocks = append(stocks, st)
	}
	m.Stocks = stocks
}
