package rect

import (
	"bytes"
	"image"
	"math"

	"github.com/btmura/ponzi2/internal/app/gfx"
)

// gapFudge is how much to extend the borders to close gaps in OpenGL rendering.
const gapFudge = 1

// Rounded corner VAOs used to render the rounded rectangle corners.
var (
	roundedCornerNWFaces = gfx.PLYVAO(bytes.NewReader(_escFSMustByte(false, "/data/roundedcorner_faces.ply")))
	roundedCornerNWEdges = gfx.PLYVAO(bytes.NewReader(_escFSMustByte(false, "/data/roundedcorner_edges.ply")))
)

// Bubble is a rounded rectangle with a fill and stroke color.
type Bubble struct {
	// rounding is how rounded the corners of the bubble are.
	rounding int

	// bounds is the bounds to draw within.
	bounds image.Rectangle
}

// NewBubble returns a new Bubble.
func NewBubble(rounding int) *Bubble {
	return &Bubble{rounding: rounding}
}

// SetBounds sets the bounds to draw within.
func (b *Bubble) SetBounds(bounds image.Rectangle) {
	b.bounds = bounds
}

// Render renders the bubble.
func (b *Bubble) Render(_ float32) {
	if b.bounds.Empty() {
		return
	}
	fillRoundedRect(b.bounds, b.rounding)
	strokeRoundedRect(b.bounds, b.rounding)
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
