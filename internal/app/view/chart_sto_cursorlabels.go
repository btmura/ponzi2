package view

import "image"

type chartStochasticsCursorLabels struct {
	// bounds is the rectangle with global coords that should be drawn within.
	bounds image.Rectangle

	// labelRect is the rectangle where the axis labels are drawn.
	labelRect image.Rectangle

	// mousePos is the current mouse position.
	mousePos image.Point
}

func newChartStochasticsCursorLabels() *chartStochasticsCursorLabels {
	return &chartStochasticsCursorLabels{}
}

// ProcessInput processes input.
func (ch *chartStochasticsCursorLabels) ProcessInput(ic inputContext, labelRect image.Rectangle) {
	ch.bounds = ic.Bounds
	ch.labelRect = labelRect
	ch.mousePos = ic.MousePos
}

func (ch *chartStochasticsCursorLabels) Render(fudge float32) {
	renderCursorLines(ch.bounds, ch.mousePos)

	if !ch.mousePos.In(ch.bounds) {
		return
	}

	perc := float32(ch.mousePos.Y-ch.bounds.Min.Y) / float32(ch.bounds.Dy())
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
