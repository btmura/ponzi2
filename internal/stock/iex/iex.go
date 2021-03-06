// Package iex provides a client to get stock data using the IEX API.
package iex

import (
	"bytes"
	"context"
	"errors"
	"expvar"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"time"

	"github.com/btmura/ponzi2/internal/logger"
)

var (
	// now is a function to get the current time. Mocked out in tests to return a fixed time.
	now = time.Now

	// loc is the timezone to use when parsing dates.
	loc = mustLoadLocation("America/New_York")
)

var (
	// validTokenRegexp is a regexp that accepts valid IEX API tokens.
	validTokenRegexp = regexp.MustCompile("^[A-Za-z0-9_]{1,}$")

	// validSymbolRegexp is a regexp that accepts valid stock symbols. Examples: X, FB, SPY, AAPL
	validSymbolRegexp = regexp.MustCompile("^[A-Z]{1,5}$")
)

var cacheClientVar = expvar.NewMap("iex-client-stats")

// ErrMissingAPIToken is the error returned when a request does not have an API token.
var ErrMissingAPIToken = errors.New("missing API token")

// Client is used to make IEX API requests.
type Client struct {
	// chartCache caches chart responses for GetCharts.
	chartCache iexChartCacheInterface

	// dumpAPIResponses dumps API responses into text files.
	dumpAPIResponses bool
}

type iexChartCacheInterface interface {
	Get(ctx context.Context, key ChartCacheKey) (*ChartCacheValue, error)
	Put(ctx context.Context, key ChartCacheKey, val *ChartCacheValue) error
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
	defer func() {
		if err := file.Close(); err != nil {
			logger.Error(err)
		}
	}()

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(file, "%s", b); err != nil {
		return nil, err
	}

	return ioutil.NopCloser(bytes.NewBuffer(b)), nil
}

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		logger.Fatalf("time.LoadLocation(%s) failed: %v", name, err)
	}
	return loc
}
