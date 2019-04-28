package chart

import (
	"image"

	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view/color"
	"github.com/btmura/ponzi2/internal/app/view/rect"
)

// chartPriceCursor renders crosshairs at the mouse pointer
// with corresponding price labels on the y-axis.
type chartPriceCursor struct {
	// renderable is true if this should be rendered.
	renderable bool

	// priceRange is the inclusive range from min to max price.
	priceRange [2]float32

	// priceRect is the rectangle where the price candleticks are drawn.
	priceRect image.Rectangle

	// labelRect is the rectangle where the axis labels are drawn.
	labelRect image.Rectangle

	// mousePos is the current mouse position.
	mousePos image.Point
}

func (ch *chartPriceCursor) SetData(ts *model.TradingSessionSeries) {
	// Reset everything.
	ch.Close()

	// Bail out if there is no data yet.
	if ts == nil {
		return
	}

	ch.priceRange = priceRange(ts.TradingSessions)

	ch.renderable = true
}

// ProcessInput processes input.
func (ch *chartPriceCursor) ProcessInput(priceRect, labelRect image.Rectangle, mousePos image.Point) {
	ch.priceRect = priceRect
	ch.labelRect = labelRect
	ch.mousePos = mousePos
}

func (ch *chartPriceCursor) Render(fudge float32) {
	if !ch.renderable {
		return
	}

	renderCursorLines(ch.priceRect, ch.mousePos)

	if !ch.mousePos.In(ch.priceRect) {
		return
	}

	perc := float32(ch.mousePos.Y-ch.priceRect.Min.Y) / float32(ch.priceRect.Dy())
	v := ch.priceRange[0] + (ch.priceRange[1]-ch.priceRange[0])*perc
	l := makeChartPriceLabel(v)

	tp := image.Point{
		X: ch.labelRect.Max.X - l.size.X,
		Y: ch.labelRect.Min.Y + int(float32(ch.labelRect.Dy())*perc) - l.size.Y/2,
	}

	rect.RenderBubble(tp, l.size, chartAxisLabelBubbleSpec)
	chartAxisLabelTextRenderer.Render(l.text, tp, color.White)
}

func (ch *chartPriceCursor) Close() {
	ch.renderable = false
}
