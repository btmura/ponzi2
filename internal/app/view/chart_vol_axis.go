package view

import (
	"image"

	"github.com/btmura/ponzi2/internal/app/model"
)

type chartVolumeAxis struct {
	// renderable is true if this should be rendered.
	renderable bool

	// maxVolume is the maximum volume used for rendering measurements.
	maxVolume int

	// labels bundle rendering measurements for volume labels.
	labels []chartVolumeLabel

	// labelRect is the rectangle with global coords that should be drawn within.
	labelRect image.Rectangle
}

func (ch *chartVolumeAxis) SetData(ts *model.TradingSessionSeries) {
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

	// Create Y-axis labels for key percentages.
	ch.labels = []chartVolumeLabel{
		makeChartVolumeLabel(ch.maxVolume, .8),
		makeChartVolumeLabel(ch.maxVolume, .2),
	}
}

func (ch *chartVolumeAxis) ProcessInput(labelRect image.Rectangle) {
	ch.labelRect = labelRect
}

func (ch *chartVolumeAxis) Render(fudge float32) {
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

func (ch *chartVolumeAxis) Close() {
	ch.renderable = false
}
