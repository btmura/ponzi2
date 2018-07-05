package alpha

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"sort"
	"strings"
)

// GetStochastics returns Stochastics or an error.
func (a *AlphaVantage) GetStochastics(ctx context.Context, req *GetStochasticsRequest) (*Stochastics, error) {
	if req.Symbol == "" {
		return nil, fmt.Errorf("stock: stoch request missing symbol: %v", req)
	}

	v := url.Values{}
	v.Set("function", "STOCH")
	v.Set("symbol", req.Symbol)
	v.Set("interval", intervalReqParam(req.Interval))
	v.Set("fastkperiod", "14")
	v.Set("slowkperiod", "3")
	v.Set("slowdperiod", "3")
	v.Set("apikey", a.apiKey)

	debugID := fmt.Sprintf("debug-stoch-%s-%v", req.Symbol, req.Interval)

	var stoch *Stochastics
	if err := a.query(ctx, v, debugID, func(r io.Reader) error {
		s, err := decodeStochasticsResponse(r)
		if err != nil {
			return err
		}
		stoch = s
		return nil
	}); err != nil {
		return nil, fmt.Errorf("stock: querying stoch failed: %v", err)
	}

	return stoch, nil
}

func decodeStochasticsResponse(r io.Reader) (*Stochastics, error) {
	type DataPoint struct {
		SlowK string
		SlowD string
	}

	type Data struct {
		Information       string               `json:"Information"`
		TechnicalAnalysis map[string]DataPoint `json:"Technical Analysis: STOCH"`
	}

	var data Data
	dec := json.NewDecoder(r)
	if err := dec.Decode(&data); err != nil {
		return nil, fmt.Errorf("stock: decoding stoch json failed: %v", err)
	}

	if strings.Contains(data.Information, callFrequencyInfo) {
		return nil, errCallFrequencyInfo
	}

	if data.Information != "" {
		return nil, fmt.Errorf("stock: stoch call returned info: %q", data.Information)
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
