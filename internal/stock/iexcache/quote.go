package iexcache

import (
	"context"

	"github.com/btmura/ponzi2/internal/stock/iex"
)

// GetQuotes gets quotes for stock symbols.
func (c *Client) GetQuotes(ctx context.Context, req *iex.GetQuotesRequest) ([]*iex.Quote, error) {
	cacheClientVar.Add("get-quotes-requests", 1)
	return c.client.GetQuotes(ctx, req)
}
