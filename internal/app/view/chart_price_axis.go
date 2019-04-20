package view

import (
	"image"

	"github.com/btmura/ponzi2/internal/app/model"
)

type chartPriceAxis struct {
	// renderable is true if this should be rendered.
	renderable bool

	// priceRange represents the inclusive range from min to max price.
	priceRange [2]float32

	// MaxLabelSize is the maximum label size useful for rendering measurements.
	MaxLabelSize image.Point

	// labelRect is the rectangle with global coords that should be drawn within.
	labelRect image.Rectangle
}

func (ch *chartPriceAxis) SetData(ts *model.TradingSessionSeries) {
	// Reset everything.
	ch.Close()

	// Bail out if there is no data yet.
	if ts == nil {
		return
	}

	ch.priceRange = priceRange(ts.TradingSessions)

	// Measure the max label size by creating a label with the max value.
	ch.MaxLabelSize = makeChartPriceLabel(ch.priceRange[1]).size

	ch.renderable = true
}

// ProcessInput processes input.
func (ch *chartPriceAxis) ProcessInput(labelRect image.Rectangle) {
	ch.labelRect = labelRect
}

func (ch *chartPriceAxis) Render(fudge float32) {
	if !ch.renderable {
		return
	}

	r := ch.labelRect

	labelPaddingY := ch.MaxLabelSize.Y / 2
	pricePerPixel := (ch.priceRange[1] - ch.priceRange[0]) / float32(r.Dy())

	// Start at top and decrement one label with top and bottom padding.
	pt := r.Max
	dp := image.Pt(0, labelPaddingY+ch.MaxLabelSize.Y+labelPaddingY)

	// Start at top with max price and decrement change in price of a label with padding.
	v := ch.priceRange[1]
	dv := pricePerPixel * float32(dp.Y)

	// Offets to the cursor and price value when drawing.
	dpy := labelPaddingY + ch.MaxLabelSize.Y   // Puts point at the baseline of the text.
	dvy := labelPaddingY + ch.MaxLabelSize.Y/2 // Uses value in the middle of the label.

	for {
		{
			l := makeChartPriceLabel(v - pricePerPixel*float32(dvy))

			pt := image.Pt(pt.X-l.size.X, pt.Y-dpy)
			if pt.Y < r.Min.Y {
				break
			}

			chartAxisLabelTextRenderer.Render(l.text, pt, white)
		}

		pt = pt.Sub(dp)
		v -= dv
	}
}

func (ch *chartPriceAxis) Close() {
	ch.renderable = false
}
