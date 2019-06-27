package iex

import (
	"context"
	"os"
	"os/user"
	"path/filepath"
	"time"
)

// CacheClient is used to make IEX API requests with caching.
type CacheClient struct {
	// client is used to make IEX API requests without caching.
	client *Client

	// quoteCache caches quote responses.
	quoteCache *quoteCache
}

// NewCacheClient returns a new CacheClient.
func NewCacheClient(client *Client) *CacheClient {
	return &CacheClient{
		client:     client,
		quoteCache: newQuoteCache(),
	}
}

// GetQuotes gets quotes for stock symbols.
func (c *CacheClient) GetQuotes(ctx context.Context, req *GetQuotesRequest) ([]*Quote, error) {
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
		return v.deepCopy()
	}
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
	chartPoints    []*ChartPoint
	lastUpdateTime time.Time
}

func (c *chartCacheValue) deepCopy() *chartCacheValue {
	copy := *c
	copy.chartPoints = nil
	for _, cp := range c.chartPoints {
		copy.chartPoints = append(copy.chartPoints, cp)
	}
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
