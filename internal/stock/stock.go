package stock

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
)

// Get esc from github.com/mjibson/esc. It's used to embed resources into the binary.
//go:generate esc -o bindata.go -pkg stock -include ".*(txt)" -modtime 1337 -private data

var logger = log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)

// loc is the timezone to set on parsed dates.
var loc = mustLoadLocation("America/New_York")

// Interval specifies an interval on requests to get stock data.
//go:generate stringer -type=Interval stock.go
type Interval int

// Values for Interval.
const (
	Daily Interval = iota
	Weekly
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

// GetMovingAverageRequest is a request for a stock's moving average.
type GetMovingAverageRequest struct {
	// Symbol is the stock's symbol like "SPY".
	Symbol string

	// TimePeriod is the number of data points to calculate each value.
	TimePeriod int
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

// GetStochasticsRequest is a request for a stock's stochastics.
type GetStochasticsRequest struct {
	// Symbol is the stock's symbol like "SPY".
	Symbol string

	// Interval is the interval like "daily" or "weekly".
	Interval Interval
}

// Stochastics is a time series of stochastic values.
type Stochastics struct {
	// Values are the stochastic values with earlier values in front.
	Values []*StochasticValue
}

// StochasticValue are the stochastic values for some date.
type StochasticValue struct {
	// Date is the start date of the time span covered by this value.
	Date time.Time

	// K measures the stock's momentum.
	K float32

	// D is some moving average of K.
	D float32
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

func dumpResponse(fileName string, r io.Reader) (io.ReadCloser, error) {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0660)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(file, "%s", b)

	return ioutil.NopCloser(bytes.NewBuffer(b)), nil
}

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		log.Fatalf("time.LoadLocation(%s) failed: %v", name, err)
	}
	return loc
}
