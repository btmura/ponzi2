package gfx

import (
	"flag"
	"image"

	"github.com/go-gl/gl/v4.5-core/gl"

	math2 "github.com/btmura/ponzi2/internal/math"
)

// Get go-bindata from github.com/jteeuwen/go-bindata. It's used to embed resources into the binary.
//go:generate go-bindata -pkg gfx -prefix data -ignore ".*blend.*" data

func init() {
	flag.Parse() // Avoid glog errors about logging before flag.Parse.
}

// Locations for the vertex and fragment shaders.
const (
	projectionViewMatrixLocation = iota
	modelMatrixLocation
	normalMatrixLocation

	ambientLightColorLocation
	directionalLightColorLocation
	directionalLightVectorLocation

	positionLocation
	normalLocation
	texCoordLocation
	colorLocation

	textureLocation
	colorMixAmountLocation
	textColorLocation
)

func Init() error {
	p, err := program(string(MustAsset("shader.vert")), string(MustAsset("shader.frag")))
	if err != nil {
		return err
	}
	gl.UseProgram(p)
	return nil
}

func SetProjectionViewMatrix(m math2.Matrix4) {
	gl.UniformMatrix4fv(projectionViewMatrixLocation, 1, false, &m[0])
}

func SetModelMatrix(m math2.Matrix4) {
	gl.UniformMatrix4fv(modelMatrixLocation, 1, false, &m[0])
}

func SetModelMatrixRect(r image.Rectangle) {
	m := math2.ScaleMatrix(float32(r.Dx()/2), float32(r.Dy()/2), 1)
	m = m.Mult(math2.TranslationMatrix(float32(r.Min.X+r.Dx()/2), float32(r.Min.Y+r.Dy()/2), 0))
	gl.UniformMatrix4fv(modelMatrixLocation, 1, false, &m[0])
}

func SetModelMatrixOrtho(pt, sz image.Point) {
	m := math2.ScaleMatrix(float32(sz.X), float32(sz.Y), 1)
	m = m.Mult(math2.TranslationMatrix(float32(pt.X), float32(pt.Y), 0))
	gl.UniformMatrix4fv(modelMatrixLocation, 1, false, &m[0])
}

func SetNormalMatrix(m math2.Matrix4) {
	gl.UniformMatrix4fv(normalMatrixLocation, 1, false, &m[0])
}

func SetAmbientLightColor(color [3]float32) {
	gl.Uniform3fv(ambientLightColorLocation, 1, &color[0])
}

func SetDirectionalLightColor(color [3]float32) {
	gl.Uniform3fv(directionalLightColorLocation, 1, &color[0])
}

func SetDirectionalLightVector(vector [3]float32) {
	gl.Uniform3fv(directionalLightVectorLocation, 1, &vector[0])
}

// TODO(btmura): rename to SetTextureColorAmount
func SetColorMixAmount(amount float32) {
	gl.Uniform1f(colorMixAmountLocation, amount)
}
