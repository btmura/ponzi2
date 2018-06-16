package stock

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/golang/glog"
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
	// Date is the start date for a daily or weekly time span.
	Date time.Time

	// K tries to measure the momentum.
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
		log.Fatalf("can't parse url")
	}
	u.RawQuery = v.Encode()

	glog.Info(u)
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("stock: http get for stoch failed: %v", err)
	}
	defer resp.Body.Close()

	// https://www.alphavantage.co/query?function=STOCH&symbol=MSFT&interval=daily&apikey=demo
	//
	// {
	//     "Meta Data": {
	//         "1: Symbol": "MSFT",
	//         "2: Indicator": "Stochastic (STOCH)",
	//         "3: Last Refreshed": "2018-06-13",
	//         "4: Interval": "daily",
	//         "5.1: FastK Period": 5,
	//         "5.2: SlowK Period": 3,
	//         "5.3: SlowK MA Type": 0,
	//         "5.4: SlowD Period": 3,
	//         "5.5: SlowD MA Type": 0,
	//         "6: Time Zone": "US/Eastern Time"
	//     },
	//     "Technical Analysis: STOCH": {
	//         "2018-06-13": {
	//             "SlowK": "29.8701",
	//             "SlowD": "38.2982"
	//         },
	//         "2018-06-12": {
	//             "SlowK": "41.1255",
	//             "SlowD": "50.5565"
	//         },
	//         "2018-06-11": {
	//             "SlowK": "43.8988",
	//             "SlowD": "63.8097"
	//         }
	//     }
	// }

	type DataPoint struct {
		SlowK string
		SlowD string
	}

	type Data struct {
		TechnicalAnalysis map[string]DataPoint `json:"Technical Analysis: STOCH"`
	}

	var data Data
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&data); err != nil {
		return nil, fmt.Errorf("stock: decoding stoch json failed: %v", err)
	}

	var vs []*StochasticValue
	for dt, pt := range data.TechnicalAnalysis {
		date, err := time.Parse("2006-01-02", dt)
		if err != nil {
			return nil, fmt.Errorf("stock: parsing stoch time (%v) failed: %v", dt, err)
		}

		k, err := parseFloat(pt.SlowK)
		if err != nil {
			return nil, fmt.Errorf("stock: parsing stoch k (%f) failed: %v", k, err)
		}

		d, err := parseFloat(pt.SlowD)
		if err != nil {
			return nil, fmt.Errorf("stock: parsing stoch d (%f) failed: %v", d, err)
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

	// TODO(btmura): add test

	return &Stochastics{Values: vs}, nil
}
