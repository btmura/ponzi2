package gfx

import (
	"github.com/go-gl/gl/v4.5-core/gl"

	gl2 "github.com/btmura/ponzi2/internal/gl"
)

// VAO is a Vertex Array Object (VAO) for one element type (lines, triangles, etc).
type VAO struct {
	array uint32 // array is the VAO name for gl.BindVertexArray.
	mode  uint32 // mode is like gl.LINES or gl.TRIANGLES passed to gl.DrawElements.
	count int32  // count is the number of elements for gl.DrawElements.
}

func CreateVAO(mode uint32, vertices, colors []float32, indices []uint16) *VAO {
	if len(vertices) == 0 || len(colors) == 0 || len(indices) == 0 {
		return nil // Can't create empty buffer objects. Bail out if nothing to render.
	}

	vbo := gl2.ArrayBuffer(vertices)
	cbo := gl2.ArrayBuffer(colors)
	ibo := gl2.ElementArrayBuffer(indices)

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

	return &VAO{
		array: array,
		mode:  mode,
		count: int32(len(indices)),
	}
}

func (v *VAO) Render() {
	if v == nil {
		return
	}
	gl.BindVertexArray(v.array)
	gl.DrawElements(v.mode, v.count, gl.UNSIGNED_SHORT, gl.Ptr(nil))
	gl.BindVertexArray(0)
}

func (v *VAO) Close() {
	if v == nil {
		return
	}
	gl.DeleteVertexArrays(1, &v.array)
}
