package iexcache

import (
	"expvar"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"time"

	"github.com/btmura/ponzi2/internal/stock/iex"
)

// now is a function to get the current time. Mocked out in tests to return a fixed time.
var now = time.Now

// validSymbolRegexp is a regexp that accepts valid stock symbols. Examples: X, FB, SPY, AAPL
var validSymbolRegexp = regexp.MustCompile("^[A-Z]{1,5}$")

var cacheClientVar = expvar.NewMap("iexcache-client-stats")

// Client is used to make IEX API requests and cache the results.
type Client struct {
	// client is used to make IEX API requests without caching.
	client *iex.Client

	// chartCache caches chart responses.
	chartCache *chartCache
}

// Wrap returns a new CacheClient.
func Wrap(client *iex.Client) (*Client, error) {
	c, err := loadChartCache()
	if err != nil {
		return nil, err
	}

	return &Client{
		client:     client,
		chartCache: c,
	}, nil
}

func userCacheDir() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}

	p := filepath.Join(u.HomeDir, ".cache", "ponzi")
	if err := os.MkdirAll(p, 0755); err != nil {
		return "", err
	}
	return p, nil
}

// timeKey converts a time into a key usable in maps
// by normalizing the location and stripping the monotonic clock.
func timeKey(t time.Time) time.Time {
	return t.UTC().Round(0)
}

// midnight strips the hours, minutes, seconds, and nanoseconds from the given time.
func midnight(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		log.Fatalf("time.LoadLocation(%s) failed: %v", name, err)
	}
	return loc
}
