package stock

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDecodeHistoryResponse(t *testing.T) {
	for _, tt := range []struct {
		desc  string
		input string
		want  *History
	}{
		{
			desc: "demo",
			input: `timestamp,open,high,low,close,volume
2018-06-29,272.1200,273.6600,271.1495,271.2800,97332077
2018-06-28,269.2900,271.7500,268.4900,270.8900,76650517
2018-06-27,272.2600,273.8650,269.1800,269.3500,104960655
`,
			want: &History{
				TradingSessions: []*TradingSession{
					{
						Date:   mustParseDate("2018-06-27"),
						Open:   272.2600,
						High:   273.8650,
						Low:    269.1800,
						Close:  269.3500,
						Volume: 104960655,
					},
					{
						Date:   mustParseDate("2018-06-28"),
						Open:   269.2900,
						High:   271.7500,
						Low:    268.4900,
						Close:  270.8900,
						Volume: 76650517,
					},
					{
						Date:   mustParseDate("2018-06-29"),
						Open:   272.1200,
						High:   273.6600,
						Low:    271.1495,
						Close:  271.2800,
						Volume: 97332077,
					},
				},
			},
		},
	} {
		t.Run(tt.desc, func(t *testing.T) {
			got, gotErr := decodeHistoryResponse(strings.NewReader(tt.input))
			if gotErr != nil {
				t.Fatalf("decodeHistoryResponse returned an error (%v), want success", gotErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("differs:\n%s", diff)
			}
		})
	}
}
