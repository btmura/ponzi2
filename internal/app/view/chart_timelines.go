package view

import (
	"fmt"
	"image"

	"gitlab.com/btmura/ponzi2/internal/app/gfx"
	"gitlab.com/btmura/ponzi2/internal/app/model"
)

type chartTimeLines struct {
	vao *gfx.VAO
}

func newChartTimeLines() *chartTimeLines {
	return &chartTimeLines{}
}

func (ch *chartTimeLines) SetData(r model.Range, ts *model.TradingSessionSeries) error {
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
			return nil, fmt.Errorf("bad range: %v", r)
		}

		values = append(values, float32(i)/float32(len(ts)))
	}

	return values, nil
}

// Render renders the chart lines.
func (ch *chartTimeLines) Render(r image.Rectangle) {
	if ch.vao == nil {
		return
	}
	gfx.SetModelMatrixRect(r)
	ch.vao.Render()
}

// Close frees the resources backing the chart lines.
func (ch *chartTimeLines) Close() {
	if ch.vao != nil {
		ch.vao.Delete()
		ch.vao = nil
	}
}
