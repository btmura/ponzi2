package iexcache

import (
	"expvar"
	"os"
	"os/user"
	"path/filepath"
	"regexp"

	"github.com/btmura/ponzi2/internal/stock/iex"
)

// validSymbolRegexp is a regexp that accepts valid stock symbols. Examples: X, FB, SPY, AAPL
var validSymbolRegexp = regexp.MustCompile("^[A-Z]{1,5}$")

var cacheClientVar = expvar.NewMap("iexcache-client-stats")

// Client is used to make IEX API requests and cache the results.
type Client struct {
	// client is used to make IEX API requests without caching.
	client *iex.Client

	// quoteCache caches quote responses.
	quoteCache *quoteCache

	// chartCache caches chart responses.
	chartCache *chartCache
}

// NewClient returns a new CacheClient.
func NewClient(client *iex.Client) (*Client, error) {
	q, err := loadQuoteCache()
	if err != nil {
		return nil, err
	}

	c, err := loadChartCache()
	if err != nil {
		return nil, err
	}

	return &Client{
		client:     client,
		quoteCache: q,
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
