package ponzi

import (
	"image"
	"time"

	"github.com/btmura/ponzi2/internal/gfx"
)

// ChartLines shows weekly lines for a single stock.
type ChartLines struct {
	stock               *ModelStock
	lastStockUpdateTime time.Time
	renderable          bool
	weekLines           *gfx.VAO
}

// NewChartLines creates a new chart lines instance.
func NewChartLines(stock *ModelStock) *ChartLines {
	return &ChartLines{
		stock: stock,
	}
}

// Update updates the chart lines.
func (ch *ChartLines) Update() {
	if ch.lastStockUpdateTime == ch.stock.lastUpdateTime {
		return
	}
	ch.lastStockUpdateTime = ch.stock.lastUpdateTime
	if ch.weekLines != nil {
		ch.weekLines.Delete()
	}
	ch.weekLines = createChartWeekLinesVAO(ch.stock.dailySessions)
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

		_, pwk := ds[i-1].date.ISOWeek()
		_, wk := s.date.ISOWeek()
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
	gfx.SetColorMixAmount(1)
	gfx.SetModelMatrixRect(r)
	ch.weekLines.Render()
}

// Close frees the resources backing the chart lines.
func (ch *ChartLines) Close() {
	ch.renderable = false
	ch.weekLines.Delete()
}
