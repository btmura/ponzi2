package chart

import (
	"image"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view/vao"
	"github.com/btmura/ponzi2/internal/errors"
)

type chartTimeline struct {
	// renderable is true if this is ready to be rendered.
	renderable bool

	// lineVAO has the vertical lines to be rendered under some technicals.
	lineVAO *gfx.VAO

	// lineRect is the rectangle to draw the lines within.
	lineRect image.Rectangle
}

func (ch *chartTimeline) SetData(r model.Range, ts *model.TradingSessionSeries) error {
	// Reset everything.
	ch.Close()

	// Bail out if there is no data yet.
	if ts == nil {
		return nil
	}

	vals, err := weekLineValues(r, ts.TradingSessions)
	if err != nil {
		return err
	}

	ch.lineVAO = vao.VertRuleSet(vals, [2]float32{0, 1}, gray)

	ch.renderable = true

	return nil
}

func weekLineValues(r model.Range, ts []*model.TradingSession) ([]float32, error) {
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
			return nil, errors.Errorf("bad range: %v", r)
		}

		values = append(values, float32(i)/float32(len(ts)))
	}

	return values, nil
}

// ProcessInput processes input.
func (ch *chartTimeline) ProcessInput(lineRect image.Rectangle) {
	ch.lineRect = lineRect
}

// Render renders the chart lines.
func (ch *chartTimeline) Render(fudge float32) {
	if !ch.renderable {
		return
	}

	gfx.SetModelMatrixRect(ch.lineRect)
	ch.lineVAO.Render()
}

// Close frees the resources backing the chart lines.
func (ch *chartTimeline) Close() {
	ch.renderable = false
	if ch.lineVAO != nil {
		ch.lineVAO.Delete()
		ch.lineVAO = nil
	}
}
