package chart

import (
	"image"
	"math"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view"
	"github.com/btmura/ponzi2/internal/logger"
)

// price shows the candlesticks and price labels for a single stock.
type price struct {
	// renderable is whether the prices can be rendered.
	renderable bool

	// priceRange represents the inclusive range from min to max price.
	priceRange [2]float32

	// priceStyle is the price style whether bars or candlesticks.
	priceStyle PriceStyle

	// faders has the faders needed to fade in and out the bars and candlesticks.
	faders map[PriceStyle]*view.Fader

	// barLines is the VAO with the price bar lines.
	barLines *gfx.VAO

	// stickLines is the VAO with the volume lines.
	stickLines *gfx.VAO

	// stickRects is the VAO with the volume bars.
	stickRects *gfx.VAO

	// bounds is the rectangle with global coords that should be drawn within.
	bounds image.Rectangle
}

func newPrice(priceStyle PriceStyle) *price {
	if priceStyle == PriceStyleUnspecified {
		logger.Error("unspecified price style")
		return nil
	}

	return &price{
		priceStyle: priceStyle,
		faders: map[PriceStyle]*view.Fader{
			Bar:         view.NewStoppedFader(1 * view.FPS),
			Candlestick: view.NewStoppedFader(1 * view.FPS),
		},
	}
}

// SetPriceStyle sets the style whether bars or candlesticks.
func (p *price) SetStyle(newStyle PriceStyle) {
	if newStyle == PriceStyleUnspecified {
		logger.Error("unspecified price style")
		return
	}

	if newStyle == p.priceStyle {
		return
	}

	p.priceStyle = newStyle
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

	p.barLines = priceBarVAO(ts.TradingSessions, p.priceRange)

	p.stickLines, p.stickRects = priceCandlestickVAOs(ts.TradingSessions, p.priceRange)

	p.renderable = true
}

func (p *price) SetBounds(bounds image.Rectangle) {
	p.bounds = bounds
}

func (p *price) Update() (dirty bool) {
	for s, fader := range p.faders {
		if s == p.priceStyle {
			fader.FadeIn()
		} else {
			fader.FadeOut()
		}
	}

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

	gfx.SetModelMatrixRect(p.bounds)

	for style, fader := range p.faders {
		fader.Render(fudge, func() {
			switch style {
			case Bar:
				p.barLines.Render()

			case Candlestick:
				p.stickLines.Render()
				p.stickRects.Render()
			}
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
	for _, s := range ts {
		if s.Low != 0 && s.Low < low {
			low = s.Low
		}
		if s.High != 0 && s.High > high {
			high = s.High
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

func pricePercent(priceRange [2]float32, value float32) (percent float32) {
	log := func(value float32) float64 {
		if value == 0 {
			return 0
		}
		return math.Log(float64(value))
	}
	percent = float32((log(value) - log(priceRange[0])) / (log(priceRange[1]) - log(priceRange[0])))
	if percent >= 0 {
		return percent
	}
	return 0
}

func priceValue(priceRange [2]float32, percent float32) (value float32) {
	log := func(value float32) float64 {
		if value == 0 {
			return 0
		}
		return math.Log(float64(value))
	}
	return float32(
		math.Pow(
			math.E,
			float64(percent)*(log(priceRange[1])-log(priceRange[0]))+log(priceRange[0]),
		),
	)
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
		return 2*pricePercent(priceRange, value) - 1
	}

	for _, s := range ts {
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
			c = view.Blue
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

	return gfx.NewVAO(
		&gfx.VAOVertexData{
			Mode:     gfx.Lines,
			Vertices: vertices,
			Colors:   colors,
			Indices:  lineIndices,
		},
	)
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
		return 2*pricePercent(priceRange, value) - 1
	}

	for _, s := range ts {
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
			c = view.Blue
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
