package chart

import (
	"image"
	"time"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view"
	"github.com/btmura/ponzi2/internal/logger"
)

// longTime is a time that takes the most display width for measuring purposes.
var longTime = time.Date(2019, time.December, 31, 23, 59, 0, 0, time.UTC)

// timelineAxis renders the time labels for a single stock.
type timelineAxis struct {
	// renderable is true if this should be rendered.
	renderable bool

	// interval is the interval of the data being presented.
	interval model.Interval

	// MaxLabelSize is the maximum label size useful for rendering measurements.
	MaxLabelSize image.Point

	// labels bundle rendering measurements for time labels.
	labels []timelineLabel

	// timelineRect is the rectangle with global coords that should be drawn within.
	timelineRect image.Rectangle
}

type timelineAxisData struct {
	Interval             model.Interval
	TradingSessionSeries *model.TradingSessionSeries
}

func (t *timelineAxis) SetData(data timelineAxisData) {
	// Reset everything.
	t.Close()

	// Bail out if there is no data yet.
	ts := data.TradingSessionSeries
	if ts == nil {
		return
	}

	t.interval = data.Interval

	txt := timelineLabelText(t.interval, longTime)
	t.MaxLabelSize = axisLabelTextRenderer.Measure(txt)

	t.labels = makeTimelineLabels(t.interval, ts.TradingSessions)

	t.renderable = true
}

func (t *timelineAxis) SetBounds(timelineRect image.Rectangle) {
	t.timelineRect = timelineRect
}

func (t *timelineAxis) Render(float32) {
	if !t.renderable {
		return
	}

	r := t.timelineRect
	for _, l := range t.labels {
		tp := image.Point{
			X: r.Min.X + int(float32(r.Dx())*l.percent) - l.size.X/2,
			Y: r.Min.Y + r.Dy()/2 - l.size.Y/2,
		}
		axisLabelTextRenderer.Render(l.text, tp, gfx.TextColor(view.White))
	}
}

func (t *timelineAxis) Close() {
	t.renderable = false
}

type timelineLabel struct {
	percent float32
	text    string
	size    image.Point
}

func timelineLabelText(interval model.Interval, t time.Time) string {
	switch interval {
	case model.Intraday:
		return t.Format("3:04")
	case model.Daily:
		return t.Format("Jan")
	default:
		logger.Errorf("bad interval: %v", interval)
		return ""
	}
}

func makeTimelineLabels(interval model.Interval, ts []*model.TradingSession) []timelineLabel {
	var ls []timelineLabel

	for i := range ts {
		// Skip if we can't check the previous value.
		if i == 0 {
			continue
		}

		// Skip if the values being printed aren't changing.
		switch interval {
		case model.Intraday:
			prev := ts[i-1].Date.Hour()
			curr := ts[i].Date.Hour()
			if prev == curr {
				continue
			}

		case model.Daily:
			pm := ts[i-1].Date.Month()
			m := ts[i].Date.Month()
			if pm == m {
				continue
			}

		default:
			logger.Errorf("bad interval: %v", interval)
			return nil
		}

		// Generate the label text and its position.

		txt := timelineLabelText(interval, ts[i].Date)

		ls = append(ls, timelineLabel{
			percent: float32(i) / float32(len(ts)),
			text:    txt,
			size:    axisLabelTextRenderer.Measure(txt),
		})
	}

	return ls
}
