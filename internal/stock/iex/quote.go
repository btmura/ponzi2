package iex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/btmura/ponzi2/internal/errors"
)

// Quote is a stock quote.
type Quote struct {
	Symbol        string
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

// DeepCopy returns a deep copy of the quote.
func (q *Quote) DeepCopy() *Quote {
	copy := *q
	return &copy
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

// GetQuotesRequest is the request for GetQuotes.
type GetQuotesRequest struct {
	Symbols []string
}

// GetQuotes gets quotes for stock symbols.
func (c *Client) GetQuotes(ctx context.Context, req *GetQuotesRequest) ([]*Quote, error) {
	if len(req.Symbols) == 0 {
		return nil, nil
	}

	u, err := url.Parse("https://cloud.iexapis.com/stable/stock/market/batch")
	if err != nil {
		return nil, err
	}

	v := url.Values{}
	v.Set("token", c.token)
	v.Set("symbols", strings.Join(req.Symbols, ","))
	v.Set("types", "quote")
	v.Set("filter", strings.Join([]string{
		"companyName",
		"latestPrice",
		"latestSource",
		"latestTime",
		"latestUpdate",
		"latestVolume",
		"open",
		"high",
		"low",
		"close",
		"change",
		"changePercent",
	}, ","))
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
		ss := make([]string, len(req.Symbols))
		copy(ss, req.Symbols)
		sort.Strings(ss)

		rr, err := dumpResponse(fmt.Sprintf("iex-quote-%s.txt", strings.Join(ss, "-")), r)
		if err != nil {
			return nil, errors.Errorf("iex: failed to dump quote resp: %v", err)
		}
		r = rr
	}

	quotes, err := decodeQuotes(r)
	if err != nil {
		return nil, errors.Errorf("iex: failed to decode quote resp: %v", err)
	}
	return quotes, nil
}

func decodeQuotes(r io.Reader) ([]*Quote, error) {
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

	type stock struct {
		Quote *quote `json:"quote"`
	}

	var m map[string]stock
	dec := json.NewDecoder(r)
	if err := dec.Decode(&m); err != nil {
		return nil, errors.Errorf("quote json decode failed: %v", err)
	}

	var quotes []*Quote

	for sym, st := range m {
		q := st.Quote
		if q == nil {
			continue
		}

		src, err := quoteSource(q.LatestSource)
		if err != nil {
			return nil, err
		}

		date, err := quoteDate(src, q.LatestTime)
		if err != nil {
			return nil, err
		}

		quotes = append(quotes, &Quote{
			Symbol:        sym,
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
		})
	}

	return quotes, nil
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
		return SourceUnspecified, errors.Errorf("unrecognized source: %q", latestSource)
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
		return time.Time{}, errors.Errorf("couldn't parse quote date with source(%q) and time(%q)", latestSource, latestTime)
	}
}