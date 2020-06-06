package chart

import (
	"image"
	"math"
	"sort"
	"strconv"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view/color"
	"github.com/btmura/ponzi2/internal/app/view/vao"
)

// priceHorizLine is the horizontal lines rendered behind the candlesticks.
var priceHorizLine = vao.HorizLine(color.Gray)

// price shows the candlesticks and price labels for a single stock.
type price struct {
	// renderable is whether the prices can be rendered.
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

type priceData struct {
	TradingSessionSeries *model.TradingSessionSeries
}

func (p *price) SetData(data priceData) {
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

	p.stickLines, p.stickRects = priceCandlestickVAOs(ts.TradingSessions, p.priceRange)

	p.renderable = true
}

func (p *price) SetBounds(bounds image.Rectangle) {
	p.bounds = bounds
}

func (p *price) Render(float32) {
	if !p.renderable {
		return
	}

	r := p.bounds

	labelPaddingY := p.MaxLabelSize.Y / 2
	y := r.Max.Y - labelPaddingY - p.MaxLabelSize.Y/2
	dy := p.MaxLabelSize.Y + labelPaddingY*2

	for {
		{
			if y < r.Min.Y {
				break
			}

			gfx.SetModelMatrixRect(image.Rect(r.Min.X, y, r.Max.X, y))
			priceHorizLine.Render()
		}
		y -= dy
	}

	gfx.SetModelMatrixRect(r)
	p.stickLines.Render()
	p.stickRects.Render()
}

func (p *price) Close() {
	p.renderable = false
	if p.stickLines != nil {
		p.stickLines.Delete()
	}
	if p.stickRects != nil {
		p.stickRects.Delete()
	}
}

func priceRange(ts []*model.TradingSession) [2]float32 {
	if len(ts) == 0 {
		return [2]float32{0, 0}
	}

	var low float32 = math.MaxFloat32
	var high float32
	ohlc := make([]float32, 4)
	for _, s := range ts {
		if s.Skip() {
			continue
		}

		// Use all values to handle incorrect data.
		ohlc[0] = s.Open
		ohlc[1] = s.High
		ohlc[2] = s.Low
		ohlc[3] = s.Close
		sort.Slice(ohlc, func(i, j int) bool {
			return ohlc[i] < ohlc[2]
		})

		if ohlc[0] < low {
			low = ohlc[0]
		}
		if ohlc[3] > high {
			high = ohlc[3]
		}
	}

	if low > high {
		return [2]float32{0, 0}
	}

	// Pad the high and low, so the candlesticks have space around them.
	padding := (high - low) * .05
	low -= padding
	high += padding

	return [2]float32{low, high}
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

func priceCandlestickVAOs(ds []*model.TradingSession, priceRange [2]float32) (stickLines, stickRects *gfx.VAO) {
	// Calculate vertices and indices for the candlesticks.
	var vertices []float32
	var colors []float32
	var lineIndices []uint16
	var triangleIndices []uint16

	stickWidth := 2.0 / float32(len(ds)) // (-1 to 1) on X-axis
	leftX := -1.0 + stickWidth*0.1
	midX := -1.0 + stickWidth*0.5
	rightX := -1.0 + stickWidth*0.9

	// Move the X coordinates one stick over.
	moveOver := func() {
		leftX += stickWidth
		midX += stickWidth
		rightX += stickWidth
	}

	calcY := func(value float32) float32 {
		return 2*(value-priceRange[0])/(priceRange[1]-priceRange[0]) - 1
	}

	for _, s := range ds {
		if s.Skip() {
			moveOver()
			continue
		}

		// Figure out Y coordinates of the key levels.
		lowY, highY, openY, closeY := calcY(s.Low), calcY(s.High), calcY(s.Open), calcY(s.Close)

		// Figure out the top and bottom of the candlestick.
		topY, botY := openY, closeY
		if openY < closeY {
			topY, botY = closeY, openY
		}

		// Add the vertices needed to create the candlestick.
		idxOffset := len(vertices) / 3
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
		var c color.RGBA
		switch {
		case s.Close > s.Open:
			c = color.Green
		case s.Close < s.Open:
			c = color.Red
		default:
			c = color.White
		}

		colors = append(colors,
			c[0], c[1], c[2], c[3], // 0
			c[0], c[1], c[2], c[3], // 1
			c[0], c[1], c[2], c[3], // 2
			c[0], c[1], c[2], c[3], // 3
			c[0], c[1], c[2], c[3], // 4
			c[0], c[1], c[2], c[3], // 5
			c[0], c[1], c[2], c[3], // 6
			c[0], c[1], c[2], c[3], // 7
		)

		// idx is a function to refer to the vertices above.
		idx := func(j uint16) uint16 {
			return uint16(idxOffset) + j
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

		moveOver()
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
