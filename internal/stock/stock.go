package stock

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

var logger = log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)

// Interval specifies an interval on requests to get stock data.
type Interval int

// Values for Interval.
const (
	Daily Interval = iota
	Weekly
)

// reqParam returns request parameter value for the Interval to use in queries.
func (i Interval) reqParam() string {
	switch i {
	case Daily:
		return "daily"
	case Weekly:
		return "weekly"
	default:
		return ""
	}
}

// AlphaVantage uses AlphaVantage to get stock data.
type AlphaVantage struct {
	// apiKey is the API key registered on the Alpha Vantage site.
	apiKey string

	// waiter is used to wait a second between API requests.
	waiter
}

// NewAlphaVantage returns a new AlphaVantage.
func NewAlphaVantage(apiKey string) *AlphaVantage {
	return &AlphaVantage{
		apiKey: apiKey,
	}
}

func (av *AlphaVantage) httpGet(ctx context.Context, url string) (*http.Response, error) {
	av.wait(time.Second) // Alpha Vantage suggests 1 second delay.
	logger.Print(url)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req.WithContext(ctx))
}

type waiter struct {
	time.Time
	sync.Mutex
}

func (w *waiter) wait(d time.Duration) {
	w.Lock()
	if elapsed := time.Since(w.Time); elapsed < d {
		time.Sleep(d - elapsed)
	}
	w.Time = time.Now()
	w.Unlock()
}

func parseDate(s string) (time.Time, error) {
	// All except one value will be a historical date without a timestamp.
	t, err := time.Parse("2006-01-02", s)
	if err == nil {
		return t, nil
	}

	// One value may have a timestamp if the market is open.
	return time.Parse("2006-01-02 15:04:05", s)
}

func parseFloat(value string) (float32, error) {
	f64, err := strconv.ParseFloat(value, 32)
	return float32(f64), err
}

func parseInt(value string) (int, error) {
	i64, err := strconv.ParseInt(value, 10, 64)
	return int(i64), err
}
