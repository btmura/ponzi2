package chart

import (
	"image"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view"
	"github.com/btmura/ponzi2/internal/app/view/color"
)

// volumeCursor renders crosshairs at the mouse pointer
// with the corresponding volume on the y-axis.
type volumeCursor struct {
	// renderable is true if this should be rendered.
	renderable bool

	// maxVolume is the maximum volume used for rendering measurements.
	maxVolume int

	// volRect is the rectangle where the volume bars are drawn.
	volRect image.Rectangle

	// labelRect is the rectangle where the axis labels are drawn.
	labelRect image.Rectangle

	// mouseMoved is a non-nil mouse move event when the mouse is moving.
	mouseMoved *view.MouseMoveEvent
}

func (v *volumeCursor) SetData(ts *model.TradingSessionSeries) {
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

	v.renderable = true
}

func (v *volumeCursor) SetBounds(volRect, labelRect image.Rectangle) {
	v.volRect = volRect
	v.labelRect = labelRect
}

func (v *volumeCursor) ProcessInput(input *view.Input) {
	v.mouseMoved = input.MouseMoved
}

func (v *volumeCursor) Render(fudge float32) {
	if !v.renderable {
		return
	}

	if v.mouseMoved == nil {
		return
	}

	renderCursorLines(v.volRect, v.mouseMoved.Pos)

	if !v.mouseMoved.In(v.volRect) {
		return
	}

	perc := float32(v.mouseMoved.Pos.Y-v.volRect.Min.Y) / float32(v.volRect.Dy())
	l := makeVolumeLabel(v.maxVolume, perc)
	textPt := image.Point{
		X: v.labelRect.Max.X - l.size.X,
		Y: v.labelRect.Min.Y + int(float32(v.labelRect.Dy())*l.percent) - l.size.Y/2,
	}
	bubbleRect := image.Rectangle{
		Min: textPt,
		Max: textPt.Add(l.size),
	}
	bubbleRect = bubbleRect.Inset(-axisLabelPadding)

	axisLabelBubble.SetBounds(bubbleRect)
	axisLabelBubble.Render(fudge)
	axisLabelTextRenderer.Render(l.text, textPt, gfx.TextColor(color.White))
}

func (v *volumeCursor) Close() {
	v.renderable = false
}
