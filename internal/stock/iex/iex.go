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
	"strconv"
	"strings"
	"time"
)

// now is a function to get the current time. Mocked out in tests to return a fixed time.
var now = time.Now

// loc is the timezone to use when parsing dates.
var loc = mustLoadLocation("America/New_York")

// Stock is the response from calling GetStocks.
type Stock struct {
	Symbol string
	Quote  *Quote
	Chart  []*ChartPoint
}

// Quote is a stock quote.
type Quote struct {
	CompanyName   string
	LatestPrice   float32
	LatestSource  Source
	LatestTime    time.Time
	LatestUpdate  time.Time
	LatestVolume  int
	Open          float32
	High          float32
	Low           float32
	Close         float32
	Change        float32
	ChangePercent float32
}

// Source is the quote data source.
type Source int

// Source values.
//go:generate stringer -type=Source
const (
	SourceUnspecified Source = iota
	IEXRealTimePrice
	FifteenMinuteDelayedPrice
	Close
	PreviousClose
)

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

// ClientOption is an option for NewClient.
type ClientOption func(c *Client) error

// DumpAPIResponses enables dumping API responses into text files.
func DumpAPIResponses() ClientOption {
	return func(c *Client) error {
		c.dumpAPIResponses = true
		return nil
	}
}

// NewClient returns a new Client.
func NewClient(opts ...ClientOption) (*Client, error) {
	c := &Client{}
	for _, o := range opts {
		if err := o(c); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// stockRequest is the request for GetStocks.
type stockRequest struct {
	rangeVal  string
	chartLast int
}

// StockOption is an option for GetStocks.
type StockOption func(req *stockRequest) error

// Range is the range to specify in the request.
type Range int

// Range values.
//go:generate stringer -type=Range
const (
	RangeUnspecified Range = iota
	OneDay
	TwoYears
)

// WithRange returns an option that requests a data range.
func WithRange(r Range) StockOption {
	return func(req *stockRequest) error {
		switch r {
		case OneDay:
			req.rangeVal = "1d"
		case TwoYears:
			req.rangeVal = "2y"
		default:
			return fmt.Errorf("iex: unsupported range for chart req: %s", r)
		}
		return nil
	}
}

// WithChartLast returns an option that requests the last N chart elements.
func WithChartLast(chartLast int) StockOption {
	return func(req *stockRequest) error {
		if chartLast < 0 {
			return errors.New("iex: chart last must be greater than or equal to zero")
		}
		req.chartLast = chartLast
		return nil
	}
}

// GetStocks gets a series of trading sessions for a stock symbol.
func (c *Client) GetStocks(ctx context.Context, symbols []string, opts ...StockOption) ([]*Stock, error) {
	if len(symbols) == 0 {
		return nil, nil
	}

	req := &stockRequest{}
	for _, o := range opts {
		if err := o(req); err != nil {
			return nil, err
		}
	}

	u, err := url.Parse("https://api.iextrading.com/1.0/stock/market/batch")
	if err != nil {
		return nil, err
	}

	v := url.Values{}
	v.Set("symbols", strings.Join(symbols, ","))
	v.Set("types", "quote,chart")
	v.Set("range", req.rangeVal)
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
	if req.chartLast > 0 {
		v.Set("chartLast", strconv.Itoa(req.chartLast))
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
		rr, err := dumpResponse(fmt.Sprintf("iex-%s.txt", strings.Join(symbols, "-")), r)
		if err != nil {
			return nil, fmt.Errorf("iex: failed to dump resp: %v", err)
		}
		r = rr
	}

	resp, err := decodeStocks(r)
	if err != nil {
		return nil, fmt.Errorf("iex: failed to decode resp: %v", err)
	}

	return resp, nil
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
		return nil, fmt.Errorf("json decode failed: %v", err)
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
				return nil, fmt.Errorf("parsing date (%s) failed: %v", pt.Date, err)
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

func quoteSource(latestSource string) (Source, error) {
	switch latestSource {
	case "IEX real time price":
		return IEXRealTimePrice, nil
	case "15 minute delayed price":
		return FifteenMinuteDelayedPrice, nil
	case "Close":
		return Close, nil
	case "Previous close":
		return PreviousClose, nil
	default:
		return SourceUnspecified, fmt.Errorf("unrecognized source: %q", latestSource)
	}
}

func quoteDate(latestSource Source, latestTime string) (time.Time, error) {
	switch latestSource {
	case IEXRealTimePrice, FifteenMinuteDelayedPrice:
		t, err := time.ParseInLocation("3:04:05 PM", latestTime, loc)
		if err != nil {
			return time.Time{}, err
		}

		n := now()
		return time.Date(
			n.Year(), n.Month(), n.Day(),
			t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location()), nil

	case PreviousClose, Close:
		return time.ParseInLocation("January 2, 2006", latestTime, loc)

	default:
		return time.Time{}, fmt.Errorf("couldn't parse quote date with source(%q) and time(%q)", latestSource, latestTime)
	}
}

func chartDate(date, minute string) (time.Time, error) {
	if minute != "" {
		return time.ParseInLocation("20060102 15:04", date+" "+minute, loc)
	}
	return time.ParseInLocation("2006-01-02", date, loc)
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
