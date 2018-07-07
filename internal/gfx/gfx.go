package gfx

import (
	"image"

	"github.com/go-gl/gl/v4.5-core/gl"

	math2 "github.com/btmura/ponzi2/internal/math"
)

// Get esc from github.com/mjibson/esc. It's used to embed resources into the binary.
//go:generate esc -o bindata.go -pkg gfx -include ".*(frag|ply|vert)" -modtime 1337 -private data

// Locations for the vertex and fragment shaders.
const (
	projectionViewMatrixLocation = iota
	modelMatrixLocation

	positionLocation
	colorLocation
	texCoordLocation

	fragModeLocation
	textureLocation
	textColorLocation
)

type fragMode int32

const (
	fragColorMode fragMode = iota
	fragTextureMode
	fragTextColorMode
)

// InitProgram loads and uses the shader program.
func InitProgram() error {
	p, err := program(_escFSMustString(false, "/data/shader.vert"), _escFSMustString(false, "/data/shader.frag"))
	if err != nil {
		return err
	}
	gl.UseProgram(p)
	return nil
}

// SetProjectionViewMatrix sets the projection view matrix.
func SetProjectionViewMatrix(m math2.Matrix4) {
	gl.UniformMatrix4fv(projectionViewMatrixLocation, 1, false, &m[0])
}

// SetModelMatrixRect sets the model matrix to the rectangle.
func SetModelMatrixRect(r image.Rectangle) {
	m := math2.ScaleMatrix(float32(r.Dx()/2), float32(r.Dy()/2), 1)
	m = m.Mult(math2.TranslationMatrix(float32(r.Min.X+r.Dx()/2), float32(r.Min.Y+r.Dy()/2), 0))
	gl.UniformMatrix4fv(modelMatrixLocation, 1, false, &m[0])
}

// SetModelMatrixRotatedRect sets the model matrix to the rotated rectangle.
func SetModelMatrixRotatedRect(r image.Rectangle, radians float32) {
	m := math2.RotationMatrix(radians)
	m = m.Mult(math2.ScaleMatrix(float32(r.Dx()/2), float32(r.Dy()/2), 1))
	m = m.Mult(math2.TranslationMatrix(float32(r.Min.X+r.Dx()/2), float32(r.Min.Y+r.Dy()/2), 0))
	gl.UniformMatrix4fv(modelMatrixLocation, 1, false, &m[0])
}

func setModelMatrixOrtho(pt, sz image.Point) {
	m := math2.ScaleMatrix(float32(sz.X), float32(sz.Y), 1)
	m = m.Mult(math2.TranslationMatrix(float32(pt.X), float32(pt.Y), 0))
	gl.UniformMatrix4fv(modelMatrixLocation, 1, false, &m[0])
}

func setFragMode(mode fragMode) {
	gl.Uniform1i(fragModeLocation, int32(mode))
}

func setTextColor(color TextColor) {
	gl.Uniform3fv(textColorLocation, 1, &color[0])
}
