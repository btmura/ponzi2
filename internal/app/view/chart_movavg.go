package view

import (
	"image"

	"gitlab.com/btmura/ponzi2/internal/app/model"
	"gitlab.com/btmura/ponzi2/internal/app/gfx"
)

type chartMovingAverage struct {
	color [3]float32
	line  *gfx.VAO
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
	ch.line = dataLineVAO(values, priceRange(ts.TradingSessions), ch.color)
}

func (ch *chartMovingAverage) Render(r image.Rectangle) {
	if ch.line == nil {
		return
	}
	gfx.SetModelMatrixRect(r)
	ch.line.Render()
}

func (ch *chartMovingAverage) Close() {
	if ch.line != nil {
		ch.line.Delete()
		ch.line = nil
	}
}
