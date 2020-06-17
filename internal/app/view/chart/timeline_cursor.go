package chart

import (
	"image"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view"
	"github.com/btmura/ponzi2/internal/logger"
)

// timelineCursor renders the time corresponding to the mouse pointer
// on the x-axis.
type timelineCursor struct {
	// data is the data necessary to render.
	data timelineCursorData

	// renderable is true if this should be rendered.
	renderable bool

	// layout is the layout to use when printing times.
	layout string

	// bounds is the rectangle with global coords that should be drawn within.
	bounds image.Rectangle

	// labelRect is the rectangle where the axis labels are drawn.
	labelRect image.Rectangle

	// mousePos is the current mouse position. Nil for no mouse input.
	mousePos *view.MousePosition
}

type timelineCursorData struct {
	Range                model.Range
	TradingSessionSeries *model.TradingSessionSeries
}

func (t *timelineCursor) SetData(data timelineCursorData) {
	// Reset everything.
	t.Close()

	// Bail out if there is no data yet.
	ts := data.TradingSessionSeries
	if ts == nil {
		return
	}

	t.data = data

	switch data.Range {
	case model.OneDay:
		t.layout = "03:04"
	case model.OneYear:
		t.layout = "1/2/06"
	default:
		logger.Errorf("bad range: %v", data.Range)
		return
	}

	t.renderable = true
}

func (t *timelineCursor) SetBounds(timelineRect, labelRect image.Rectangle) {
	t.bounds = timelineRect
	t.labelRect = labelRect
}

// ProcessInput processes input.
func (t *timelineCursor) ProcessInput(input *view.Input) {
	t.mousePos = input.MousePos
}

func (t *timelineCursor) Render(fudge float32) {
	if !t.renderable {
		return
	}

	if !t.mousePos.WithinX(t.bounds) {
		return
	}

	_, ts := tradingSessionAtX(t.data.TradingSessionSeries.TradingSessions, t.bounds, t.mousePos.X)
	text := ts.Date.Format(t.layout)
	textSize := axisLabelTextRenderer.Measure(text)

	textPt := image.Point{
		X: t.mousePos.X - textSize.X/2,
		Y: t.labelRect.Min.Y + t.labelRect.Dy()/2 - textSize.Y/2,
	}
	bubbleRect := image.Rectangle{
		Min: textPt,
		Max: textPt.Add(textSize),
	}
	bubbleRect = bubbleRect.Inset(-axisLabelPadding)

	axisLabelBubble.SetBounds(bubbleRect)
	axisLabelBubble.Render(fudge)
	axisLabelTextRenderer.Render(text, textPt, gfx.TextColor(view.White))
}

func (t *timelineCursor) Close() {
	t.renderable = false
}
