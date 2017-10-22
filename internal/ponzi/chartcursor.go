package ponzi

import (
	"image"

	"github.com/btmura/ponzi2/internal/gfx"
)

// ChartCursor renders crosshairs on the chart.
type ChartCursor struct {
}

// NewChartCursor creates a new ChartCursor.
func NewChartCursor() *ChartCursor {
	return &ChartCursor{}
}

// RenderHorizLine renders the ChartCursor's horizontal line.
func (ch *ChartCursor) RenderHorizLine(vc ViewContext) {
	if !vc.MousePos.In(vc.Bounds) {
		return
	}
	gfx.SetModelMatrixRect(image.Rect(vc.Bounds.Min.X, vc.MousePos.Y, vc.Bounds.Max.X, vc.MousePos.Y))
	horizLine.Render()
}
