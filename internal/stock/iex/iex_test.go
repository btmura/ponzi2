package iex

import (
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestDecodeStocks(t *testing.T) {
	old := now
	defer func() { now = old }()
	now = func() time.Time { return time.Date(2018, time.October, 11, 0, 0, 0, 0, loc) }

	for _, tt := range []struct {
		desc    string
		data    string
		want    []*Stock
		wantErr error
	}{
		{
			desc: "real time quote",
			data: `{"CEF": {"quote":{"companyName":"Sprott Physical Gold and Silver Trust Units","latestPrice":11.71,"latestSource":"IEX real time price","latestTime":"12:45:40 PM","latestUpdate":1538153140524,"latestVolume":478088,"open":11.61,"high":11.72,"low":11.61,"close":11.54,"change":0.17,"changePercent":0.01473}}}`,
			want: []*Stock{
				{
					Symbol: "CEF",
					Quote: &Quote{
						CompanyName:   "Sprott Physical Gold and Silver Trust Units",
						LatestPrice:   11.71,
						LatestSource:  SourceIEXRealTimePrice,
						LatestTime:    time.Date(2018, time.October, 11, 12, 45, 40, 0, loc),
						LatestUpdate:  time.Unix(1538153140, 524000000),
						LatestVolume:  478088,
						Open:          11.61,
						High:          11.72,
						Low:           11.61,
						Close:         11.54,
						Change:        0.17,
						ChangePercent: 0.01473,
					},
				},
			},
		},
		{
			desc: "delayed quote",
			data: `{"UUP":{"quote":{"companyName":"Invesco DB USD Index Bullish Fund","latestPrice":25.234,"latestSource":"15 minute delayed price","latestTime":"12:32:11 PM","latestUpdate":1538152331455,"latestVolume":1000000,"open":25.3,"high":25.314,"low":25.22,"close":25.2,"change":0.034,"changePercent":0.00135}}}`,
			want: []*Stock{
				{
					Symbol: "UUP",
					Quote: &Quote{
						CompanyName:   "Invesco DB USD Index Bullish Fund",
						LatestPrice:   25.234,
						LatestSource:  Source15MinuteDelayedPrice,
						LatestTime:    time.Date(2018, time.October, 11, 12, 32, 11, 0, loc),
						LatestUpdate:  time.Unix(1538152331, 455000000),
						LatestVolume:  1000000,
						Open:          25.3,
						High:          25.314,
						Low:           25.22,
						Close:         25.2,
						Change:        0.034,
						ChangePercent: 0.00135,
					},
				},
			},
		},
		{
			desc: "one day quote and chart",
			data: `{
				"AAPL": {
					"quote":{"companyName":"Apple Inc.","latestPrice":225.74,"latestSource":"Close","latestTime":"September 28, 2018","latestUpdate":1538164800414,"latestVolume":22067409,"open":224.8,"high":225.84,"low":224.02,"close":225.74,"change":0.79,"changePercent":0.00351},
					"chart": [
						{"date":"20180918","minute":"15:57","open":218.44,"high":218.49,"low":218.37,"close":218.49,"volume":2607},
						{"date":"20180918","minute":"15:58","open":218.46,"high":218.5,"low":218.435,"close":218.44,"volume":3680},
						{"date":"20180918","minute":"15:59","open":218.45,"high":218.49,"low":218.34,"close":218.34,"volume":26153}
					]
				}
			}`,
			want: []*Stock{
				{
					Symbol: "AAPL",
					Quote: &Quote{
						CompanyName:   "Apple Inc.",
						LatestPrice:   225.74,
						LatestSource:  SourceClose,
						LatestTime:    time.Date(2018, time.September, 28, 0, 0, 0, 0, loc),
						LatestUpdate:  time.Unix(1538164800, 414000000),
						LatestVolume:  22067409,
						Open:          224.8,
						High:          225.84,
						Low:           224.02,
						Close:         225.74,
						Change:        0.79,
						ChangePercent: 0.00351,
					},
					Chart: []*ChartPoint{
						{
							Date:   mustChartDate("20180918", "15:57"),
							Open:   218.44,
							High:   218.49,
							Low:    218.37,
							Close:  218.49,
							Volume: 2607,
						},
						{
							Date:   mustChartDate("20180918", "15:58"),
							Open:   218.46,
							High:   218.5,
							Low:    218.435,
							Close:  218.44,
							Volume: 3680,
						},
						{
							Date:   mustChartDate("20180918", "15:59"),
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
			desc: "daily quote and chart",
			data: `{
				"MSFT": {
					"quote":{"companyName":"Microsoft Corporation","latestPrice":114.37,"latestSource":"Close","latestTime":"September 28, 2018","latestUpdate":1538164800600,"latestVolume":20491683,"open":114.17,"high":114.57,"low":113.68,"close":114.37,"change":-0.04,"changePercent":-0.00035},
					"chart": [
						{"date":"2017-07-05","open":66.948,"high":68.1103,"low":66.9136,"close":67.7572,"volume":21176272,"change":0.892575,"changePercent":1.335},
						{"date":"2017-07-06","open":66.9627,"high":67.4629,"low":66.8156,"close":67.2569,"volume":21117572,"change":-0.500233,"changePercent":-0.738},
						{"date":"2017-07-07","open":67.3845,"high":68.5026,"low":67.3845,"close":68.1299,"volume":16878317,"change":0.872957,"changePercent":1.298}
					]
				}
			}`,
			want: []*Stock{
				{
					Symbol: "MSFT",
					Quote: &Quote{
						CompanyName:   "Microsoft Corporation",
						LatestPrice:   114.37,
						LatestSource:  SourceClose,
						LatestTime:    time.Date(2018, time.September, 28, 0, 0, 0, 0, loc),
						LatestUpdate:  time.Unix(1538164800, 600000000),
						LatestVolume:  20491683,
						Open:          114.17,
						High:          114.57,
						Low:           113.68,
						Close:         114.37,
						Change:        -0.04,
						ChangePercent: -0.00035,
					},
					Chart: []*ChartPoint{
						{
							Date:          mustChartDate("2017-07-05", ""),
							Open:          66.948,
							High:          68.1103,
							Low:           66.9136,
							Close:         67.7572,
							Volume:        21176272,
							Change:        0.892575,
							ChangePercent: 1.335,
						},
						{
							Date:          mustChartDate("2017-07-06", ""),
							Open:          66.9627,
							High:          67.4629,
							Low:           66.8156,
							Close:         67.2569,
							Volume:        21117572,
							Change:        -0.500233,
							ChangePercent: -0.738,
						},
						{
							Date:          mustChartDate("2017-07-07", ""),
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
			got, gotErr := decodeStocks(strings.NewReader(tt.data))

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("resp differs:\n%s", diff)
			}

			if diff := cmp.Diff(fmt.Sprint(tt.wantErr), fmt.Sprint(gotErr)); diff != "" {
				t.Errorf("error differs:\n%s", diff)
			}
		})
	}
}

func TestQuoteDate(t *testing.T) {
	old := now
	defer func() { now = old }()
	now = func() time.Time { return time.Date(2018, time.October, 11, 0, 0, 0, 0, loc) }

	for _, tt := range []struct {
		desc              string
		inputLatestSource Source
		inputLatestTime   string
		want              time.Time
	}{
		{
			desc:              "IEX real time price",
			inputLatestSource: SourceIEXRealTimePrice,
			inputLatestTime:   "2:52:11 PM",
			want:              time.Date(2018, time.October, 11, 14, 52, 11, 0, loc),
		},
		{
			desc:              "15 minute delayed price",
			inputLatestSource: Source15MinuteDelayedPrice,
			inputLatestTime:   "12:32:11 PM",
			want:              time.Date(2018, time.October, 11, 12, 32, 11, 0, loc),
		},
		{
			desc:              "previous close",
			inputLatestSource: SourcePreviousClose,
			inputLatestTime:   "September 25, 2018",
			want:              time.Date(2018, time.September, 25, 0, 0, 0, 0, loc),
		},
		{
			desc:              "close",
			inputLatestSource: SourceClose,
			inputLatestTime:   "September 25, 2018",
			want:              time.Date(2018, time.September, 25, 0, 0, 0, 0, loc),
		},
	} {
		t.Run(tt.desc, func(t *testing.T) {
			got, gotErr := quoteDate(tt.inputLatestSource, tt.inputLatestTime)
			if gotErr != nil {
				t.Fatalf("returned error (%v), want success", gotErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("differs:\n%s", diff)
			}
		})
	}
}

func TestChartDate(t *testing.T) {
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
			got, gotErr := chartDate(tt.inputDate, tt.inputMinute)
			if gotErr != nil {
				t.Fatalf("returned error (%v), want success", gotErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("differs:\n%s", diff)
			}
		})
	}
}

func mustChartDate(date, minute string) time.Time {
	t, err := chartDate(date, minute)
	if err != nil {
		log.Fatal(err)
	}
	return t
}
