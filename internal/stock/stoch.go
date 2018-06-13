package stock

import "time"

// StochasticRequest is a request for a stock's stochastics.
type StochasticRequest struct {
	// Symbol is the stock's symbol like "SPY".
	Symbol string
}

// Stochastics is a time series of stochastic values.
type Stochastics struct {
	// Values are the stochastic values with earlier values in front.
	Values []*StochasticValue
}

// StochasticValue are the stochastic values for some date.
type StochasticValue struct {
	// Date is the start date for a daily or weekly time span.
	Date time.Time

	// K tries to measure the momentum.
	K float32

	// D is some moving average of K.
	D float32
}

// GetStochastics returns Stochastics or an error.
func (a *AlphaVantage) GetStochastics(req *StochasticRequest) (*Stochastics, error) {
	return &Stochastics{}, nil
}
