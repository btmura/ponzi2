package stock

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"time"
)

var logger = log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)

// HistoryRequest is a request for a stock's trading history.
type HistoryRequest struct {
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

// AlphaVantage uses AlphaVantage to get stock data.
type AlphaVantage struct {
	apiKey string
}

// NewAlphaVantage returns a new AlphaVantage.
func NewAlphaVantage(apiKey string) *AlphaVantage {
	return &AlphaVantage{apiKey: apiKey}
}

// GetHistory returns stock data or an error.
func (a *AlphaVantage) GetHistory(req *HistoryRequest) (*History, error) {
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

	logger.Printf("GET %s", u)
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("stock: can't get data: %v", err)
	}
	defer resp.Body.Close()

	var ts []*TradingSession

	r := csv.NewReader(resp.Body)
	for i := 0; ; i++ {
		rec, err := r.Read()
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

func parseDate(dstr string) (time.Time, error) {
	return time.Parse("2006-01-02", dstr)
}

func parseFloat(value string) (float32, error) {
	f64, err := strconv.ParseFloat(value, 32)
	return float32(f64), err
}

func parseInt(value string) (int, error) {
	i64, err := strconv.ParseInt(value, 10, 64)
	return int(i64), err
}
