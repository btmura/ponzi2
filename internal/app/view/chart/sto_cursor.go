package chart

import (
	"image"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/view"
	"github.com/btmura/ponzi2/internal/app/view/color"
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

	// mousePos is the current mouse position. Nil for no mouse input.
	mousePos *view.MousePosition
}

func (s *stochasticCursor) SetBounds(stoRect, labelRect image.Rectangle) {
	s.stoRect = stoRect
	s.labelRect = labelRect
}

func (s *stochasticCursor) ProcessInput(input *view.Input) {
	s.mousePos = input.MousePos
}

func (s *stochasticCursor) SetData() {
	s.renderable = true
}

func (s *stochasticCursor) Render(fudge float32) {
	if !s.renderable {
		return
	}

	if s.mousePos == nil {
		return
	}

	renderCursorLines(s.stoRect, s.mousePos.Point)

	if !s.mousePos.In(s.stoRect) {
		return
	}

	perc := float32(s.mousePos.Y-s.stoRect.Min.Y) / float32(s.stoRect.Dy())
	l := makeStochasticLabel(perc)
	tp := image.Point{
		X: s.labelRect.Max.X - l.size.X,
		Y: s.labelRect.Min.Y + int(float32(s.labelRect.Dy())*l.percent) - l.size.Y/2,
	}

	bubbleRect := image.Rectangle{Min: tp, Max: tp.Add(l.size)}
	bubbleRect = bubbleRect.Inset(-axisLabelPadding)

	if s.mousePos.In(bubbleRect) {
		tp.X = s.labelRect.Min.X
		bubbleRect = image.Rectangle{Min: tp, Max: tp.Add(l.size)}
		bubbleRect = bubbleRect.Inset(-axisLabelPadding)
	}

	axisLabelBubble.SetBounds(bubbleRect)
	axisLabelBubble.Render(fudge)
	axisLabelTextRenderer.Render(l.text, tp, gfx.TextColor(color.White))
}

func (s *stochasticCursor) Close() {
	s.renderable = false
}
