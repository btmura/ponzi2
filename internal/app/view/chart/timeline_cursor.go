package chart

import (
	"image"
	"math"
	"time"

	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view/color"
	"github.com/btmura/ponzi2/internal/app/view/rect"
	"github.com/btmura/ponzi2/internal/errors"
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

	// mousePos is the current mouse position.
	mousePos image.Point
}

func (t *timelineCursor) SetData(r model.Range, ts *model.TradingSessionSeries) error {
	// Reset everything.
	t.Close()

	// Bail out if there is no data yet.
	if ts == nil {
		return nil
	}

	t.dates = nil
	for _, s := range ts.TradingSessions {
		t.dates = append(t.dates, s.Date)
	}

	switch r {
	case model.OneDay:
		t.layout = "03:04"
	case model.OneYear:
		t.layout = "1/2/06"
	default:
		return errors.Errorf("bad range: %v", r)
	}

	t.renderable = true

	return nil
}

// ProcessInput processes input.
func (t *timelineCursor) ProcessInput(timelineRect, labelRect image.Rectangle, mousePos image.Point) {
	t.bounds = timelineRect
	t.labelRect = labelRect
	t.mousePos = mousePos
}

func (t *timelineCursor) Render(fudge float32) {
	if !t.renderable {
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
	size := axisLabelTextRenderer.Measure(text)

	tp := image.Point{
		X: t.mousePos.X - size.X/2,
		Y: t.labelRect.Min.Y + t.labelRect.Dy()/2 - size.Y/2,
	}

	rect.RenderBubble(tp, size, axisLabelBubbleSpec)
	axisLabelTextRenderer.Render(text, tp, color.White)
}

func (t *timelineCursor) Close() {
	t.renderable = false
}
