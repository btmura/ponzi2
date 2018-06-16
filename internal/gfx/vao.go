package gfx

import (
	"image"

	"github.com/go-gl/gl/v4.5-core/gl"
)

// VAO is a Vertex Array Object for one element type (lines, triangles, etc).
type VAO struct {
	// initData is used to initialize the VAO when needed. Nil after it is used.
	initData *VAOVertexData

	// array is the VAO name for gl.BindVertexArray.
	array uint32

	// mode is like gl.LINES or gl.TRIANGLES passed to gl.DrawElements.
	mode uint32

	// count is the number of elements for gl.DrawElements.
	count int32

	// texture is the texture name for gl.BindTexture.
	texture uint32

	// hasTexture is true if a texture was binded.
	hasTexture bool
}

// VAODrawMode is analogous to the mode argument to gl.DrawElements
// like gl.LINES or gl.TRIANGLES.
type VAODrawMode int

// Enum values for VAODrawMode.
const (
	Unspecified VAODrawMode = iota
	Triangles
	Lines
)

// VAOVertexData is a bunch of slices filled with vertex data to create a VAO.
type VAOVertexData struct {
	// Vertices is a required slice of flattened (x, y, z) vertices.
	Vertices []float32

	// TexCoords is an optional slice of flattened (s, t) coords.
	TexCoords []float32

	// Colors is an optional slice of flattened (r, g, b) values.
	Colors []float32

	// Indices is a required slice of indices into all the buffers.
	Indices []uint16

	// TextureRGBA is the optional texture to use.
	TextureRGBA *image.RGBA
}

// NewVAO creates a VAO with the given data buffers and a drawing mode.
//
// Since NewVAO defers the creation of the actual OpenGL VAO till the first
// rendering, callers may call NewVAO at the package scope to simplify code.
func NewVAO(mode VAODrawMode, data *VAOVertexData) *VAO {
	logger.Printf("NewVAO: v(%d) tc(%d) c(%d) i(%d)", len(data.Vertices), len(data.TexCoords), len(data.Colors), len(data.Indices))
	if len(data.Indices) == 0 {
		return &VAO{} // OpenGL doesn't allow empty buffer objects. Return VAO with zero count.
	}

	// TODO(btmura): store draw mode instead of the GL value and use unspecified in Delete
	var glMode uint32
	switch mode {
	case Triangles:
		glMode = gl.TRIANGLES
	case Lines:
		glMode = gl.LINES
	default:
		logger.Fatalf("NewVAO: unsupported mode: %v", mode)
	}

	return &VAO{
		initData: data,
		mode:     glMode,
	}
}

// Render renders the VAO.
func (v *VAO) Render() {
	if v.initData != nil {
		v.array, v.count = createVAO(v.initData)
		if v.initData.TextureRGBA != nil {
			v.texture = texture(v.initData.TextureRGBA)
			v.hasTexture = true
		}
		v.initData = nil // Reset initData to indicate the data was used to create the VAO.
	}
	if v.count == 0 {
		return // No buffer data. Nothing to render.
	}
	if v.hasTexture {
		setFragMode(fragTextureMode)
		gl.BindTexture(gl.TEXTURE_2D, v.texture)
		defer func() {
			setFragMode(fragColorMode)
			gl.BindTexture(gl.TEXTURE_2D, 0)
		}()
	}
	gl.BindVertexArray(v.array)
	gl.DrawElements(v.mode, v.count, gl.UNSIGNED_SHORT, gl.Ptr(nil))
	gl.BindVertexArray(0)
}

// createVAO blindly creates the VAO using the given vertex data.
func createVAO(data *VAOVertexData) (array uint32, count int32) {
	vbo := arrayBuffer(data.Vertices)

	var tbo uint32
	if len(data.TexCoords) != 0 {
		tbo = arrayBuffer(data.TexCoords)
	}

	var cbo uint32
	if len(data.Colors) != 0 {
		cbo = arrayBuffer(data.Colors)
	}

	ibo := elementArrayBuffer(data.Indices)

	gl.GenVertexArrays(1, &array)
	gl.BindVertexArray(array)
	{
		gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
		gl.EnableVertexAttribArray(positionLocation)
		gl.VertexAttribPointer(positionLocation, 3, gl.FLOAT, false, 0, gl.PtrOffset(0))

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

	return array, int32(len(data.Indices))
}

// Delete deletes the VAO. Don't call Render after calling this.
func (v *VAO) Delete() {
	logger.Printf("Delete: array(%d) mode(%v) count(%d) texture(%d) hasTexture(%t)", v.array, v.mode, v.count, v.texture, v.hasTexture)
	defer func() {
		v.initData = nil
		v.array = 0
		v.mode = 0
		v.count = 0
		v.texture = 0
		v.hasTexture = false
	}()

	if v.initData != nil {
		return // Data never used to create the VAO. Bail out.
	}
	if v.count != 0 {
		gl.DeleteVertexArrays(1, &v.array)
	}
	if v.hasTexture {
		gl.DeleteTextures(1, &v.texture)
	}
}
