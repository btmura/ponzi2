package chart

import (
	"image"
	"math"
	"sort"
	"strconv"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view"
	"github.com/btmura/ponzi2/internal/app/view/vao"
	"github.com/btmura/ponzi2/internal/logger"
)

// priceHorizLine is the horizontal lines rendered behind the candlesticks.
var priceHorizLine = vao.HorizLine(view.TransparentGray, view.Gray)

// price shows the candlesticks and price labels for a single stock.
type price struct {
	// renderable is whether the prices can be rendered.
	renderable bool

	// priceRange represents the inclusive range from min to max price.
	priceRange [2]float32

	// MaxLabelSize is the maximum label size useful for rendering measurements.
	MaxLabelSize image.Point

	// style is the chart style whether bars or candlesticks.
	style Style

	// faders has the faders needed to fade in and out the bars and candlesticks.
	faders map[Style]*view.Fader

	// barLines is the VAO with the price bar lines.
	barLines *gfx.VAO

	// stickLines is the VAO with the vertical candlestick lines.
	stickLines *gfx.VAO

	// stickRects ithe VAO with the candlestick boxes without the lines.
	stickRects *gfx.VAO

	// bounds is the rectangle with global coords that should be drawn within.
	bounds image.Rectangle
}

func newPrice() *price {
	return &price{
		style: StyleBar,
		faders: map[Style]*view.Fader{
			StyleBar:         view.NewStartedFader(1.5 * view.FPS),
			StyleCandlestick: view.NewStoppedFader(1.5 * view.FPS),
		},
	}
}

// SetStyle sets the style whether bars or candlesticks.
func (p *price) SetStyle(newStyle Style) {
	if newStyle == StyleUnspecified {
		logger.Error("unspecified style")
		return
	}

	if newStyle == p.style {
		return
	}

	p.style = newStyle

	for style, fader := range p.faders {
		if style == p.style {
			fader.FadeIn()
		} else {
			fader.FadeOut()
		}
	}
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

	p.barLines = priceBarVAO(ts.TradingSessions, p.priceRange)

	p.stickLines, p.stickRects = priceCandlestickVAOs(ts.TradingSessions, p.priceRange)

	p.renderable = true
}

func (p *price) SetBounds(bounds image.Rectangle) {
	p.bounds = bounds
}

func (p *price) Update() (dirty bool) {
	for _, fader := range p.faders {
		if fader.Update() {
			dirty = true
		}
	}
	return dirty
}

func (p *price) Render(fudge float32) {
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

	switch p.style {
	case StyleBar:
		p.faders[StyleBar].Render(fudge, func() {
			p.barLines.Render()
		})

	case StyleCandlestick:
		p.faders[StyleCandlestick].Render(fudge, func() {
			p.stickLines.Render()
			p.stickRects.Render()
		})
	}
}

func (p *price) Close() {
	p.renderable = false
	if p.barLines != nil {
		p.barLines.Delete()
	}
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
			return ohlc[i] < ohlc[j]
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

func priceBarVAO(ts []*model.TradingSession, priceRange [2]float32) *gfx.VAO {
	var vertices []float32
	var colors []float32
	var lineIndices []uint16

	stickWidth := 2.0 / float32(len(ts)) // (-1 to 1) on X-axis
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

	for _, s := range ts {
		if s.Skip() {
			moveOver()
			continue
		}

		// Figure out Y coordinates of the key levels.
		lowY, highY, openY, closeY := calcY(s.Low), calcY(s.High), calcY(s.Open), calcY(s.Close)

		// Add the vertices needed to create the candlestick.
		idxOffset := len(vertices) / 3
		vertices = append(vertices,
			midX, highY, 0, // 0
			midX, lowY, 0, // 1
			leftX, openY, 0, // 2
			midX, openY, 0, // 3
			rightX, closeY, 0, // 4
			midX, closeY, 0, // 5
		)

		// Add the colors corresponding to the vertices.
		var c view.Color
		switch {
		case s.Source == model.RealTimePrice:
			c = view.Yellow
		case s.Change > 0:
			c = view.Green
		case s.Change < 0:
			c = view.Red
		default:
			c = view.White
		}

		colors = append(colors,
			c[0], c[1], c[2], c[3], // 0
			c[0], c[1], c[2], c[3], // 1
			c[0], c[1], c[2], c[3], // 2
			c[0], c[1], c[2], c[3], // 3
			c[0], c[1], c[2], c[3], // 4
			c[0], c[1], c[2], c[3], // 5
		)

		// idx is a function to refer to the vertices above.
		idx := func(j uint16) uint16 {
			return uint16(idxOffset) + j
		}

		// Add the vertex indices to render the candlestick.
		lineIndices = append(lineIndices,
			idx(0), idx(1),
			idx(2), idx(3),
			idx(4), idx(5),
		)

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

	return lineVAO
}

func priceCandlestickVAOs(ts []*model.TradingSession, priceRange [2]float32) (stickLines, stickRects *gfx.VAO) {
	var vertices []float32
	var colors []float32
	var lineIndices []uint16
	var triangleIndices []uint16

	stickWidth := 2.0 / float32(len(ts)) // (-1 to 1) on X-axis
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

	for _, s := range ts {
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
		var c view.Color
		switch {
		case s.Source == model.RealTimePrice:
			c = view.Yellow
		case s.Close > s.Open:
			c = view.Green
		case s.Close < s.Open:
			c = view.Red
		default:
			c = view.White
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
