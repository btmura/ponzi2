package ponzi

import (
	"image"

	"github.com/go-gl/gl/v4.5-core/gl"
)

// vao is a Vertex Array Object (VAO) for one element type (lines, triangles, etc).
type vao struct {
	array uint32 // array is the VAO name for gl.BindVertexArray.
	mode  uint32 // mode is like gl.LINES or gl.TRIANGLES passed to gl.DrawElements.
	count int32  // count is the number of elements for gl.DrawElements.
}

func createVAO(mode uint32, vertices, colors []float32, indices []uint16) *vao {
	vbo := createArrayBuffer(vertices)
	cbo := createArrayBuffer(colors)
	ibo := createElementArrayBuffer(indices)

	var array uint32
	gl.GenVertexArrays(1, &array)
	gl.BindVertexArray(array)
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

	return &vao{
		array: array,
		mode:  mode,
		count: int32(len(indices)),
	}
}

func (v *vao) render(r image.Rectangle) {
	setModelMatrixRectangle(r)
	gl.BindVertexArray(v.array)
	gl.DrawElements(v.mode, v.count, gl.UNSIGNED_SHORT, gl.Ptr(nil))
	gl.BindVertexArray(0)
}

func (v *vao) close() {
	gl.DeleteVertexArrays(1, &v.array)
}
