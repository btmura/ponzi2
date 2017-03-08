package ponzi

import (
	"fmt"
	"image"
	"math"
	"strconv"

	"github.com/go-gl/gl/v4.5-core/gl"
)

// Colors used by the chart.
var (
	green  = [3]float32{0.25, 1, 0}
	red    = [3]float32{1, 0.3, 0}
	yellow = [3]float32{1, 1, 0}
	purple = [3]float32{0.5, 0, 1}
	white  = [3]float32{1, 1, 1}
)

const chartLabelPadding = 2

type chart struct {
	stock *modelStock

	minPrice float32
	maxPrice float32

	symbolQuoteText *dynamicText
	labelText       *dynamicText
	frameBorder     *vao
	frameDivider    *vao
	stickLines      *vao
	stickRects      *vao
	volRects        *vao
	dailyStoLines   *vao
	weeklyStoLines  *vao
}

func createChart(stock *modelStock, symbolQuoteText, labelText *dynamicText) *chart {
	return &chart{
		stock:           stock,
		symbolQuoteText: symbolQuoteText,
		labelText:       labelText,
		frameBorder:     createStrokedRectVAO(white, white, white, white),
		frameDivider:    createLineVAO(white, white),
	}
}

func (ch *chart) update() {
	if ch == nil {
		return
	}
	ch.updatePrices()
	ch.updateVolume()
	ch.updateStochastics()
}

