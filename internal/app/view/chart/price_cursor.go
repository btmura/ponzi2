package chart

import (
	"image"

	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view"
)

// priceCursor renders crosshairs at the mouse pointer
// with corresponding price labels on the y-axis.
type priceCursor struct {
	// data is the data necessary to render
	data priceCursorData

	// renderable is true if this should be rendered.
	renderable bool

	// priceRange is the inclusive range from min to max price.
	priceRange [2]float32

	// priceRect is the rectangle where the price candlesticks are drawn.
	priceRect image.Rectangle

	// labelRect is the rectangle where the axis labels are drawn.
	labelRect image.Rectangle

	// mousePos is the current mouse position. Nil for no mouse input.
	mousePos *view.MousePosition
}

type priceCursorData struct {
	TradingSessionSeries *model.TradingSessionSeries
}

func (p *priceCursor) SetData(data priceCursorData) {
	p.data = data

	// Reset everything.
	p.Close()

	// Bail out if there is no data yet.
	ts := data.TradingSessionSeries
	if ts == nil {
		return
	}

	p.priceRange = priceRange(ts.TradingSessions)

	p.renderable = true
}

func (p *priceCursor) SetBounds(priceRect, labelRect image.Rectangle) {
	p.priceRect = priceRect
	p.labelRect = labelRect
}

func (p *priceCursor) ProcessInput(input *view.Input) {
	p.mousePos = input.MousePos
}

func (p *priceCursor) Render(fudge float32) {
	if !p.renderable {
		return
	}

	if p.mousePos == nil {
		return
	}

	renderCursorLines(p.priceRect, p.mousePos)

	if p.mousePos.In(p.priceRect) {
		renderPriceLabel(fudge, p.priceRange, p.labelRect, p.mousePos.Point, true)
	}
}

func (p *priceCursor) Close() {
	p.renderable = false
}
