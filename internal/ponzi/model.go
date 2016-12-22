package ponzi

// model is the state of the program separate from the view.
type model struct {
	dow    *quote
	sap    *quote
	nasdaq *quote
}

func (m *model) load() error {
	const (
		dowSymbol    = ".DJI"
		sapSymbol    = ".INX"
		nasdaqSymbol = ".IXIC"
	)

	indexSymbols := []string{
		dowSymbol,
		sapSymbol,
		nasdaqSymbol,
	}

	resp, err := listQuotes(&listQuotesRequest{indexSymbols})
	if err != nil {
		return err
	}

	m.dow = resp.quotes[dowSymbol]
	m.sap = resp.quotes[sapSymbol]
	m.nasdaq = resp.quotes[nasdaqSymbol]

	return nil
}
