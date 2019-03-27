package controller

import (
	"testing"
	"time"

	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/stock/iex"
	"github.com/google/go-cmp/cmp"
)

func TestModelOneDayChart(t *testing.T) {
	for _, tt := range []struct {
		desc    string
		input   *iex.Stock
		want    *model.Chart
		wantErr error
	}{
		{
			input: &iex.Stock{
				Quote: &iex.Quote{CompanyName: "Apple Inc."},
				Chart: []*iex.ChartPoint{
					{
						Date:   time.Date(2018, time.September, 18, 15, 57, 0, 0, time.UTC),
						Open:   218.44,
						High:   218.49,
						Low:    218.37,
						Close:  218.49,
						Volume: 2607,
					},
				},
			},
			want: &model.Chart{
				Quote: &model.Quote{CompanyName: "Apple Inc."},
				Range: model.OneDay,
				TradingSessionSeries: &model.TradingSessionSeries{
					TradingSessions: []*model.TradingSession{
						{
							Date:   time.Date(2018, time.September, 18, 15, 57, 0, 0, time.UTC),
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
		got, gotErr := modelOneDayChart(tt.input)

		if diff := cmp.Diff(tt.want, got); diff != "" {
			t.Errorf("diff (-want, +got)\n%s", diff)
		}

		if diff := cmp.Diff(tt.wantErr, gotErr); diff != "" {
			t.Errorf("diff (-wantErr, +gotErr)\n%s", diff)
		}
	}
}
