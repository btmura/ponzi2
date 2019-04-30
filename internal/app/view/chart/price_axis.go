package chart

import (
	"image"

	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view/color"
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

func (p *priceAxis) SetData(ts *model.TradingSessionSeries) {
	// Reset everything.
	p.Close()

	// Bail out if there is no data yet.
	if ts == nil {
		return
	}

	p.priceRange = priceRange(ts.TradingSessions)

	// Measure the max label size by creating a label with the max value.
	p.MaxLabelSize = makePriceLabel(p.priceRange[1]).size

	p.renderable = true
}

// ProcessInput processes input.
func (p *priceAxis) ProcessInput(labelRect image.Rectangle) {
	p.labelRect = labelRect
}

func (p *priceAxis) Render(fudge float32) {
	if !p.renderable {
		return
	}

	r := p.labelRect

	labelPaddingY := p.MaxLabelSize.Y / 2
	pricePerPixel := (p.priceRange[1] - p.priceRange[0]) / float32(r.Dy())

	// Start at top and decrement one label with top and bottom padding.
	pt := r.Max
	dp := image.Pt(0, labelPaddingY+p.MaxLabelSize.Y+labelPaddingY)

	// Start at top with max price and decrement change in price of a label with padding.
	v := p.priceRange[1]
	dv := pricePerPixel * float32(dp.Y)

	// Offets to the cursor and price value when drawing.
	dpy := labelPaddingY + p.MaxLabelSize.Y   // Puts point at the baseline of the text.
	dvy := labelPaddingY + p.MaxLabelSize.Y/2 // Uses value in the middle of the label.

	for {
		{
			l := makePriceLabel(v - pricePerPixel*float32(dvy))

			pt := image.Pt(pt.X-l.size.X, pt.Y-dpy)
			if pt.Y < r.Min.Y {
				break
			}

			chartAxisLabelTextRenderer.Render(l.text, pt, color.White)
		}

		pt = pt.Sub(dp)
		v -= dv
	}
}

func (p *priceAxis) Close() {
	p.renderable = false
}
