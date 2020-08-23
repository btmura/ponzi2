package chart

import (
	"image"
	"strconv"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view"
	"github.com/btmura/ponzi2/internal/app/view/vao"
)

// priceHorizLine is the horizontal lines rendered behind the candlesticks.
var priceHorizLine = vao.HorizLine(view.TransparentGray, view.Gray)

// priceLevel renders the horizontal price lines and their labels.
type priceLevel struct {
	// renderable is true if this should be rendered.
	renderable bool

	// priceRange represents the inclusive range from min to max price.
	priceRange [2]float32

	// MaxLabelSize is the maximum label size useful for rendering measurements.
	MaxLabelSize image.Point

	// lineBounds is the rectangle where the horizontal lines should be drawn within.
	lineBounds image.Rectangle

	// labelBounds is the rectangle where the labels for the lines should be drawn within.
	labelBounds image.Rectangle
}

func newPriceLevel() *priceLevel {
	return new(priceLevel)
}

type priceLevelData struct {
	TradingSessionSeries *model.TradingSessionSeries
}

func (p *priceLevel) SetData(data priceLevelData) {
	// Reset everything.
	p.Close()

	// Bail out if there is no data yet.
	ts := data.TradingSessionSeries
	if ts == nil {
		return
	}

	p.priceRange = priceRange(ts.TradingSessions)

	// Measure the max label size by creating a label with the max value.
	p.MaxLabelSize = makePriceLabel(p.priceRange[1]).size

	p.renderable = true
}

func (p *priceLevel) SetBounds(lineBounds, labelBounds image.Rectangle) {
	p.lineBounds = lineBounds
	p.labelBounds = labelBounds
}

func (p *priceLevel) Render(fudge float32) {
	if !p.renderable {
		return
	}

	r := p.lineBounds
	for _, y := range p.labelYPositions(r) {
		gfx.SetModelMatrixRect(image.Rect(r.Min.X, y, r.Max.X, y))
		priceHorizLine.Render()
	}

	r = p.labelBounds
	for _, y := range p.labelYPositions(r) {
		renderLabel(fudge, p.priceRange, r, image.Pt(0, y), false)
	}
}

func (p *priceLevel) labelYPositions(r image.Rectangle) []int {
	labelPaddingY := p.MaxLabelSize.Y / 2
	firstY := r.Max.Y - labelPaddingY - p.MaxLabelSize.Y/2
	dy := p.MaxLabelSize.Y * 2

	var yPositions []int
	for y := firstY; y >= r.Min.Y; y -= dy {
		yPositions = append(yPositions, y)
	}
	return yPositions
}

func (p *priceLevel) Close() {
	p.renderable = false
}

type priceLabel struct {
	text string
	size image.Point
}

func makePriceLabel(v float32) priceLabel {
	t := strconv.FormatFloat(float64(v), 'f', 2, 32)
	return priceLabel{
		text: t,
		size: axisLabelTextRenderer.Measure(t),
	}
}

func renderLabel(fudge float32, priceRange [2]float32, r image.Rectangle, pt image.Point, includeBubble bool) {
	yPercent := float32(pt.Y-r.Min.Y) / float32(r.Dy())
	value := priceValue(priceRange, yPercent)
	label := makePriceLabel(value)

	textPt := image.Point{
		X: r.Max.X - label.size.X,
		Y: r.Min.Y + int(float32(r.Dy())*yPercent) - label.size.Y/2,
	}
	bubbleRect := image.Rectangle{
		Min: textPt,
		Max: textPt.Add(label.size),
	}.Inset(-axisLabelPadding)

	// Move the label to the left if the mouse is overlapping.
	if pt.In(bubbleRect) {
		textPt.X = r.Min.X
		bubbleRect = image.Rectangle{
			Min: textPt,
			Max: textPt.Add(label.size),
		}.Inset(-axisLabelPadding)
	}

	if includeBubble {
		axisLabelBubble.SetBounds(bubbleRect)
		axisLabelBubble.Render(fudge)
	}
	axisLabelTextRenderer.Render(label.text, textPt, gfx.TextColor(view.White))
}
