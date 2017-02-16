package ponzi

import (
	"image"

	"github.com/go-gl/gl/v4.5-core/gl"
)

// vaoLine is a Vertex Array Object (VAO) for a line segment from (-1, 0) to (1, 0).
type vaoLine struct {
	vao   uint32 // vao is the VAO name for gl.BindVertexArray.
	count int32  // count is the number of elements for gl.DrawElements.
}

func createVAOLine(lColor, rColor [3]float32) *vaoLine {
	vertices := []float32{
		-1, 0,
		1, 0,
	}

	colors := []float32{
		lColor[0], lColor[1], lColor[2],
		rColor[0], rColor[1], rColor[2],
	}

	indices := []uint16{
		0, 1,
	}

	vbo := createArrayBuffer(vertices)
	cbo := createArrayBuffer(colors)
	ibo := createElementArrayBuffer(indices)

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	{
		gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
		gl.EnableVertexAttribArray(positionLocation)
		gl.VertexAttribPointer(positionLocation, 2, gl.FLOAT, false, 0, gl.PtrOffset(0))

		gl.BindBuffer(gl.ARRAY_BUFFER, cbo)
		gl.EnableVertexAttribArray(colorLocation)
		gl.VertexAttribPointer(colorLocation, 3, gl.FLOAT, false, 0, gl.PtrOffset(0))

		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ibo)
	}
	gl.BindVertexArray(0)

	return &vaoLine{
		vao:   vao,
		count: int32(len(indices)),
	}
}

func (v *vaoLine) render(r image.Rectangle) {
	setModelMatrixRectangle(r)
	gl.BindVertexArray(v.vao)
	gl.DrawElements(gl.LINES, v.count, gl.UNSIGNED_SHORT, gl.Ptr(nil))
	gl.BindVertexArray(0)
}

func (v *vaoLine) close() {
	gl.DeleteVertexArrays(1, &v.vao)
}
