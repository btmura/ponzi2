package ponzi

import (
	"fmt"
	"image"
	"time"

	"github.com/go-gl/gl/v4.5-core/gl"

	"github.com/btmura/ponzi2/internal/gfx"
)

type chartStochasticType int32

const (
	daily chartStochasticType = iota
	weekly
)

type chartStochastics struct {
	stock               *modelStock
	lastStockUpdateTime time.Time
	renderable          bool
	labelText           *gfx.DynamicText
	stoType             chartStochasticType
	stoLines            *gfx.VAO
	labelLine           *gfx.VAO
}

func createChartStochastics(stock *modelStock, labelText *gfx.DynamicText, stoType chartStochasticType) *chartStochastics {
	return &chartStochastics{
		stock:     stock,
		labelText: labelText,
		stoType:   stoType,
	}
}

func (ch *chartStochastics) update() {
	if ch.lastStockUpdateTime == ch.stock.lastUpdateTime {
		return
	}
	ch.lastStockUpdateTime = ch.stock.lastUpdateTime

	ss, dColor := ch.stock.dailySessions, yellow
	if ch.stoType == weekly {
		ss, dColor = ch.stock.weeklySessions, purple
	}

	ch.stoLines.Close()
	ch.stoLines = createStochasticVAOs(ss, dColor)

	ch.labelLine.Close()
	ch.labelLine = gfx.CreateLineVAO(gray, gray)

	ch.renderable = true
}

func createStochasticVAOs(ss []*modelTradingSession, dColor [3]float32) (stoLines *gfx.VAO) {
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

	return gfx.CreateVAO(gl.LINES, vertices, colors, indices)
}

func (ch *chartStochastics) render(r image.Rectangle) {
	if !ch.renderable {
		return
	}

	gfx.SetColorMixAmount(1)
	gfx.SetModelMatrixRect(r)
	ch.stoLines.Render()

	for _, yLocPercent := range []float32{0.3, 0.7} {
		y := r.Min.Y + int(float32(r.Dy())*yLocPercent)
		gfx.SetModelMatrixRect(image.Rect(r.Min.X, y, r.Max.X, y))
		ch.labelLine.Render()
	}
}

func (ch *chartStochastics) renderLabels(r image.Rectangle) (maxLabelWidth int) {
	if !ch.renderable {
		return
	}

	t1, s1 := ch.stochasticLabelText(.7)
	t2, s2 := ch.stochasticLabelText(.3)

	render := func(t string, s image.Point, yLocPercent float32) {
		x := r.Max.X - s.X
		y := r.Min.Y + int(float32(r.Dy())*yLocPercent) - s.Y/2
		ch.labelText.Render(t, image.Pt(x, y), white)
	}

	render(t1, s1, .7)
	render(t2, s2, .3)

	s := s1
	if s.X < s2.X {
		s = s2
	}
	return s.X
}

func (ch *chartStochastics) stochasticLabelText(percent float32) (text string, size image.Point) {
	t := fmt.Sprintf("%.f%%", percent*100)
	return t, ch.labelText.Measure(t)
}

func (ch *chartStochastics) close() {
	ch.renderable = false
	ch.stoLines.Close()
	ch.stoLines = nil
}
