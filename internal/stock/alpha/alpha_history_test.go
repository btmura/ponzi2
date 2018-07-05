package alpha

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDecodeHistoryResponse(t *testing.T) {
	for _, tt := range []struct {
		desc    string
		input   string
		want    *History
		wantErr error
	}{
		{
			desc: "demo",
			input: `
			{
				"Meta Data": {
					"1. Information": "Daily Prices (open, high, low, close) and Volumes",
					"2. Symbol": "MSFT",
					"3. Last Refreshed": "2018-07-02",
					"4. Output Size": "Compact",
					"5. Time Zone": "US/Eastern"
				},
				"Time Series (Daily)": {
					"2018-07-02": {
						"1. open": "98.1000",
						"2. high": "100.0600",
						"3. low": "98.0000",
						"4. close": "100.0100",
						"5. volume": "18850112"
					},
					"2018-06-29": {
						"1. open": "98.9300",
						"2. high": "99.9100",
						"3. low": "98.3300",
						"4. close": "98.6100",
						"5. volume": "28053214"
					},
					"2018-06-28": {
						"1. open": "97.3800",
						"2. high": "99.1100",
						"3. low": "97.2600",
						"4. close": "98.6300",
						"5. volume": "26650671"
					}
				}
			}`,
			want: &History{
				TradingSessions: []*TradingSession{
					{
						Date:   mustParseDate("2018-06-28"),
						Open:   97.3800,
						High:   99.1100,
						Low:    97.2600,
						Close:  98.6300,
						Volume: 26650671,
					},
					{
						Date:   mustParseDate("2018-06-29"),
						Open:   98.9300,
						High:   99.9100,
						Low:    98.3300,
						Close:  98.6100,
						Volume: 28053214,
					},
					{
						Date:   mustParseDate("2018-07-02"),
						Open:   98.1000,
						High:   100.0600,
						Low:    98.0000,
						Close:  100.0100,
						Volume: 18850112,
					},
				},
			},
		},
		{
			desc: "info",
			input: `
			{
				"Information": "Please consider optimizing your API call frequency."
			}`,
			wantErr: errCallFrequencyInfo,
		},
	} {
		t.Run(tt.desc, func(t *testing.T) {
			got, gotErr := decodeHistoryResponse(strings.NewReader(tt.input))

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("resp differs:\n%s", diff)
			}

			if diff := cmp.Diff(fmt.Sprint(tt.wantErr), fmt.Sprint(gotErr)); diff != "" {
				t.Errorf("error differs:\n%s", diff)
			}
		})
	}
}
