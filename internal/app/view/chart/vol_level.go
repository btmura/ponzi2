package chart

import (
	"fmt"
	"image"
	"strconv"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view"
	"github.com/btmura/ponzi2/internal/app/view/vao"
)

// volumeHorizLine is the horizontal lines rendered behind the volume bars.
var volumeHorizLine = vao.HorizLine(view.TransparentGray, view.Gray)

type volumeLevel struct {
	// renderable is true if this should be rendered.
	renderable bool

	// volumeRange represents the inclusive range from min to max volume.
	volumeRange [2]int

	// MaxLabelSize is the maximum label size useful for rendering measurements.
	MaxLabelSize image.Point

	// lineBounds is the rectangle where the horizontal lines should be drawn within.
	lineBounds image.Rectangle

	// labelBounds is the rectangle where the labels for the lines should be drawn within.
	labelBounds image.Rectangle
}

func newVolumeLevel() *volumeLevel {
	return new(volumeLevel)
}

type volumeLevelData struct {
	TradingSessionSeries *model.TradingSessionSeries
}

func (v *volumeLevel) SetData(data volumeLevelData) {
	// Reset everything.
	v.Close()

	// Bail out if there is no data yet.
	ts := data.TradingSessionSeries
	if ts == nil {
		return
	}

	v.volumeRange = volumeRange(ts.TradingSessions)

	// Measure the max label size by creating a label with the max value.
	v.MaxLabelSize = makeVolumeLabel(v.volumeRange[1]).size

	v.renderable = true
}

func (v *volumeLevel) SetBounds(lineBounds, labelBounds image.Rectangle) {
	v.lineBounds = lineBounds
	v.labelBounds = labelBounds
}

func (v *volumeLevel) Render(fudge float32) {
	if !v.renderable {
		return
	}

	r := v.lineBounds
	for _, y := range v.labelYPositions(r) {
		gfx.SetModelMatrixRect(image.Rect(r.Min.X, y, r.Max.X, y))
		volumeHorizLine.Render()
	}

	r = v.labelBounds
	for _, y := range v.labelYPositions(r) {
		renderVolumeLabel(fudge, v.volumeRange, r, image.Pt(0, y), false)
	}
}

func (v *volumeLevel) labelYPositions(r image.Rectangle) []int {
	labelPaddingY := v.MaxLabelSize.Y / 2
	firstY := r.Max.Y - labelPaddingY - v.MaxLabelSize.Y/2
	dy := v.MaxLabelSize.Y * 2

	var yPositions []int
	for y := firstY; y >= r.Min.Y; y -= dy {
		yPositions = append(yPositions, y)
	}
	return yPositions
}

func (v *volumeLevel) Close() {
	v.renderable = false
}

// volumeLabel is a right-justified Y-axis label with the volume.
type volumeLabel struct {
	text string
	size image.Point
}

func makeVolumeLabel(value int) volumeLabel {
	t := volumeText(value)
	return volumeLabel{
		text: t,
		size: axisLabelTextRenderer.Measure(t),
	}
}

func volumeText(v int) string {
	var t string
	switch {
	case v > 1000000000:
		t = fmt.Sprintf("%.1fB", float32(v)/1e9)
	case v > 1000000:
		t = fmt.Sprintf("%.1fM", float32(v)/1e6)
	case v > 1000:
		t = fmt.Sprintf("%.1fK", float32(v)/1e3)
	default:
		t = strconv.Itoa(v)
	}
	return t
}

func renderVolumeLabel(fudge float32, volRange [2]int, r image.Rectangle, pt image.Point, includeBubble bool) {
	yPercent := float32(pt.Y-r.Min.Y) / float32(r.Dy())
	value := volumeValue(volRange, yPercent)
	label := makeVolumeLabel(value)

	textPt := image.Point{
		X: r.Max.X - label.size.X,
		Y: r.Min.Y + int(float32(r.Dy())*yPercent) - label.size.Y/2,
	}
	bubbleRect := image.Rectangle{
		Min: textPt,
		Max: textPt.Add(label.size),
	}.Inset(-axisLabelPadding)

	// Move the label to the left if the mouse is overlapping.
	if pt.In(bubbleRect) {
		textPt.X = r.Min.X
		bubbleRect = image.Rectangle{
			Min: textPt,
			Max: textPt.Add(label.size),
		}.Inset(-axisLabelPadding)
	}

	if includeBubble {
		axisLabelBubble.SetBounds(bubbleRect)
		axisLabelBubble.Render(fudge)
	}
	axisLabelTextRenderer.Render(label.text, textPt, gfx.TextColor(view.White))
}
