package view

import (
	"image"

	"github.com/btmura/ponzi2/internal/app/model"
)

type chartVolumeAxisLabels struct {
	// maxVolume is the maximum volume used for rendering measurements.
	maxVolume int

	// labels bundle rendering measurements for volume labels.
	labels []chartVolumeLabel

	// bounds is the rectangle with global coords that should be drawn within.
	bounds image.Rectangle
}

func newChartVolumeAxisLabels() *chartVolumeAxisLabels {
	return &chartVolumeAxisLabels{}
}

func (ch *chartVolumeAxisLabels) SetData(ts *model.TradingSessionSeries) {
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

func (ch *chartVolumeAxisLabels) ProcessInput(ic inputContext) {
	ch.bounds = ic.Bounds
}

func (ch *chartVolumeAxisLabels) Render(fudge float32) {
	r := ch.bounds
	for _, l := range ch.labels {
		x := r.Max.X - l.size.X
		y := r.Min.Y + int(float32(r.Dy())*l.percent) - l.size.Y/2
		chartAxisLabelTextRenderer.Render(l.text, image.Pt(x, y), white)
	}
}
