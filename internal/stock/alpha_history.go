package stock

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"sort"
	"strings"
)

// GetHistory returns stock data or an error.
func (a *AlphaVantage) GetHistory(ctx context.Context, req *GetHistoryRequest) (*History, error) {
	if req.Symbol == "" {
		return nil, fmt.Errorf("stock: history request missing symbol: %v", req)
	}

	v := url.Values{}
	v.Set("function", "TIME_SERIES_DAILY")
	v.Set("symbol", req.Symbol)
	v.Set("outputsize", "compact")
	v.Set("datatype", "json")
	v.Set("apikey", a.apiKey)

	debugID := fmt.Sprintf("debug-hist-%s", req.Symbol)

	var hist *History
	if err := a.query(ctx, v, debugID, func(r io.Reader) error {
		h, err := decodeHistoryResponse(r)
		if err != nil {
			return err
		}
		hist = h
		return nil
	}); err != nil {
		return nil, fmt.Errorf("stock: querying hist failed: %v", err)
	}

	return hist, nil
}

func decodeHistoryResponse(r io.Reader) (*History, error) {
	type DataPoint struct {
		Open   string `json:"1. open"`
		High   string `json:"2. high"`
		Low    string `json:"3. low"`
		Close  string `json:"4. close"`
		Volume string `json:"5. volume"`
	}

	type Data struct {
		Information string               `json:"Information"`
		TimeSeries  map[string]DataPoint `json:"Time Series (Daily)"`
	}

	var data Data
	dec := json.NewDecoder(r)
	if err := dec.Decode(&data); err != nil {
		return nil, fmt.Errorf("stock: decoding hist json failed: %v", err)
	}

	if strings.Contains(data.Information, callFrequencyInfo) {
		return nil, errCallFrequencyInfo
	}

	if data.Information != "" {
		return nil, fmt.Errorf("stock: hist call returned info: %q", data.Information)
	}

	var ts []*TradingSession
	for dstr, pt := range data.TimeSeries {
		date, err := parseDate(dstr)
		if err != nil {
			return nil, fmt.Errorf("stock: parsing hist time (%v) failed: %v", dstr, err)
		}

		open, err := parseFloat(pt.Open)
		if err != nil {
			return nil, fmt.Errorf("stock: can't parse open: %v", err)
		}

		high, err := parseFloat(pt.High)
		if err != nil {
			return nil, fmt.Errorf("stock: can't parse high: %v", err)
		}

		low, err := parseFloat(pt.Low)
		if err != nil {
			return nil, fmt.Errorf("stock: can't parse low: %v", err)
		}

		close, err := parseFloat(pt.Close)
		if err != nil {
			return nil, fmt.Errorf("stock: can't parse close: %v", err)
		}

		volume, err := parseInt(pt.Volume)
		if err != nil {
			return nil, fmt.Errorf("stock: can't parse volume: %v", err)
		}

		ts = append(ts, &TradingSession{
			Date:   date,
			Open:   open,
			High:   high,
			Low:    low,
			Close:  close,
			Volume: volume,
		})
	}

	sort.Slice(ts, func(i, j int) bool {
		return ts[i].Date.Before(ts[j].Date)
	})

	return &History{TradingSessions: ts}, nil
}
