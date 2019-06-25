package iex

import (
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
		wantErr bool
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
						LatestSource:  IEXRealTimePrice,
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
						LatestSource:  FifteenMinuteDelayedPrice,
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
						{"date":"2018-09-18","minute":"15:57","open":218.44,"high":218.49,"low":218.37,"close":218.49,"volume":2607},
						{"date":"2018-09-18","minute":"15:58","open":218.46,"high":218.5,"low":218.435,"close":218.44,"volume":3680},
						{"date":"2018-09-18","minute":"15:59","open":218.45,"high":218.49,"low":218.34,"close":218.34,"volume":26153}
					]
				}
			}`,
			want: []*Stock{
				{
					Symbol: "AAPL",
					Quote: &Quote{
						CompanyName:   "Apple Inc.",
						LatestPrice:   225.74,
						LatestSource:  Close,
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
							Date:   time.Date(2018, time.September, 18, 15, 57, 0, 0, loc),
							Open:   218.44,
							High:   218.49,
							Low:    218.37,
							Close:  218.49,
							Volume: 2607,
						},
						{
							Date:   time.Date(2018, time.September, 18, 15, 58, 0, 0, loc),
							Open:   218.46,
							High:   218.5,
							Low:    218.435,
							Close:  218.44,
							Volume: 3680,
						},
						{
							Date:   time.Date(2018, time.September, 18, 15, 59, 0, 0, loc),
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
						LatestSource:  Close,
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
							Date:          time.Date(2017, time.July, 5, 0, 0, 0, 0, loc),
							Open:          66.948,
							High:          68.1103,
							Low:           66.9136,
							Close:         67.7572,
							Volume:        21176272,
							Change:        0.892575,
							ChangePercent: 1.335,
						},
						{
							Date:          time.Date(2017, time.July, 6, 0, 0, 0, 0, loc),
							Open:          66.9627,
							High:          67.4629,
							Low:           66.8156,
							Close:         67.2569,
							Volume:        21117572,
							Change:        -0.500233,
							ChangePercent: -0.738,
						},
						{
							Date:          time.Date(2017, time.July, 7, 0, 0, 0, 0, loc),
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
				t.Errorf("diff (-want, +got)\n%s", diff)
			}

			if (gotErr != nil) != tt.wantErr {
				t.Errorf("got error: %v, wanted err: %t", gotErr, tt.wantErr)
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
		wantErr     bool
	}{
		{
			desc:      "date",
			inputDate: "2018-06-28",
			want:      time.Date(2018, 6, 28, 0, 0, 0, 0, loc),
		},
		{
			desc:        "date and time",
			inputDate:   "2018-06-28",
			inputMinute: "14:53",
			want:        time.Date(2018, 6, 28, 14, 53, 0, 0, loc),
		},
	} {
		t.Run(tt.desc, func(t *testing.T) {
			got, gotErr := chartDate(tt.inputDate, tt.inputMinute)

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("diff (-want, +got)\n%s", diff)
			}

			if (gotErr != nil) != tt.wantErr {
				t.Errorf("got error: %v, wanted err: %t", gotErr, tt.wantErr)
			}
		})
	}
}
