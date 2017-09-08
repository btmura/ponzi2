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

// renderRoundedRect renders a rounded rectangle using the given rectangular bounds.
func renderRoundedRect(r image.Rectangle, roundAmount int) {
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

	// fudge is how much to extend the borders to close gaps in OpenGL rendering.
	const fudge = 2

	// TOP border
	hMinX, hMaxX := r.Min.X+roundAmount-fudge, r.Max.X-roundAmount+fudge
	gfx.SetModelMatrixRect(image.Rect(hMinX, r.Max.Y, hMaxX, r.Max.Y))
	horizLine.Render()

	// BOTTOM border
	gfx.SetModelMatrixRect(image.Rect(hMinX, r.Min.Y, hMaxX, r.Min.Y))
	horizLine.Render()

	// LEFT border
	vMinX, vMaxX := r.Min.Y+roundAmount-fudge, r.Max.Y-roundAmount+fudge
	gfx.SetModelMatrixRect(image.Rect(r.Min.X, vMinX, r.Min.X, vMaxX))
	vertLine.Render()

	// RIGHT border
	gfx.SetModelMatrixRect(image.Rect(r.Max.X, vMinX, r.Max.X, vMaxX))
	vertLine.Render()
}

// renderHorizDividers horizontally cuts a rectangle from the bottom at the given percentages,
// renders dividers at those percentages, and returns the n+1 rectangles given n percentages.
func renderHorizDividers(r image.Rectangle, percentages ...float32) []image.Rectangle {
	gfx.SetColorMixAmount(1)

	rects := sliceRect(r, percentages...)
	for _, r := range rects {
		gfx.SetModelMatrixRect(image.Rect(r.Min.X, r.Max.Y, r.Max.X, r.Max.Y))
		horizLine.Render()
	}
	return rects
}

// sliceRect horizontally cuts a rectangle from the bottom at the given percentages.
// It returns n+1 rectangles given n percentages.
func sliceRect(r image.Rectangle, percentages ...float32) []image.Rectangle {
	var rs []image.Rectangle
	addRect := func(minY, maxY int) {
		rs = append(rs, image.Rect(r.Min.X, minY, r.Max.X, maxY))
	}

	ry := r.Dy()  // Remaining Y to distribute.
	cy := r.Min.Y // Start at the bottom and cut horizontally up.
	for _, p := range percentages {
		dy := int(float32(r.Dy()) * p)
		addRect(cy, cy+dy)
		cy += dy // Bump upwards.
		ry -= dy // Subtract from remaining.
	}
	addRect(cy, cy+ry) // Use remaining Y for last rect.

	return rs
}
