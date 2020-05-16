package chart

import (
	"image"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view/color"
	"github.com/btmura/ponzi2/internal/app/view/vao"
)

type movingAverage struct {
	renderable bool
	color      color.RGBA
	line       *gfx.VAO
	bounds     image.Rectangle
}

func newMovingAverage(color color.RGBA) *movingAverage {
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

	// Create the graph line VAO.
	var values []float32
	for _, m := range ms.MovingAverages {
		values = append(values, m.Value)
	}
	m.line = vao.DataLine(values, priceRange(ts.TradingSessions), m.color)

	m.renderable = true
}

func (m *movingAverage) SetBounds(bounds image.Rectangle) {
	m.bounds = bounds
}

func (m *movingAverage) Render(fudge float32) {
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
