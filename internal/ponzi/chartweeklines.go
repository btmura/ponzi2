package ponzi

import (
	"image"

	"github.com/btmura/ponzi2/internal/gfx"
)

// ChartWeekLines renders the weekly lines for a single stock.
type ChartWeekLines struct {
	renderable bool
	lines      *gfx.VAO
}

// NewChartWeekLines creates a new ChartWeekLines.
func NewChartWeekLines() *ChartWeekLines {
	return &ChartWeekLines{}
}

// SetStock sets the ChartWeekLines' stock.
func (ch *ChartWeekLines) SetStock(st *ModelStock) {
	// Reset everything.
	ch.Close()

	// Bail out if there is no data yet.
	if st.LastUpdateTime.IsZero() {
		return // Stock has no data yet.
	}

	ch.lines = createChartWeekLinesVAO(st.DailySessions)
	ch.renderable = true
}

func createChartWeekLinesVAO(ds []*ModelTradingSession) *gfx.VAO {
	data := &gfx.VAOVertexData{}
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

	return gfx.NewVAO(gfx.Lines, data)
}

// Render renders the chart lines.
func (ch *ChartWeekLines) Render(r image.Rectangle) {
	if !ch.renderable {
		return
	}

	gfx.SetModelMatrixRect(r)
	ch.lines.Render()
}

// Close frees the resources backing the chart lines.
func (ch *ChartWeekLines) Close() {
	ch.renderable = false
	if ch.lines != nil {
		ch.lines.Delete()
		ch.lines = nil
	}
}
