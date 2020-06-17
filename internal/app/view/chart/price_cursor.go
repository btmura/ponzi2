package chart

import (
	"image"

	"github.com/btmura/ponzi2/internal/app/gfx"
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

	if !p.mousePos.In(p.priceRect) {
		return
	}

	yPercent := float32(p.mousePos.Y-p.priceRect.Min.Y) / float32(p.priceRect.Dy())
	p.renderLabel(fudge, yPercent)
}

func (p *priceCursor) renderLabel(fudge float32, yPercent float32) {
	v := p.priceRange[0] + (p.priceRange[1]-p.priceRange[0])*yPercent
	l := makePriceLabel(v)

	textPt := image.Point{
		X: p.labelRect.Max.X - l.size.X,
		Y: p.labelRect.Min.Y + int(float32(p.labelRect.Dy())*yPercent) - l.size.Y/2,
	}
	bubbleRect := image.Rectangle{
		Min: textPt,
		Max: textPt.Add(l.size),
	}.Inset(-axisLabelPadding)

	// Move the label to the left if the mouse is overlapping.
	if p.mousePos.In(bubbleRect) {
		textPt.X = p.labelRect.Min.X
		bubbleRect = image.Rectangle{
			Min: textPt,
			Max: textPt.Add(l.size),
		}.Inset(-axisLabelPadding)
	}

	axisLabelBubble.SetBounds(bubbleRect)
	axisLabelBubble.Render(fudge)
	axisLabelTextRenderer.Render(l.text, textPt, gfx.TextColor(view.White))
}

func (p *priceCursor) Close() {
	p.renderable = false
}
