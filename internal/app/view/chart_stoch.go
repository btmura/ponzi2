package view

import (
	"fmt"
	"image"

	"github.com/btmura/ponzi2/internal/app/model"
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
	interval ChartInterval

	// renderable is whether the ChartStochastics can be rendered.
	renderable bool

	// MaxLabelSize is the maximum label size useful for rendering measurements.
	MaxLabelSize image.Point

	// labels bundle rendering measurements for stochastic labels.
	labels []chartStochasticLabel

	// stoLines is the VAO with the colored stochastic lines.
	stoLines *gfx.VAO
}

// NewChartStochastics creates a new ChartStochastics.
func NewChartStochastics(interval ChartInterval) *ChartStochastics {
	return &ChartStochastics{interval: interval}
}

// SetStock sets the ChartStochastics' stock.
func (ch *ChartStochastics) SetStock(st *model.Stock) {
	// Reset everything.
	ch.Close()

	// Bail out if there is no data yet.
	if st.LastUpdateTime.IsZero() {
		return // Stock has no data yet.
	}

	// Measure the max label size by creating a label with the max value.
	ch.MaxLabelSize = makeChartStochasticLabel(1).size

	// Create Y-axis labels for key percentages.
	ch.labels = []chartStochasticLabel{
		makeChartStochasticLabel(.8),
		makeChartStochasticLabel(.2),
	}

	var ss *model.Stochastics
	var dColor [3]float32
	switch ch.interval {
	case DailyInterval:
		ss, dColor = st.DailyStochastics, yellow
	case WeeklyInterval:
		ss, dColor = st.WeeklyStochastics, purple
	default:
		logger.Fatalf("SetStock: unsupported interval: %v", ch.interval)
	}

	ch.stoLines = chartStochasticVAO(ss, dColor)

	ch.renderable = true
}

// Render renders the stochastic lines.
func (ch *ChartStochastics) Render(r image.Rectangle) {
	if !ch.renderable {
		return
	}

	// Render lines for the 20% and 80% levels.
	renderSlicedRectDividers(r, chartGridHorizLine, 0.2, 0.6)

	// Render the stochastic lines.
	gfx.SetModelMatrixRect(r)
	ch.stoLines.Render()
}

// RenderAxisLabels renders the Y-axis labels for the stochastic lines.
func (ch *ChartStochastics) RenderAxisLabels(r image.Rectangle) {
	if !ch.renderable {
		return
	}

	for _, l := range ch.labels {
		x := r.Max.X - l.size.X
		y := r.Min.Y + int(float32(r.Dy())*l.percent) - l.size.Y/2
		chartAxisLabelTextRenderer.Render(l.text, image.Pt(x, y), white)
	}
}

// RenderCursorLabels renders a label for the value under the mouse cursor.
func (ch *ChartStochastics) RenderCursorLabels(mainRect, labelRect image.Rectangle, mousePos image.Point) {
	if !ch.renderable {
		return
	}

	if !mousePos.In(mainRect) {
		return
	}

	perc := float32(mousePos.Y-mainRect.Min.Y) / float32(mainRect.Dy())
	l := makeChartStochasticLabel(perc)

	var tp image.Point
	tp.X = labelRect.Max.X - l.size.X
	tp.Y = labelRect.Min.Y + int(float32(labelRect.Dy())*l.percent) - l.size.Y/2

	br := image.Rectangle{Min: tp, Max: tp.Add(l.size)}
	br = br.Inset(-chartAxisLabelBubblePadding)

	if mousePos.In(br) {
		tp.X = labelRect.Min.X
		br = image.Rectangle{Min: tp, Max: tp.Add(l.size)}
		br = br.Inset(-chartAxisLabelBubblePadding)
	}

	fillRoundedRect(br, chartAxisLabelBubbleRounding)
	strokeRoundedRect(br, chartAxisLabelBubbleRounding)
	chartAxisLabelTextRenderer.Render(l.text, tp, white)
}

// Close frees the resources backing the ChartStochastics.
func (ch *ChartStochastics) Close() {
	ch.renderable = false
	if ch.stoLines != nil {
		ch.stoLines.Delete()
	}
}

// chartStochasticLabel is a right-justified Y-axis label with the value.
type chartStochasticLabel struct {
	percent float32
	text    string
	size    image.Point
}

func makeChartStochasticLabel(perc float32) chartStochasticLabel {
	t := fmt.Sprintf("%.f%%", perc*100)
	return chartStochasticLabel{
		percent: perc,
		text:    t,
		size:    chartAxisLabelTextRenderer.Measure(t),
	}
}

func chartStochasticVAO(ss *model.Stochastics, dColor [3]float32) *gfx.VAO {
	data := &gfx.VAOVertexData{}
	var v uint16 // vertex index

	dx := 2.0 / float32(len(ss.Values)) // (-1 to 1) on X-axis
	calcX := func(i int) float32 {
		return -1.0 + dx*float32(i) + dx*0.5
	}
	calcY := func(v float32) float32 {
		return 2.0*float32(v) - 1.0
	}

	// Add vertices and indices for k percent lines.
	first := true
	for i, s := range ss.Values {
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
	for i, s := range ss.Values {
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
