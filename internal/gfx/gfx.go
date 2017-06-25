package gfx

import (
	"image"

	"github.com/go-gl/gl/v4.5-core/gl"

	"github.com/btmura/ponzi2/internal/math2"
)

// Get go-bindata from github.com/jteeuwen/go-bindata. It's used to embed resources into the binary.
//go:generate go-bindata -pkg gfx -prefix data -ignore ".*blend.*" data

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

func SetModelMatrixRect(r image.Rectangle) {
	m := math2.NewScaleMatrix(float32(r.Dx()/2), float32(r.Dy()/2), 1)
	m = m.Mult(math2.NewTranslationMatrix(float32(r.Min.X+r.Dx()/2), float32(r.Min.Y+r.Dy()/2), 0))
	gl.UniformMatrix4fv(modelMatrixLocation, 1, false, &m[0])
}
