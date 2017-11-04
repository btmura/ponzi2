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
func (ch *ChartCursor) Render(vc ViewContext) {
	if vc.MousePos.In(vc.Bounds) {
		gfx.SetModelMatrixRect(image.Rect(vc.Bounds.Min.X, vc.MousePos.Y, vc.Bounds.Max.X, vc.MousePos.Y))
		cursorHorizLine.Render()
	}

	if vc.MousePos.X >= vc.Bounds.Min.X && vc.MousePos.X <= vc.Bounds.Max.X {
		gfx.SetModelMatrixRect(image.Rect(vc.MousePos.X, vc.Bounds.Min.Y, vc.MousePos.X, vc.Bounds.Max.Y))
		cursorVertLine.Render()
	}
}
