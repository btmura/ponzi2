package iex

import (
	"context"
	"encoding/gob"
	"os"
	"path/filepath"
	"time"

	"github.com/btmura/ponzi2/internal/errors"
)

// GetCharts gets charts for stock symbols.
func (c *CacheClient) GetCharts(ctx context.Context, req *GetChartsRequest) ([]*Chart, error) {
	cacheClientVar.Add("get-charts-requests", 1)

	symbol2Chart := map[string]*Chart{}
	var missingSymbols []string
	for _, sym := range req.Symbols {
		k := chartCacheKey{Symbol: sym, DataRange: req.Range}
		v := c.chartCache.get(k)
		if v != nil && v.fresh(now()) {
			symbol2Chart[sym] = v.Chart.DeepCopy()
			continue
		}
		missingSymbols = append(missingSymbols, sym)
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
		k := chartCacheKey{Symbol: ch.Symbol, DataRange: req.Range}
		v := &chartCacheValue{
			Chart:          ch.DeepCopy(),
			LastUpdateTime: now(),
		}
		symbol2Chart[ch.Symbol] = ch.DeepCopy()
		if err := c.chartCache.put(k, v); err != nil {
			return nil, err
		}
	}

	if err := saveChartCache(c.chartCache); err != nil {
		return nil, err
	}

	var charts []*Chart
	for _, sym := range req.Symbols {
		ch := symbol2Chart[sym]
		if ch == nil {
			ch = &Chart{Symbol: sym}
		}
		charts = append(charts, ch)
	}
	return charts, nil
}

// chartCache caches data from the chart endpoint.
// Fields are exported for gob encoding and decoding.
type chartCache struct {
	Data map[chartCacheKey]*chartCacheValue
}

type chartCacheKey struct {
	Symbol    string
	DataRange Range
}

type chartCacheValue struct {
	Chart          *Chart
	LastUpdateTime time.Time
}

func (c *chartCacheValue) fresh(now time.Time) bool {
	return now.Year() == c.LastUpdateTime.Year() &&
		now.Month() == c.LastUpdateTime.Month() &&
		now.Day() == c.LastUpdateTime.Day()
}

func (c *chartCacheValue) deepCopy() *chartCacheValue {
	copy := *c
	copy.Chart = copy.Chart.DeepCopy()
	return &copy
}

func (c *chartCache) get(key chartCacheKey) *chartCacheValue {
	cacheClientVar.Add("chart-cache-gets", 1)

	v := c.Data[key]
	if v != nil {
		cacheClientVar.Add("chart-cache-hits", 1)
		return v.deepCopy()
	}
	cacheClientVar.Add("chart-cache-misses", 1)
	return nil
}

func (c *chartCache) put(key chartCacheKey, val *chartCacheValue) error {
	cacheClientVar.Add("chart-cache-puts", 1)

	if !validSymbolRegexp.MatchString(key.Symbol) {
		return errors.Errorf("bad symbol: got %s, want: %v", key.Symbol, validSymbolRegexp)
	}

	if c.Data == nil {
		c.Data = map[chartCacheKey]*chartCacheValue{}
	}
	c.Data[key] = val.deepCopy()
	c.Data[key].LastUpdateTime = now()

	return nil
}

func loadChartCache() (*chartCache, error) {
	t := now()
	defer func() {
		cacheClientVar.Set("chart-cache-load-time", time.Since(t))
	}()

	path, err := chartCachePath()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return &chartCache{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	q := &chartCache{}
	dec := gob.NewDecoder(file)
	if err := dec.Decode(q); err != nil {
		return nil, err
	}
	return q, nil
}

func saveChartCache(q *chartCache) error {
	t := now()
	defer func() {
		cacheClientVar.Set("chart-cache-save-time", time.Since(t))
	}()

	path, err := chartCachePath()
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

func chartCachePath() (string, error) {
	dir, err := userCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "iex-chart-cache.gob"), nil
}
