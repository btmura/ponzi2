package stock

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
)

func init() {
	flag.Parse() // Avoid glog errors about logging before flag.Parse.
}

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
	return googleGetTradingHistory(req)
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

		// Skip the header row.
		if i == 0 {
			continue
		}

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

		close, err := parseRecordFloat(4)
		if err != nil {
			return nil, wrapErr(err)
		}

		// Open, high, and low can be reported as "-" causing parse errors,
		// so set them to the close by default to fix graph rendering.

		open, high, low := close, close, close

		if record[1] != "-" {
			open, err = parseRecordFloat(1)
			if err != nil {
				return nil, wrapErr(err)
			}
		}

		if record[2] != "-" {
			high, err = parseRecordFloat(2)
			if err != nil {
				return nil, wrapErr(err)
			}
		}

		if record[3] != "-" {
			low, err = parseRecordFloat(3)
			if err != nil {
				return nil, wrapErr(err)
			}
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

	// Most recent trading sessions at the back.
	sortByTradingSessionDate(history.Sessions)

	return history, nil
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

func sortByTradingSessionDate(ss []*TradingSession) {
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Date.Before(ss[j].Date)
	})
}
