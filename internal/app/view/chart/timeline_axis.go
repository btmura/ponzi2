package chart

import (
	"image"
	"time"

	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view/color"
	"github.com/btmura/ponzi2/internal/errors"
)

// longTime is a time that takes the most display width for measuring purposes.
var longTime = time.Date(2019, time.December, 31, 23, 59, 0, 0, time.UTC)

// timelineAxis renders the time labels for a single stock.
type timelineAxis struct {
	// renderable is true if this should be rendered.
	renderable bool

	// dataRange is range of the data being presented.
	dataRange model.Range

	// MaxLabelSize is the maximum label size useful for rendering measurements.
	MaxLabelSize image.Point

	// labels bundle rendering measurements for time labels.
	labels []timelineLabel

	// timelineRect is the rectangle with global coords that should be drawn within.
	timelineRect image.Rectangle
}

func (t *timelineAxis) SetData(r model.Range, ts *model.TradingSessionSeries) error {
	// Reset everything.
	t.Close()

	// Bail out if there is no data yet.
	if ts == nil {
		return nil
	}

	t.dataRange = r

	txt, err := timelineLabelText(r, longTime)
	if err != nil {
		return err
	}
	t.MaxLabelSize = chartAxisLabelTextRenderer.Measure(txt)

	labels, err := makeTimelineLabels(r, ts.TradingSessions)
	if err != nil {
		return err
	}
	t.labels = labels

	return nil
}

// ProcessInput processes input.
func (t *timelineAxis) ProcessInput(timelineRect image.Rectangle) {
	t.timelineRect = timelineRect
}

func (t *timelineAxis) Render(fudge float32) {
	if !t.renderable {
		return
	}

	r := t.timelineRect
	for _, l := range t.labels {
		tp := image.Point{
			X: r.Min.X + int(float32(r.Dx())*l.percent) - l.size.X/2,
			Y: r.Min.Y + r.Dy()/2 - l.size.Y/2,
		}
		chartAxisLabelTextRenderer.Render(l.text, tp, color.White)
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

func timelineLabelText(r model.Range, t time.Time) (string, error) {
	switch r {
	case model.OneDay:
		return t.Format("3:04"), nil
	case model.OneYear:
		return t.Format("Jan"), nil
	default:
		return "", errors.Errorf("bad range: %v", r)
	}
}

func makeTimelineLabels(r model.Range, ts []*model.TradingSession) ([]timelineLabel, error) {
	var ls []timelineLabel

	for i := range ts {
		// Skip if we can't check the previous value.
		if i == 0 {
			continue
		}

		// Skip if the values being printed aren't changing.
		switch r {
		case model.OneDay:
			prev := ts[i-1].Date.Hour()
			curr := ts[i].Date.Hour()
			if prev == curr {
				continue
			}

		case model.OneYear:
			pm := ts[i-1].Date.Month()
			m := ts[i].Date.Month()
			if pm == m {
				continue
			}

		default:
			return nil, errors.Errorf("bad range: %v", r)
		}

		// Generate the label text and its position.

		txt, err := timelineLabelText(r, ts[i].Date)
		if err != nil {
			return nil, err
		}

		ls = append(ls, timelineLabel{
			percent: float32(i) / float32(len(ts)),
			text:    txt,
			size:    chartAxisLabelTextRenderer.Measure(txt),
		})
	}

	return ls, nil
}
