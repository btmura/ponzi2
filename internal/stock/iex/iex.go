// Package iex provides a client to get stock data using the IEX API.
package iex

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/btmura/ponzi2/internal/errors"
)

// Internal package variables used for the implementation.
var (
	// now is a function to get the current time. Mocked out in tests to return a fixed time.
	now = time.Now

	// loc is the timezone to use when parsing dates.
	loc = mustLoadLocation("America/New_York")
)

// GetStocksRequest is the request for GetStocks.
type GetStocksRequest struct {
	Symbols   []string
	Range     Range
	ChartLast int
}

// Range is the range to specify in the request.
type Range int

// Range values.
//go:generate stringer -type=Range
const (
	RangeUnspecified Range = iota
	OneDay
	TwoYears
)

// Stock is the response from calling GetStocks.
type Stock struct {
	Symbol string
	Quote  *Quote
	Chart  []*ChartPoint
}

// Client is used to make IEX API requests.
type Client struct {
	// token is the API token required on all requests.
	token string

	// dumpAPIResponses dumps API responses into text files.
	dumpAPIResponses bool
}

// NewClient returns a new Client.
func NewClient(token string, dumpAPIResponses bool) *Client {
	return &Client{token: token, dumpAPIResponses: dumpAPIResponses}
}

// GetStocks gets a series of trading sessions for a stock symbol.
func (c *Client) GetStocks(ctx context.Context, req *GetStocksRequest) ([]*Stock, error) {
	if len(req.Symbols) == 0 {
		return nil, nil
	}

	if req.Range == RangeUnspecified {
		return nil, errors.Errorf("iex: missing range for chart req")
	}

	var rangeStr string
	switch req.Range {
	case OneDay:
		rangeStr = "1d"
	case TwoYears:
		rangeStr = "2y"
	default:
		return nil, errors.Errorf("iex: unsupported range for chart req: %s", req.Range)
	}

	if req.ChartLast < 0 {
		return nil, errors.Errorf("iex: last must be greater than or equal to zero")
	}

	u, err := url.Parse("https://cloud.iexapis.com/stable/stock/market/batch")
	if err != nil {
		return nil, err
	}

	v := url.Values{}
	v.Set("token", c.token)
	v.Set("symbols", strings.Join(req.Symbols, ","))
	v.Set("types", "quote,chart")
	v.Set("range", rangeStr)
	v.Set("filter", strings.Join([]string{
		// Keys for quote.
		"companyName",
		"latestPrice",
		"latestSource",
		"latestTime",
		"latestUpdate",
		"latestVolume",

		// Keys for chart and quote.
		"date",
		"minute",
		"open",
		"high",
		"low",
		"close",
		"volume",
		"change",
		"changePercent",
	}, ","))
	if req.ChartLast > 0 {
		v.Set("chartLast", strconv.Itoa(req.ChartLast))
	}
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
		rr, err := dumpResponse(fmt.Sprintf("iex-%s-%v.txt", strings.Join(req.Symbols, "-"), rangeStr), r)
		if err != nil {
			return nil, errors.Errorf("iex: failed to dump resp: %v", err)
		}
		r = rr
	}

	stocks, err := decodeStocks(r)
	if err != nil {
		return nil, errors.Errorf("iex: failed to decode resp: %v", err)
	}
	return stocks, nil
}

func decodeStocks(r io.Reader) ([]*Stock, error) {
	type quote struct {
		CompanyName   string  `json:"companyName"`
		LatestPrice   float64 `json:"latestPrice"`
		LatestSource  string  `json:"latestSource"`
		LatestTime    string  `json:"latestTime"`
		LatestUpdate  int64   `json:"latestUpdate"`
		LatestVolume  int64   `json:"latestVolume"`
		Open          float64 `json:"open"`
		High          float64 `json:"high"`
		Low           float64 `json:"low"`
		Close         float64 `json:"close"`
		Change        float64 `json:"change"`
		ChangePercent float64 `json:"changePercent"`
	}

	type chartPoint struct {
		Date          string  `json:"date"`
		Minute        string  `json:"minute"`
		Open          float64 `json:"open"`
		High          float64 `json:"high"`
		Low           float64 `json:"low"`
		Close         float64 `json:"close"`
		Volume        float64 `json:"volume"`
		Change        float64 `json:"change"`
		ChangePercent float64 `json:"changePercent"`
	}

	type stock struct {
		Quote *quote        `json:"quote"`
		Chart []*chartPoint `json:"chart"`
	}

	var m map[string]stock
	dec := json.NewDecoder(r)
	if err := dec.Decode(&m); err != nil {
		return nil, errors.Errorf("json decode failed: %v", err)
	}

	var chs []*Stock

	for s, d := range m {
		ch := &Stock{Symbol: s}

		if q := d.Quote; q != nil {
			src, err := quoteSource(q.LatestSource)
			if err != nil {
				return nil, err
			}

			date, err := quoteDate(src, q.LatestTime)
			if err != nil {
				return nil, err
			}

			ch.Quote = &Quote{
				CompanyName:   q.CompanyName,
				LatestPrice:   float32(q.LatestPrice),
				LatestSource:  src,
				LatestTime:    date,
				LatestUpdate:  millisToTime(q.LatestUpdate),
				LatestVolume:  int(q.LatestVolume),
				Open:          float32(q.Open),
				High:          float32(q.High),
				Low:           float32(q.Low),
				Close:         float32(q.Close),
				Change:        float32(q.Change),
				ChangePercent: float32(q.ChangePercent),
			}
		}

		for _, pt := range d.Chart {
			date, err := chartDate(pt.Date, pt.Minute)
			if err != nil {
				return nil, errors.Errorf("parsing date (%s) failed: %v", pt.Date, err)
			}

			ch.Chart = append(ch.Chart, &ChartPoint{
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
		sort.Slice(ch.Chart, func(i, j int) bool {
			return ch.Chart[i].Date.Before(ch.Chart[j].Date)
		})
		chs = append(chs, ch)
	}

	return chs, nil
}

func millisToTime(ms int64) time.Time {
	sec := ms / 1e3
	nsec := ms*1e6 - sec*1e9
	return time.Unix(sec, nsec)
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
