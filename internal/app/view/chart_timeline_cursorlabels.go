package view

import (
	"image"
	"math"
	"time"

	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/status"
)

type chartTimelineCursorLabels struct {
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

func newChartTimelineCursorLabels() *chartTimelineCursorLabels {
	return &chartTimelineCursorLabels{}
}

func (ch *chartTimelineCursorLabels) SetData(r model.Range, ts *model.TradingSessionSeries) error {
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
		return status.Errorf("bad range: %v", r)
	}

	return nil
}

// ProcessInput processes input.
func (ch *chartTimelineCursorLabels) ProcessInput(ic inputContext, labelRect image.Rectangle) {
	ch.bounds = ic.Bounds
	ch.labelRect = labelRect
	ch.mousePos = ic.MousePos
}

func (ch *chartTimelineCursorLabels) Render(fudge float32) error {
	if ch.mousePos.X < ch.bounds.Min.X || ch.mousePos.X > ch.bounds.Max.X {
		return nil
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

	renderBubble(tp, size, chartAxisLabelBubbleSpec)
	chartAxisLabelTextRenderer.Render(text, tp, white)

	return nil
}
