package iex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/btmura/ponzi2/internal/errors"
)

// Chart has points for a stock chart.
type Chart struct {
	Symbol      string
	ChartPoints []*ChartPoint
}

// DeepCopy retuns a deep copy of the chart.
func (c *Chart) DeepCopy() *Chart {
	copy := *c
	copy.ChartPoints = make([]*ChartPoint, len(c.ChartPoints))
	for i := range c.ChartPoints {
		copy.ChartPoints[i] = c.ChartPoints[i].DeepCopy()
	}
	return &copy
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

// DeepCopy returns a deep copy of the chart point.
func (c *ChartPoint) DeepCopy() *ChartPoint {
	copy := *c
	return &copy
}

// GetChartsRequest is the request for GetCharts.
type GetChartsRequest struct {
	Symbols   []string
	Range     Range
	ChartLast int
}

// GetCharts gets charts for stock symbols.
func (c *Client) GetCharts(ctx context.Context, req *GetChartsRequest) ([]*Chart, error) {
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
		return nil, errors.Errorf("iex: chart last must be greater than or equal to zero")
	}

	u, err := url.Parse("https://cloud.iexapis.com/stable/stock/market/batch")
	if err != nil {
		return nil, err
	}

	v := url.Values{}
	v.Set("token", c.token)
	v.Set("symbols", strings.Join(req.Symbols, ","))
	v.Set("types", "chart")
	v.Set("range", rangeStr)
	v.Set("filter", strings.Join([]string{
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
		ss := make([]string, len(req.Symbols))
		copy(ss, req.Symbols)
		sort.Strings(ss)

		rr, err := dumpResponse(fmt.Sprintf("iex-chart-%s-%v.txt", strings.Join(ss, "-"), rangeStr), r)
		if err != nil {
			return nil, errors.Errorf("iex: failed to dump chart resp: %v", err)
		}
		r = rr
	}

	charts, err := decodeCharts(r)
	if err != nil {
		return nil, errors.Errorf("iex: failed to decode chart resp: %v", err)
	}
	return charts, nil
}

func decodeCharts(r io.Reader) ([]*Chart, error) {
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
		Chart []*chartPoint `json:"chart"`
	}

	var m map[string]stock
	dec := json.NewDecoder(r)
	if err := dec.Decode(&m); err != nil {
		return nil, errors.Errorf("chart json decode failed: %v", err)
	}

	var charts []*Chart

	for sym, st := range m {
		ch := &Chart{Symbol: sym}
		charts = append(charts, ch)

		for _, pt := range st.Chart {
			date, err := chartDate(pt.Date, pt.Minute)
			if err != nil {
				return nil, errors.Errorf("parsing date (%s) failed: %v", pt.Date, err)
			}

			ch.ChartPoints = append(ch.ChartPoints, &ChartPoint{
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
		sort.Slice(ch.ChartPoints, func(i, j int) bool {
			return ch.ChartPoints[i].Date.Before(ch.ChartPoints[j].Date)
		})
	}

	return charts, nil
}

func chartDate(date, minute string) (time.Time, error) {
	if minute != "" {
		return time.ParseInLocation("2006-01-02 15:04", date+" "+minute, loc)
	}
	return time.ParseInLocation("2006-01-02", date, loc)
}
