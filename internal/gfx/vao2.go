package gfx

import (
	gl2 "github.com/btmura/ponzi2/internal/gl"
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/golang/glog"
)

// VAO2 is a Vertex Array Object (VAO) for one element type (lines, triangles, etc).
type VAO2 struct {
	array uint32 // array is the VAO name for gl.BindVertexArray.
	mode  uint32 // mode is like gl.LINES or gl.TRIANGLES passed to gl.DrawElements.
	count int32  // count is the number of elements for gl.DrawElements.
}

// VAOBufferData is a bunch of slices filled with vertex data to create a VAO.
type VAOBufferData struct {
	Vertices  []float32 // Vertices is a required slice of flattened (x, y, z) vertices.
	Normals   []float32 // Normals is an optional slice of flattened (nx, ny, nz) normals.
	TexCoords []float32 // TexCoords is an optional slice of flattened (s, t) coords.
	Colors    []float32 // Colors is an optional slice of flattened (r, g, b) values.
	Indices   []uint16  // Indices is a required slice of indices into all the buffers.
}

// VAOMode is analogous to the mode argument to gl.DrawElements like gl.LINES or gl.TRIANGLES.
type VAOMode int

// VAOMode enums.
const (
	Triangles VAOMode = iota
	Lines
)

// NewVAO creates a VAO out of the given data buffers and drawing mode.
func NewVAO(data *VAOBufferData, mode VAOMode) *VAO2 {
	if len(data.Vertices) == 0 || len(data.Indices) == 0 {
		return &VAO2{} // OpenGL doesn't allow empty buffer objects. Return VAO with zero count.
	}

	vbo := gl2.ArrayBuffer(data.Vertices)

	var nbo uint32
	if len(data.Normals) != 0 {
		nbo = gl2.ArrayBuffer(data.Normals)
	}

	var tbo uint32
	if len(data.TexCoords) != 0 {
		tbo = gl2.ArrayBuffer(data.TexCoords)
	}

	var cbo uint32
	if len(data.Colors) != 0 {
		cbo = gl2.ArrayBuffer(data.Colors)
	}

	ibo := gl2.ElementArrayBuffer(data.Indices)

	var array uint32

	gl.GenVertexArrays(1, &array)
	gl.BindVertexArray(array)
	{
		gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
		gl.EnableVertexAttribArray(positionLocation)
		gl.VertexAttribPointer(positionLocation, 3, gl.FLOAT, false, 0, gl.PtrOffset(0))

		if len(data.Normals) != 0 {
			gl.BindBuffer(gl.ARRAY_BUFFER, nbo)
			gl.EnableVertexAttribArray(normalLocation)
			gl.VertexAttribPointer(normalLocation, 3, gl.FLOAT, false, 0, gl.PtrOffset(0))
		}

		if len(data.TexCoords) != 0 {
			gl.BindBuffer(gl.ARRAY_BUFFER, tbo)
			gl.EnableVertexAttribArray(texCoordLocation)
			gl.VertexAttribPointer(texCoordLocation, 2, gl.FLOAT, false, 0, gl.PtrOffset(0))
		}

		if len(data.Colors) != 0 {
			gl.BindBuffer(gl.ARRAY_BUFFER, cbo)
			gl.EnableVertexAttribArray(colorLocation)
			gl.VertexAttribPointer(colorLocation, 3, gl.FLOAT, false, 0, gl.PtrOffset(0))
		}

		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ibo)
	}
	gl.BindVertexArray(0)

	var glMode uint32
	switch mode {
	case Triangles:
		glMode = gl.TRIANGLES
	case Lines:
		glMode = gl.LINES
	default:
		glog.Fatalf("gfx.NewVAO: unsupported mode: %v", mode)
	}

	return &VAO2{
		array: array,
		mode:  glMode,
		count: int32(len(data.Indices)),
	}
}

// Render renders the VAO.
func (v *VAO2) Render() {
	if v.count == 0 {
		return // No buffer data. Nothing to render.
	}
	gl.BindVertexArray(v.array)
	gl.DrawElements(v.mode, v.count, gl.UNSIGNED_SHORT, gl.Ptr(nil))
	gl.BindVertexArray(0)
}

// Delete deletes the VAO. Don't call Render after calling this.
func (v *VAO2) Delete() {
	if v.count == 0 {
		return // No buffer data. Nothing to delete.
	}
	gl.DeleteVertexArrays(1, &v.array)
}
