package gfx

import (
	"image"

	"github.com/go-gl/gl/v4.5-core/gl"
)

// VAO is a Vertex Array Object for one element type (lines, triangles, etc).
type VAO struct {
	// loadData loads data to create a vao.
	loadData func() *VAOVertexData

	// vao is the actual VAO that can be rendered.
	vao *vao
}

// VAODrawMode is analogous to the mode argument to gl.DrawElements.
type VAODrawMode int

// Enum values for VAODrawMode.
const (
	Triangles VAODrawMode = gl.TRIANGLES
	Lines                 = gl.LINES
)

// VAOVertexData is a bunch of slices filled with vertex data to create a VAO.
type VAOVertexData struct {
	// Mode is the mode like triangles or lines.
	Mode VAODrawMode

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

// NewVAO creates a VAO with the given vertex data.
//
// Since NewVAO defers the creation of the actual OpenGL VAO till the first
// rendering, callers may call NewVAO at the package scope to simplify code.
func NewVAO(data *VAOVertexData) *VAO {
	return &VAO{loadData: func() *VAOVertexData { return data }}
}

// NewVAOLoadData creates a VAO with a data loader function
// that provides vertex data.
//
// Use this over NewVAO at the package scope if loading the vertex data
// requires flags to be parsed or some other requirement.
func NewVAOLoadData(loadData func() *VAOVertexData) *VAO {
	return &VAO{loadData: loadData}
}

// Render renders the VAO.
func (v *VAO) Render() {
	if v.vao == nil {
		v.vao = newVAO(v.loadData())
	}
	v.vao.render()
}

// Delete deletes the VAO. Don't call Render after calling this.
func (v *VAO) Delete() {
	if v.vao != nil {
		v.vao.delete()
	}
}
