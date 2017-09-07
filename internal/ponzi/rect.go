package ponzi

import (
	"bytes"
	"image"

	"github.com/btmura/ponzi2/internal/gfx"
)

// White horizontal and vertical line VAOs that can be reused anywhere.
var (
	horizLine = gfx.HorizColoredLineVAO(white, white)
	vertLine  = gfx.VertColoredLineVAO(white, white)
)

// VAOs for each corner. Rotation could possibly be used to simplify this...
var (
	roundedCornerNW = gfx.ReadPLYVAO(bytes.NewReader(MustAsset("roundedCornerNW.ply")))
	roundedCornerNE = gfx.ReadPLYVAO(bytes.NewReader(MustAsset("roundedCornerNE.ply")))
	roundedCornerSE = gfx.ReadPLYVAO(bytes.NewReader(MustAsset("roundedCornerSE.ply")))
	roundedCornerSW = gfx.ReadPLYVAO(bytes.NewReader(MustAsset("roundedCornerSW.ply")))
)

const (
	roundAmount = 10 // roundAmount is the size of the square that the rounded corner is rendered in.
	roundFudge  = 2  // roundFudge is how much to extend the borders to close gaps in OpenGL rendering.
)

// renderRoundedRect renders a rounded rectangle using the given rectangular bounds.
func renderRoundedRect(r image.Rectangle) {
	gfx.SetColorMixAmount(1)

	// NORTHWEST Corner
	gfx.SetModelMatrixRect(image.Rect(r.Min.X, r.Max.Y-roundAmount, r.Min.X+roundAmount, r.Max.Y))
	roundedCornerNW.Render()

	// NORTHEAST Corner
	gfx.SetModelMatrixRect(image.Rect(r.Max.X-roundAmount, r.Max.Y-roundAmount, r.Max.X, r.Max.Y))
	roundedCornerNE.Render()

	// SOUTHEAST Corner
	gfx.SetModelMatrixRect(image.Rect(r.Max.X-roundAmount, r.Min.Y, r.Max.X, r.Min.Y+roundAmount))
	roundedCornerSE.Render()

	gfx.SetModelMatrixRect(image.Rect(r.Min.X, r.Min.Y, r.Min.X+roundAmount, r.Min.Y+roundAmount))
	roundedCornerSW.Render()

	// TOP border
	hMinX, hMaxX := r.Min.X+roundAmount-roundFudge, r.Max.X-roundAmount+roundFudge
	gfx.SetModelMatrixRect(image.Rect(hMinX, r.Max.Y, hMaxX, r.Max.Y))
	horizLine.Render()

	// BOTTOM border
	gfx.SetModelMatrixRect(image.Rect(hMinX, r.Min.Y, hMaxX, r.Min.Y))
	horizLine.Render()

	// LEFT border
	vMinX, vMaxX := r.Min.Y+roundAmount-roundFudge, r.Max.Y-roundAmount+roundFudge
	gfx.SetModelMatrixRect(image.Rect(r.Min.X, vMinX, r.Min.X, vMaxX))
	vertLine.Render()

	// RIGHT border
	gfx.SetModelMatrixRect(image.Rect(r.Max.X, vMinX, r.Max.X, vMaxX))
	vertLine.Render()
}
