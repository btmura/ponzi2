package chart

import (
	"image"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view"
)

type volumeAxis struct {
	// renderable is true if this should be rendered.
	renderable bool

	// volumeRange represents the inclusive range from min to max volume.
	volumeRange [2]int

	// labels bundle rendering measurements for volume labels.
	labels []volumeLabel

	// labelRect is the rectangle with global coords that should be drawn within.
	labelRect image.Rectangle
}

type volumeAxisData struct {
	TradingSessionSeries *model.TradingSessionSeries
}

func (v *volumeAxis) SetData(data volumeAxisData) {
	// Reset everything.
	v.Close()

	// Bail out if there is no data yet.
	ts := data.TradingSessionSeries
	if ts == nil {
		return
	}

	v.volumeRange = volumeRange(ts.TradingSessions)

	// Create Y-axis labels for key percentages.
	v.labels = []volumeLabel{
		makeVolumeLabel(volumeValue(v.volumeRange, .8), .8),
		makeVolumeLabel(volumeValue(v.volumeRange, .2), .2),
	}

	v.renderable = true
}

func (v *volumeAxis) SetBounds(labelRect image.Rectangle) {
	v.labelRect = labelRect
}

func (v *volumeAxis) Render(float32) {
	if !v.renderable {
		return
	}

	r := v.labelRect
	for _, l := range v.labels {
		x := r.Max.X - l.size.X
		y := r.Min.Y + int(float32(r.Dy())*l.percent) - l.size.Y/2
		axisLabelTextRenderer.Render(l.text, image.Pt(x, y), gfx.TextColor(view.White))
	}
}

func (v *volumeAxis) Close() {
	v.renderable = false
}
