package stock

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"sort"
	"time"
)

// GetMovingAverageRequest is a request for a stock's moving average.
type GetMovingAverageRequest struct {
	// Symbol is the stock's symbol like "SPY".
	Symbol string
}

// MovingAverage is a time series of moving average values.
type MovingAverage struct {
	// Values are the moving average values with earlier values in front.
	Values []*MovingAverageValue
}

// MovingAverageValue is a moving average data value for some date.
type MovingAverageValue struct {
	// Date is the start date of the time span covered by this value.
	Date time.Time

	// Average is the average value.
	Average float32
}

// GetMovingAverage returns MovingAverage data or an error.
func (a *AlphaVantage) GetMovingAverage(req *GetMovingAverageRequest) (*MovingAverage, error) {
	v := url.Values{}
	v.Set("function", "SMA")
	v.Set("symbol", req.Symbol)
	v.Set("interval", "daily")
	v.Set("time_period", "50")
	v.Set("series_type", "close")
	v.Set("apikey", a.apiKey)

	u, err := url.Parse("https://www.alphavantage.co/query")
	if err != nil {
		logger.Fatalf("can't parse url")
	}
	u.RawQuery = v.Encode()

	resp, err := a.httpGet(u.String())
	if err != nil {
		return nil, fmt.Errorf("stock: http get for movavg failed: %v", err)
	}
	defer resp.Body.Close()

	return decodeMovingAverageResponse(resp.Body)
}

func decodeMovingAverageResponse(r io.Reader) (*MovingAverage, error) {
	type DataPoint struct {
		SMA string
	}

	type Data struct {
		TechnicalAnalysis map[string]DataPoint `json:"Technical Analysis: SMA"`
	}

	var data Data
	dec := json.NewDecoder(r)
	if err := dec.Decode(&data); err != nil {
		return nil, fmt.Errorf("stock: decoding moving avg json failed: %v", err)
	}

	var vs []*MovingAverageValue
	for dstr, pt := range data.TechnicalAnalysis {
		date, err := parseDate(dstr)
		if err != nil {
			return nil, fmt.Errorf("stock: parsing moving avg time (%v) failed: %v", dstr, err)
		}

		avg, err := parseFloat(pt.SMA)
		if err != nil {
			return nil, fmt.Errorf("stock: parsing moving avg value (%s) failed: %v", pt.SMA, err)
		}

		vs = append(vs, &MovingAverageValue{
			Date:    date,
			Average: avg,
		})
	}

	sort.Slice(vs, func(i, j int) bool {
		return vs[i].Date.Before(vs[j].Date)
	})

	return &MovingAverage{Values: vs}, nil
}
