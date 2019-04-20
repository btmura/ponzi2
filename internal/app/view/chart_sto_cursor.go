package view

import "image"

// chartStochasticCursor renders crosshairs at the mouse pointer
// with the corresponding stochastic on the y-axis.
type chartStochasticCursor struct {
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
func (ch *chartStochasticCursor) ProcessInput(stoRect, labelRect image.Rectangle, mousePos image.Point) {
	ch.stoRect = stoRect
	ch.labelRect = labelRect
	ch.mousePos = mousePos
}

func (ch *chartStochasticCursor) SetData() {
	ch.renderable = true
}

func (ch *chartStochasticCursor) Render(fudge float32) {
	if !ch.renderable {
		return
	}

	renderCursorLines(ch.stoRect, ch.mousePos)

	if !ch.mousePos.In(ch.stoRect) {
		return
	}

	perc := float32(ch.mousePos.Y-ch.stoRect.Min.Y) / float32(ch.stoRect.Dy())
	l := makeChartStochasticLabel(perc)

	var tp image.Point
	tp.X = ch.labelRect.Max.X - l.size.X
	tp.Y = ch.labelRect.Min.Y + int(float32(ch.labelRect.Dy())*l.percent) - l.size.Y/2

	br := image.Rectangle{Min: tp, Max: tp.Add(l.size)}
	br = br.Inset(-chartAxisLabelBubblePadding)

	if ch.mousePos.In(br) {
		tp.X = ch.labelRect.Min.X
		br = image.Rectangle{Min: tp, Max: tp.Add(l.size)}
		br = br.Inset(-chartAxisLabelBubblePadding)
	}

	fillRoundedRect(br, chartAxisLabelBubbleRounding)
	strokeRoundedRect(br, chartAxisLabelBubbleRounding)
	chartAxisLabelTextRenderer.Render(l.text, tp, white)
}

func (ch *chartStochasticCursor) Close() {
	ch.renderable = false
}
