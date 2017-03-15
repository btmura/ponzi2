package ponzi

import (
	"image"

	"github.com/go-gl/gl/v4.5-core/gl"
)

type chartLines struct {
	stock     *modelStock
	weekLines *vao
}

func createChartLines(stock *modelStock) *chartLines {
	return &chartLines{
		stock: stock,
	}
}

func (ch *chartLines) update() {
	if ch == nil || ch.stock.dailySessions == nil {
		return
	}

	ch.weekLines = ch.createWeekLineVAO()
}

func (ch *chartLines) createWeekLineVAO() *vao {
	var vertices []float32
	var colors []float32
	var indices []uint16

	// Amount to move on X-axis for one session.
	dx := 2.0 / float32(len(ch.stock.dailySessions))

	// Render lines whenever the week number changes.
	for i, s := range ch.stock.dailySessions {
		if i == 0 {
			continue // Can't check previous week.
		}

		_, pwk := ch.stock.dailySessions[i-1].date.ISOWeek()
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

	return createVAO(gl.LINES, vertices, colors, indices)
}

func (ch *chartLines) render(r image.Rectangle) {
	gl.Uniform1f(colorMixAmountLocation, 1)
	setModelMatrixRectangle(r)
	ch.weekLines.render()
}

func (ch *chartLines) close() {
	if ch == nil {
		return
	}

	ch.weekLines.close()
	ch.weekLines = nil
}
