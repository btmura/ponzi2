package controller

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/btmura/ponzi2/internal/app/model"

	"gitlab.com/btmura/ponzi2/internal/stock/iex"
)

func TestModelMinuteChart(t *testing.T) {
	for _, tt := range []struct {
		desc    string
		input   *iex.Stock
		want    *model.MinuteChart
		wantErr error
	}{
		{
			input: &iex.Stock{
				Quote: &iex.Quote{CompanyName: "Apple Inc."},
				Chart: []*iex.ChartPoint{
					{
						Date:   time.Date(2018, time.September, 18, 15, 57, 0, 0, loc),
						Open:   218.44,
						High:   218.49,
						Low:    218.37,
						Close:  218.49,
						Volume: 2607,
					},
				},
			},
			want: &model.MinuteChart{
				Quote: &model.Quote{CompanyName: "Apple Inc."},
				TradingSessionSeries: &model.TradingSessionSeries{
					TradingSessions: []*model.TradingSession{
						{
							Date:   time.Date(2018, time.September, 18, 15, 57, 0, 0, loc),
							Open:   218.44,
							High:   218.49,
							Low:    218.37,
							Close:  218.49,
							Volume: 2607,
						},
					},
				},
			},
		},
	} {
		got, gotErr := modelMinuteChart(tt.input)

		if diff := cmp.Diff(tt.want, got); diff != "" {
			t.Errorf("diff (-want, +got)\n%s", diff)
		}

		if diff := cmp.Diff(tt.wantErr, gotErr); diff != "" {
			t.Errorf("diff (-wantErr, +gotErr)\n%s", diff)
		}
	}
}