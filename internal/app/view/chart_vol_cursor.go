package view

import (
	"image"

	"github.com/btmura/ponzi2/internal/app/model"
)

// chartVolumeCursor renders crosshairs at the mouse pointer
// with the corresponding volume on the y-axis.
type chartVolumeCursor struct {
	// renderable is true if this should be rendered.
	renderable bool

	// maxVolume is the maximum volume used for rendering measurements.
	maxVolume int

	// volRect is the rectangle where the volume bars are drawn.
	volRect image.Rectangle

	// labelRect is the rectangle where the axis labels are drawn.
	labelRect image.Rectangle

	// mousePos is the current mouse position.
	mousePos image.Point
}

func (ch *chartVolumeCursor) SetData(ts *model.TradingSessionSeries) {
	// Reset everything.
	ch.Close()

	// Bail out if there is no data yet.
	if ts == nil {
		return
	}

	// Find the maximum volume.
	ch.maxVolume = 0
	for _, s := range ts.TradingSessions {
		if ch.maxVolume < s.Volume {
			ch.maxVolume = s.Volume
		}
	}

	ch.renderable = true
}

// ProcessInput processes input.
func (ch *chartVolumeCursor) ProcessInput(volRect, labelRect image.Rectangle, mousePos image.Point) {
	ch.volRect = volRect
	ch.labelRect = labelRect
	ch.mousePos = mousePos
}

func (ch *chartVolumeCursor) Render(fudge float32) {
	if !ch.renderable {
		return
	}

	renderCursorLines(ch.volRect, ch.mousePos)

	if !ch.mousePos.In(ch.volRect) {
		return
	}

	perc := float32(ch.mousePos.Y-ch.volRect.Min.Y) / float32(ch.volRect.Dy())
	l := makeChartVolumeLabel(ch.maxVolume, perc)
	tp := image.Point{
		X: ch.labelRect.Max.X - l.size.X,
		Y: ch.labelRect.Min.Y + int(float32(ch.labelRect.Dy())*l.percent) - l.size.Y/2,
	}

	renderBubble(tp, l.size, chartAxisLabelBubbleSpec)
	chartAxisLabelTextRenderer.Render(l.text, tp, white)
}

func (ch *chartVolumeCursor) Close() {
	ch.renderable = false
}
