package ponzi

// ModelSidebar models the left sidebar with the user's saved stocks.
type ModelSidebar struct {
	// Stocks are the ordered stocks from top to bottom.
	Stocks []*ModelStock
}

// AddStock adds a stock to the sidebar.
func (m *ModelSidebar) AddStock(symbol string) bool {
	if m.Stock(symbol) != nil {
		return false // Already have it.
	}
	m.Stocks = append(m.Stocks, NewModelStock(symbol))
	return true
}

// RemoveStock removes a stock from the sidebar.
func (m *ModelSidebar) RemoveStock(symbol string) bool {
	if m.Stock(symbol) == nil {
		return false // Don't have it.
	}

	var ss []*ModelStock
	for _, st := range m.Stocks {
		if st.symbol == symbol {
			continue
		}
		ss = append(ss, st)
	}
	m.Stocks = ss
	return true
}

// Stock returns the stock with the symbol or nil if the sidebar doesn't have it.
func (m *ModelSidebar) Stock(symbol string) *ModelStock {
	for _, st := range m.Stocks {
		if st.symbol == symbol {
			return st
		}
	}
	return nil
}
