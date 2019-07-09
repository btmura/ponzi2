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
func NewCacheClient(client *Client) *CacheClient {
	return &CacheClient{
		client:     client,
		quoteCache: newQuoteCache(),
		chartCache: newChartCache(),
	}
}

// GetQuotes gets quotes for stock symbols.
func (c *CacheClient) GetQuotes(ctx context.Context, req *GetQuotesRequest) ([]*Quote, error) {
	cacheClientVar.Add("get-quotes-requests", 1)

	var missingSymbols []string
	for _, sym := range req.Symbols {
		k := quoteCacheKey{symbol: sym}
		v := c.quoteCache.get(k)
		if v == nil {
			missingSymbols = append(missingSymbols, sym)
		}
	}

	r := &GetQuotesRequest{Symbols: missingSymbols}
	missingQuotes, err := c.client.GetQuotes(ctx, r)
	if err != nil {
		return nil, err
	}

	for _, q := range missingQuotes {
		k := quoteCacheKey{symbol: q.Symbol}
		v := &quoteCacheValue{
			quote:          q.DeepCopy(),
			lastUpdateTime: now(),
		}
		c.quoteCache.put(k, v)
	}

	var quotes []*Quote
	for _, sym := range req.Symbols {
		k := quoteCacheKey{symbol: sym}
		v := c.quoteCache.get(k)
		quotes = append(quotes, v.quote.DeepCopy())
	}
	return quotes, nil
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

// quoteCache caches data from the quote endpoint.
type quoteCache struct {
	data map[quoteCacheKey]*quoteCacheValue
}

type quoteCacheKey struct {
	symbol string
}

type quoteCacheValue struct {
	quote          *Quote
	lastUpdateTime time.Time
}

func (q *quoteCacheValue) deepCopy() *quoteCacheValue {
	copy := *q
	copy.quote = copy.quote.DeepCopy()
	return &copy
}

func newQuoteCache() *quoteCache {
	return &quoteCache{data: map[quoteCacheKey]*quoteCacheValue{}}
}

func (q *quoteCache) get(key quoteCacheKey) *quoteCacheValue {
	v := q.data[key]
	if v != nil {
		cacheClientVar.Add("quote-cache-hits", 1)
		return v.deepCopy()
	}
	cacheClientVar.Add("quote-cache-misses", 1)
	return nil
}

func (q *quoteCache) put(key quoteCacheKey, val *quoteCacheValue) {
	q.data[key] = val.deepCopy()
	q.data[key].lastUpdateTime = now()
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
