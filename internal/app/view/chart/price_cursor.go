package chart

import (
	"image"

	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view/color"
	"github.com/btmura/ponzi2/internal/app/view/rect"
)

// priceCursor renders crosshairs at the mouse pointer
// with corresponding price labels on the y-axis.
type priceCursor struct {
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

func (p *priceCursor) SetData(ts *model.TradingSessionSeries) {
	// Reset everything.
	p.Close()

	// Bail out if there is no data yet.
	if ts == nil {
		return
	}

	p.priceRange = priceRange(ts.TradingSessions)

	p.renderable = true
}

// ProcessInput processes input.
func (p *priceCursor) ProcessInput(priceRect, labelRect image.Rectangle, mousePos image.Point) {
	p.priceRect = priceRect
	p.labelRect = labelRect
	p.mousePos = mousePos
}

func (p *priceCursor) Render(fudge float32) {
	if !p.renderable {
		return
	}

	renderCursorLines(p.priceRect, p.mousePos)

	if !p.mousePos.In(p.priceRect) {
		return
	}

	perc := float32(p.mousePos.Y-p.priceRect.Min.Y) / float32(p.priceRect.Dy())
	v := p.priceRange[0] + (p.priceRange[1]-p.priceRange[0])*perc
	l := makePriceLabel(v)

	tp := image.Point{
		X: p.labelRect.Max.X - l.size.X,
		Y: p.labelRect.Min.Y + int(float32(p.labelRect.Dy())*perc) - l.size.Y/2,
	}

	rect.RenderBubble(tp, l.size, axisLabelBubbleSpec)
	axisLabelTextRenderer.Render(l.text, tp, color.White)
}

func (p *priceCursor) Close() {
	p.renderable = false
}