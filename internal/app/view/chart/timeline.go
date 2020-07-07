package chart

import (
	"image"
	"time"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view"
	"github.com/btmura/ponzi2/internal/app/view/vao"
	"github.com/btmura/ponzi2/internal/logger"
)

// timeline renders the vertical lines behind a chart's technicals.
type timeline struct {
	// majorTopColor is the top color of the major lines.
	majorTopColor view.Color

	// majorBottomColor is the bottom color of the major lines.
	majorBottomColor view.Color

	// minorTopColor is the top color of the minor lines.
	minorTopColor view.Color

	// minorBottomColor is the bottom color of the minor lines.
	minorBottomColor view.Color

	// majorLineVAO has the vertical lines to be rendered under some technicals.
	majorLineVAO *gfx.VAO

	// minorLineVAO has the vertical lines to be rendered under some technicals.
	minorLineVAO *gfx.VAO

	// renderable is true if this is ready to be rendered.
	renderable bool

	// bounds is the rectangle to draw the lines within.
	bounds image.Rectangle
}

func newTimeline(majorTopColor, majorBottomColor, minorTopColor, minorBottomColor view.Color) *timeline {
	return &timeline{
		majorTopColor:    majorTopColor,
		majorBottomColor: majorBottomColor,
		minorTopColor:    minorTopColor,
		minorBottomColor: minorBottomColor,
	}
}

type timelineData struct {
	Interval             model.Interval
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

	majorValues, minorValues := weekLineValues(data.Interval, ts.TradingSessions)
	t.majorLineVAO = vao.VertRuleSet(majorValues, [2]float32{0, 1}, t.majorBottomColor, t.majorTopColor)
	t.minorLineVAO = vao.VertRuleSet(minorValues, [2]float32{0, 1}, t.minorBottomColor, t.minorTopColor)

	t.renderable = true
}

func weekLineValues(interval model.Interval, ts []*model.TradingSession) (majorValues, minorValues []float32) {
	pendingMonthChanged := false
	for i := range ts {
		curr := ts[i].Date

		var prev time.Time
		if i == 0 {
			prev = curr.Add(-time.Hour * 24)
		} else {
			prev = ts[i-1].Date
		}

		ph := prev.Hour()
		ch := curr.Hour()
		hourChanged := ph != ch

		_, pw := prev.ISOWeek()
		_, cw := curr.ISOWeek()
		weekChanged := pw != cw

		if prev.Month() != curr.Month() {
			pendingMonthChanged = true
		}

		addMajor := func() {
			majorValues = append(majorValues, float32(i)/float32(len(ts)))
		}

		addMinor := func() {
			minorValues = append(minorValues, float32(i)/float32(len(ts)))
		}

		switch interval {
		case model.Intraday:
			if hourChanged {
				continue
			}

			addMinor()

		case model.Daily:
			if !weekChanged {
				continue
			}

			if pendingMonthChanged {
				addMajor()
				pendingMonthChanged = false
			} else {
				addMinor()
			}

		default:
			logger.Errorf("bad interval: %v", interval)
		}
	}

	return majorValues, minorValues
}

func (t *timeline) SetBounds(bounds image.Rectangle) {
	t.bounds = bounds
}

// Render renders the chart lines.
func (t *timeline) Render(float32) {
	if !t.renderable {
		return
	}

	gfx.SetModelMatrixRect(t.bounds)
	t.majorLineVAO.Render()
	t.minorLineVAO.Render()
}

// Close frees the resources backing the chart lines.
func (t *timeline) Close() {
	t.renderable = false
	if t.majorLineVAO != nil {
		t.majorLineVAO.Delete()
	}
	if t.minorLineVAO != nil {
		t.minorLineVAO.Delete()
	}
}