func (ch *chart) updatePrices() {
	if ch.stock.dailySessions == nil {
		return
	}

	if ch.stickLines != nil {
		return
	}

	// Find the min and max prices for the y-axis.
	ch.minPrice, ch.maxPrice = math.MaxFloat32, 0
	for _, s := range ch.stock.dailySessions {
		if ch.minPrice > s.low {
			ch.minPrice = s.low
		}
		if ch.maxPrice < s.high {
			ch.maxPrice = s.high
		}
	}

	// Calculate vertices and indices for the candlesticks.
	var vertices []float32
	var colors []float32
	var lineIndices []uint16
	var triangleIndices []uint16

	stickWidth := 2.0 / float32(len(ch.stock.dailySessions)) // (-1 to 1) on X-axis
	leftX := -1.0 + stickWidth*0.1
	midX := -1.0 + stickWidth*0.5
	rightX := -1.0 + stickWidth*0.9

	calcY := func(value float32) float32 {
		return 2*(value-ch.minPrice)/(ch.maxPrice-ch.minPrice) - 1
	}

	for i, s := range ch.stock.dailySessions {
		// Figure out Y coordinates of the key levels.
		lowY, highY, openY, closeY := calcY(s.low), calcY(s.high), calcY(s.open), calcY(s.close)

		// Figure out the top and bottom of the candlestick.
		topY, botY := openY, closeY
		if openY < closeY {
			topY, botY = closeY, openY
		}

		// Add the vertices needed to create the candlestick.
		vertices = append(vertices,
			midX, highY, // 0
			midX, topY, // 1
			midX, lowY, // 2
			midX, botY, // 3
			leftX, topY, // 4 - Upper left of box
			rightX, topY, // 5 - Upper right of box
			leftX, botY, // 6 - Bottom left of box
			rightX, botY, // 7 - Bottom right of box
		)

		// Add the colors corresponding to the vertices.
		var c [3]float32
		switch {
		case s.close > s.open:
			c = green
		case s.close < s.open:
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

		if s.close > s.open {
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

	ch.stickLines = createVAO(gl.LINES, vertices, colors, lineIndices)
	ch.stickRects = createVAO(gl.TRIANGLES, vertices, colors, triangleIndices)
}

func (ch *chart) updateVolume() {
	if ch.stock.dailySessions == nil {
		return
	}

	if ch.volRects != nil {
		return
	}

	// Find the max volume for the y-axis.
	var max int
	for _, s := range ch.stock.dailySessions {
		if s.volume > max {
			max = s.volume
		}
	}

	// Calculate vertices and indices for the volume bars.
	var vertices []float32
	var colors []float32
	var indices []uint16

	barWidth := 2.0 / float32(len(ch.stock.dailySessions)) // (-1 to 1) on X-axis
	leftX := -1.0 + barWidth*0.2
	rightX := -1.0 + barWidth*0.8

	calcY := func(value int) float32 {
		return 2*float32(value)/float32(max) - 1
	}

	for i, s := range ch.stock.dailySessions {
		topY := calcY(s.volume)
		botY := calcY(0)

		// Add the vertices needed to create the volume bar.
		vertices = append(vertices,
			leftX, topY, // UL
			rightX, topY, // UR
			leftX, botY, // BL
			rightX, botY, // BR
		)

		// Add the colors corresponding to the volume bar.
		switch {
		case s.close > s.open:
			colors = append(colors,
				green[0], green[1], green[2],
				green[0], green[1], green[2],
				green[0], green[1], green[2],
				green[0], green[1], green[2],
			)

		case s.close < s.open:
			colors = append(colors,
				red[0], red[1], red[2],
				red[0], red[1], red[2],
				red[0], red[1], red[2],
				red[0], red[1], red[2],
			)

		default:
			colors = append(colors,
				yellow[0], yellow[1], yellow[2],
				yellow[0], yellow[1], yellow[2],
				yellow[0], yellow[1], yellow[2],
				yellow[0], yellow[1], yellow[2],
			)
		}

		// idx is function to refer to the vertices above.
		idx := func(j uint16) uint16 {
			return uint16(i)*4 + j
		}

		// Use triangles for filled candlestick on lower closes.
		indices = append(indices,
			idx(0), idx(2), idx(1),
			idx(1), idx(2), idx(3),
		)

		// Move the X coordinates one bar over.
		leftX += barWidth
		rightX += barWidth
	}

	ch.volRects = createVAO(gl.TRIANGLES, vertices, colors, indices)
}

func (ch *chart) updateStochastics() {
	if ch == nil {
		return
	}

	if ch.stock.dailySessions == nil {
		return
	}

	if ch.dailyStoLines != nil {
		return
	}

	ch.dailyStoLines = ch.createStoLines(ch.stock.dailySessions, yellow)
	ch.weeklyStoLines = ch.createStoLines(ch.stock.weeklySessions, purple)
}

func (ch *chart) createStoLines(ss []*modelTradingSession, dColor [3]float32) *vao {
	var vertices []float32
	var colors []float32
	var indices []uint16

	width := 2.0 / float32(len(ss)) // (-1 to 1) on X-axis
	calcX := func(i int) float32 {
		return -1.0 + width*0.5 + width*float32(i)
	}
	calcY := func(value float32) float32 {
		return 2*float32(value) - 1
	}

	var v uint16 // vertex index

	// Add vertices and indices for d percent lines.
	first := true
	for i, s := range ss {
		if s.d == 0.0 {
			continue
		}

		vertices = append(vertices, calcX(i), calcY(s.d))
		colors = append(colors, dColor[0], dColor[1], dColor[2])
		if !first {
			indices = append(indices, v, v-1)
		}
		v++
		first = false
	}

	// Add vertices and indices for k percent lines.
	first = true
	for i, s := range ss {
		if s.k == 0.0 {
			continue
		}

		vertices = append(vertices, calcX(i), calcY(s.k))
		colors = append(colors, red[0], red[1], red[2])
		if !first {
			indices = append(indices, v, v-1)
		}
		v++
		first = false
	}

	return createVAO(gl.LINES, vertices, colors, indices)
}

func (ch *chart) render(r image.Rectangle) {
	if ch == nil {
		return
	}
	const pad = 3
	subRects := ch.renderFrame(r)
	ch.renderPrices(subRects[3].Inset(pad))
	ch.renderVolume(subRects[2].Inset(pad))
	ch.renderStochastics(ch.dailyStoLines, subRects[1].Inset(pad))
	ch.renderStochastics(ch.weeklyStoLines, subRects[0].Inset(pad))
}

func (ch *chart) renderFrame(r image.Rectangle) []image.Rectangle {
	if ch == nil {
		return nil
	}

	// Start rendering from the top left. Track position with point.
	pt := image.Pt(r.Min.X, r.Max.Y)

	//
	// Render the frame around the chart.
	//

	gl.Uniform1f(colorMixAmountLocation, 1)
	setModelMatrixRectangle(r)
	ch.frameBorder.render()

	//
	// Render the symbol and its quote.
	//

	const pad = 10
	s := ch.symbolQuoteText.measure(ch.stock.symbol)
	pt.Y -= pad + s.Y
	{
		c := pt
		c.X += pad
		c = c.Add(ch.symbolQuoteText.render(ch.stock.symbol, c))
		c = c.Add(ch.symbolQuoteText.render(formatQuote(ch.stock.quote), c))
	}
	pt.Y -= pad

	//
	// Render the dividers between the sections.
	//

	r.Max.Y = pt.Y
	gl.Uniform1f(colorMixAmountLocation, 1)

	rects := sliceRectangle(r, 0.13, 0.13, 0.13, 0.6)
	for _, r := range rects {
		setModelMatrixRectangle(image.Rect(r.Min.X, r.Max.Y, r.Max.X, r.Max.Y))
		ch.frameDivider.render()
	}
	return rects
}

func (ch *chart) renderPrices(r image.Rectangle) {
	if ch == nil {
		return
	}

	if ch.stock.dailySessions != nil {
		makeLabel := func(v float32) string {
			return strconv.FormatFloat(float64(v), 'f', 2, 32)
		}

		labelSize := ch.labelText.measure(makeLabel(ch.maxPrice))
		labelPaddingX, labelPaddingY := 4, labelSize.Y/2
		pricePerPixel := (ch.maxPrice - ch.minPrice) / float32(r.Dy())

		// Start at top and decrement one label with top and bottom padding.
		c := r.Max
		dc := image.Pt(0, labelPaddingY+labelSize.Y+labelPaddingY)

		// Start at top with max price and decrement change in price of a label with padding.
		v := ch.maxPrice
		dv := pricePerPixel * float32(dc.Y)

		// Offets to the cursor and price value when drawing.
		dcy := labelPaddingY + labelSize.Y   // Puts cursor at the baseline of the text.
		dvy := labelPaddingY + labelSize.Y/2 // Uses value in the middle of the label.

		for {
			{
				c := image.Pt(c.X, c.Y-dcy)
				if c.Y < r.Min.Y {
					break
				}

				v := v - pricePerPixel*float32(dvy)
				l := makeLabel(v)
				s := ch.labelText.measure(l)
				c.X -= s.X + labelPaddingX
				ch.labelText.render(l, c)
			}

			c = c.Sub(dc)
			v -= dv
		}

		r.Max.X -= labelSize.X + labelPaddingX*2
	}

	gl.Uniform1f(colorMixAmountLocation, 1)
	setModelMatrixRectangle(r)
	ch.stickLines.render()
	ch.stickRects.render()
}

func (ch *chart) renderVolume(r image.Rectangle) {
	if ch == nil {
		return
	}

	r = ch.renderVolumeLabels(r)
	gl.Uniform1f(colorMixAmountLocation, 1)
	setModelMatrixRectangle(r)
	ch.volRects.render()
}

func (ch *chart) renderVolumeLabels(r image.Rectangle) image.Rectangle {
	if ch == nil {
		return r
	}

	if ch.stock.dailySessions == nil {
		return r
	}

	var maxVol int
	for _, ds := range ch.stock.dailySessions {
		if maxVol < ds.volume {
			maxVol = ds.volume
		}
	}

	makeLabel := func(v int) string {
		switch {
		case v > 1000000000:
			return fmt.Sprintf("%dB", v/1000000000)
		case v > 1000000:
			return fmt.Sprintf("%dM", v/1000000)
		case v > 1000:
			return fmt.Sprintf("%dK", v/1000)
		}
		return strconv.Itoa(v)
	}

	labelSize := ch.labelText.measure(makeLabel(maxVol))
	labelPadX, labelPadY := 4, labelSize.Y/2

	volPerPixel := float32(maxVol) / float32(r.Dy())
	volOffset := int(float32(labelPadY+labelSize.Y/2) * volPerPixel)

	var maxTextWidth int
	render := func(v, y int) {
		l := makeLabel(v)
		s := ch.labelText.measure(l)
		if maxTextWidth < s.X {
			maxTextWidth = s.X
		}
		x := r.Max.X - s.X - labelPadX
		ch.labelText.render(l, image.Pt(x, y))
	}

	render(maxVol-volOffset, r.Max.Y-labelPadY-labelSize.Y)
	render(volOffset, r.Min.Y+labelPadY)

	r.Max.X -= maxTextWidth + labelPadX*2
	return r
}

func (ch *chart) renderStochastics(vao *vao, r image.Rectangle) {
	if ch == nil {
		return
	}

	r.Max.X -= ch.renderStochasticLabels(r)
	gl.Uniform1f(colorMixAmountLocation, 1)
	setModelMatrixRectangle(r)
	vao.render()
}

func (ch *chart) renderStochasticLabels(r image.Rectangle) (maxLabelWidth int) {
	if ch == nil {
		return
	}

	if ch.stock.dailySessions == nil {
		return
	}

	render := func(percent float32) (width int) {
		t := fmt.Sprintf("%.f%%", percent*100)
		s := ch.labelText.measure(t)
		p := image.Pt(r.Max.X-s.X-chartLabelPadding, r.Min.Y+int(float32(r.Dy())*percent)-s.Y/2)
		ch.labelText.render(t, p)
		return s.X + chartLabelPadding*2
	}

	w1, w2 := render(.3), render(.7)
	if w1 > w2 {
		return w1
	}
	return w2
}

func (ch *chart) close() {
	if ch == nil {
		return
	}
	ch.frameDivider.close()
	ch.frameBorder.close()
	ch.stickLines.close()
	ch.stickRects.close()
	ch.volRects.close()
	ch.dailyStoLines.close()
	ch.weeklyStoLines.close()
}
