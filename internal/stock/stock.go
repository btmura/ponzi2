package stock

import (
	"log"
	"os"
	"strconv"
	"time"
)

var logger = log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)

// AlphaVantage uses AlphaVantage to get stock data.
type AlphaVantage struct {
	apiKey string
}

// NewAlphaVantage returns a new AlphaVantage.
func NewAlphaVantage(apiKey string) *AlphaVantage {
	return &AlphaVantage{apiKey: apiKey}
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
