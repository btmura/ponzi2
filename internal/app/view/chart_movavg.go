package view

import (
	"image"

	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/gfx"
)

type chartMovingAverage struct {
	color [3]float32
	vao   *gfx.VAO
}

func newChartMovingAverage(color [3]float32) *chartMovingAverage {
	return &chartMovingAverage{color: color}
}

func (ch *chartMovingAverage) SetData(ts *model.TradingSessionSeries, ms *model.MovingAverageSeries) {
	// Reset everything.
	ch.Close()

	// Bail out if there is not enough data yet.
	if ts == nil || ms == nil {
		return
	}

	// Create the graph line VAO.
	var values []float32
	for _, m := range ms.MovingAverages {
		values = append(values, m.Value)
	}
	ch.vao = dataLineVAO(values, priceRange(ts.TradingSessions), ch.color)
}

func (ch *chartMovingAverage) Render(r image.Rectangle) {
	if ch.vao == nil {
		return
	}
	gfx.SetModelMatrixRect(r)
	ch.vao.Render()
}

func (ch *chartMovingAverage) Close() {
	if ch.vao != nil {
		ch.vao.Delete()
		ch.vao = nil
	}
}
