package alpha

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDecodeMovingAverageResponse(t *testing.T) {
	for _, tt := range []struct {
		desc    string
		input   string
		want    *MovingAverage
		wantErr error
	}{
		{
			desc: "demo",
			// https://www.alphavantage.co/query?function=SMA&symbol=MSFT&interval=15min&time_period=10&series_type=close&apikey=demo
			input: `
			{
				"Meta Data": {
					"1: Symbol": "MSFT",
					"2: Indicator": "Simple Moving Average (SMA)",
					"3: Last Refreshed": "2018-06-15",
					"4: Interval": "weekly",
					"5: Time Period": 10,
					"6: Series Type": "open",
					"7: Time Zone": "US/Eastern"
				},
				"Technical Analysis: SMA": {
					"2000-03-31": {
						"SMA": "99.4970"
					},
					"2000-03-24": {
						"SMA": "99.9010"
					},
					"2000-03-17": {
						"SMA": "101.3700"
					}
				}
			}`,
			want: &MovingAverage{
				Values: []*MovingAverageValue{
					{
						Date:    mustParseDate("2000-03-17"),
						Average: 101.3700,
					},
					{
						Date:    mustParseDate("2000-03-24"),
						Average: 99.9010,
					},
					{
						Date:    mustParseDate("2000-03-31"),
						Average: 99.4970,
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
			got, gotErr := decodeMovingAverageResponse(strings.NewReader(tt.input))

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("resp differs:\n%s", diff)
			}

			if diff := cmp.Diff(fmt.Sprint(tt.wantErr), fmt.Sprint(gotErr)); diff != "" {
				t.Errorf("error differs:\n%s", diff)
			}
		})
	}
}
