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

func createLineVAO(lColor, rColor [3]float32) *vao {
	return createVAO(
		gl.LINES,
		[]float32{
			-1, 0, // L
			+1, 0, // R
		},
		[]float32{
			lColor[0], lColor[1], lColor[2],
			rColor[0], rColor[1], rColor[2],
		},
		[]uint16{
			0, 1,
		},
	)
}

func createStrokedRectVAO(ulColor, urColor, blColor, brColor [3]float32) *vao {
	return createVAO(
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

func createFilledRectVAO(ulColor, urColor, blColor, brColor [3]float32) *vao {
	return createVAO(
		gl.TRIANGLES,
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
			0, 2, 1,
			1, 2, 3,
		},
	)
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
