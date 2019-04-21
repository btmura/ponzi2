package view

import (
	"image"
	"math"
	"strconv"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
)

// chartPrices shows the candlesticks and price labels for a single stock.
type chartPrices struct {
	// renderable is whether the ChartPrices can be rendered.
	renderable bool

	// priceRange represents the inclusive range from min to max price.
	priceRange [2]float32

	// MaxLabelSize is the maximum label size useful for rendering measurements.
	MaxLabelSize image.Point

	// stickLines is the VAO with the vertical candlestick lines.
	stickLines *gfx.VAO

	// stickRects ithe VAO with the candlestick boxes without the lines.
	stickRects *gfx.VAO

	// bounds is the rectangle with global coords that should be drawn within.
	bounds image.Rectangle
}

func (ch *chartPrices) SetData(ts *model.TradingSessionSeries) {
	// Reset everything.
	ch.Close()

	// Bail out if there is no data yet.
	if ts == nil {
		return
	}

	ch.priceRange = priceRange(ts.TradingSessions)

	// Measure the max label size by creating a label with the max value.
	ch.MaxLabelSize = makeChartPriceLabel(ch.priceRange[1]).size

	ch.stickLines, ch.stickRects = chartPriceCandlestickVAOs(ts.TradingSessions, ch.priceRange)

	ch.renderable = true
}

// ProcessInput processes input.
func (ch *chartPrices) ProcessInput(ic inputContext) {
	ch.bounds = ic.Bounds
}

func (ch *chartPrices) Render(fudge float32) {
	if !ch.renderable {
		return
	}

	r := ch.bounds

	labelPaddingY := ch.MaxLabelSize.Y / 2
	y := r.Max.Y - labelPaddingY - ch.MaxLabelSize.Y/2
	dy := ch.MaxLabelSize.Y + labelPaddingY*2

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

func (ch *chartPrices) Close() {
	ch.renderable = false
	if ch.stickLines != nil {
		ch.stickLines.Delete()
	}
	if ch.stickRects != nil {
		ch.stickRects.Delete()
	}
}

func priceRange(ts []*model.TradingSession) [2]float32 {
	if len(ts) == 0 {
		return [2]float32{0, 0}
	}

	var low float32 = math.MaxFloat32
	var high float32
	for _, s := range ts {
		if s.Empty() {
			continue
		}
		if s.Low < low {
			low = s.Low
		}
		if s.High > high {
			high = s.High
		}
	}

	if low > high {
		return [2]float32{0, 0}
	}

	padding := (high - low) * chartAxisVerticalPaddingPercent
	low -= padding
	high += padding

	return [2]float32{low, high}
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

func chartPriceCandlestickVAOs(ds []*model.TradingSession, priceRange [2]float32) (stickLines, stickRects *gfx.VAO) {
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
		&gfx.VAOVertexData{
			Mode:     gfx.Lines,
			Vertices: vertices,
			Colors:   colors,
			Indices:  lineIndices,
		},
	)

	triangleVAO := gfx.NewVAO(
		&gfx.VAOVertexData{
			Mode:     gfx.Triangles,
			Vertices: vertices,
			Colors:   colors,
			Indices:  triangleIndices,
		},
	)

	return lineVAO, triangleVAO
}
