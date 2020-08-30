package iex

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/btmura/ponzi2/internal/errs"
	"github.com/btmura/ponzi2/internal/logger"
)

// Chart has points for a stock chart.
type Chart struct {
	Symbol      string
	ChartPoints []*ChartPoint
}

// DeepCopy returns a deep copy of the chart.
func (c *Chart) DeepCopy() *Chart {
	if c == nil {
		return nil
	}
	deep := *c
	if len(deep.ChartPoints) != 0 {
		deep.ChartPoints = make([]*ChartPoint, len(c.ChartPoints))
		for i, cp := range c.ChartPoints {
			deep.ChartPoints[i] = cp.DeepCopy()
		}
	}
	return &deep
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
	if c == nil {
		return nil
	}
	deep := *c
	return &deep
}

// GetChartsRequest is the request for GetCharts.
type GetChartsRequest struct {
	Token     string
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

// GetCharts gets charts for stock symbols.
func (c *Client) GetCharts(ctx context.Context, req *GetChartsRequest) ([]*Chart, error) {
	cacheClientVar.Add("get-charts-requests", 1)

	if req.Token == "" {
		return nil, ErrMissingAPIToken
	}

	if len(req.Symbols) == 0 {
		return nil, nil
	}

	if req.Range != TwoYears {
		return nil, errs.Errorf("only the two years range is supported")
	}

	fixedNow := now()
	today := midnight(fixedNow)

	type data struct {
		// cacheChart is the chart found in the cache. Nil if not in cache.
		cacheChart *Chart

		// minChartLast is the minimum chartLast value to complete the data set.
		// 0 means make a request for the range's default data.
		// -1 means don't make any request at all.
		minChartLast int

		// responseChart is the chart data from calling the API. Nil if API not called.
		responseChart *Chart

		// finalChart is the non-nil final chart to cache and return.
		finalChart *Chart
	}

	symbol2Data := map[string]*data{}

	for _, sym := range req.Symbols {
		k := ChartCacheKey{req.Token, sym, DailyInterval}
		v, err := c.chartCache.Get(ctx, k)
		if err != nil {
			return nil, err
		}
		if v == nil {
			symbol2Data[sym] = &data{minChartLast: 0}
			continue
		}

		ps := v.Chart.ChartPoints

		// If cached value has no data, then consider this missing.
		if len(ps) == 0 {
			symbol2Data[sym] = &data{
				cacheChart:   v.Chart.DeepCopy(),
				minChartLast: 0,
			}
			continue
		}

		// Compute the number of points required to be combined with the cached value
		// by counting business days between the latest point's date and today's date.
		minChartLast := -1

		latest := midnight(ps[len(ps)-1].Date)

		for {
			latest = latest.AddDate(0, 0, 1 /* day */)

			// Don't ask for data in the future. :)
			if !latest.Before(today) {
				break
			}

			// Don't ask for data for weekends, since the market is closed.
			// Keep iterating though.
			if latest.Weekday() != time.Saturday && latest.Weekday() != time.Sunday {
				if minChartLast == -1 {
					minChartLast = 0
				}
				minChartLast++
			}
		}

		symbol2Data[sym] = &data{
			cacheChart:   v.Chart.DeepCopy(),
			minChartLast: minChartLast,
		}
	}

	token := req.Token
	chartLast2Request := map[int]*GetChartsRequest{}
	for sym, data := range symbol2Data {
		if data.minChartLast == -1 {
			continue
		}
		req := chartLast2Request[data.minChartLast]
		if req == nil {
			req = &GetChartsRequest{
				Token:     token,
				Range:     TwoYears,
				ChartLast: data.minChartLast,
			}
			chartLast2Request[data.minChartLast] = req
		}
		req.Symbols = append(req.Symbols, sym)
	}

	var reqs []*GetChartsRequest
	for _, req := range chartLast2Request {
		reqs = append(reqs, req)
	}

	responses := make([][]*Chart, len(reqs))

	g, gCtx := errgroup.WithContext(ctx)
	for i, req := range reqs {
		i, req := i, req
		g.Go(func() error {
			resp, err := c.noCacheGetCharts(gCtx, req)
			if err != nil {
				return err
			}
			responses[i] = resp
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	for _, charts := range responses {
		for _, ch := range charts {
			data := symbol2Data[ch.Symbol]
			data.responseChart = ch
		}
	}

	for sym, data := range symbol2Data {
		switch data.minChartLast {
		case -1:
			data.finalChart = data.cacheChart

		case 0:
			data.finalChart = data.responseChart

		default:
			date2Point := map[time.Time]*ChartPoint{}
			for _, pt := range data.cacheChart.ChartPoints {
				date2Point[timeKey(pt.Date)] = pt
			}
			for _, pt := range data.responseChart.ChartPoints {
				date2Point[timeKey(pt.Date)] = pt
			}

			var pts []*ChartPoint
			for _, pt := range date2Point {
				pts = append(pts, pt)
			}
			sort.Slice(pts, func(i, j int) bool {
				return pts[i].Date.Before(pts[j].Date)
			})

			data.finalChart = &Chart{
				Symbol:      sym,
				ChartPoints: pts,
			}
		}
	}

	for sym, data := range symbol2Data {
		k := ChartCacheKey{req.Token, sym, DailyInterval}
		v := &ChartCacheValue{
			Chart:          data.finalChart,
			LastUpdateTime: fixedNow,
		}
		if err := c.chartCache.Put(ctx, k, v); err != nil {
			return nil, err
		}
	}

	var charts []*Chart
	for _, sym := range req.Symbols {
		data := symbol2Data[sym]
		charts = append(charts, data.finalChart)
	}
	return charts, nil
}

func (c *Client) noCacheGetCharts(ctx context.Context, req *GetChartsRequest) ([]*Chart, error) {
	if req.Token == "" {
		return nil, ErrMissingAPIToken
	}

	if len(req.Symbols) == 0 {
		return nil, nil
	}

	if req.Range == RangeUnspecified {
		return nil, errs.Errorf("iex: missing range for chart req")
	}

	var rangeStr string
	switch req.Range {
	case OneDay:
		rangeStr = "1d"
	case TwoYears:
		rangeStr = "2y"
	default:
		return nil, errs.Errorf("iex: unsupported range for chart req: %s", req.Range)
	}

	if req.ChartLast < 0 {
		return nil, errs.Errorf("iex: chart last must be greater than or equal to zero")
	}

	u, err := url.Parse("https://cloud.iexapis.com/stable/stock/market/batch")
	if err != nil {
		return nil, err
	}

	v := url.Values{}
	v.Set("token", req.Token)
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
	defer func() {
		if err := httpResp.Body.Close(); err != nil {
			logger.Error(err)
		}
	}()

	r := httpResp.Body
	if c.dumpAPIResponses {
		ss := make([]string, len(req.Symbols))
		copy(ss, req.Symbols)
		sort.Strings(ss)

		rr, err := dumpResponse(fmt.Sprintf("iex-chart-%s-%v.txt", strings.Join(ss, "-"), rangeStr), r)
		if err != nil {
			return nil, errs.Errorf("iex: failed to dump chart resp: %v", err)
		}
		r = rr
	}

	charts, err := decodeCharts(r)
	if err != nil {
		return nil, errs.Errorf("iex: failed to decode chart resp: %v", err)
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

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, errs.Errorf("reading chart json failed: %v", err)
	}

	var m map[string]stock
	dec := json.NewDecoder(bytes.NewReader(b))
	if err := dec.Decode(&m); err != nil {
		return nil, errs.Errorf("chart json decode failed: %v, got: %s", err, string(b))
	}

	var charts []*Chart

	for sym, st := range m {
		ch := &Chart{Symbol: sym}
		charts = append(charts, ch)

		for _, pt := range st.Chart {
			date, err := chartDate(pt.Date, pt.Minute)
			if err != nil {
				return nil, errs.Errorf("parsing date (%s) failed: %v", pt.Date, err)
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

// timeKey converts a time into a key usable in maps
// by normalizing the location and stripping the monotonic clock.
func timeKey(t time.Time) time.Time {
	return t.UTC().Round(0)
}

// midnight strips the hours, minutes, seconds, and nanoseconds from the given time.
func midnight(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
