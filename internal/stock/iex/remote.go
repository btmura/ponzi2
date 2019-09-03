package iex

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"net/http"
	"net/url"
)

// RemoteClient connects to the server.
type RemoteClient struct {
	url string
}

// NewRemoteClient returns a new RemoteClient.
func NewRemoteClient(url string) *RemoteClient {
	return &RemoteClient{url: url}
}

// GetQuotes sends the request for quotes to the server.
func (c *RemoteClient) GetQuotes(ctx context.Context, req *GetQuotesRequest) ([]*Quote, error) {
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

	var resp []*Quote
	dec := gob.NewDecoder(httpResp.Body)
	if err := dec.Decode(&resp); err != nil {
		return nil, err
	}

	return resp, nil
}

// GetCharts sends the charts request to the server.
func (c *RemoteClient) GetCharts(ctx context.Context, req *GetChartsRequest) ([]*Chart, error) {
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

	var resp []*Chart
	dec := gob.NewDecoder(httpResp.Body)
	if err := dec.Decode(&resp); err != nil {
		return nil, err
	}

	return resp, nil
}
