package view

import (
	"image"

	"github.com/btmura/ponzi2/internal/app/model"
)

type chartPriceCursorLabels struct {
	// priceRange represents the inclusive range from min to max price.
	priceRange [2]float32

	// bounds is the rectangle with global coords that should be drawn within.
	bounds image.Rectangle

	labelRect image.Rectangle

	mousePos image.Point
}

func newChartPriceCursorLabels() *chartPriceCursorLabels {
	return &chartPriceCursorLabels{}
}

func (ch *chartPriceCursorLabels) SetData(ts *model.TradingSessionSeries) {
	// Bail out if there is no data yet.
	if ts == nil {
		return
	}

	ch.priceRange = priceRange(ts.TradingSessions)
}

// ProcessInput processes input.
func (ch *chartPriceCursorLabels) ProcessInput(ic inputContext, labelRect image.Rectangle) {
	ch.bounds = ic.Bounds
	ch.labelRect = labelRect
	ch.mousePos = ic.MousePos
}

func (ch *chartPriceCursorLabels) Render(fudge float32) {
	if !ch.mousePos.In(ch.bounds) {
		return
	}

	perc := float32(ch.mousePos.Y-ch.bounds.Min.Y) / float32(ch.bounds.Dy())
	v := ch.priceRange[0] + (ch.priceRange[1]-ch.priceRange[0])*perc
	l := makeChartPriceLabel(v)

	tp := image.Point{
		X: ch.labelRect.Max.X - l.size.X,
		Y: ch.labelRect.Min.Y + int(float32(ch.labelRect.Dy())*perc) - l.size.Y/2,
	}

	renderBubble(tp, l.size, chartAxisLabelBubbleSpec)
	chartAxisLabelTextRenderer.Render(l.text, tp, white)
}
