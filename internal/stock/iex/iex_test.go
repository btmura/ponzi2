package iex

import (
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestDecodeChart(t *testing.T) {
	for _, tt := range []struct {
		desc    string
		data    string
		want    []*Chart
		wantErr error
	}{
		{
			desc: "one day",
			data: `{
				"AAPL": {
					"chart": [
						{"date":"20180918","minute":"15:57","open":218.44,"high":218.49,"low":218.37,"close":218.49,"volume":2607},
						{"date":"20180918","minute":"15:58","open":218.46,"high":218.5,"low":218.435,"close":218.44,"volume":3680},
						{"date":"20180918","minute":"15:59","open":218.45,"high":218.49,"low":218.34,"close":218.34,"volume":26153}
					]
				}
			}`,
			want: []*Chart{
				{
					Symbol: "AAPL",
					Points: []*ChartPoint{
						{
							Date:   mustParseDateMinute("20180918", "15:57"),
							Open:   218.44,
							High:   218.49,
							Low:    218.37,
							Close:  218.49,
							Volume: 2607,
						},
						{
							Date:   mustParseDateMinute("20180918", "15:58"),
							Open:   218.46,
							High:   218.5,
							Low:    218.435,
							Close:  218.44,
							Volume: 3680,
						},
						{
							Date:   mustParseDateMinute("20180918", "15:59"),
							Open:   218.45,
							High:   218.49,
							Low:    218.34,
							Close:  218.34,
							Volume: 26153,
						},
					},
				},
			},
		},
		{
			desc: "daily",
			data: `{
				"MSFT": {
					"chart": [
						{"date":"2017-07-05","open":66.948,"high":68.1103,"low":66.9136,"close":67.7572,"volume":21176272,"change":0.892575,"changePercent":1.335},
						{"date":"2017-07-06","open":66.9627,"high":67.4629,"low":66.8156,"close":67.2569,"volume":21117572,"change":-0.500233,"changePercent":-0.738},
						{"date":"2017-07-07","open":67.3845,"high":68.5026,"low":67.3845,"close":68.1299,"volume":16878317,"change":0.872957,"changePercent":1.298}
					]
				}
			}`,
			want: []*Chart{
				{
					Symbol: "MSFT",
					Points: []*ChartPoint{
						{
							Date:          mustParseDateMinute("2017-07-05", ""),
							Open:          66.948,
							High:          68.1103,
							Low:           66.9136,
							Close:         67.7572,
							Volume:        21176272,
							Change:        0.892575,
							ChangePercent: 1.335,
						},
						{
							Date:          mustParseDateMinute("2017-07-06", ""),
							Open:          66.9627,
							High:          67.4629,
							Low:           66.8156,
							Close:         67.2569,
							Volume:        21117572,
							Change:        -0.500233,
							ChangePercent: -0.738,
						},
						{
							Date:          mustParseDateMinute("2017-07-07", ""),
							Open:          67.3845,
							High:          68.5026,
							Low:           67.3845,
							Close:         68.1299,
							Volume:        16878317,
							Change:        0.872957,
							ChangePercent: 1.298,
						},
					},
				},
			},
		},
	} {
		t.Run(tt.desc, func(t *testing.T) {
			got, gotErr := decodeChart(strings.NewReader(tt.data))

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("resp differs:\n%s", diff)
			}

			if diff := cmp.Diff(fmt.Sprint(tt.wantErr), fmt.Sprint(gotErr)); diff != "" {
				t.Errorf("error differs:\n%s", diff)
			}
		})
	}
}

func TestParseDateMinute(t *testing.T) {
	for _, tt := range []struct {
		desc        string
		inputDate   string
		inputMinute string
		want        time.Time
	}{
		{
			desc:      "date",
			inputDate: "2018-06-28",
			want:      time.Date(2018, 6, 28, 0, 0, 0, 0, loc),
		},
		{
			desc:        "date and time",
			inputDate:   "20180628",
			inputMinute: "14:53",
			want:        time.Date(2018, 6, 28, 14, 53, 0, 0, loc),
		},
	} {
		t.Run(tt.desc, func(t *testing.T) {
			got, gotErr := parseDateMinute(tt.inputDate, tt.inputMinute)
			if gotErr != nil {
				t.Fatalf("returned error (%v), want success", gotErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("differs:\n%s", diff)
			}
		})
	}
}

func mustParseDateMinute(date, minute string) time.Time {
	t, err := parseDateMinute(date, minute)
	if err != nil {
		log.Fatal(err)
	}
	return t
}
