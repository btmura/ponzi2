package chart

import (
	"image"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view/color"
	"github.com/btmura/ponzi2/internal/app/view/vao"
	"github.com/btmura/ponzi2/internal/log"
)

// timeline renders the vertical lines behind a chart's technicals.
type timeline struct {
	// topColor is the color of the line at the top.
	topColor color.RGBA

	// bottomColor is the color of the line at the bottom.
	bottomColor color.RGBA

	// renderable is true if this is ready to be rendered.
	renderable bool

	// lineVAO has the vertical lines to be rendered under some technicals.
	lineVAO *gfx.VAO

	// lineRect is the rectangle to draw the lines within.
	lineRect image.Rectangle
}

func newTimeline(topColor, bottomColor color.RGBA) *timeline {
	return &timeline{
		topColor:    topColor,
		bottomColor: bottomColor,
	}
}

type timelineData struct {
	Range                model.Range
	TradingSessionSeries *model.TradingSessionSeries
}

func (t *timeline) SetData(data timelineData) {
	// Reset everything.
	t.Close()

	// Bail out if there is no data yet.
	ts := data.TradingSessionSeries
	if ts == nil {
		return
	}

	vals := weekLineValues(data.Range, ts.TradingSessions)

	t.lineVAO = vao.VertRuleSet(vals, [2]float32{0, 1}, t.bottomColor, t.topColor)

	t.renderable = true
}

func weekLineValues(r model.Range, ts []*model.TradingSession) []float32 {
	var values []float32

	for i := range ts {
		// Skip if we can't check the previous value.
		if i == 0 {
			continue
		}

		// Skip if the values used for the lines aren't changing.
		switch r {
		case model.OneDay:
			prev := ts[i-1].Date.Hour()
			curr := ts[i].Date.Hour()
			if prev == curr {
				continue
			}

		case model.OneYear:
			_, prev := ts[i-1].Date.ISOWeek()
			_, curr := ts[i].Date.ISOWeek()
			if prev == curr {
				continue
			}

		default:
			log.Errorf("bad range: %v", r)
		}

		values = append(values, float32(i)/float32(len(ts)))
	}

	return values
}

func (t *timeline) SetBounds(lineRect image.Rectangle) {
	t.lineRect = lineRect
}

// Render renders the chart lines.
func (t *timeline) Render(float32) {
	if !t.renderable {
		return
	}

	gfx.SetModelMatrixRect(t.lineRect)
	t.lineVAO.Render()
}

// Close frees the resources backing the chart lines.
func (t *timeline) Close() {
	t.renderable = false
	if t.lineVAO != nil {
		t.lineVAO.Delete()
		t.lineVAO = nil
	}
}
