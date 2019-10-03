// Package iex provides a client to get stock data using the IEX API.
package iex

import (
	"bytes"
	"context"
	"expvar"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"time"
)

var (
	// now is a function to get the current time. Mocked out in tests to return a fixed time.
	now = time.Now

	// loc is the timezone to use when parsing dates.
	loc = mustLoadLocation("America/New_York")
)

// validSymbolRegexp is a regexp that accepts valid stock symbols. Examples: X, FB, SPY, AAPL
var validSymbolRegexp = regexp.MustCompile("^[A-Z]{1,5}$")

var cacheClientVar = expvar.NewMap("iex-client-stats")

// Client is used to make IEX API requests.
type Client struct {
	// chartCache caches chart responses for GetCharts.
	chartCache iexChartCacheInterface

	// dumpAPIResponses dumps API responses into text files.
	dumpAPIResponses bool
}

type iexChartCacheInterface interface {
	Get(ctx context.Context, key ChartCacheKey) (*ChartCacheValue, error)
	Put(ctx context.Context , key ChartCacheKey, val *ChartCacheValue) error
}

// NewClient returns a new Client.
func NewClient(chartCache iexChartCacheInterface, dumpAPIResponses bool) *Client {
	return &Client{
		chartCache:       chartCache,
		dumpAPIResponses: dumpAPIResponses,
	}
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
