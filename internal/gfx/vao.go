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

func CreateStrokedRectVAO(ulColor, urColor, blColor, brColor [3]float32) *VAO {
	return CreateVAO(
		gl.LINES,
		[]float32{
			-1, +1, // UL - 0
			+1, +1, // UR - 1
			-1, -1, // BL - 2
			+1, -1, // BR - 3
		},
		[]float32{
			ulColor[0], ulColor[1], ulColor[2],
			urColor[0], urColor[1], urColor[2],
			blColor[0], blColor[1], blColor[2],
			brColor[0], brColor[1], brColor[2],
		},
		[]uint16{
			0, 1,
			1, 3,
			3, 2,
			2, 0,
		},
	)
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
