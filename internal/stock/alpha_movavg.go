package stock

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"sort"
	"strconv"
	"strings"
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

	debugID := fmt.Sprintf("debug-movavg-%s-%d", req.Symbol, req.TimePeriod)

	var ma *MovingAverage
	if err := a.query(ctx, v, debugID, func(r io.Reader) error {
		m, err := decodeMovingAverageResponse(r)
		if err != nil {
			return err
		}
		ma = m
		return nil
	}); err != nil {
		return nil, fmt.Errorf("stock: querying movavg failed: %v", err)
	}

	return ma, nil
}

func decodeMovingAverageResponse(r io.Reader) (*MovingAverage, error) {
	type DataPoint struct {
		SMA string
	}

	type Data struct {
		Information       string               `json:"Information"`
		TechnicalAnalysis map[string]DataPoint `json:"Technical Analysis: SMA"`
	}

	var data Data
	dec := json.NewDecoder(r)
	if err := dec.Decode(&data); err != nil {
		return nil, fmt.Errorf("stock: decoding movavg json failed: %v", err)
	}

	if strings.Contains(data.Information, callFrequencyInfo) {
		return nil, errCallFrequencyInfo
	}

	if data.Information != "" {
		return nil, fmt.Errorf("stock: movavg call returned info: %q", data.Information)
	}

	var vs []*MovingAverageValue
	for dstr, pt := range data.TechnicalAnalysis {
		date, err := parseDate(dstr)
		if err != nil {
			return nil, fmt.Errorf("stock: parsing movavg time (%v) failed: %v", dstr, err)
		}

		avg, err := parseFloat(pt.SMA)
		if err != nil {
			return nil, fmt.Errorf("stock: parsing movavg value (%s) failed: %v", pt.SMA, err)
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
