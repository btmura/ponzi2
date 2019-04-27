package chart

import (
	"image"

	"github.com/btmura/ponzi2/internal/app/model"
)

type chartStochasticAxis struct {
	// renderable is true if this should be rendered.
	renderable bool

	labels []chartStochasticsLabel

	// labelRect is the rectangle with global coords that should be drawn within.
	labelRect image.Rectangle
}

func (ch *chartStochasticAxis) SetData(ss *model.StochasticSeries) {
	// Reset everything.
	ch.Close()

	// Bail out if there is not enough data yet.
	if ss == nil {
		return
	}

	// Create Y-axis labels for key percentages.
	ch.labels = []chartStochasticsLabel{
		makeChartStochasticLabel(.8),
		makeChartStochasticLabel(.2),
	}

	ch.renderable = true
}

// ProcessInput processes input.
func (ch *chartStochasticAxis) ProcessInput(labelRect image.Rectangle) {
	ch.labelRect = labelRect
}

func (ch *chartStochasticAxis) Render(fudge float32) {
	if !ch.renderable {
		return
	}

	r := ch.labelRect
	for _, l := range ch.labels {
		x := r.Max.X - l.size.X
		y := r.Min.Y + int(float32(r.Dy())*l.percent) - l.size.Y/2
		chartAxisLabelTextRenderer.Render(l.text, image.Pt(x, y), white)
	}
}

func (ch *chartStochasticAxis) Close() {
	ch.renderable = false
}
