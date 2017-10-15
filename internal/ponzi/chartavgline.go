package ponzi

import (
	"image"

	"github.com/btmura/ponzi2/internal/gfx"
)

// ChartAvgLine renders a moving average line for a single stock.
type ChartAvgLine struct {
	renderable bool
	lines      *gfx.VAO
}

// SetStock sets the ChartAvgLine's stock.
func (ch *ChartAvgLine) SetStock(st *ModelStock) {
	// Reset everything.
	ch.Close()

	// Bail out if there is no data yet.
	if st.LastUpdateTime.IsZero() {
		return // Stock has no data yet.
	}

	ch.lines = createChartAvgLinesVAO(st.DailySessions)
	ch.renderable = true
}

func createChartAvgLinesVAO(ds []*ModelTradingSession) *gfx.VAO {
	data := &gfx.VAOVertexData{}

	// Amount to move on X-axis for one session.
	dx := 2.0 / float32(len(ds))

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

		x := -1 + dx*float32(i)
		data.Vertices = append(data.Vertices,
			x, -1, 0,
			x, +1, 0,
		)
		data.Colors = append(data.Colors,
			gray[0], gray[1], gray[2],
			gray[0], gray[1], gray[2],
		)
		data.Indices = append(data.Indices,
			uint16(len(data.Vertices)-1),
			uint16(len(data.Vertices)-2),
		)
	}

	return gfx.NewVAO(gfx.Lines, data)
}

// Render renders the ChartAvgLine.
func (ch *ChartAvgLine) Render(r image.Rectangle) {
	if !ch.renderable {
		return
	}

	gfx.SetModelMatrixRect(r)
	ch.lines.Render()
}

// Close frees the resources backing the ChartAvgLine.
func (ch *ChartAvgLine) Close() {
	ch.renderable = false
	if ch.lines != nil {
		ch.lines.Delete()
		ch.lines = nil
	}
}
