package iexcache

import (
	"bytes"
	"context"
	"encoding/gob"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/btmura/ponzi2/internal/errors"
	"github.com/btmura/ponzi2/internal/stock/iex"
)

// GetQuotes gets quotes for stock symbols.
func (c *Client) GetQuotes(ctx context.Context, req *iex.GetQuotesRequest) ([]*iex.Quote, error) {
	cacheClientVar.Add("get-quotes-requests", 1)

	symbol2Quote := map[string]*iex.Quote{}
	var missingSymbols []string
	for _, sym := range req.Symbols {
		k := quoteCacheKey{Symbol: sym}
		v := c.quoteCache.get(k)
		if v != nil && v.fresh(now()) {
			symbol2Quote[sym] = v.Quote.DeepCopy()
			continue
		}
		missingSymbols = append(missingSymbols, sym)
	}

	r := &iex.GetQuotesRequest{Symbols: missingSymbols}
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
		if err := c.quoteCache.put(k, v); err != nil {
			return nil, err
		}
	}

	if err := saveQuoteCache(c.quoteCache); err != nil {
		return nil, err
	}

	var quotes []*iex.Quote
	for _, sym := range req.Symbols {
		q := symbol2Quote[sym]
		if q == nil {
			q = &iex.Quote{Symbol: sym}
		}
		quotes = append(quotes, q)
	}
	return quotes, nil
}

// quoteCache caches data from the quote endpoint.
// Fields are exported for gob encoding and decoding.
type quoteCache struct {
	Data map[quoteCacheKey]*quoteCacheValue
	sync.Mutex
}

type quoteCacheKey struct {
	Symbol string
}

type quoteCacheValue struct {
	Quote          *iex.Quote
	LastUpdateTime time.Time
}

func (q *quoteCacheValue) fresh(now time.Time) bool {
	return now.Year() == q.LastUpdateTime.Year() &&
		now.Month() == q.LastUpdateTime.Month() &&
		now.Day() == q.LastUpdateTime.Day()
}

func (q *quoteCacheValue) deepCopy() *quoteCacheValue {
	copy := *q
	copy.Quote = copy.Quote.DeepCopy()
	return &copy
}

func (q *quoteCache) get(key quoteCacheKey) *quoteCacheValue {
	q.Lock()
	defer q.Unlock()

	cacheClientVar.Add("quote-cache-gets", 1)

	v := q.Data[key]
	if v != nil {
		cacheClientVar.Add("quote-cache-hits", 1)
		return v.deepCopy()
	}
	cacheClientVar.Add("quote-cache-misses", 1)
	return nil
}

func (q *quoteCache) put(key quoteCacheKey, val *quoteCacheValue) error {
	q.Lock()
	defer q.Unlock()

	cacheClientVar.Add("quote-cache-puts", 1)

	if !validSymbolRegexp.MatchString(key.Symbol) {
		return errors.Errorf("bad symbol: got %s, want: %v", key.Symbol, validSymbolRegexp)
	}

	if q.Data == nil {
		q.Data = map[quoteCacheKey]*quoteCacheValue{}
	}
	q.Data[key] = val.deepCopy()
	q.Data[key].LastUpdateTime = now()

	return nil
}

// encodableQuoteCache is an gob encodable version of quoteCache.
// Fields are exported for gob encoding and decoding.
type encodableQuoteCache struct {
	Version int
	Data    map[quoteCacheKey]*quoteCacheValue
}

// GobDecode implements the GobDecoder interface.
func (q *quoteCache) GobDecode(b []byte) error {
	e := &encodableQuoteCache{}

	dec := gob.NewDecoder(bytes.NewReader(b))
	if err := dec.Decode(e); err != nil {
		return err
	}
	q.Data = e.Data

	return nil
}

// GobEncode implements the GobEncoder interface.
func (q *quoteCache) GobEncode() ([]byte, error) {
	e := &encodableQuoteCache{
		Version: 1,
		Data:    q.Data,
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(e); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func loadQuoteCache() (*quoteCache, error) {
	t := now()
	defer func() {
		cacheClientVar.Set("quote-cache-load-time", time.Since(t))
	}()

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
	t := now()
	defer func() {
		cacheClientVar.Set("quote-cache-save-time", time.Since(t))
	}()

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
