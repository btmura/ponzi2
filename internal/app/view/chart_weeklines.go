package view

import (
	"image"

	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/gfx"
)

// chartWeekLines renders the weekly lines for a single stock.
type chartWeekLines struct {
	vao *gfx.VAO
}

func newChartWeekLines() *chartWeekLines {
	return &chartWeekLines{}
}

func (ch *chartWeekLines) SetData(ts *model.TradingSessionSeries) {
	// Reset everything.
	ch.Close()

	// Bail out if there is no data yet.
	if ts == nil {
		return
	}

	// Create the line VAO.
	ch.vao = verticalRuleSetVAO(weeklineXValues(ts.TradingSessions), [2]float32{0, 1}, gray)
}

func weeklineXValues(ts []*model.TradingSession) []float32 {
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

func createChartWeekLinesVAO(ds []*model.TradingSession) *gfx.VAO {
	data := &gfx.VAOVertexData{Mode: gfx.Lines}
	var v uint16 // vertex index

	dx := 2.0 / float32(len(ds)) // (-1 to 1) on X-axis
	calcX := func(i int) float32 {
		return -1.0 + dx*float32(i)
	}

	// Render lines whenever the week number changes.
	for i, s := range ds {
		if i == 0 {
			continue // Can't check previous week.
		}

		_, pwk := ds[i-1].Date.ISOWeek()
		_, wk := s.Date.ISOWeek()
		if pwk == wk {
			continue
		}

		data.Vertices = append(data.Vertices,
			calcX(i), -1, 0,
			calcX(i), +1, 0,
		)
		data.Colors = append(data.Colors,
			gray[0], gray[1], gray[2],
			gray[0], gray[1], gray[2],
		)
		data.Indices = append(data.Indices, v, v+1)
		v += 2
	}

	return gfx.NewVAO(data)
}

// Render renders the chart lines.
func (ch *chartWeekLines) Render(r image.Rectangle) {
	if ch.vao == nil {
		return
	}
	gfx.SetModelMatrixRect(r)
	ch.vao.Render()
}

// Close frees the resources backing the chart lines.
func (ch *chartWeekLines) Close() {
	if ch.vao != nil {
		ch.vao.Delete()
		ch.vao = nil
	}
}
