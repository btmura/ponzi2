package iex

import (
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestDecodeQuotes(t *testing.T) {
	old := now
	defer func() { now = old }()
	now = func() time.Time { return time.Date(2018, time.October, 11, 0, 0, 0, 0, loc) }

	for _, tt := range []struct {
		desc    string
		data    string
		want    []*Quote
		wantErr bool
	}{
		{
			desc: "real time quote",
			data: `{"CEF": {"quote":{"companyName":"Sprott Physical Gold and Silver Trust Units","latestPrice":11.71,"latestSource":"IEX real time price","latestTime":"12:45:40 PM","latestUpdate":1538153140524,"latestVolume":478088,"open":11.61,"high":11.72,"low":11.61,"close":11.54,"change":0.17,"changePercent":0.01473}}}`,
			want: []*Quote{
				{
					Symbol:        "CEF",
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
		{
			desc: "delayed quote",
			data: `{"UUP":{"quote":{"companyName":"Invesco DB USD Index Bullish Fund","latestPrice":25.234,"latestSource":"15 minute delayed price","latestTime":"12:32:11 PM","latestUpdate":1538152331455,"latestVolume":1000000,"open":25.3,"high":25.314,"low":25.22,"close":25.2,"change":0.034,"changePercent":0.00135}}}`,
			want: []*Quote{
				{
					Symbol:        "UUP",
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
		{
			desc: "one day quote and chart",
			data: `{"AAPL": {"quote":{"companyName":"Apple Inc.","latestPrice":225.74,"latestSource":"Close","latestTime":"September 28, 2018","latestUpdate":1538164800414,"latestVolume":22067409,"open":224.8,"high":225.84,"low":224.02,"close":225.74,"change":0.79,"changePercent":0.00351}}}`,
			want: []*Quote{
				{
					Symbol:        "AAPL",
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
			},
		},
		{
			desc: "daily quote and chart",
			data: `{"MSFT": {"quote":{"companyName":"Microsoft Corporation","latestPrice":114.37,"latestSource":"Close","latestTime":"September 28, 2018","latestUpdate":1538164800600,"latestVolume":20491683,"open":114.17,"high":114.57,"low":113.68,"close":114.37,"change":-0.04,"changePercent":-0.00035}}}`,
			want: []*Quote{
				{
					Symbol:        "MSFT",
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
			},
		},
	} {
		t.Run(tt.desc, func(t *testing.T) {
			got, gotErr := decodeQuotes(strings.NewReader(tt.data))

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("diff (-want, +got)\n%s", diff)
			}

			if (gotErr != nil) != tt.wantErr {
				t.Errorf("got error: %v, wanted err: %t", gotErr, tt.wantErr)
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
		wantErr           bool
	}{
		{
			desc:              "IEX real time price",
			inputLatestSource: IEXRealTimePrice,
			inputLatestTime:   "2:52:11 PM",
			want:              time.Date(2018, time.October, 11, 14, 52, 11, 0, loc),
		},
		{
			desc:              "15 minute delayed price",
			inputLatestSource: FifteenMinuteDelayedPrice,
			inputLatestTime:   "12:32:11 PM",
			want:              time.Date(2018, time.October, 11, 12, 32, 11, 0, loc),
		},
		{
			desc:              "previous close",
			inputLatestSource: PreviousClose,
			inputLatestTime:   "September 25, 2018",
			want:              time.Date(2018, time.September, 25, 0, 0, 0, 0, loc),
		},
		{
			desc:              "close",
			inputLatestSource: Close,
			inputLatestTime:   "September 25, 2018",
			want:              time.Date(2018, time.September, 25, 0, 0, 0, 0, loc),
		},
	} {
		t.Run(tt.desc, func(t *testing.T) {
			got, gotErr := quoteDate(tt.inputLatestSource, tt.inputLatestTime)

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("diff (-want, +got)\n%s", diff)
			}

			if (gotErr != nil) != tt.wantErr {
				t.Errorf("got error: %v, wanted err: %t", gotErr, tt.wantErr)
			}
		})
	}
}
