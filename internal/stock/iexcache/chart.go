package iexcache

import (
	"context"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/btmura/ponzi2/internal/errors"
	"github.com/btmura/ponzi2/internal/stock/iex"
)

const debugChart = false

// GetCharts gets charts for stock symbols.
func (c *Client) GetCharts(ctx context.Context, req *iex.GetChartsRequest) ([]*iex.Chart, error) {
	cacheClientVar.Add("get-charts-requests", 1)

	if len(req.Symbols) == 0 {
		return nil, nil
	}

	if req.Range != iex.TwoYears {
		return nil, errors.Errorf("only the two years range is supported")
	}

	fixedNow := now()
	today := toDate(fixedNow)

	type data struct {
		// cacheChart is the chart found in the cache. Nil if not in cache.
		cacheChart *iex.Chart

		// minChartLast is the minimum chartLast value to complete the data set.
		// 0 means make a request for the range's default data.
		// -1 means don't make any request at all.
		minChartLast int

		// responseChart is the chart data from calling the API. Nil if API not called.
		responseChart *iex.Chart

		// finalChart is the non-nil final chart to cache and return.
		finalChart *iex.Chart
	}

	symbol2Data := map[string]*data{}

	dump := func(i int) {
		if !debugChart {
			return
		}
		for sym, data := range symbol2Data {
			fmt.Printf("[%d] %s: %v\n", i, sym, data)
		}
	}

	for _, sym := range req.Symbols {
		k := newChartCacheKey(sym, daily)
		v := c.chartCache.get(k)
		if v == nil {
			symbol2Data[sym] = &data{minChartLast: 0}
			continue
		}

		ps := v.Chart.ChartPoints

		// If cached value has no data, then consider this missing.
		if len(ps) == 0 {
			symbol2Data[sym] = &data{
				cacheChart:   v.Chart.DeepCopy(),
				minChartLast: 0,
			}
			continue
		}

		// Compute the number of points required to be combined with the cached value
		// by counting business days between the latest point's date and today's date.
		minChartLast := -1

		latest := toDate(ps[len(ps)-1].Date)

		for {
			if debugChart {
				fmt.Printf("%s: l: %v t: %v\n", sym, latest, today)
			}

			latest = latest.AddDate(0, 0, 1 /* day */)

			// Don't ask for data in the future. :)
			if !latest.Before(today) {
				break
			}

			// Don't ask for data for weekends, since the market is closed.
			// Keep iterating though.
			if latest.Weekday() != time.Saturday && latest.Weekday() != time.Sunday {
				if minChartLast == -1 {
					minChartLast = 0
				}
				minChartLast++
			}
		}

		symbol2Data[sym] = &data{
			cacheChart:   v.Chart.DeepCopy(),
			minChartLast: minChartLast,
		}
	}

	dump(0)

	chartLast2Request := map[int]*iex.GetChartsRequest{}
	for sym, data := range symbol2Data {
		if data.minChartLast == -1 {
			continue
		}
		req := chartLast2Request[data.minChartLast]
		if req == nil {
			req = &iex.GetChartsRequest{
				Range:     iex.TwoYears,
				ChartLast: data.minChartLast,
			}
			chartLast2Request[data.minChartLast] = req
		}
		req.Symbols = append(req.Symbols, sym)
	}

	var reqs []*iex.GetChartsRequest
	for _, req := range chartLast2Request {
		reqs = append(reqs, req)
	}

	responses := make([][]*iex.Chart, len(reqs))

	g, gCtx := errgroup.WithContext(ctx)
	for i, req := range reqs {
		i, req := i, req
		if debugChart {
			fmt.Printf("%d: api request: %v\n", i, req)
		}
		g.Go(func() error {
			resp, err := c.client.GetCharts(gCtx, req)
			if err != nil {
				return err
			}
			responses[i] = resp
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	for _, charts := range responses {
		for _, ch := range charts {
			data := symbol2Data[ch.Symbol]
			data.responseChart = ch
		}
	}

	dump(1)

	for sym, data := range symbol2Data {
		switch data.minChartLast {
		case -1:
			data.finalChart = data.cacheChart

		case 0:
			data.finalChart = data.responseChart

		default:
			date2Point := map[time.Time]*iex.ChartPoint{}
			for _, pt := range data.cacheChart.ChartPoints {
				date2Point[timeKey(pt.Date)] = pt
			}
			for _, pt := range data.responseChart.ChartPoints {
				date2Point[timeKey(pt.Date)] = pt
			}

			var pts []*iex.ChartPoint
			for _, pt := range date2Point {
				pts = append(pts, pt)
			}
			sort.Slice(pts, func(i, j int) bool {
				return pts[i].Date.Before(pts[j].Date)
			})

			data.finalChart = &iex.Chart{
				Symbol:      sym,
				ChartPoints: pts,
			}
		}
	}

	dump(2)

	for sym, data := range symbol2Data {
		k := newChartCacheKey(sym, daily)
		v := &chartCacheValue{
			Chart:          data.finalChart,
			LastUpdateTime: fixedNow,
		}
		if err := c.chartCache.put(k, v); err != nil {
			return nil, err
		}
	}

	if err := saveChartCache(c.chartCache); err != nil {
		return nil, err
	}

	var charts []*iex.Chart
	for _, sym := range req.Symbols {
		data := symbol2Data[sym]
		charts = append(charts, data.finalChart)
	}
	return charts, nil
}

// chartCache caches data from the chart endpoint.
// Fields are exported for gob encoding and decoding.
type chartCache struct {
	Data map[chartCacheKey]*chartCacheValue
	mu   sync.Mutex
}

type chartCacheKey struct {
	Symbol string
	Type   chartType
}

type chartType int

const (
	chartTypeUnspecified chartType = iota
	minute
	daily
)

func newChartCacheKey(symbol string, chType chartType) chartCacheKey {
	return chartCacheKey{Symbol: symbol, Type: chType}
}

type chartCacheValue struct {
	Chart          *iex.Chart
	LastUpdateTime time.Time
}

func (c *chartCacheValue) deepCopy() *chartCacheValue {
	copy := *c
	copy.Chart = copy.Chart.DeepCopy()
	return &copy
}

func (c *chartCache) get(key chartCacheKey) *chartCacheValue {
	c.mu.Lock()
	defer c.mu.Unlock()

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
	c.mu.Lock()
	defer c.mu.Unlock()

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

	return gob.NewEncoder(file).Encode(q)
}

func chartCachePath() (string, error) {
	dir, err := userCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "iex-chart-cache.gob"), nil
}

func timeKey(t time.Time) time.Time {
	return t.UTC().Round(0)
}

func toDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
