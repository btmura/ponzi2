package chart

import (
	"image"

	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view"
)

// volumeCursor renders crosshairs at the mouse pointer
// with the corresponding volume on the y-axis.
type volumeCursor struct {
	// data is the data necessary to render.
	data volumeCursorData

	// renderable is true if this should be rendered.
	renderable bool

	// volumeRange represents the inclusive range from min to max volume.
	volumeRange [2]int

	// volRect is the rectangle where the volume barLines are drawn.
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

	v.volumeRange = volumeRange(ts.TradingSessions)

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

	if v.mousePos.In(v.volRect) {
		renderVolumeLabel(fudge, v.volumeRange, v.labelRect, v.mousePos.Point, true)
	}
}

func (v *volumeCursor) Close() {
	v.renderable = false
}
