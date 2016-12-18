package ponzi

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

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

// listTradingSessionsRequest is the request for listTradingSessions.
type listTradingSessionsRequest struct {
	symbol    string
	startDate time.Time
	endDate   time.Time
}

// listTradingSessionsResponse is the response for listTradingSessions.
type listTradingSessionsResponse struct {
	sessions []*tradingSession
}

// listTradingSessions lists the trading sessions matching the request criteria.
func listTradingSessions(req *listTradingSessionsRequest) (*listTradingSessionsResponse, error) {
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

	listResp := new(listTradingSessionsResponse)
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
				f64, err := strconv.ParseFloat(strings.Replace(record[i], ",", "", -1), 32)
				return float32(f64), err
			}

			parseRecordInt := func(i int) (int, error) {
				i64, err := strconv.ParseInt(record[i], 10, 32)
				return int(i64), err
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

			listResp.sessions = append(listResp.sessions, &tradingSession{
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
	sort.Reverse(bySessionDate(listResp.sessions))

	return listResp, nil
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
