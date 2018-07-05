package stock

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"
)

// loc is the timezone to set on parsed dates.
var loc = mustLoadLocation("America/New_York")

// callFrequencyInfo is the info message returned by Alpha Vantage when the API is overloaded.
const callFrequencyInfo = "Please consider optimizing your API call frequency."

// errCallFrequencyInfo is used internally to decide whether to retry.
var errCallFrequencyInfo = errors.New("stock: api returned call frequency info message")

// maxRetries is the number of retries if getting errCallFrequencyInfo.
const maxRetries = 3

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

func (av *AlphaVantage) query(ctx context.Context, v url.Values, debugID string, decodeFunc func(io.Reader) error) error {
	u, err := url.Parse("https://www.alphavantage.co/query")
	if err != nil {
		return err
	}
	u.RawQuery = v.Encode()

	for i := 0; i < maxRetries; i++ {
		av.wait(time.Duration(i+1) * time.Second)

		req, err := http.NewRequest(http.MethodGet, u.String(), nil)
		if err != nil {
			return err
		}

		logger.Printf("[%s] %d: querying", debugID, i)
		resp, err := http.DefaultClient.Do(req.WithContext(ctx))
		if err != nil {
			return fmt.Errorf("http get failed: %v", err)
		}
		defer resp.Body.Close()

		r := resp.Body
		if av.dumpAPIResponses {
			fileName := fmt.Sprintf("%s.txt", debugID)
			rr, err := dumpResponse(fileName, r)
			if err != nil {
				return fmt.Errorf("dumping resp to %q failed: %v", fileName, err)
			}
			r = rr
		}

		err = decodeFunc(r)

		if err == errCallFrequencyInfo {
			if i+1 == maxRetries {
				return err
			}
			logger.Printf("[%s] %d: retrying", debugID, i)
			continue
		}

		if err != nil {
			return err
		}

		return nil
	}

	return err
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
