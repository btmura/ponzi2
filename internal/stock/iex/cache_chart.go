package iex

import (
	"context"
	"time"
)

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
