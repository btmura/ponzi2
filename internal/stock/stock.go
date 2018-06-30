package stock

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

// Get esc from github.com/mjibson/esc. It's used to embed resources into the binary.
//go:generate esc -o bindata.go -pkg stock -include ".*(txt)" -modtime 1337 -private data

//go:generate stringer -type=Interval

var logger = log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)

// loc is the timezone to set on parsed dates.
var loc = mustLoadLocation("America/New_York")

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
