package rect

import (
	"bytes"
	"image"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/view/color"
	"github.com/btmura/ponzi2/internal/app/view/vao"
)

// Embed resources into the application. Get esc from github.com/mjibson/esc.
//go:generate esc -o bindata.go -pkg rect -include ".*(ply|png)" -modtime 1337 -private data

// White horizontal and vertical line VAOs that can be reused anywhere.
var (
	horizLine   = vao.HorizLine(color.White, color.White)
	vertLine    = vao.VertLine(color.White, color.White)
	squarePlane = gfx.PLYVAO(bytes.NewReader(_escFSMustByte(false, "/data/squareplane.ply")))
)

// RenderLineAtTop renders a VAO in a single pixel horizontal rectangle
// at the top edge of the rectangle.
func RenderLineAtTop(r image.Rectangle) {
	gfx.SetModelMatrixRect(image.Rect(r.Min.X, r.Max.Y, r.Max.X, r.Max.Y))
	horizLine.Render()
}

// FromCenterPointAndSize returns a rectangle of the given size centered at the given point.
func FromCenterPointAndSize(centerPt, size image.Point) image.Rectangle {
	return image.Rect(
		centerPt.X-size.X/2, centerPt.Y-size.Y/2,
		centerPt.X+size.X/2, centerPt.Y+size.Y/2,
	)
}

// CenterPoint returns a point at the center of the rectangle.
func CenterPoint(r image.Rectangle) image.Point {
	return image.Pt(r.Min.X+r.Dx()/2, r.Min.Y+r.Dy()/2)
}

// Slice horizontally cuts a rectangle from the bottom at the percentage
// amounts. It returns n+1 rectangles given n percentages.
func Slice(r image.Rectangle, percentages ...float32) []image.Rectangle {
	var rs []image.Rectangle
	addRect := func(minY, maxY int) {
		rs = append(rs, image.Rect(r.Min.X, minY, r.Max.X, maxY))
	}

	y := r.Min.Y // Start at the bottom and cut horizontally up.
	for _, p := range percentages {
		dy := int(float32(r.Dy()) * p)
		addRect(y, y+dy)
		y += dy // Bump upwards.
	}
	addRect(y, r.Max.Y)

	return rs
}

// Translate returns a rectangle translated by the dx and dy amounts.
func Translate(r image.Rectangle, dx, dy int) image.Rectangle {
	return r.Add(image.Pt(dx, dy))
}
