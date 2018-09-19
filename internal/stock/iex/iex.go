// Package iex provides a client to get stock data using the IEX API.
package iex

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"time"
)

// loc is the timezone to use when parsing dates.
var loc = mustLoadLocation("America/New_York")

// ChartRange is the range to specify in the request.
type ChartRange string

// ChartRange values.
const (
	ChartRangeOneDay   ChartRange = "1d"
	ChartRangeTwoYears            = "2y"
)

// GetChartRequest is the request for GetChart.
type GetChartRequest struct {
	Symbol string
	Range  ChartRange
}

// Chart is the response from calling GetChart.
type Chart struct {
	Symbol string
	Points []*ChartPoint
}

// ChartPoint is a single point on the chart.
type ChartPoint struct {
	Date          time.Time
	Open          float32
	High          float32
	Low           float32
	Close         float32
	Volume        int
	Change        float32
	ChangePercent float32
}

// Client is used to make IEX API requests.
type Client struct {
	// dumpAPIResponses dumps API responses into text files.
	dumpAPIResponses bool
}

// NewClient returns a new Client.
func NewClient(dumpAPIResponses bool) *Client {
	return &Client{dumpAPIResponses: dumpAPIResponses}
}

// GetChart gets a series of trading sessions for a stock symbol.
func (c *Client) GetChart(ctx context.Context, req *GetChartRequest) (*Chart, error) {
	if req.Symbol == "" {
		return nil, errors.New("iex: missing symbol for chart req")
	}
	if req.Range == "" {
		return nil, errors.New("iex: missing range for chart req")
	}

	u, err := url.Parse(fmt.Sprintf("https://api.iextrading.com/1.0/stock/%s/chart/%s", req.Symbol, req.Range))
	if err != nil {
		return nil, err
	}

	v := url.Values{}
	v.Set("filter", "date,open,high,low,close,volume,change,changePercent")

	u.RawQuery = v.Encode()

	httpReq, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	httpResp, err := http.DefaultClient.Do(httpReq.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	r := httpResp.Body
	if c.dumpAPIResponses {
		rr, err := dumpResponse(fmt.Sprintf("iex-chart-%s.txt", req.Symbol), r)
		if err != nil {
			return nil, fmt.Errorf("iex: failed to dump resp: %v", err)
		}
		r = rr
	}

	resp, err := decodeChart(req.Symbol, r)
	if err != nil {
		return nil, fmt.Errorf("iex: failed to decode resp: %v", err)
	}

	return resp, nil
}

func decodeChart(symbol string, r io.Reader) (*Chart, error) {
	type DataPoint struct {
		Date          string  `json:"date"`
		Open          float64 `json:"open"`
		High          float64 `json:"high"`
		Low           float64 `json:"low"`
		Close         float64 `json:"close"`
		Volume        float64 `json:"volume"`
		Change        float64 `json:"change"`
		ChangePercent float64 `json:"changePercent"`
	}

	var data []DataPoint
	dec := json.NewDecoder(r)
	if err := dec.Decode(&data); err != nil {
		return nil, fmt.Errorf("json decode failed: %v", err)
	}

	var ps []*ChartPoint
	for _, pt := range data {
		date, err := parseDate(pt.Date)
		if err != nil {
			return nil, fmt.Errorf("parsing date (%s) failed: %v", pt.Date, err)
		}

		ps = append(ps, &ChartPoint{
			Date:          date,
			Open:          float32(pt.Open),
			High:          float32(pt.High),
			Low:           float32(pt.Low),
			Close:         float32(pt.Close),
			Volume:        int(pt.Volume),
			Change:        float32(pt.Change),
			ChangePercent: float32(pt.ChangePercent),
		})
	}

	sort.Slice(ps, func(i, j int) bool {
		return ps[i].Date.Before(ps[j].Date)
	})

	return &Chart{Symbol: symbol, Points: ps}, nil
}

func parseDate(s string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", s, loc)
}

func dumpResponse(fileName string, r io.Reader) (io.ReadCloser, error) {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0660)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(file, "%s", b)

	return ioutil.NopCloser(bytes.NewBuffer(b)), nil
}

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		log.Fatalf("time.LoadLocation(%s) failed: %v", name, err)
	}
	return loc
}
