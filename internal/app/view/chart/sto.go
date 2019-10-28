package chart

import (
	"fmt"
	"image"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view/color"
	"github.com/btmura/ponzi2/internal/app/view/vao"
)

var stochasticHorizRuleSet = vao.HorizRuleSet([]float32{0.2, 0.8}, [2]float32{0, 1}, color.Gray)

type stochastic struct {
	renderable   bool
	dColor       [4]float32
	lineKVAO     *gfx.VAO
	lineDVAO     *gfx.VAO
	MaxLabelSize image.Point
	bounds       image.Rectangle
}

func newStochastic(dColor [4]float32) *stochastic {
	return &stochastic{dColor: dColor}
}

func (s *stochastic) SetData(ss *model.StochasticSeries) {
	// Reset everything.
	s.Close()

	// Bail out if there is not enough data yet.
	if ss == nil {
		return
	}

	// Measure the max label size by creating a label with the max value.
	s.MaxLabelSize = makeStochasticLabel(1).size

	// Create the graph line VAOs.
	var kvals, dvals []float32
	for _, s := range ss.Stochastics {
		kvals = append(kvals, s.K)
		dvals = append(dvals, s.D)
	}
	valRange := [2]float32{0, 1}
	s.lineKVAO = vao.DataLine(kvals, valRange, color.Red)
	s.lineDVAO = vao.DataLine(dvals, valRange, s.dColor)

	s.renderable = true
}

// ProcessInput processes input.
func (s *stochastic) ProcessInput(bounds image.Rectangle) {
	s.bounds = bounds
}

func (s *stochastic) Render(fudge float32) {
	if !s.renderable {
		return
	}

	gfx.SetModelMatrixRect(s.bounds)

	// Render lines for the 20% and 80% levels.
	stochasticHorizRuleSet.Render()

	// Render the stochastic lines.
	s.lineKVAO.Render()
	s.lineDVAO.Render()
}

func (s *stochastic) Close() {
	s.renderable = false
	if s.lineKVAO != nil {
		s.lineKVAO.Delete()
	}
	if s.lineDVAO != nil {
		s.lineDVAO.Delete()
	}
}

// stochasticLabel is a right-justified Y-axis label with the value.
type stochasticLabel struct {
	percent float32
	text    string
	size    image.Point
}

func makeStochasticLabel(perc float32) stochasticLabel {
	t := fmt.Sprintf("%.f%%", perc*100)
	return stochasticLabel{
		percent: perc,
		text:    t,
		size:    axisLabelTextRenderer.Measure(t),
	}
}
