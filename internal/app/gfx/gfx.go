// Package gfx provides APIs for shaders, VAO creation, and text rendering.
package gfx

import (
	"image"

	"github.com/go-gl/gl/v4.5-core/gl"

	"gitlab.com/btmura/ponzi2/internal/matrix"
)

// Get esc from github.com/mjibson/esc. It's used to embed resources into the binary.
//go:generate esc -o bindata.go -pkg gfx -include ".*(frag|ply|vert)" -modtime 1337 -private data

// Locations for the vertex and fragment shaders.
const (
	projectionViewMatrixLocation = 0
	modelMatrixLocation          = 1

	positionLocation = 2
	colorLocation    = 3
	texCoordLocation = 4

	fragModeLocation  = 5
	textureLocation   = 6
	textColorLocation = 7
	alphaLocation     = 8
)

type fragMode int32

const (
	fragColorMode     fragMode = 0
	fragTextureMode            = 1
	fragTextColorMode          = 2
)

var program uint32

// InitProgram loads and uses the shader program.
func InitProgram() error {
	p, err := glProgram(_escFSMustString(false, "/data/shader.vert"), _escFSMustString(false, "/data/shader.frag"))
	if err != nil {
		return err
	}
	gl.UseProgram(p)
	program = p
	return nil
}

// SetProjectionViewMatrix sets the projection view matrix.
func SetProjectionViewMatrix(m matrix.Matrix4) {
	gl.UniformMatrix4fv(projectionViewMatrixLocation, 1, false, &m[0])
}

// SetModelMatrixRect sets the model matrix to the rectangle.
func SetModelMatrixRect(r image.Rectangle) {
	m := matrix.Scale(float32(r.Dx()/2), float32(r.Dy()/2), 1)
	m = m.Mult(matrix.Translation(float32(r.Min.X+r.Dx()/2), float32(r.Min.Y+r.Dy()/2), 0))
	gl.UniformMatrix4fv(modelMatrixLocation, 1, false, &m[0])
}

// SetModelMatrixRotatedRect sets the model matrix to the rotated rectangle.
func SetModelMatrixRotatedRect(r image.Rectangle, radians float32) {
	m := matrix.Rotation(radians)
	m = m.Mult(matrix.Scale(float32(r.Dx()/2), float32(r.Dy()/2), 1))
	m = m.Mult(matrix.Translation(float32(r.Min.X+r.Dx()/2), float32(r.Min.Y+r.Dy()/2), 0))
	gl.UniformMatrix4fv(modelMatrixLocation, 1, false, &m[0])
}

func setModelMatrixOrtho(pt, sz image.Point) {
	m := matrix.Scale(float32(sz.X), float32(sz.Y), 1)
	m = m.Mult(matrix.Translation(float32(pt.X), float32(pt.Y), 0))
	gl.UniformMatrix4fv(modelMatrixLocation, 1, false, &m[0])
}

func setFragMode(mode fragMode) {
	gl.Uniform1i(fragModeLocation, int32(mode))
}

func setTextColor(color TextColor) {
	gl.Uniform3fv(textColorLocation, 1, &color[0])
}

// SetAlpha sets the alpha amount.
func SetAlpha(alpha float32) {
	gl.Uniform1f(alphaLocation, alpha)
}

// Alpha returns the current alpha amount.
func Alpha() (alpha float32) {
	gl.GetUniformfv(program, alphaLocation, &alpha)
	return alpha
}
