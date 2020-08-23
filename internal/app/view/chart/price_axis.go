package chart

import (
	"image"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view"
)

type priceAxis struct {
	// renderable is true if this should be rendered.
	renderable bool

	// priceRange represents the inclusive range from min to max price.
	priceRange [2]float32

	// MaxLabelSize is the maximum label size useful for rendering measurements.
	MaxLabelSize image.Point

	// labelRect is the rectangle with global coords that should be drawn within.
	labelRect image.Rectangle
}

type priceAxisData struct {
	TradingSessionSeries *model.TradingSessionSeries
}

func (p *priceAxis) SetData(data priceAxisData) {
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

func (p *priceAxis) SetBounds(labelRect image.Rectangle) {
	p.labelRect = labelRect
}

func (p *priceAxis) Render(float32) {
	if !p.renderable {
		return
	}

	r := p.labelRect
	labelPaddingY := p.MaxLabelSize.Y / 2
	firstY := r.Max.Y - labelPaddingY - p.MaxLabelSize.Y/2
	dy := p.MaxLabelSize.Y * 2

	for y := firstY; y >= r.Min.Y; y -= dy {
		yPercent := float32(y-r.Min.Y) / float32(r.Dy())
		value := priceValue(p.priceRange, yPercent)
		label := makePriceLabel(value)

		textPt := image.Point{
			X: r.Max.X - label.size.X,
			Y: y - p.MaxLabelSize.Y/2,
		}
		axisLabelTextRenderer.Render(label.text, textPt, gfx.TextColor(view.White))
	}
}

func (p *priceAxis) Close() {
	p.renderable = false
}
