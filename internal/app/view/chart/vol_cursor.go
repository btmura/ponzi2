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
	// data is the data necessary to render.
	data volumeCursorData

	// renderable is true if this should be rendered.
	renderable bool

	// maxVolume is the maximum volume used for rendering measurements.
	maxVolume int

	// volRect is the rectangle where the volume bars are drawn.
	volRect image.Rectangle

	// labelRect is the rectangle where the axis labels are drawn.
	labelRect image.Rectangle

	// mousePos is the current mouse position. Nil for no mouse input.
	mousePos *view.MousePosition
}

type volumeCursorData struct {
	TradingSessionSeries *model.TradingSessionSeries
}

func (v *volumeCursor) SetData(data volumeCursorData) {
	v.data = data

	// Reset everything.
	v.Close()

	// Bail out if there is no data yet.
	ts := data.TradingSessionSeries
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
	v.mousePos = input.MousePos
}

func (v *volumeCursor) Render(fudge float32) {
	if !v.renderable {
		return
	}

	if v.mousePos == nil {
		return
	}

	renderCursorLines(v.volRect, v.mousePos)

	if !v.mousePos.In(v.volRect) {
		return
	}

	yPercent := float32(v.mousePos.Y-v.volRect.Min.Y) / float32(v.volRect.Dy())
	v.renderLabel(fudge, yPercent)
}

func (v *volumeCursor) renderLabel(fudge float32, yPercent float32) {
	l := makeVolumeLabel(v.maxVolume, yPercent)

	textPt := image.Point{
		X: v.labelRect.Max.X - l.size.X,
		Y: v.labelRect.Min.Y + int(float32(v.labelRect.Dy())*l.percent) - l.size.Y/2,
	}
	bubbleRect := image.Rectangle{
		Min: textPt,
		Max: textPt.Add(l.size),
	}.Inset(-axisLabelPadding)

	// Move the label to the left if the mouse is overlapping.
	if v.mousePos.In(bubbleRect) {
		textPt.X = v.labelRect.Min.X
		bubbleRect = image.Rectangle{
			Min: textPt,
			Max: textPt.Add(l.size),
		}.Inset(-axisLabelPadding)
	}

	axisLabelBubble.SetBounds(bubbleRect)
	axisLabelBubble.Render(fudge)
	axisLabelTextRenderer.Render(l.text, textPt, gfx.TextColor(color.White))
}

func (v *volumeCursor) Close() {
	v.renderable = false
}
