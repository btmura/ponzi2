package ponzi

// ModelSidebar models the left sidebar with the user's saved stocks.
type ModelSidebar struct {
	// Stocks are the ordered stocks from top to bottom.
	Stocks []*ModelStock
}

// HasStock returns true if the sidebar has the stock.
func (m *ModelSidebar) HasStock(symbol string) bool {
	for _, st := range m.Stocks {
		if st.symbol == symbol {
			return true
		}
	}
	return false
}

// AddStock adds a stock to the sidebar.
func (m *ModelSidebar) AddStock(symbol string) {
	if m.HasStock(symbol) {
		return
	}
	m.Stocks = append(m.Stocks, NewModelStock(symbol))
}

// RemoveStock removes a stock from the sidebar.
func (m *ModelSidebar) RemoveStock(symbol string) {
	if !m.HasStock(symbol) {
		return
	}

	var ss []*ModelStock
	for _, st := range m.Stocks {
		if st.symbol == symbol {
			continue
		}
		ss = append(ss, st)
	}
	m.Stocks = ss
}
