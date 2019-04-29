package chart

import (
	"image"

	"github.com/btmura/ponzi2/internal/app/view/color"
	"github.com/btmura/ponzi2/internal/app/view/rect"
)

// stochasticCursor renders crosshairs at the mouse pointer
// with the corresponding stochastic on the y-axis.
type stochasticCursor struct {
	// renderable is true if this should be rendered.
	renderable bool

	// stoRect is the rectangle where the stochastics are drawn.
	stoRect image.Rectangle

	// labelRect is the rectangle where the axis labels are drawn.
	labelRect image.Rectangle

	// mousePos is the current mouse position.
	mousePos image.Point
}

// ProcessInput processes input.
func (s *stochasticCursor) ProcessInput(stoRect, labelRect image.Rectangle, mousePos image.Point) {
	s.stoRect = stoRect
	s.labelRect = labelRect
	s.mousePos = mousePos
}

func (s *stochasticCursor) SetData() {
	s.renderable = true
}

func (s *stochasticCursor) Render(fudge float32) {
	if !s.renderable {
		return
	}

	renderCursorLines(s.stoRect, s.mousePos)

	if !s.mousePos.In(s.stoRect) {
		return
	}

	perc := float32(s.mousePos.Y-s.stoRect.Min.Y) / float32(s.stoRect.Dy())
	l := makeStochasticLabel(perc)

	var tp image.Point
	tp.X = s.labelRect.Max.X - l.size.X
	tp.Y = s.labelRect.Min.Y + int(float32(s.labelRect.Dy())*l.percent) - l.size.Y/2

	br := image.Rectangle{Min: tp, Max: tp.Add(l.size)}
	br = br.Inset(-chartAxisLabelBubblePadding)

	if s.mousePos.In(br) {
		tp.X = s.labelRect.Min.X
		br = image.Rectangle{Min: tp, Max: tp.Add(l.size)}
		br = br.Inset(-chartAxisLabelBubblePadding)
	}

	rect.FillRoundedRect(br, chartAxisLabelBubbleRounding)
	rect.StrokeRoundedRect(br, chartAxisLabelBubbleRounding)
	chartAxisLabelTextRenderer.Render(l.text, tp, color.White)
}

func (s *stochasticCursor) Close() {
	s.renderable = false
}
