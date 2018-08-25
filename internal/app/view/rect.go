package view

import (
	"bytes"
	"image"
	"math"

	"github.com/btmura/ponzi2/internal/gfx"
)

// White horizontal and vertical line VAOs that can be reused anywhere.
var (
	horizLine   = horizLineVAO(white)
	vertLine    = vertLineVAO(white)
	squarePlane = gfx.PLYVAO(bytes.NewReader(_escFSMustByte(false, "/data/squareplane.ply")))
)

// Rounded corner VAOs used to render the rounded rectangle corners.
var (
	roundedCornerNWFaces = gfx.PLYVAO(bytes.NewReader(_escFSMustByte(false, "/data/roundedcorner_faces.ply")))
	roundedCornerNWEdges = gfx.PLYVAO(bytes.NewReader(_escFSMustByte(false, "/data/roundedcorner_edges.ply")))
)

// gapFudge is how much to extend the borders to close gaps in OpenGL rendering.
const gapFudge = 2

// TODO(btmura): change from bubbleSpec type to bubble type
type bubbleSpec struct {
	// rounding is how much rounding of the bubble's rounded rectangle.
	rounding int

	// padding is how much padding of the bubble's text.
	padding int
}

func renderBubble(pt, sz image.Point, bs bubbleSpec) {
	br := image.Rectangle{
		Min: pt,
		Max: pt.Add(sz),
	}
	br = br.Inset(-bs.padding)
	fillRoundedRect(br, bs.rounding)
	strokeRoundedRect(br, bs.rounding)
}

// fillRoundedRect renders a filled rounded rectangle within r.
func fillRoundedRect(r image.Rectangle, rounding int) {
	// [+] Render 2 overlapping rects to fill in everything except the corners.

	// 1. [|] Render filled rect from bottom to top but less on the X-Axis.
	gfx.SetModelMatrixRect(image.Rect(r.Min.X+rounding-gapFudge, r.Min.Y, r.Max.X-rounding+gapFudge, r.Max.Y))
	squarePlane.Render()

	// 2. [-] Render filled rect from left to right but less on the Y-Axis.
	gfx.SetModelMatrixRect(image.Rect(r.Min.X, r.Min.Y+rounding-gapFudge, r.Max.X, r.Max.Y-rounding+gapFudge))
	squarePlane.Render()

	// Render the four corners.
	renderRoundedRectCorners(r, roundedCornerNWFaces, rounding)
}

// strokeRoundedRect renders a stroked rounded rectangle within r.
func strokeRoundedRect(r image.Rectangle, rounding int) {
	// TOP Border
	hMinX, hMaxX := r.Min.X+rounding-gapFudge, r.Max.X-rounding+gapFudge
	gfx.SetModelMatrixRect(image.Rect(hMinX, r.Max.Y, hMaxX, r.Max.Y))
	horizLine.Render()

	// BOTTOM Border
	gfx.SetModelMatrixRect(image.Rect(hMinX, r.Min.Y, hMaxX, r.Min.Y))
	horizLine.Render()

	// LEFT Border
	vMinX, vMaxX := r.Min.Y+rounding-gapFudge, r.Max.Y-rounding+gapFudge
	gfx.SetModelMatrixRect(image.Rect(r.Min.X, vMinX, r.Min.X, vMaxX))
	vertLine.Render()

	// RIGHT Border
	gfx.SetModelMatrixRect(image.Rect(r.Max.X, vMinX, r.Max.X, vMaxX))
	vertLine.Render()

	// Render the four corners.
	renderRoundedRectCorners(r, roundedCornerNWEdges, rounding)
}

// renderRoundedRectCorners is a helper function for fillRoundedRect
// and strokeRoundedRect that renders a VAO at r's corners.
func renderRoundedRectCorners(r image.Rectangle, nwCornerVAO *gfx.VAO, rounding int) {
	// NORTHWEST Corner
	gfx.SetModelMatrixRect(image.Rect(r.Min.X, r.Max.Y-rounding, r.Min.X+rounding, r.Max.Y))
	nwCornerVAO.Render()

	// NORTHEAST Corner
	gfx.SetModelMatrixRotatedRect(image.Rect(r.Max.X-rounding, r.Max.Y-rounding, r.Max.X, r.Max.Y), -math.Pi/2)
	nwCornerVAO.Render()

	// SOUTHEAST Corner
	gfx.SetModelMatrixRotatedRect(image.Rect(r.Max.X-rounding, r.Min.Y, r.Max.X, r.Min.Y+rounding), -math.Pi)
	nwCornerVAO.Render()

	// SOUTHWEST Corner
	gfx.SetModelMatrixRotatedRect(image.Rect(r.Min.X, r.Min.Y, r.Min.X+rounding, r.Min.Y+rounding), -3*math.Pi/2)
	nwCornerVAO.Render()
}

// renderSlicedRectDividers horizontally cuts a rectangle from the bottom at
// the percentage amounts and draws the VAO at those percentages.
func renderSlicedRectDividers(r image.Rectangle, dividerVAO *gfx.VAO, percentages ...float32) {
	rects := sliceRect(r, percentages...)
	for _, r := range rects[:len(rects)-1] {
		renderRectTopDivider(r, dividerVAO)
	}
}

// renderRectTopDivider renders a VAO in a single pixel horizontal rectangle
// at the top edge of the rectangle.
func renderRectTopDivider(r image.Rectangle, dividerVAO *gfx.VAO) {
	gfx.SetModelMatrixRect(image.Rect(r.Min.X, r.Max.Y, r.Max.X, r.Max.Y))
	dividerVAO.Render()
}

// sliceRect horizontally cuts a rectangle from the bottom at the percentage
// amounts. It returns n+1 rectangles given n percentages.
func sliceRect(r image.Rectangle, percentages ...float32) []image.Rectangle {
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

// transRect returns a rectangle translated by the dx and dy amounts.
func transRect(r image.Rectangle, dx, dy int) image.Rectangle {
	return r.Add(image.Pt(dx, dy))
}
