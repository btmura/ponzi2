package view

import (
	"image"

	"github.com/btmura/ponzi2/internal/app/model"
)

type chartVolumeCursorLabels struct {
	// maxVolume is the maximum volume used for rendering measurements.
	maxVolume int

	// bounds is the rectangle with global coords that should be drawn within.
	bounds image.Rectangle

	// labelRect is the rectangle where the axis labels are drawn.
	labelRect image.Rectangle

	// mousePos is the current mouse position.
	mousePos image.Point
}

func newChartVolumeCursorLabels() *chartVolumeCursorLabels {
	return &chartVolumeCursorLabels{}
}

func (ch *chartVolumeCursorLabels) SetData(ts *model.TradingSessionSeries) {
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
}

// ProcessInput processes input.
func (ch *chartVolumeCursorLabels) ProcessInput(ic inputContext, labelRect image.Rectangle) {
	ch.bounds = ic.Bounds
	ch.labelRect = labelRect
	ch.mousePos = ic.MousePos
}

func (ch *chartVolumeCursorLabels) Render(fudge float32) {
	if !ch.mousePos.In(ch.bounds) {
		return
	}

	perc := float32(ch.mousePos.Y-ch.bounds.Min.Y) / float32(ch.bounds.Dy())
	l := makeChartVolumeLabel(ch.maxVolume, perc)
	tp := image.Point{
		X: ch.labelRect.Max.X - l.size.X,
		Y: ch.labelRect.Min.Y + int(float32(ch.labelRect.Dy())*l.percent) - l.size.Y/2,
	}

	renderBubble(tp, l.size, chartAxisLabelBubbleSpec)
	chartAxisLabelTextRenderer.Render(l.text, tp, white)
}
