package ponzi

import (
	"fmt"
	"image"

	"github.com/golang/glog"

	"github.com/btmura/ponzi2/internal/gfx"
)

// ChartInterval is the data interval like daily or weekly.
type ChartInterval int32

// ChartInterval enums.
const (
	DailyInterval ChartInterval = iota
	WeeklyInterval
)

// ChartStochastics renders the stochastic lines and labels for a single stock.
type ChartStochastics struct {
	interval   ChartInterval
	renderable bool
	labels     []chartStochasticLabel
	stoLines   *gfx.VAO
}

type chartStochasticLabel struct {
	percent float32
	text    string
	size    image.Point
}

// NewChartStochastics creates a new ChartStochastics.
func NewChartStochastics(interval ChartInterval) *ChartStochastics {
	return &ChartStochastics{interval: interval}
}

// SetStock sets the ChartStochastics' stock.
func (ch *ChartStochastics) SetStock(st *ModelStock) {
	// Reset everything.
	ch.Close()

	// Bail out if there is no data yet.
	if st.LastUpdateTime.IsZero() {
		return // Stock has no data yet.
	}

	// Create Y-axis labels for key percentages.
	makeLabel := func(perc float32) chartStochasticLabel {
		t := fmt.Sprintf("%.f%%", perc*100)
		return chartStochasticLabel{
			percent: perc,
			text:    t,
			size:    chartAxisLabelTextRenderer.Measure(t),
		}
	}

	ch.labels = append(ch.labels, makeLabel(.7))
	ch.labels = append(ch.labels, makeLabel(.3))

	var ss []*ModelTradingSession
	var dColor [3]float32
	switch ch.interval {
	case DailyInterval:
		ss, dColor = st.DailySessions, yellow
	case WeeklyInterval:
		ss, dColor = st.WeeklySessions, purple
	default:
		glog.Fatalf("SetStock: unsupported interval: %v", ch.interval)
	}

	ch.stoLines = createStochasticVAOs(ss, dColor)
	ch.renderable = true
}

func createStochasticVAOs(ss []*ModelTradingSession, dColor [3]float32) (stoLines *gfx.VAO) {
	width := 2.0 / float32(len(ss)) // (-1 to 1) on X-axis
	calcX := func(i int) float32 {
		return -1.0 + width*float32(i) + width*0.5
	}
	calcY := func(value float32) float32 {
		return 2.0*float32(value) - 1.0
	}

	data := &gfx.VAOVertexData{}
	var v uint16 // vertex index

	// Add vertices and indices for k percent lines.
	first := true
	for i, s := range ss {
		if s.K == 0 {
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

	// Add vertices and indices for d percent lines.
	first = true
	for i, s := range ss {
		if s.D == 0 {
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

	return gfx.NewVAO(gfx.Lines, data)
}

// Render renders the stochastic lines.
func (ch *ChartStochastics) Render(r image.Rectangle) {
	if !ch.renderable {
		return
	}

	sliceRenderHorizDividers(r, chartGridHorizLine, 0.3, 0.4)

	gfx.SetModelMatrixRect(r)
	ch.stoLines.Render()
}

// RenderLabels renders the Y-axis labels for the stochastic lines.
func (ch *ChartStochastics) RenderLabels(r image.Rectangle) (maxLabelWidth int) {
	if !ch.renderable {
		return
	}

	var maxWidth int
	renderLabel := func(l chartStochasticLabel) {
		x := r.Max.X - l.size.X
		y := r.Min.Y + int(float32(r.Dy())*l.percent) - l.size.Y/2
		chartAxisLabelTextRenderer.Render(l.text, image.Pt(x, y), white)
		if maxWidth < l.size.X {
			maxWidth = l.size.X
		}
	}
	for _, l := range ch.labels {
		renderLabel(l)
	}
	return maxWidth
}

// Close frees the resources backing the ChartStochastics.
func (ch *ChartStochastics) Close() {
	ch.renderable = false
	ch.labels = nil
	if ch.stoLines != nil {
		ch.stoLines.Delete()
		ch.stoLines = nil
	}
}
