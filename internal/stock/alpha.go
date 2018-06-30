package stock

import (
	"context"
	"net/http"
	"strconv"
	"time"
)

// AlphaVantage uses AlphaVantage to get stock data.
type AlphaVantage struct {
	// apiKey is the API key registered on the Alpha Vantage site.
	apiKey string

	// dumpAPIResponses dumps API responses into text files.
	dumpAPIResponses bool

	// waiter is used to wait a second between API requests.
	waiter
}

// NewAlphaVantage returns a new AlphaVantage.
func NewAlphaVantage(apiKey string, dumpAPIResponses bool) *AlphaVantage {
	return &AlphaVantage{
		apiKey:           apiKey,
		dumpAPIResponses: dumpAPIResponses,
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

func intervalReqParam(i Interval) string {
	switch i {
	case Daily:
		return "daily"
	case Weekly:
		return "weekly"
	default:
		return ""
	}
}

func parseDate(s string) (time.Time, error) {
	// All except one value will be a historical date without a timestamp.
	t, err := time.ParseInLocation("2006-01-02", s, loc)
	if err == nil {
		return t, nil
	}

	// One value may have a timestamp if the market is open.
	return time.ParseInLocation("2006-01-02 15:04:05", s, loc)
}

func parseFloat(value string) (float32, error) {
	f64, err := strconv.ParseFloat(value, 32)
	return float32(f64), err
}

func parseInt(value string) (int, error) {
	i64, err := strconv.ParseInt(value, 10, 64)
	return int(i64), err
}
