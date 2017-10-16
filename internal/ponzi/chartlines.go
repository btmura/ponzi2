package ponzi

import (
	"image"

	"github.com/btmura/ponzi2/internal/gfx"
)

// ChartLines renders the weekly lines for a single stock.
type ChartLines struct {
	renderable bool
	weekLines  *gfx.VAO
}

// NewChartLines creates a new ChartLines.
func NewChartLines() *ChartLines {
	return &ChartLines{}
}

// SetStock sets the ChartLines' stock.
func (ch *ChartLines) SetStock(st *ModelStock) {
	// Reset everything.
	ch.Close()

	// Bail out if there is no data yet.
	if st.LastUpdateTime.IsZero() {
		return // Stock has no data yet.
	}

	ch.weekLines = createChartWeekLinesVAO(st.DailySessions)
	ch.renderable = true
}

func createChartWeekLinesVAO(ds []*ModelTradingSession) *gfx.VAO {
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

// Render renders the chart lines.
func (ch *ChartLines) Render(r image.Rectangle) {
	if !ch.renderable {
		return
	}

	gfx.SetModelMatrixRect(r)
	ch.weekLines.Render()
}

// Close frees the resources backing the chart lines.
func (ch *ChartLines) Close() {
	ch.renderable = false
	if ch.weekLines != nil {
		ch.weekLines.Delete()
		ch.weekLines = nil
	}
}
