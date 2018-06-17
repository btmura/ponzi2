package stock

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"sort"
	"time"
)

// GetStochasticsRequest is a request for a stock's stochastics.
type GetStochasticsRequest struct {
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
	// Date is the start date of the time span covered by this value.
	Date time.Time

	// K measures the stock's momentum.
	K float32

	// D is some moving average of K.
	D float32
}

// GetStochastics returns Stochastics or an error.
func (a *AlphaVantage) GetStochastics(req *GetStochasticsRequest) (*Stochastics, error) {
	v := url.Values{}
	v.Set("function", "STOCH")
	v.Set("symbol", req.Symbol)
	v.Set("interval", "daily")
	v.Set("apikey", a.apiKey)

	u, err := url.Parse("https://www.alphavantage.co/query")
	if err != nil {
		logger.Fatalf("can't parse url")
	}
	u.RawQuery = v.Encode()

	resp, err := a.httpGet(u.String())
	if err != nil {
		return nil, fmt.Errorf("stock: http get for stoch failed: %v", err)
	}
	defer resp.Body.Close()

	return decodeStochasticsResponse(resp.Body)
}

func decodeStochasticsResponse(r io.Reader) (*Stochastics, error) {
	type DataPoint struct {
		SlowK string
		SlowD string
	}

	type Data struct {
		TechnicalAnalysis map[string]DataPoint `json:"Technical Analysis: STOCH"`
	}

	var data Data
	dec := json.NewDecoder(r)
	if err := dec.Decode(&data); err != nil {
		return nil, fmt.Errorf("stock: decoding stoch json failed: %v", err)
	}

	var vs []*StochasticValue
	for dstr, pt := range data.TechnicalAnalysis {
		date, err := parseDate(dstr)
		if err != nil {
			return nil, fmt.Errorf("stock: parsing stoch time (%v) failed: %v", dstr, err)
		}

		k, err := parseFloat(pt.SlowK)
		if err != nil {
			return nil, fmt.Errorf("stock: parsing stoch k (%s) failed: %v", pt.SlowK, err)
		}

		d, err := parseFloat(pt.SlowD)
		if err != nil {
			return nil, fmt.Errorf("stock: parsing stoch d (%s) failed: %v", pt.SlowD, err)
		}

		vs = append(vs, &StochasticValue{
			Date: date,
			K:    k,
			D:    d,
		})
	}

	sort.Slice(vs, func(i, j int) bool {
		return vs[i].Date.Before(vs[j].Date)
	})

	return &Stochastics{Values: vs}, nil
}
