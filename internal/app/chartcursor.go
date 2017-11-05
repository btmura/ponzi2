package app

import (
	"image"

	"github.com/btmura/ponzi2/internal/gfx"
)

var (
	cursorHorizLine = gfx.HorizColoredLineVAO(lightGray, lightGray)
	cursorVertLine  = gfx.VertColoredLineVAO(lightGray, lightGray)
)

// ChartCursor renders crosshairs on the chart.
type ChartCursor struct{}

// NewChartCursor creates a new ChartCursor.
func NewChartCursor() *ChartCursor {
	return &ChartCursor{}
}

// Render renders the ChartCursor.
func (ch *ChartCursor) Render(r image.Rectangle, mousePos image.Point) {
	if mousePos.In(r) {
		gfx.SetModelMatrixRect(image.Rect(r.Min.X, mousePos.Y, r.Max.X, mousePos.Y))
		cursorHorizLine.Render()
	}

	if mousePos.X >= r.Min.X && mousePos.X <= r.Max.X {
		gfx.SetModelMatrixRect(image.Rect(mousePos.X, r.Min.Y, mousePos.X, r.Max.Y))
		cursorVertLine.Render()
	}
}
