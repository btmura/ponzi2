package ponzi

import (
	"image"

	"github.com/go-gl/gl/v4.5-core/gl"
)

type chartLines struct {
	stock      *modelStock
	monthLines *vao
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

	var vertices []float32
	var colors []float32
	var indices []uint16

	dx := 2.0 / float32(len(ch.stock.dailySessions))
	calcX := func(i int) float32 {
		return -1 + dx*float32(i)
	}

	for i, s := range ch.stock.dailySessions {
		if i > 0 && ch.stock.dailySessions[i-1].date.Month() != s.date.Month() {
			x := calcX(i)
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
	}

	ch.monthLines = createVAO(gl.LINES, vertices, colors, indices)
}

func (ch *chartLines) render(r image.Rectangle) {
	gl.Uniform1f(colorMixAmountLocation, 1)
	setModelMatrixRectangle(r)
	ch.monthLines.render()
}

func (ch *chartLines) close() {
	if ch == nil {
		return
	}

	ch.monthLines.close()
	ch.monthLines = nil
}
