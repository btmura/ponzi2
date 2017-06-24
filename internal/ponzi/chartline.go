package ponzi

import (
	"image"
	"time"

	"github.com/go-gl/gl/v4.5-core/gl"

	"github.com/btmura/ponzi2/internal/gfx"
)

type chartLines struct {
	stock               *modelStock
	lastStockUpdateTime time.Time
	renderable          bool
	weekLines           *gfx.VAO
}

func createChartLines(stock *modelStock) *chartLines {
	return &chartLines{
		stock: stock,
	}
}

func (ch *chartLines) update() {
	if ch.lastStockUpdateTime == ch.stock.lastUpdateTime {
		return
	}
	ch.lastStockUpdateTime = ch.stock.lastUpdateTime
	ch.weekLines.Close()
	ch.weekLines = createChartWeekLinesVAO(ch.stock.dailySessions)
	ch.renderable = true
}

func createChartWeekLinesVAO(ds []*modelTradingSession) *gfx.VAO {
	var vertices []float32
	var colors []float32
	var indices []uint16

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
		vertices = append(vertices,
			x, -1,
			x, +1,
		)
		colors = append(colors,
			gray[0], gray[1], gray[2],
			gray[0], gray[1], gray[2],
		)
		indices = append(indices,
			uint16(len(vertices)-1),
			uint16(len(vertices)-2),
		)
	}

	return gfx.CreateVAO(gl.LINES, vertices, colors, indices)
}

func (ch *chartLines) render(r image.Rectangle) {
	if !ch.renderable {
		return
	}
	gl.Uniform1f(colorMixAmountLocation, 1)
	setModelMatrixRectangle(r)
	ch.weekLines.Render()
}

func (ch *chartLines) close() {
	ch.renderable = false
	ch.weekLines.Close()
	ch.weekLines = nil
}
