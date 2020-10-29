package chart

import (
	"image"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view"
	"github.com/btmura/ponzi2/internal/app/view/vao"
)

type movingAverage struct {
	renderable bool
	color      view.Color
	line       *gfx.VAO
	bounds     image.Rectangle
}

func newMovingAverage(color view.Color) *movingAverage {
	return &movingAverage{color: color}
}

type movingAverageData struct {
	TradingSessionSeries *model.TradingSessionSeries
	MovingAverageSeries  *model.MovingAverageSeries
}

func (m *movingAverage) SetData(data movingAverageData) {
	// Reset everything.
	m.Close()

	// Bail out if there is not enough data yet.
	ts := data.TradingSessionSeries
	if ts == nil {
		return
	}

	ms := data.MovingAverageSeries
	if ms == nil {
		return
	}

	yRange := priceRange(ts.TradingSessions)

	m.line = movingAverageDataLine(ms.Values, yRange, m.color)

	m.renderable = true
}

func (m *movingAverage) SetBounds(bounds image.Rectangle) {
	m.bounds = bounds
}

func (m *movingAverage) Render(float32) {
	if m.line == nil {
		return
	}
	gfx.SetModelMatrixRect(m.bounds)
	m.line.Render()
}

func (m *movingAverage) Close() {
	m.renderable = false
	if m.line != nil {
		m.line.Delete()
	}
}

func movingAverageDataLine(ms []*model.MovingAverageValue, yRange [2]float32, color view.Color) *gfx.VAO {
	var yPercentValues []float32
	for _, m := range ms {
		yPercentValues = append(yPercentValues, pricePercent(yRange, m.Value))
	}
	return vao.DataLine(yPercentValues, color)
}
