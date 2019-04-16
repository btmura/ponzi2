package view

import (
	"image"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/status"
)

type chartTimeline struct {
	vao    *gfx.VAO
	bounds image.Rectangle
}

func newChartTimeline() *chartTimeline {
	return &chartTimeline{}
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

	// Create the line VAO.
	ch.vao = vertRuleSetVAO(vals, [2]float32{0, 1}, gray)

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
			return nil, status.Errorf("bad range: %v", r)
		}

		values = append(values, float32(i)/float32(len(ts)))
	}

	return values, nil
}

// ProcessInput processes input.
func (ch *chartTimeline) ProcessInput(ic inputContext) {
	ch.bounds = ic.Bounds
}

// Render renders the chart lines.
func (ch *chartTimeline) Render(fudge float32) {
	if ch.vao == nil {
		return
	}
	gfx.SetModelMatrixRect(ch.bounds)
	ch.vao.Render()
}

// Close frees the resources backing the chart lines.
func (ch *chartTimeline) Close() {
	if ch.vao != nil {
		ch.vao.Delete()
		ch.vao = nil
	}
}
