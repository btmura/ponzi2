package ponzi

import (
	"image"

	"github.com/go-gl/gl/v4.5-core/gl"
)

type chart struct {
	vao   uint32
	count int32
}

func createChart(sessions []*modelTradingSession) *chart {
	ns := len(sessions)
	ws := 2.0 / float32(ns) // -1 to 1 on X-axis

	x := -1.0 + ws/2.0
	var vertices []float32
	var indices []uint16
	for i := 0; i < ns; i++ {
		vertices = append(vertices, x, -1, x, 1)
		indices = append(indices, uint16(i*2), uint16(i*2+1))
		x += ws
	}

	vbo := createArrayBuffer(vertices)
	ibo := createElementArrayBuffer(indices)

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.EnableVertexAttribArray(positionLocation)
	gl.VertexAttribPointer(positionLocation, 2, gl.FLOAT, false, 0, gl.PtrOffset(0))

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ibo)
	gl.BindVertexArray(0)

	return &chart{
		vao:   vao,
		count: int32(len(indices)),
	}
}

func (c *chart) render(r image.Rectangle) {
	m := newScaleMatrix(float32(r.Dx()/2), float32(r.Dy()/2), 1)
	m = m.mult(newTranslationMatrix(float32(r.Min.X+r.Dx()/2), float32(r.Min.Y+r.Dy()/2), 0))
	gl.UniformMatrix4fv(modelMatrixLocation, 1, false, &m[0])
	gl.BindTexture(gl.TEXTURE_2D, 0 /* dummy texture */)

	gl.BindVertexArray(c.vao)
	gl.DrawElements(gl.LINES, c.count, gl.UNSIGNED_SHORT, gl.Ptr(nil))
	gl.BindVertexArray(0)
}
