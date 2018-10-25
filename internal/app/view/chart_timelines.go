package view

import (
	"image"

	"gitlab.com/btmura/ponzi2/internal/app/model"
	"gitlab.com/btmura/ponzi2/internal/app/gfx"
)

type chartTimeLines struct {
	vao *gfx.VAO
}

func newChartTimeLines() *chartTimeLines {
	return &chartTimeLines{}
}

func (ch *chartTimeLines) SetData(ts *model.TradingSessionSeries) {
	// Reset everything.
	ch.Close()

	// Bail out if there is no data yet.
	if ts == nil {
		return
	}

	// Create the line VAO.
	ch.vao = vertRuleSetVAO(weekLineValues(ts.TradingSessions), [2]float32{0, 1}, gray)
}

func weekLineValues(ts []*model.TradingSession) []float32 {
	var values []float32
	for i := range ts {
		if i == 0 {
			continue // Can't check previous week.
		}

		_, pwk := ts[i-1].Date.ISOWeek()
		_, wk := ts[i].Date.ISOWeek()
		if pwk == wk {
			continue
		}

		values = append(values, float32(i)/float32(len(ts)))
	}
	return values
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
