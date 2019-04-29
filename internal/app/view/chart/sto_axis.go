package chart

import (
	"image"

	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view/color"
)

type stochasticAxis struct {
	// renderable is true if this should be rendered.
	renderable bool

	labels []stochasticLabel

	// labelRect is the rectangle with global coords that should be drawn within.
	labelRect image.Rectangle
}

func (s *stochasticAxis) SetData(ss *model.StochasticSeries) {
	// Reset everything.
	s.Close()

	// Bail out if there is not enough data yet.
	if ss == nil {
		return
	}

	// Create Y-axis labels for key percentages.
	s.labels = []stochasticLabel{
		makeStochasticLabel(.8),
		makeStochasticLabel(.2),
	}

	s.renderable = true
}

// ProcessInput processes input.
func (s *stochasticAxis) ProcessInput(labelRect image.Rectangle) {
	s.labelRect = labelRect
}

func (s *stochasticAxis) Render(fudge float32) {
	if !s.renderable {
		return
	}

	r := s.labelRect
	for _, l := range s.labels {
		x := r.Max.X - l.size.X
		y := r.Min.Y + int(float32(r.Dy())*l.percent) - l.size.Y/2
		chartAxisLabelTextRenderer.Render(l.text, image.Pt(x, y), color.White)
	}
}

func (s *stochasticAxis) Close() {
	s.renderable = false
}
