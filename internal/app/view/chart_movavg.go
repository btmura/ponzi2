package view

import (
	"image"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
)

type chartMovingAverage struct {
	color  [3]float32
	line   *gfx.VAO
	bounds image.Rectangle
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

// ProcessInput processes input.
func (ch *chartMovingAverage) ProcessInput(bounds image.Rectangle) {
	ch.bounds = bounds
}

func (ch *chartMovingAverage) Render(fudge float32) {
	if ch.line == nil {
		return
	}
	gfx.SetModelMatrixRect(ch.bounds)
	ch.line.Render()
}

func (ch *chartMovingAverage) Close() {
	if ch.line != nil {
		ch.line.Delete()
		ch.line = nil
	}
}
