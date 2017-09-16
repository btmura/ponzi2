package ponzi

import (
	"fmt"
	"image"
	"time"

	"github.com/btmura/ponzi2/internal/gfx"
)

// ChartInterval is the data interval like daily or weekly.
type ChartInterval int32

// ChartInterval enums.
const (
	DailyInterval ChartInterval = iota
	WeeklyInterval
)

// ChartStochastics shows the stochastic lines for a single stock.
type ChartStochastics struct {
	stock               *ModelStock
	lastStockUpdateTime time.Time
	renderable          bool
	stoInterval         ChartInterval
	stoLines            *gfx.VAO
}

// NewChartStochastics creates a new ChartStochastics.
func NewChartStochastics(stock *ModelStock, stoInterval ChartInterval) *ChartStochastics {
	return &ChartStochastics{
		stock:       stock,
		stoInterval: stoInterval,
	}
}

// Update updates the ChartStochastics.
func (ch *ChartStochastics) Update() {
	if ch.lastStockUpdateTime == ch.stock.LastUpdateTime {
		return
	}
	ch.lastStockUpdateTime = ch.stock.LastUpdateTime

	ss, dColor := ch.stock.DailySessions, yellow
	if ch.stoInterval == WeeklyInterval {
		ss, dColor = ch.stock.WeeklySessions, purple
	}

	if ch.stoLines != nil {
		ch.stoLines.Delete()
	}
	ch.stoLines = createStochasticVAOs(ss, dColor)

	ch.renderable = true
}

func createStochasticVAOs(ss []*ModelTradingSession, dColor [3]float32) (stoLines *gfx.VAO) {
	data := &gfx.VAOVertexData{}

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
		if s.D == 0.0 {
			continue
		}

		data.Vertices = append(data.Vertices, calcX(i), calcY(s.D), 0)
		data.Colors = append(data.Colors, dColor[0], dColor[1], dColor[2])
		if !first {
			data.Indices = append(data.Indices, v, v-1)
		}
		v++
		first = false
	}

	// Add vertices and indices for k percent lines.
	first = true
	for i, s := range ss {
		if s.K == 0.0 {
			continue
		}

		data.Vertices = append(data.Vertices, calcX(i), calcY(s.K), 0)
		data.Colors = append(data.Colors, red[0], red[1], red[2])
		if !first {
			data.Indices = append(data.Indices, v, v-1)
		}
		v++
		first = false
	}

	return gfx.NewVAO(gfx.Lines, data)
}

// Render renders the stochastic lines.
func (ch *ChartStochastics) Render(r image.Rectangle) {
	if !ch.renderable {
		return
	}

	gfx.SetModelMatrixRect(r)
	ch.stoLines.Render()

	renderHorizDividers(r, chartGridHorizLine, 0.3, 0.4)
}

// RenderLabels renders the Y-axis labels for the stochastic lines.
func (ch *ChartStochastics) RenderLabels(r image.Rectangle) (maxLabelWidth int) {
	if !ch.renderable {
		return
	}

	t1, s1 := ch.stochasticLabelText(.7)
	t2, s2 := ch.stochasticLabelText(.3)

	render := func(t string, s image.Point, yLocPercent float32) {
		x := r.Max.X - s.X
		y := r.Min.Y + int(float32(r.Dy())*yLocPercent) - s.Y/2
		chartAxisLabelTextRenderer.Render(t, image.Pt(x, y), white)
	}

	render(t1, s1, .7)
	render(t2, s2, .3)

	s := s1
	if s.X < s2.X {
		s = s2
	}
	return s.X
}

func (ch *ChartStochastics) stochasticLabelText(percent float32) (text string, size image.Point) {
	t := fmt.Sprintf("%.f%%", percent*100)
	return t, chartAxisLabelTextRenderer.Measure(t)
}

// Close frees the resources backing the ChartStochastics.
func (ch *ChartStochastics) Close() {
	ch.renderable = false
	ch.stoLines.Delete()
	ch.stoLines = nil
}
