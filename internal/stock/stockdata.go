package stock

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
)

// TradingHistory is a list of trading sessions spanning some time.
type TradingHistory struct {
	Symbol     string
	StartDate  time.Time
	EndDate    time.Time
	DataSource DataSource
	Sessions   []*TradingSession
}

// TradingSession contains stats from a single trading session.
// It often spans a day, but it could span any time period.
type TradingSession struct {
	Date   time.Time
	Open   float32
	High   float32
	Low    float32
	Close  float32
	Volume int
}

// DataSource is the data source to query for tradingSession data.
type DataSource int

// dataSource values.
const (
	DefaultSource DataSource = iota
	Google
	Yahoo
)

// GetTradingHistoryRequest is the request for getTradingHistory.
type GetTradingHistoryRequest struct {
	Symbol     string
	StartDate  time.Time
	EndDate    time.Time
	DataSource DataSource
}

// GetTradingHistory gets the trading history matching the request criteria.
func GetTradingHistory(req *GetTradingHistoryRequest) (*TradingHistory, error) {
	switch req.DataSource {
	case Yahoo:
		return yahooGetTradingHistory(req)

	default:
		return googleGetTradingHistory(req)
	}
}

func googleGetTradingHistory(req *GetTradingHistoryRequest) (*TradingHistory, error) {
	formatTime := func(date time.Time) string {
		return date.Format("Jan 02, 2006")
	}

	v := url.Values{}
	v.Set("q", req.Symbol)
	v.Set("startdate", formatTime(req.StartDate))
	v.Set("enddate", formatTime(req.EndDate))
	v.Set("output", "csv")

	u, err := url.Parse("http://www.google.com/finance/historical")
	if err != nil {
		return nil, err
	}
	u.RawQuery = v.Encode()
	glog.Infof("GET %s", u)

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	history := &TradingHistory{
		Symbol:     req.Symbol,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		DataSource: req.DataSource,
	}
	r := csv.NewReader(resp.Body)
	for i := 0; ; i++ {
		record, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		// wrapErr adds the error to the record for debugging.
		wrapErr := func(err error) error {
			return fmt.Errorf("parsing %q: %v", strings.Join(record, ","), err)
		}

		// format: Date, Open, High, Low, Close, Volume
		if len(record) != 6 {
			return nil, fmt.Errorf("record length should be 6, got %d", len(record))
		}

		// skip header row
		if i != 0 {
			parseRecordTime := func(i int) (time.Time, error) {
				return time.Parse("2-Jan-06", record[i])
			}

			parseRecordFloat := func(i int) (float32, error) {
				return parseFloat(record[i])
			}

			parseRecordInt := func(i int) (int, error) {
				return parseInt(record[i])
			}

			date, err := parseRecordTime(0)
			if err != nil {
				return nil, wrapErr(err)
			}

			open, err := parseRecordFloat(1)
			if err != nil {
				return nil, wrapErr(err)
			}

			high, err := parseRecordFloat(2)
			if err != nil {
				return nil, wrapErr(err)
			}

			low, err := parseRecordFloat(3)
			if err != nil {
				return nil, wrapErr(err)
			}

			close, err := parseRecordFloat(4)
			if err != nil {
				return nil, wrapErr(err)
			}

			volume, err := parseRecordInt(5)
			if err != nil {
				return nil, wrapErr(err)
			}

			history.Sessions = append(history.Sessions, &TradingSession{
				Date:   date,
				Open:   open,
				High:   high,
				Low:    low,
				Close:  close,
				Volume: volume,
			})
		}
	}

	// Most recent trading sessions at the back.
	sort.Sort(byTradingSessionDate(history.Sessions))

	return history, nil
}

