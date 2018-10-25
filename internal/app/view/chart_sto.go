package view

import (
	"fmt"
	"image"

	"gitlab.com/btmura/ponzi2/internal/app/gfx"
	"gitlab.com/btmura/ponzi2/internal/app/model"
)

var chartStochasticHorizRuleSet = horizRuleSetVAO([]float32{0.2, 0.8}, [2]float32{0, 1}, gray)

type chartStochastics struct {
	dColor       [3]float32
	lineKVAO     *gfx.VAO
	lineDVAO     *gfx.VAO
	labels       []chartStochasticsLabel
	MaxLabelSize image.Point
}

func newChartStochastics(dColor [3]float32) *chartStochastics {
	return &chartStochastics{dColor: dColor}
}

func (ch *chartStochastics) SetData(ss *model.StochasticSeries) {
	// Reset everything.
	ch.Close()

	// Bail out if there is not enough data yet.
	if ss == nil {
		return
	}

	// Measure the max label size by creating a label with the max value.
	ch.MaxLabelSize = makeChartStochasticLabel(1).size

	// Create Y-axis labels for key percentages.
	ch.labels = []chartStochasticsLabel{
		makeChartStochasticLabel(.8),
		makeChartStochasticLabel(.2),
	}

	// Create the graph line VAOs.
	var kvals, dvals []float32
	for _, s := range ss.Stochastics {
		kvals = append(kvals, s.K)
		dvals = append(dvals, s.D)
	}
	valRange := [2]float32{0, 1}
	ch.lineKVAO = dataLineVAO(kvals, valRange, red)
	ch.lineDVAO = dataLineVAO(dvals, valRange, ch.dColor)
}

func (ch *chartStochastics) Render(r image.Rectangle) {
	if !ch.renderable() {
		return
	}

	gfx.SetModelMatrixRect(r)

	// Render lines for the 20% and 80% levels.
	chartStochasticHorizRuleSet.Render()

	// Render the stochastic lines.
	ch.lineKVAO.Render()
	ch.lineDVAO.Render()
}

func (ch *chartStochastics) RenderAxisLabels(r image.Rectangle) {
	if !ch.renderable() {
		return
	}

	for _, l := range ch.labels {
		x := r.Max.X - l.size.X
		y := r.Min.Y + int(float32(r.Dy())*l.percent) - l.size.Y/2
		chartAxisLabelTextRenderer.Render(l.text, image.Pt(x, y), white)
	}
}

func (ch *chartStochastics) RenderCursorLabels(mainRect, labelRect image.Rectangle, mousePos image.Point) {
	if !ch.renderable() {
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

func (ch *chartStochastics) Close() {
	if ch.lineKVAO != nil {
		ch.lineKVAO.Delete()
		ch.lineKVAO = nil
	}
	if ch.lineDVAO != nil {
		ch.lineDVAO.Delete()
		ch.lineDVAO = nil
	}
}

func (ch *chartStochastics) renderable() bool {
	return ch.lineKVAO != nil && ch.lineDVAO != nil
}

// chartStochasticsLabel is a right-justified Y-axis label with the value.
type chartStochasticsLabel struct {
	percent float32
	text    string
	size    image.Point
}

func makeChartStochasticLabel(perc float32) chartStochasticsLabel {
	t := fmt.Sprintf("%.f%%", perc*100)
	return chartStochasticsLabel{
		percent: perc,
		text:    t,
		size:    chartAxisLabelTextRenderer.Measure(t),
	}
}
