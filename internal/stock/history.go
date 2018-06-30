package stock

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/url"
	"sort"
	"time"
)

// GetHistoryRequest is a request for a stock's trading history.
type GetHistoryRequest struct {
	// Symbol is the stock's symbol like "SPY".
	Symbol string
}

// History is a stock's trading history.
type History struct {
	// TradingSessions is a sorted slice of trading sessions spanning some time.
	TradingSessions []*TradingSession
}

// TradingSession contains stats from a single trading session.
// It often spans a day, but it could span any time period.
type TradingSession struct {
	Date   time.Time
	Open   float32
	High   float32
	Low    float32
	Close  float32
	Volume int
}

// GetHistory returns stock data or an error.
func (a *AlphaVantage) GetHistory(ctx context.Context, req *GetHistoryRequest) (*History, error) {
	if req.Symbol == "" {
		return nil, fmt.Errorf("stock: history request missing symbol: %v", req)
	}

	v := url.Values{}
	v.Set("function", "TIME_SERIES_DAILY")
	v.Set("symbol", req.Symbol)
	v.Set("outputsize", "compact")
	v.Set("datatype", "csv")
	v.Set("apikey", a.apiKey)

	u, err := url.Parse("https://www.alphavantage.co/query")
	if err != nil {
		logger.Fatalf("can't parse url")
	}
	u.RawQuery = v.Encode()

	resp, err := a.httpGet(ctx, u.String())
	if err != nil {
		return nil, fmt.Errorf("stock: can't get data: %v", err)
	}
	defer resp.Body.Close()

	r := resp.Body
	if a.dumpAPIResponses {
		rr, err := dumpResponse(fmt.Sprintf("debug-hist-%s.txt", req.Symbol), r)
		if err != nil {
			return nil, fmt.Errorf("stock: dumping hist resp failed: %v", err)
		}
		r = rr
	}
	return decodeHistoryResponse(r)
}

func decodeHistoryResponse(r io.Reader) (*History, error) {
	var ts []*TradingSession

	cr := csv.NewReader(r)
	for i := 0; ; i++ {
		rec, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// Skip the header row: timestamp, open, high, low, close, volume
		if i == 0 {
			continue
		}

		if len(rec) != 6 {
			return nil, fmt.Errorf("stock: rec length should be 6, got %d", len(rec))
		}

		date, err := parseDate(rec[0])
		if err != nil {
			return nil, fmt.Errorf("stock: can't parse timestamp: %v", err)
		}

		open, err := parseFloat(rec[1])
		if err != nil {
			return nil, fmt.Errorf("stock: can't parse open: %v", err)
		}

		high, err := parseFloat(rec[2])
		if err != nil {
			return nil, fmt.Errorf("stock: can't parse high: %v", err)
		}

		low, err := parseFloat(rec[3])
		if err != nil {
			return nil, fmt.Errorf("stock: can't parse low: %v", err)
		}

		close, err := parseFloat(rec[4])
		if err != nil {
			return nil, fmt.Errorf("stock: can't parse close: %v", err)
		}

		volume, err := parseInt(rec[5])
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

	// Most recent trading sessions at the back.
	sort.Slice(ts, func(i, j int) bool {
		return ts[i].Date.Before(ts[j].Date)
	})

	return &History{TradingSessions: ts}, nil
}
