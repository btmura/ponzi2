package view

import (
	"image"

	"github.com/btmura/ponzi2/internal/app/model"
)

type chartStochasticsAxisLabels struct {
	labels []chartStochasticsLabel

	// bounds is the rectangle with global coords that should be drawn within.
	bounds image.Rectangle
}

func newChartStochasticsAxisLabels() *chartStochasticsAxisLabels {
	return &chartStochasticsAxisLabels{}
}

func (ch *chartStochasticsAxisLabels) SetData(ss *model.StochasticSeries) {
	// Bail out if there is not enough data yet.
	if ss == nil {
		return
	}

	// Create Y-axis labels for key percentages.
	ch.labels = []chartStochasticsLabel{
		makeChartStochasticLabel(.8),
		makeChartStochasticLabel(.2),
	}
}

// ProcessInput processes input.
func (ch *chartStochasticsAxisLabels) ProcessInput(ic inputContext) {
	ch.bounds = ic.Bounds
}

func (ch *chartStochasticsAxisLabels) Render(fudge float32) {
	r := ch.bounds
	for _, l := range ch.labels {
		x := r.Max.X - l.size.X
		y := r.Min.Y + int(float32(r.Dy())*l.percent) - l.size.Y/2
		chartAxisLabelTextRenderer.Render(l.text, image.Pt(x, y), white)
	}
}
