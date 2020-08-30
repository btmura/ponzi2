package server

import (
	"context"
	"fmt"

	"gocloud.dev/docstore"
	"gocloud.dev/gcerrors"

	"github.com/btmura/ponzi2/internal/errs"
	"github.com/btmura/ponzi2/internal/stock/iex"
)

// TODO(btmura): add validation for function arguments
// TODO(btmura): add unit tests

// DocChartCache implements iex.iexChartCacheInterface using DocStore.
type DocChartCache struct {
	coll *docstore.Collection
}

// chartCacheDoc is the underlying document stored in the backing Docstore.
// Fields are exported for Docstore.
type chartCacheDoc struct {
	Key        string
	CacheKey   iex.ChartCacheKey
	CacheValue *iex.ChartCacheValue
}

// NewDocChartCache opens the cache at the given url.
func NewDocChartCache(coll *docstore.Collection) *DocChartCache {
	return &DocChartCache{coll: coll}
}

// Get implements the iexChartCacheInterface.
func (d *DocChartCache) Get(ctx context.Context, key iex.ChartCacheKey) (*iex.ChartCacheValue, error) {
	docKey, err := chartCacheDocKey(key)
	if err != nil {
		return nil, errs.Errorf("making get doc key failed: %v", err)
	}

	doc := &chartCacheDoc{
		Key: docKey,
	}
	err = d.coll.Get(ctx, doc)
	if gcerrors.Code(err) == gcerrors.NotFound {
		return nil, nil
	}
	if err != nil {
		return nil, errs.Errorf("getting doc for %s failed: %v", doc.Key, err)
	}
	return doc.CacheValue, nil
}

// Put implements the iexChartCacheInterface.
func (d *DocChartCache) Put(ctx context.Context, key iex.ChartCacheKey, val *iex.ChartCacheValue) error {
	docKey, err := chartCacheDocKey(key)
	if err != nil {
		return errs.Errorf("making put doc key failed: %v", err)
	}

	doc := &chartCacheDoc{
		Key:        docKey,
		CacheKey:   key,
		CacheValue: val,
	}
	if err := d.coll.Put(ctx, doc); err != nil {
		return errs.Errorf("putting doc for %s failed: %v", doc.Key, err)
	}
	return nil
}

func chartCacheDocKey(cacheKey iex.ChartCacheKey) (string, error) {
	return fmt.Sprintf("%s:%s:%v", cacheKey.Token, cacheKey.Symbol, cacheKey.Interval), nil
}
