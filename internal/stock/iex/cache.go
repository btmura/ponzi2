package iex

import (
	"context"
	"encoding/gob"
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

// GetQuotes gets quotes for stock symbols.
func (c *CacheClient) GetQuotes(ctx context.Context, req *GetQuotesRequest) ([]*Quote, error) {
	cacheClientVar.Add("get-quotes-requests", 1)

	symbol2Quote := map[string]*Quote{}
	var missingSymbols []string
	for _, sym := range req.Symbols {
		k := quoteCacheKey{Symbol: sym}
		v := c.quoteCache.get(k)
		if v != nil {
			symbol2Quote[sym] = v.Quote.DeepCopy()
		} else {
			missingSymbols = append(missingSymbols, sym)
		}
	}

	r := &GetQuotesRequest{Symbols: missingSymbols}
	missingQuotes, err := c.client.GetQuotes(ctx, r)
	if err != nil {
		return nil, err
	}

	for _, q := range missingQuotes {
		k := quoteCacheKey{Symbol: q.Symbol}
		v := &quoteCacheValue{
			Quote:          q.DeepCopy(),
			LastUpdateTime: now(),
		}
		symbol2Quote[q.Symbol] = q.DeepCopy()
		c.quoteCache.put(k, v)
	}

	if err := saveQuoteCache(c.quoteCache); err != nil {
		return nil, err
	}

	var quotes []*Quote
	for _, sym := range req.Symbols {
		q := symbol2Quote[sym]
		if q == nil {
			q = &Quote{Symbol: sym}
		}
		quotes = append(quotes, q)
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
// Fields are exported for gob encoding and decoding.
type quoteCache struct {
	Data map[quoteCacheKey]*quoteCacheValue
}

type quoteCacheKey struct {
	Symbol string
}

type quoteCacheValue struct {
	Quote          *Quote
	LastUpdateTime time.Time
}

func (q *quoteCacheValue) deepCopy() *quoteCacheValue {
	copy := *q
	copy.Quote = copy.Quote.DeepCopy()
	return &copy
}

func (q *quoteCache) get(key quoteCacheKey) *quoteCacheValue {
	v := q.Data[key]
	if v != nil {
		cacheClientVar.Add("quote-cache-hits", 1)
		return v.deepCopy()
	}
	cacheClientVar.Add("quote-cache-misses", 1)
	return nil
}

func (q *quoteCache) put(key quoteCacheKey, val *quoteCacheValue) {
	if q.Data == nil {
		q.Data = map[quoteCacheKey]*quoteCacheValue{}
	}
	q.Data[key] = val.deepCopy()
	q.Data[key].LastUpdateTime = now()
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

func loadQuoteCache() (*quoteCache, error) {
	path, err := quoteCachePath()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return &quoteCache{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	q := &quoteCache{}
	dec := gob.NewDecoder(file)
	if err := dec.Decode(q); err != nil {
		return nil, err
	}
	return q, nil
}

func saveQuoteCache(q *quoteCache) error {
	path, err := quoteCachePath()
	if err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0660)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := gob.NewEncoder(file)
	if err := enc.Encode(q); err != nil {
		return err
	}

	return nil
}

func quoteCachePath() (string, error) {
	dir, err := userCacheDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "iex-quote-cache.gob"), nil
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
