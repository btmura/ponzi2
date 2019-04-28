package chart

import (
	"image"

	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view/color"
)

type volumeAxis struct {
	// renderable is true if this should be rendered.
	renderable bool

	// maxVolume is the maximum volume used for rendering measurements.
	maxVolume int

	// labels bundle rendering measurements for volume labels.
	labels []volumeLabel

	// labelRect is the rectangle with global coords that should be drawn within.
	labelRect image.Rectangle
}

func (v *volumeAxis) SetData(ts *model.TradingSessionSeries) {
	// Reset everything.
	v.Close()

	// Bail out if there is no data yet.
	if ts == nil {
		return
	}

	// Find the maximum volume.
	v.maxVolume = 0
	for _, s := range ts.TradingSessions {
		if v.maxVolume < s.Volume {
			v.maxVolume = s.Volume
		}
	}

	// Create Y-axis labels for key percentages.
	v.labels = []volumeLabel{
		makeVolumeLabel(v.maxVolume, .8),
		makeVolumeLabel(v.maxVolume, .2),
	}

	v.renderable = true
}

func (v *volumeAxis) ProcessInput(labelRect image.Rectangle) {
	v.labelRect = labelRect
}

func (v *volumeAxis) Render(fudge float32) {
	if !v.renderable {
		return
	}

	r := v.labelRect
	for _, l := range v.labels {
		x := r.Max.X - l.size.X
		y := r.Min.Y + int(float32(r.Dy())*l.percent) - l.size.Y/2
		chartAxisLabelTextRenderer.Render(l.text, image.Pt(x, y), color.White)
	}
}

func (v *volumeAxis) Close() {
	v.renderable = false
}
