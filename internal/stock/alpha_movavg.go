package stock

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"sort"
	"strconv"
)

// GetMovingAverage returns MovingAverage data or an error.
func (a *AlphaVantage) GetMovingAverage(ctx context.Context, req *GetMovingAverageRequest) (*MovingAverage, error) {
	if req.Symbol == "" {
		return nil, fmt.Errorf("stock: movavg request missing symbol: %v", req)
	}

	if req.TimePeriod == 0 {
		return nil, fmt.Errorf("stock: movavg request missing time period: %v", req)
	}

	v := url.Values{}
	v.Set("function", "SMA")
	v.Set("symbol", req.Symbol)
	v.Set("interval", "daily")
	v.Set("time_period", strconv.Itoa(req.TimePeriod))
	v.Set("series_type", "close")
	v.Set("apikey", a.apiKey)

	u, err := url.Parse("https://www.alphavantage.co/query")
	if err != nil {
		logger.Fatalf("can't parse url")
	}
	u.RawQuery = v.Encode()

	resp, err := a.httpGet(ctx, u.String())
	if err != nil {
		return nil, fmt.Errorf("stock: http get for movavg failed: %v", err)
	}
	defer resp.Body.Close()

	r := resp.Body
	if a.dumpAPIResponses {
		rr, err := dumpResponse(fmt.Sprintf("debug-movavg-%s-%d.txt", req.Symbol, req.TimePeriod), r)
		if err != nil {
			return nil, fmt.Errorf("stock: dumping movavg resp failed: %v", err)
		}
		r = rr
	}
	return decodeMovingAverageResponse(r)
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
