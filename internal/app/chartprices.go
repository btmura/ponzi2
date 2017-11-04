package app

import (
	"image"
	"math"
	"strconv"

	"github.com/btmura/ponzi2/internal/gfx"
)

// ChartPrices shows the candlesticks and price labels for a single stock.
type ChartPrices struct {
	renderable   bool
	priceRange   [2]float32
	maxLabelSize image.Point
	stickLines   *gfx.VAO
	stickRects   *gfx.VAO
}

// NewChartPrices creates a new ChartPrices.
func NewChartPrices() *ChartPrices {
	return &ChartPrices{}
}

// SetStock sets the ChartPrices' stock.
func (ch *ChartPrices) SetStock(st *ModelStock) {
	// Reset everything.
	ch.Close()

	// Bail out if there is no data yet.
	if st.LastUpdateTime.IsZero() {
		return // Stock has no data yet.
	}

	// Find the min and max price.
	ch.priceRange = [2]float32{math.MaxFloat32, 0}
	for _, s := range st.DailySessions {
		if ch.priceRange[0] > s.Low {
			ch.priceRange[0] = s.Low
		}
		if ch.priceRange[1] < s.High {
			ch.priceRange[1] = s.High
		}
	}

	// Create Y-axis label for maximum price for rendering measurements.
	ch.maxLabelSize = makeChartPriceLabel(ch.priceRange[1]).size

	ch.stickLines, ch.stickRects = chartPriceCandlestickVAOs(st.DailySessions, ch.priceRange)

	ch.renderable = true
}

// Render renders the price candlesticks.
func (ch *ChartPrices) Render(r image.Rectangle) {
	if !ch.renderable {
		return
	}

	labelPaddingY := ch.maxLabelSize.Y / 2
	y := r.Max.Y - labelPaddingY - ch.maxLabelSize.Y/2
	dy := ch.maxLabelSize.Y + labelPaddingY*2

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

	gfx.SetModelMatrixRect(r)
	ch.stickLines.Render()
	ch.stickRects.Render()
}

// RenderLabels renders the price labels.
func (ch *ChartPrices) RenderLabels(r image.Rectangle, mousePos image.Point) (maxLabelWidth int) {
	if !ch.renderable {
		return
	}

	labelPaddingY := ch.maxLabelSize.Y / 2
	pricePerPixel := (ch.priceRange[1] - ch.priceRange[0]) / float32(r.Dy())

	// Start at top and decrement one label with top and bottom padding.
	pt := r.Max
	dp := image.Pt(0, labelPaddingY+ch.maxLabelSize.Y+labelPaddingY)

	// Start at top with max price and decrement change in price of a label with padding.
	v := ch.priceRange[1]
	dv := pricePerPixel * float32(dp.Y)

	// Offets to the cursor and price value when drawing.
	dpy := labelPaddingY + ch.maxLabelSize.Y   // Puts point at the baseline of the text.
	dvy := labelPaddingY + ch.maxLabelSize.Y/2 // Uses value in the middle of the label.

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

	if mousePos.In(r) {
		perc := float32(mousePos.Y-r.Min.Y) / float32(r.Dy())
		v = ch.priceRange[0] + (ch.priceRange[1]-ch.priceRange[0])*perc

		l := makeChartPriceLabel(v)
		x := r.Max.X - l.size.X
		y := r.Min.Y + int(float32(r.Dy())*perc) - l.size.Y/2
		chartAxisLabelTextRenderer.Render(l.text, image.Pt(x, y), white)

		r := image.Rect(x, y, r.Max.X, y+l.size.Y).Inset(-chartAxisLabelBubblePadding)
		renderRoundedRect(r, chartAxisLabelBubbleRounding)
	}

	return ch.maxLabelSize.X
}

// Close frees the resources backing the ChartPrices.
func (ch *ChartPrices) Close() {
	ch.renderable = false
	if ch.stickLines != nil {
		ch.stickLines.Delete()
	}
	if ch.stickRects != nil {
		ch.stickRects.Delete()
	}
}

type chartPriceLabel struct {
	text string
	size image.Point
}

func makeChartPriceLabel(v float32) chartPriceLabel {
	t := strconv.FormatFloat(float64(v), 'f', 2, 32)
	return chartPriceLabel{
		text: t,
		size: chartAxisLabelTextRenderer.Measure(t),
	}
}

func chartPriceCandlestickVAOs(ds []*ModelTradingSession, priceRange [2]float32) (stickLines, stickRects *gfx.VAO) {
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
		return 2*(value-priceRange[0])/(priceRange[1]-priceRange[0]) - 1
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
