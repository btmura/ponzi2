package iex

import (
	"expvar"
	"os"
	"os/user"
	"path/filepath"
)

var cacheClientVar = expvar.NewMap("iex-cache-client-stats")

// CacheClient is used to make IEX API requests with caching.
type CacheClient struct {
	// client is used to make IEX API requests without caching.
	client *Client

	// quoteCache caches quote responses.
	quoteCache *quoteCache

	// chartCache caches chart responses.
	chartCache *chartCache
}

// NewCacheClient returns a new CacheClient.
func NewCacheClient(client *Client) (*CacheClient, error) {
	q, err := loadQuoteCache()
	if err != nil {
		return nil, err
	}

	c, err := loadChartCache()
	if err != nil {
		return nil, err
	}

	return &CacheClient{
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
