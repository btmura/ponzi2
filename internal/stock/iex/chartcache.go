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
	Chart          *Chart
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

// timeKey converts a time into a key usable in maps
// by normalizing the location and stripping the monotonic clock.
func timeKey(t time.Time) time.Time {
	return t.UTC().Round(0)
}

// midnight strips the hours, minutes, seconds, and nanoseconds from the given time.
func midnight(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
