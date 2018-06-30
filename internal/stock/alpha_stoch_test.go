package stock

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDecodeStochasticsResponse(t *testing.T) {
	for _, tt := range []struct {
		desc  string
		input string
		want  *Stochastics
	}{
		{
			desc: "demo",
			// https://www.alphavantage.co/query?function=STOCH&symbol=MSFT&interval=daily&apikey=demo
			input: `
			{
				"Meta Data": {
					"1: Symbol": "MSFT",
					"2: Indicator": "Stochastic (STOCH)",
					"3: Last Refreshed": "2018-06-13",
					"4: Interval": "daily",
					"5.1: FastK Period": 5,
					"5.2: SlowK Period": 3,
					"5.3: SlowK MA Type": 0,
					"5.4: SlowD Period": 3,
					"5.5: SlowD MA Type": 0,
					"6: Time Zone": "US/Eastern Time"
				},
				"Technical Analysis: STOCH": {
					"2018-06-13": {
						"SlowK": "29.8701",
						"SlowD": "38.2982"
					},
					"2018-06-12": {
						"SlowK": "41.1255",
						"SlowD": "50.5565"
					},
					"2018-06-11": {
						"SlowK": "43.8988",
						"SlowD": "63.8097"
					}
				}
			}`,
			want: &Stochastics{
				Values: []*StochasticValue{
					{
						Date: mustParseDate("2018-06-11"),
						K:    43.8988,
						D:    63.8097,
					},
					{
						Date: mustParseDate("2018-06-12"),
						K:    41.1255,
						D:    50.5565,
					},
					{
						Date: mustParseDate("2018-06-13"),
						K:    29.8701,
						D:    38.2982,
					},
				},
			},
		},
	} {
		t.Run(tt.desc, func(t *testing.T) {
			got, gotErr := decodeStochasticsResponse(strings.NewReader(tt.input))
			if gotErr != nil {
				t.Fatalf("decodeStochasticsResponse returned an error (%v), want success", gotErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("differs:\n%s", diff)
			}
		})
	}
}
