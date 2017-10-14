package ponzi

import (
	"image"
	"math"
	"strconv"

	"github.com/btmura/ponzi2/internal/gfx"
)

// ChartPrices shows the candlesticks and price labels for a single stock.
type ChartPrices struct {
	renderable  bool
	minPrice    float32
	maxPrice    float32
	labelHeight int
	stickLines  *gfx.VAO
	stickRects  *gfx.VAO
}

// SetState sets the ChartPrices' state.
func (ch *ChartPrices) SetState(state *ChartState) {
	// Reset everything.
	ch.Close()

	// Bail out if there is no data yet.
	if state.Stock.LastUpdateTime.IsZero() {
		return // Stock has no data yet.
	}

	// Find the min and max price.
	ch.minPrice = math.MaxFloat32
	ch.maxPrice = 0
	for _, s := range state.Stock.DailySessions {
		if ch.minPrice > s.Low {
			ch.minPrice = s.Low
		}
		if ch.maxPrice < s.High {
			ch.maxPrice = s.High
		}
	}

	ch.stickLines, ch.stickRects = createChartCandlestickVAOs(state.Stock.DailySessions, ch.minPrice, ch.maxPrice)

	_, labelSize := priceLabelText(ch.maxPrice)
	ch.labelHeight = labelSize.Y

	ch.renderable = true
}

func createChartCandlestickVAOs(ds []*ModelTradingSession, minPrice, maxPrice float32) (stickLines, stickRects *gfx.VAO) {
	// Calculate vertices and indices for the candlesticks.
	var vertices []float32
	var colors []float32
	var lineIndices []uint16
	var triangleIndices []uint16

	stickWidth := 2.0 / float32(len(ds)) // (-1 to 1) on X-axis
	leftX := -1.0 + stickWidth*0.1
	midX := -1.0 + stickWidth*0.5
	rightX := -1.0 + stickWidth*0.9

	calcY := func(value float32) float32 {
		return 2*(value-minPrice)/(maxPrice-minPrice) - 1
	}

	for i, s := range ds {
		// Figure out Y coordinates of the key levels.
		lowY, highY, openY, closeY := calcY(s.Low), calcY(s.High), calcY(s.Open), calcY(s.Close)

		// Figure out the top and bottom of the candlestick.
		topY, botY := openY, closeY
		if openY < closeY {
			topY, botY = closeY, openY
		}

		// Add the vertices needed to create the candlestick.
		vertices = append(vertices,
			midX, highY, 0, // 0
			midX, topY, 0, // 1
			midX, lowY, 0, // 2
			midX, botY, 0, // 3
			leftX, topY, 0, // 4 - Upper left of box
			rightX, topY, 0, // 5 - Upper right of box
			leftX, botY, 0, // 6 - Bottom left of box
			rightX, botY, 0, // 7 - Bottom right of box
		)

		// Add the colors corresponding to the vertices.
		var c [3]float32
		switch {
		case s.Close > s.Open:
			c = green
		case s.Close < s.Open:
			c = red
		default:
			c = yellow
		}

		colors = append(colors,
			c[0], c[1], c[2],
			c[0], c[1], c[2],
			c[0], c[1], c[2],
			c[0], c[1], c[2],
			c[0], c[1], c[2],
			c[0], c[1], c[2],
			c[0], c[1], c[2],
			c[0], c[1], c[2],
		)

		// idx is function to refer to the vertices above.
		idx := func(j uint16) uint16 {
			return uint16(i)*8 + j
		}

		// Add the vertex indices to render the candlestick.
		lineIndices = append(lineIndices,
			// Top and bottom lines around the box.
			idx(0), idx(1),
			idx(2), idx(3),
		)

		if s.Close > s.Open {
			// Use lines for open candlestick on higher closes.
			lineIndices = append(lineIndices,
				idx(4), idx(5),
				idx(6), idx(7),
				idx(4), idx(6),
				idx(5), idx(7),
			)
		} else {
			// Use triangles for filled candlestick on lower closes.
			triangleIndices = append(triangleIndices,
				idx(4), idx(6), idx(5),
				idx(5), idx(6), idx(7),
			)
		}

		// Move the X coordinates one stick over.
		leftX += stickWidth
		midX += stickWidth
		rightX += stickWidth
	}

	lineVAO := gfx.NewVAO(
		gfx.Lines,
		&gfx.VAOVertexData{
			Vertices: vertices,
			Colors:   colors,
			Indices:  lineIndices,
		},
	)

	triangleVAO := gfx.NewVAO(
		gfx.Triangles,
		&gfx.VAOVertexData{
			Vertices: vertices,
			Colors:   colors,
			Indices:  triangleIndices,
		},
	)

	return lineVAO, triangleVAO
}

// Render renders the price candlesticks.
func (ch *ChartPrices) Render(r image.Rectangle) {
	if !ch.renderable {
		return
	}

	gfx.SetModelMatrixRect(r)
	ch.stickLines.Render()
	ch.stickRects.Render()

	labelPaddingY := ch.labelHeight / 2
	y := r.Max.Y - labelPaddingY - ch.labelHeight/2
	dy := ch.labelHeight + labelPaddingY*2

	for {
		{
			if y < r.Min.Y {
				break
			}

			gfx.SetModelMatrixRect(image.Rect(r.Min.X, y, r.Max.X, y))
			chartGridHorizLine.Render()
		}
		y -= dy
	}
}

// RenderLabels renders the price labels.
func (ch *ChartPrices) RenderLabels(r image.Rectangle) (maxLabelWidth int) {
	if !ch.renderable {
		return
	}

	labelPaddingY := ch.labelHeight / 2
	pricePerPixel := (ch.maxPrice - ch.minPrice) / float32(r.Dy())

	// Start at top and decrement one label with top and bottom padding.
	c := r.Max
	dc := image.Pt(0, labelPaddingY+ch.labelHeight+labelPaddingY)

	// Start at top with max price and decrement change in price of a label with padding.
	v := ch.maxPrice
	dv := pricePerPixel * float32(dc.Y)

	// Offets to the cursor and price value when drawing.
	dcy := labelPaddingY + ch.labelHeight   // Puts cursor at the baseline of the text.
	dvy := labelPaddingY + ch.labelHeight/2 // Uses value in the middle of the label.

	maxWidth := 0

	for {
		{
			c := image.Pt(c.X, c.Y-dcy)
			if c.Y < r.Min.Y {
				break
			}

			v := v - pricePerPixel*float32(dvy)
			t, s := priceLabelText(v)
			c.X -= s.X
			chartAxisLabelTextRenderer.Render(t, c, white)

			if maxWidth < s.X {
				maxWidth = s.X
			}
		}

		c = c.Sub(dc)
		v -= dv
	}

	return maxWidth
}

func priceLabelText(v float32) (text string, size image.Point) {
	t := strconv.FormatFloat(float64(v), 'f', 2, 32)
	return t, chartAxisLabelTextRenderer.Measure(t)
}

// Close frees the resources backing the ChartPrices.
func (ch *ChartPrices) Close() {
	ch.renderable = false
	ch.minPrice = 0
	ch.maxPrice = 0
	ch.labelHeight = 0
	if ch.stickLines != nil {
		ch.stickLines.Delete()
		ch.stickLines = nil
	}
	if ch.stickRects != nil {
		ch.stickRects.Delete()
		ch.stickRects = nil
	}
}
