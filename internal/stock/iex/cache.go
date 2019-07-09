package iex

import (
	"context"
	"expvar"
	"os"
	"os/user"
	"path/filepath"
	"time"
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

	return &CacheClient{
		client:     client,
		quoteCache: q,
		chartCache: newChartCache(),
	}, nil
}

// GetCharts gets charts for stock symbols.
func (c *CacheClient) GetCharts(ctx context.Context, req *GetChartsRequest) ([]*Chart, error) {
	var missingSymbols []string
	for _, sym := range req.Symbols {
		k := chartCacheKey{symbol: sym, dataRange: req.Range}
		v := c.chartCache.get(k)
		if v == nil {
			missingSymbols = append(missingSymbols, sym)
		}
	}

	r := &GetChartsRequest{
		Symbols:   missingSymbols,
		Range:     req.Range,
		ChartLast: req.ChartLast,
	}
	missingCharts, err := c.client.GetCharts(ctx, r)
	if err != nil {
		return nil, err
	}

	for _, ch := range missingCharts {
		k := chartCacheKey{symbol: ch.Symbol, dataRange: req.Range}
		v := &chartCacheValue{
			chart:          ch.DeepCopy(),
			lastUpdateTime: now(),
		}
		c.chartCache.put(k, v)
	}

	var charts []*Chart
	for _, sym := range req.Symbols {
		k := chartCacheKey{symbol: sym, dataRange: req.Range}
		v := c.chartCache.get(k)
		charts = append(charts, v.chart.DeepCopy())
	}
	return charts, nil
}

// chartCache caches data from the chart endpoint.
type chartCache struct {
	data map[chartCacheKey]*chartCacheValue
}

type chartCacheKey struct {
	symbol    string
	dataRange Range
}

type chartCacheValue struct {
	chart          *Chart
	lastUpdateTime time.Time
}

func (c *chartCacheValue) deepCopy() *chartCacheValue {
	copy := *c
	copy.chart = copy.chart.DeepCopy()
	return &copy
}

func newChartCache() *chartCache {
	return &chartCache{data: map[chartCacheKey]*chartCacheValue{}}
}

func (c *chartCache) get(key chartCacheKey) *chartCacheValue {
	v := c.data[key]
	if v != nil {
		return v.deepCopy()
	}
	return nil
}

func (c *chartCache) put(key chartCacheKey, val *chartCacheValue) {
	c.data[key] = val.deepCopy()
	c.data[key].lastUpdateTime = now()
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
