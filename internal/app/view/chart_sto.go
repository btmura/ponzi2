package view

import (
	"fmt"
	"image"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view/vao"
)

var chartStochasticHorizRuleSet = vao.HorizRuleSet([]float32{0.2, 0.8}, [2]float32{0, 1}, gray)

type chartStochastics struct {
	dColor       [3]float32
	lineKVAO     *gfx.VAO
	lineDVAO     *gfx.VAO
	MaxLabelSize image.Point
	bounds       image.Rectangle
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

	// Create the graph line VAOs.
	var kvals, dvals []float32
	for _, s := range ss.Stochastics {
		kvals = append(kvals, s.K)
		dvals = append(dvals, s.D)
	}
	valRange := [2]float32{0, 1}
	ch.lineKVAO = vao.DataLine(kvals, valRange, red)
	ch.lineDVAO = vao.DataLine(dvals, valRange, ch.dColor)
}

// ProcessInput processes input.
func (ch *chartStochastics) ProcessInput(bounds image.Rectangle) {
	ch.bounds = bounds
}

func (ch *chartStochastics) Render(fudge float32) {
	if !ch.renderable() {
		return
	}

	gfx.SetModelMatrixRect(ch.bounds)

	// Render lines for the 20% and 80% levels.
	chartStochasticHorizRuleSet.Render()

	// Render the stochastic lines.
	ch.lineKVAO.Render()
	ch.lineDVAO.Render()
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
