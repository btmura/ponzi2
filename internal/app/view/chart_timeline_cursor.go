package view

import (
	"image"
	"math"
	"time"

	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view/rect"
	"github.com/btmura/ponzi2/internal/errors"
)

// chartTimelineCursor renders the time corresponding to the mouse pointer
// on the x-axis.
type chartTimelineCursor struct {
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

func (ch *chartTimelineCursor) SetData(r model.Range, ts *model.TradingSessionSeries) error {
	// Reset everything.
	ch.Close()

	// Bail out if there is no data yet.
	if ts == nil {
		return nil
	}

	ch.dates = nil
	for _, s := range ts.TradingSessions {
		ch.dates = append(ch.dates, s.Date)
	}

	switch r {
	case model.OneDay:
		ch.layout = "03:04"
	case model.OneYear:
		ch.layout = "1/2/06"
	default:
		return errors.Errorf("bad range: %v", r)
	}

	ch.renderable = true

	return nil
}

// ProcessInput processes input.
func (ch *chartTimelineCursor) ProcessInput(timelineRect, labelRect image.Rectangle, mousePos image.Point) {
	ch.bounds = timelineRect
	ch.labelRect = labelRect
	ch.mousePos = mousePos
}

func (ch *chartTimelineCursor) Render(fudge float32) {
	if !ch.renderable {
		return
	}

	if ch.mousePos.X < ch.bounds.Min.X || ch.mousePos.X > ch.bounds.Max.X {
		return
	}

	percent := float32(ch.mousePos.X-ch.bounds.Min.X) / float32(ch.bounds.Dx())

	i := int(math.Floor(float64(len(ch.dates))*float64(percent) + 0.5))
	if i >= len(ch.dates) {
		i = len(ch.dates) - 1
	}

	text := ch.dates[i].Format(ch.layout)
	size := chartAxisLabelTextRenderer.Measure(text)

	tp := image.Point{
		X: ch.mousePos.X - size.X/2,
		Y: ch.labelRect.Min.Y + ch.labelRect.Dy()/2 - size.Y/2,
	}

	rect.RenderBubble(tp, size, chartAxisLabelBubbleSpec)
	chartAxisLabelTextRenderer.Render(text, tp, white)
}

func (ch *chartTimelineCursor) Close() {
	ch.renderable = false
}
