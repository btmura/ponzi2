package ponzi

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// tradingHistory is a list of trading sessions spanning some time.
type tradingHistory struct {
	symbol     string
	startDate  time.Time
	endDate    time.Time
	dataSource dataSource
	sessions   []*tradingSession
}

// tradingSession contains stats from a single trading session.
// It often spans a day, but it could span any time period.
type tradingSession struct {
	date   time.Time
	open   float32
	high   float32
	low    float32
	close  float32
	volume int
}

// dataSource is the data source to query for tradingSession data.
type dataSource int

// dataSource values.
const (
	defaultSource dataSource = iota
	google
	yahoo
)

// getTradingHistoryRequest is the request for getTradingHistory.
type getTradingHistoryRequest struct {
	symbol     string
	startDate  time.Time
	endDate    time.Time
	dataSource dataSource
}

// getTradingHistory gets the trading history matching the request criteria.
func getTradingHistory(req *getTradingHistoryRequest) (*tradingHistory, error) {
	switch req.dataSource {
	case yahoo:
		return yahooGetTradingHistory(req)

	default:
		return googleGetTradingHistory(req)
	}
}

func googleGetTradingHistory(req *getTradingHistoryRequest) (*tradingHistory, error) {
	formatTime := func(date time.Time) string {
		return date.Format("Jan 02, 2006")
	}

	v := url.Values{}
	v.Set("q", req.symbol)
	v.Set("startdate", formatTime(req.startDate))
	v.Set("enddate", formatTime(req.endDate))
	v.Set("output", "csv")

	u, err := url.Parse("http://www.google.com/finance/historical")
	if err != nil {
		return nil, err
	}
	u.RawQuery = v.Encode()
	log.Printf("GET %s", u)

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	history := &tradingHistory{
		symbol:     req.symbol,
		startDate:  req.startDate,
		endDate:    req.endDate,
		dataSource: req.dataSource,
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
				return nil, err
			}

			open, err := parseRecordFloat(1)
			if err != nil {
				return nil, err
			}

			high, err := parseRecordFloat(2)
			if err != nil {
				return nil, err
			}

			low, err := parseRecordFloat(3)
			if err != nil {
				return nil, err
			}

			close, err := parseRecordFloat(4)
			if err != nil {
				return nil, err
			}

			volume, err := parseRecordInt(5)
			if err != nil {
				return nil, err
			}

			history.sessions = append(history.sessions, &tradingSession{
				date:   date,
				open:   open,
				high:   high,
				low:    low,
				close:  close,
				volume: volume,
			})
		}
	}

	// Most recent trading sessions at the front.
	sort.Reverse(bySessionDate(history.sessions))

	return history, nil
}

func yahooGetTradingHistory(req *getTradingHistoryRequest) (*tradingHistory, error) {
	v := url.Values{}
	v.Set("s", req.symbol)
	v.Set("a", strconv.Itoa(int(req.startDate.Month())-1))
	v.Set("b", strconv.Itoa(req.startDate.Day()))
	v.Set("c", strconv.Itoa(req.startDate.Year()))
	v.Set("d", strconv.Itoa(int(req.endDate.Month())-1))
	v.Set("e", strconv.Itoa(req.endDate.Day()))
	v.Set("f", strconv.Itoa(req.endDate.Year()))
	v.Set("g", "d")
	v.Set("ignore", ".csv")

	u, err := url.Parse("http://ichart.yahoo.com/table.csv")
	if err != nil {
		return nil, err
	}
	u.RawQuery = v.Encode()
	log.Printf("GET %s", u)

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	history := &tradingHistory{
		symbol:     req.symbol,
		startDate:  req.startDate,
		endDate:    req.endDate,
		dataSource: req.dataSource,
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
				return nil, err
			}

			open, err := parseRecordFloat(1)
			if err != nil {
				return nil, err
			}

			high, err := parseRecordFloat(2)
			if err != nil {
				return nil, err
			}

			low, err := parseRecordFloat(3)
			if err != nil {
				return nil, err
			}

			close, err := parseRecordFloat(4)
			if err != nil {
				return nil, err
			}

			volume, err := parseRecordInt(5)
			if err != nil {
				return nil, err
			}

			// Ignore adjusted close value to keep Google and Yahoo APIs the same.

			history.sessions = append(history.sessions, &tradingSession{
				date:   date,
				open:   open,
				high:   high,
				low:    low,
				close:  close,
				volume: volume,
			})
		}
	}

	// Most recent trading sessions at the front.
	sort.Reverse(bySessionDate(history.sessions))

	return history, nil
}

// quote is a live stock quote.
type quote struct {
	symbol        string
	timestamp     time.Time
	price         float32
	change        float32
	percentChange float32
}

// listQuotesRequest is the request of listQuotes.
type listQuotesRequest struct {
	symbols []string
}

// listQuotesResponse is the response of listQuotes.
type listQuotesResponse struct {
	// quotes is a map from symbol to quote.
	quotes map[string]*quote
}

// listQuotes lists the quotes matching the request criteria.
func listQuotes(req *listQuotesRequest) (*listQuotesResponse, error) {
	v := url.Values{}
	v.Set("client", "ig")
	v.Set("q", strings.Join(req.symbols, ","))

	u, err := url.Parse("http://www.google.com/finance/info")
	if err != nil {
		return nil, err
	}
	u.RawQuery = v.Encode()
	log.Printf("GET %s", u)

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	parsed := []struct {
		T      string // ticker symbol
		L      string // price
		C      string // change
		Cp     string // percent change
		Lt_dts string // time
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

	listResp := &listQuotesResponse{
		quotes: map[string]*quote{},
	}
	for _, p := range parsed {
		timestamp, err := time.Parse("2006-01-02T15:04:05Z", p.Lt_dts)
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

		listResp.quotes[p.T] = &quote{
			symbol:        p.T,
			timestamp:     timestamp,
			price:         price,
			change:        change,
			percentChange: percentChange,
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

// bySessionDate is a sortable tradingSession slice.
type bySessionDate []*tradingSession

// Len implements sort.Interface.
func (sessions bySessionDate) Len() int {
	return len(sessions)
}

// Less implements sort.Interface.
func (sessions bySessionDate) Less(i, j int) bool {
	return sessions[i].date.Before(sessions[j].date)
}

// Swap implements sort.Interface.
func (sessions bySessionDate) Swap(i, j int) {
	sessions[i], sessions[j] = sessions[j], sessions[i]
}
