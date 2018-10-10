package gfx

import (
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/golang/glog"
)

// vao is a Vertex Array Object (VAO) that can be rendered and deleted.
type vao struct {
	// mode is like gl.LINES or gl.TRIANGLES passed to gl.DrawElements.
	mode uint32

	// array is the VAO name for gl.BindVertexArray.
	array uint32

	// count is the number of elements for gl.DrawElements.
	count int32

	// texture is the texture name for gl.BindTexture.
	texture uint32

	// hasTexture is true if a texture was binded.
	hasTexture bool
}

func newVAO(data *VAOVertexData) *vao {
	glog.V(2).Infof("creating vao: v(%d) tc(%d) c(%d) i(%d)", len(data.Vertices), len(data.TexCoords), len(data.Colors), len(data.Indices))

	if len(data.Vertices) == 0 {
		return &vao{}
	}

	if len(data.Indices) == 0 {
		return &vao{}
	}

	vbo := glArrayBuffer(data.Vertices)
	ibo := glElementArrayBuffer(data.Indices)

	var tbo uint32
	if len(data.TexCoords) != 0 {
		tbo = glArrayBuffer(data.TexCoords)
	}

	var cbo uint32
	if len(data.Colors) != 0 {
		cbo = glArrayBuffer(data.Colors)
	}

	var array uint32
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

	v := &vao{
		mode:  uint32(data.Mode),
		array: array,
		count: int32(len(data.Indices)),
	}
	if data.TextureRGBA != nil {
		v.texture = glTexture(data.TextureRGBA)
		v.hasTexture = true
	}
	return v
}

// Render renders the VAO.
func (v *vao) render() {
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

func (v *vao) delete() {
	glog.V(2).Infof("deleting vao: array(%d) mode(%v) count(%d) texture(%d) hasTexture(%t)", v.array, v.mode, v.count, v.texture, v.hasTexture)
	defer func() {
		v.array = 0
		v.mode = 0
		v.count = 0
		v.texture = 0
		v.hasTexture = false
	}()

	if v.count != 0 {
		gl.DeleteVertexArrays(1, &v.array)
	}
	if v.hasTexture {
		gl.DeleteTextures(1, &v.texture)
	}
}
