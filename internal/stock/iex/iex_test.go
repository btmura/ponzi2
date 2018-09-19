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
		symbol  string
		data    string
		want    *Chart
		wantErr error
	}{
		{
			desc:   "demo",
			symbol: "MSFT",
			data: `[
				{"date":"2017-07-05","open":66.948,"high":68.1103,"low":66.9136,"close":67.7572,"volume":21176272,"change":0.892575,"changePercent":1.335},
				{"date":"2017-07-06","open":66.9627,"high":67.4629,"low":66.8156,"close":67.2569,"volume":21117572,"change":-0.500233,"changePercent":-0.738},
				{"date":"2017-07-07","open":67.3845,"high":68.5026,"low":67.3845,"close":68.1299,"volume":16878317,"change":0.872957,"changePercent":1.298}
			]`,
			want: &Chart{
				Symbol: "MSFT",
				Points: []*ChartPoint{
					{
						Date:          mustParseDate("2017-07-05"),
						Open:          66.948,
						High:          68.1103,
						Low:           66.9136,
						Close:         67.7572,
						Volume:        21176272,
						Change:        0.892575,
						ChangePercent: 1.335,
					},
					{
						Date:          mustParseDate("2017-07-06"),
						Open:          66.9627,
						High:          67.4629,
						Low:           66.8156,
						Close:         67.2569,
						Volume:        21117572,
						Change:        -0.500233,
						ChangePercent: -0.738,
					},
					{
						Date:          mustParseDate("2017-07-07"),
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
	} {
		t.Run(tt.desc, func(t *testing.T) {
			got, gotErr := decodeChart(tt.symbol, strings.NewReader(tt.data))

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("resp differs:\n%s", diff)
			}

			if diff := cmp.Diff(fmt.Sprint(tt.wantErr), fmt.Sprint(gotErr)); diff != "" {
				t.Errorf("error differs:\n%s", diff)
			}
		})
	}
}

func TestParseDate(t *testing.T) {
	for _, tt := range []struct {
		desc  string
		input string
		want  time.Time
	}{
		{
			desc:  "date without time",
			input: "2018-06-28",
			want:  time.Date(2018, 6, 28, 0, 0, 0, 0, loc),
		},
	} {
		t.Run(tt.desc, func(t *testing.T) {
			got, gotErr := parseDate(tt.input)
			if gotErr != nil {
				t.Fatalf("parseDate returned an error (%v), want success", gotErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("differs:\n%s", diff)
			}
		})
	}
}

func mustParseDate(dstr string) time.Time {
	date, err := parseDate(dstr)
	if err != nil {
		log.Fatal(err)
	}
	return date
}
