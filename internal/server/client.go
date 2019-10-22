package server

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"net/http"
	"net/url"

	"github.com/btmura/ponzi2/internal/errors"
	"github.com/btmura/ponzi2/internal/stock/iex"
)

// Client connects to the server.
type Client struct {
	url string
}

// NewClient returns a new Client.
func NewClient(url string) *Client {
	return &Client{url: url}
}

// GetQuotes sends the request for quotes to the server.
func (c *Client) GetQuotes(ctx context.Context, req *iex.GetQuotesRequest) ([]*iex.Quote, error) {
	fmt.Println("GetQuotes")

	u, err := url.Parse(c.url + "/quote")
	if err != nil {
		return nil, err
	}

	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	if err := enc.Encode(req); err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest(http.MethodGet, u.String(), &b)
	if err != nil {
		return nil, err
	}

	httpResp, err := http.DefaultClient.Do(httpReq.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		var b bytes.Buffer
		b.ReadFrom(httpResp.Body)
		return nil, errors.Errorf("getting quotes failed: %s", b.String())
	}

	var resp []*iex.Quote
	dec := gob.NewDecoder(httpResp.Body)
	if err := dec.Decode(&resp); err != nil {
		return nil, err
	}

	return resp, nil
}

// GetCharts sends the charts request to the server.
func (c *Client) GetCharts(ctx context.Context, req *iex.GetChartsRequest) ([]*iex.Chart, error) {
	fmt.Println("GetCharts")

	u, err := url.Parse(c.url + "/chart")
	if err != nil {
		return nil, err
	}

	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	if err := enc.Encode(req); err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest(http.MethodGet, u.String(), &b)
	if err != nil {
		return nil, err
	}

	httpResp, err := http.DefaultClient.Do(httpReq.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		var b bytes.Buffer
		b.ReadFrom(httpResp.Body)
		return nil, errors.Errorf("getting charts failed: %s", b.String())
	}

	var resp []*iex.Chart
	dec := gob.NewDecoder(httpResp.Body)
	if err := dec.Decode(&resp); err != nil {
		return nil, err
	}

	return resp, nil
}