func yahooGetTradingHistory(req *GetTradingHistoryRequest) (*TradingHistory, error) {
	v := url.Values{}
	v.Set("s", req.Symbol)
	v.Set("a", strconv.Itoa(int(req.StartDate.Month())-1))
	v.Set("b", strconv.Itoa(req.StartDate.Day()))
	v.Set("c", strconv.Itoa(req.StartDate.Year()))
	v.Set("d", strconv.Itoa(int(req.EndDate.Month())-1))
	v.Set("e", strconv.Itoa(req.EndDate.Day()))
	v.Set("f", strconv.Itoa(req.EndDate.Year()))
	v.Set("g", "d")
	v.Set("ignore", ".csv")

	u, err := url.Parse("http://ichart.yahoo.com/table.csv")
	if err != nil {
		return nil, err
	}
	u.RawQuery = v.Encode()
	glog.Infof("GET %s", u)

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	history := &TradingHistory{
		Symbol:     req.Symbol,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		DataSource: req.DataSource,
	}
	r := csv.NewReader(resp.Body)
	for i := 0; ; i++ {
		record, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		// wrapErr adds the error to the record for debugging.
		wrapErr := func(err error) error {
			return fmt.Errorf("parsing %q: %v", strings.Join(record, ","), err)
		}

		// format: Date, Open, High, Low, Close, Volume, Adj. Close
		if len(record) != 7 {
			return nil, fmt.Errorf("record length should be 7, got %d", len(record))
		}

		// skip header row
		if i != 0 {
			parseRecordTime := func(i int) (time.Time, error) {
				return time.Parse("2006-01-02", record[i])
			}

			parseRecordFloat := func(i int) (float32, error) {
				return parseFloat(record[i])
			}

			parseRecordInt := func(i int) (int, error) {
				return parseInt(record[i])
			}

			date, err := parseRecordTime(0)
			if err != nil {
				return nil, wrapErr(err)
			}

			open, err := parseRecordFloat(1)
			if err != nil {
				return nil, wrapErr(err)
			}

			high, err := parseRecordFloat(2)
			if err != nil {
				return nil, wrapErr(err)
			}

			low, err := parseRecordFloat(3)
			if err != nil {
				return nil, wrapErr(err)
			}

			close, err := parseRecordFloat(4)
			if err != nil {
				return nil, wrapErr(err)
			}

			volume, err := parseRecordInt(5)
			if err != nil {
				return nil, wrapErr(err)
			}

			// Ignore adjusted close value to keep Google and Yahoo APIs the same.

			history.Sessions = append(history.Sessions, &TradingSession{
				Date:   date,
				Open:   open,
				High:   high,
				Low:    low,
				Close:  close,
				Volume: volume,
			})
		}
	}

	// Most recent trading sessions at the back.
	sort.Sort(byTradingSessionDate(history.Sessions))

	return history, nil
}

// Quote is a live stock Quote.
type Quote struct {
	Symbol        string
	Timestamp     time.Time
	Price         float32
	Change        float32
	PercentChange float32
}

// ListQuotesRequest is the request of listQuotes.
type ListQuotesRequest struct {
	Symbols []string
}

// ListQuotesResponse is the response of listQuotes.
type ListQuotesResponse struct {
	// Quotes is a map from symbol to quote.
	Quotes map[string]*Quote
}

// ListQuotes lists the quotes matching the request criteria.
func ListQuotes(req *ListQuotesRequest) (*ListQuotesResponse, error) {
	v := url.Values{}
	v.Set("client", "ig")
	v.Set("q", strings.Join(req.Symbols, ","))

	u, err := url.Parse("http://www.google.com/finance/info")
	if err != nil {
		return nil, err
	}
	u.RawQuery = v.Encode()
	glog.Infof("GET %s", u)

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	parsed := []struct {
		T     string // ticker symbol
		L     string // price
		C     string // change
		Cp    string // percent change
		LtDts string `json:"Lt_dts"` // time
	}{}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Check that data has the expected "//" comment string to trim off.
	if len(data) < 3 {
		return nil, fmt.Errorf("expected data should be larger")
	}

	// Unmarshal the data after the "//" comment string.
	if err := json.Unmarshal(data[3:], &parsed); err != nil {
		return nil, err
	}

	if len(parsed) == 0 {
		return nil, errors.New("expected at least one entry")
	}

	listResp := &ListQuotesResponse{
		Quotes: map[string]*Quote{},
	}
	for _, p := range parsed {
		timestamp, err := time.Parse("2006-01-02T15:04:05Z", p.LtDts)
		if err != nil {
			return nil, fmt.Errorf("p: %+v timestamp: %v", p, err)
		}

		price, err := parseFloat(p.L)
		if err != nil {
			return nil, fmt.Errorf("p: %+v price: %v", p, err)
		}

		var change float32
		if p.C != "" { // C is empty after market close.
			change, err = parseFloat(p.C)
			if err != nil {
				return nil, fmt.Errorf("p: %+v change: %v", p, err)
			}
		}

		var percentChange float32
		if p.Cp != "" { // Cp is empty after market close.
			percentChange, err = parseFloat(p.Cp)
			if err != nil {
				return nil, fmt.Errorf("p: %+v percentChange: %v", p, err)
			}
			percentChange /= 100.0
		}

		listResp.Quotes[p.T] = &Quote{
			Symbol:        p.T,
			Timestamp:     timestamp,
			Price:         price,
			Change:        change,
			PercentChange: percentChange,
		}
	}

	return listResp, nil
}

// parseFloat removes commas and then calls parseFloat.
func parseFloat(value string) (float32, error) {
	f64, err := strconv.ParseFloat(strings.Replace(value, ",", "", -1), 32)
	return float32(f64), err
}

// parseInt parses a string into an int.
func parseInt(value string) (int, error) {
	i64, err := strconv.ParseInt(value, 10, 64)
	return int(i64), err
}

// byTradingSessionDate is a sortable tradingSession slice.
type byTradingSessionDate []*TradingSession

// Len implements sort.Interface.
func (sessions byTradingSessionDate) Len() int {
	return len(sessions)
}

// Less implements sort.Interface.
func (sessions byTradingSessionDate) Less(i, j int) bool {
	return sessions[i].Date.Before(sessions[j].Date)
}

// Swap implements sort.Interface.
func (sessions byTradingSessionDate) Swap(i, j int) {
	sessions[i], sessions[j] = sessions[j], sessions[i]
}
