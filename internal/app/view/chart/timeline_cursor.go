package chart

import (
	"image"
	"math"
	"time"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view"
	"github.com/btmura/ponzi2/internal/app/view/color"
	"github.com/btmura/ponzi2/internal/log"
)

// timelineCursor renders the time corresponding to the mouse pointer
// on the x-axis.
type timelineCursor struct {
	// renderable is true if this should be rendered.
	renderable bool

	// dates are session dates shown for the cursor.
	dates []time.Time

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

	t.dates = nil
	for _, s := range ts.TradingSessions {
		t.dates = append(t.dates, s.Date)
	}

	switch data.Range {
	case model.OneDay:
		t.layout = "03:04"
	case model.OneYear:
		t.layout = "1/2/06"
	default:
		log.Errorf("bad range: %v", data.Range)
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

	if t.mousePos == nil {
		return
	}

	if t.mousePos.X < t.bounds.Min.X || t.mousePos.X > t.bounds.Max.X {
		return
	}

	percent := float32(t.mousePos.X-t.bounds.Min.X) / float32(t.bounds.Dx())

	i := int(math.Floor(float64(len(t.dates))*float64(percent) + 0.5))
	if i >= len(t.dates) {
		i = len(t.dates) - 1
	}

	text := t.dates[i].Format(t.layout)
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
	axisLabelTextRenderer.Render(text, textPt, gfx.TextColor(color.White))
}

func (t *timelineCursor) Close() {
	t.renderable = false
}
