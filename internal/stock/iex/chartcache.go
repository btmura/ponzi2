package iex

import (
	"encoding/gob"
	"os"
	"os/user"
	"path/filepath"
	"sync"
	"time"

	"github.com/btmura/ponzi2/internal/errors"
)

type ChartCacheKey struct {
	Symbol   string
	Interval ChartInterval
}

type ChartInterval int

const (
	ChartIntervalUnspecified ChartInterval = iota
	MinuteInterval
	DailyInterval
)

type ChartCacheValue struct {
	Chart          *Chart
	LastUpdateTime time.Time
}

func (c *ChartCacheValue) DeepCopy() *ChartCacheValue {
	copy := *c
	copy.Chart = copy.Chart.DeepCopy()
	return &copy
}

type NoOpChartCache struct{}

func (n *NoOpChartCache) Get(key ChartCacheKey) *ChartCacheValue            { return nil }
func (n *NoOpChartCache) Put(key ChartCacheKey, val *ChartCacheValue) error { return nil }

// GOBChartCache caches data from the chart endpoint.
// Fields are exported for gob encoding and decoding.
type GOBChartCache struct {
	Data map[ChartCacheKey]*ChartCacheValue
	mu   sync.Mutex
}

func OpenGOBChartCache() (*GOBChartCache, error) {
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
		return &GOBChartCache{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	c := &GOBChartCache{}
	dec := gob.NewDecoder(file)
	if err := dec.Decode(c); err != nil {
		return nil, err
	}
	return c, nil
}

func (g *GOBChartCache) Get(key ChartCacheKey) *ChartCacheValue {
	g.mu.Lock()
	defer g.mu.Unlock()

	cacheClientVar.Add("chart-cache-gets", 1)

	v := g.Data[key]
	if v != nil {
		cacheClientVar.Add("chart-cache-hits", 1)
		return v.DeepCopy()
	}
	cacheClientVar.Add("chart-cache-misses", 1)
	return nil
}

func (g *GOBChartCache) Put(key ChartCacheKey, val *ChartCacheValue) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	cacheClientVar.Add("chart-cache-puts", 1)

	if !validSymbolRegexp.MatchString(key.Symbol) {
		return errors.Errorf("bad symbol: got %s, want: %v", key.Symbol, validSymbolRegexp)
	}

	if g.Data == nil {
		g.Data = map[ChartCacheKey]*ChartCacheValue{}
	}
	g.Data[key] = val.DeepCopy()
	g.Data[key].LastUpdateTime = now()

	saveChartCache(g)

	return nil
}

func saveChartCache(g *GOBChartCache) error {
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

	return gob.NewEncoder(file).Encode(g)
}

func chartCachePath() (string, error) {
	dir, err := userCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "iex-chart-cache.gob"), nil
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
