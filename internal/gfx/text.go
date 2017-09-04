package gfx

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/golang/freetype/truetype"
	"github.com/golang/glog"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"

	math2 "github.com/btmura/ponzi2/internal/math"
)

// A TextRenderer measures and renders text.
// It is designed to render single lines of [A-Z0-9] characters.
type TextRenderer struct {
	face            font.Face              // Face to render text with.
	metrics         font.Metrics           // Custom metrics to ease vertical alignment.
	runeRendererMap map[rune]*runeRenderer // Map from rune to runeRenderer.
}

// TextColor is a RGB color of 3 floats from 0.0 to 1.0.
type TextColor [3]float32

// NewTextRenderer creates a new TextRenderer from a TTF font file and a size.
func NewTextRenderer(ttfBytes []byte, size int) *TextRenderer {
	// Callers should be able to initialize TextRenderers as globals,
	// so do not do any intialization here that requires OpenGL.

	// Parse the TTF font bytes and create a face out of it.
	ttFont, err := truetype.Parse(ttfBytes)
	if err != nil {
		glog.Fatalf("gfx.NewTextRenderer: parsing TTF bytes failed: %v", err)
	}
	face := truetype.NewFace(ttFont, &truetype.Options{
		Size:    float64(size),
		DPI:     72,
		Hinting: font.HintingFull,
	})

	// Generate custom metrics to suit single line positioning and rendering.
	// https://developer.apple.com/library/content/documentation/TextFonts/Conceptual/CocoaTextArchitecture/Art/glyph_metrics_2x.png
	bnds, _, ok := face.GlyphBounds('M') // Bounds for a square that mimics most cap letters.
	if !ok {
		glog.Fatal("gfx.NewTextRenderer: getting bounds for M failed")
	}
	a := bnds.Max.Y - bnds.Min.Y // Height of M is the ascent.
	d := face.Metrics().Descent  // Some descent for Q and J.
	m := font.Metrics{
		Ascent:  a,
		Descent: d,
		Height:  d + a + d, // Pad ascent on top and bottom to vertically center.
	}

	return &TextRenderer{
		face:            face,
		metrics:         m,
		runeRendererMap: map[rune]*runeRenderer{},
	}
}

// LineHeight returns the line height. Useful for measurements.
func (t *TextRenderer) LineHeight() int {
	return t.metrics.Height.Round()
}

// Measure returns an image.Point with the width and height of the given text.
func (t *TextRenderer) Measure(text string) image.Point {
	s := image.Pt(0, t.LineHeight())
	for _, r := range text {
		rr := t.runeRenderer(r)
		s.X += rr.size.X
	}
	return s
}

// Render renders color text at the given point that points at the origin (baseline).
func (t *TextRenderer) Render(text string, pt image.Point, color TextColor) int {
	dx := 0
	for _, r := range text {
		rr := t.runeRenderer(r)
		rr.render(pt, color)
		pt.X += rr.size.X
		dx += rr.size.X
	}
	return dx
}

func (t *TextRenderer) runeRenderer(r rune) *runeRenderer {
	if rr := t.runeRendererMap[r]; rr != nil {
		return rr
	}
	rr := newRuneRenderer(t.face, t.metrics, r)
	t.runeRendererMap[r] = rr
	return rr
}

type runeRenderer struct {
	texture uint32
	size    image.Point
}

func newRuneRenderer(face font.Face, m font.Metrics, r rune) *runeRenderer {
	fg := image.NewUniform(color.RGBA{255, 0, 0, 255})
	bg := image.Transparent
	// bg = fg // Uncomment to render rectangles.

	w := font.MeasureString(face, string(r))
	rgba := image.NewRGBA(image.Rect(0, 0, w.Round(), m.Height.Round()))
	draw.Draw(rgba, rgba.Bounds(), bg, image.ZP, draw.Src)

	d := &font.Drawer{
		Dst:  rgba,
		Src:  fg,
		Face: face,
		Dot: fixed.Point26_6{
			Y: m.Ascent + m.Descent, // Move down from top.
		},
	}
	d.DrawString(string(r))

	return &runeRenderer{
		texture: texture(rgba),
		size:    rgba.Bounds().Size(),
	}
}

// runePlaneObject is a shared Vertex Array Object that all runeRenderers use.
var runePlaneObject *VAO2

func (r *runeRenderer) render(pt image.Point, color TextColor) image.Point {
	m := math2.ScaleMatrix(float32(r.size.X), float32(r.size.Y), 1)
	m = m.Mult(math2.TranslationMatrix(float32(pt.X), float32(pt.Y), 0))
	gl.UniformMatrix4fv(modelMatrixLocation, 1, false, &m[0])

	gl.BindTexture(gl.TEXTURE_2D, r.texture)
	gl.Uniform3fv(textColorLocation, 1, &color[0])
	gl.Uniform1f(colorMixAmountLocation, 0)

	if runePlaneObject == nil {
		runePlaneObject = newPLYVAO(bytes.NewReader(MustAsset("textPlane.ply")))
	}
	runePlaneObject.Render()

	return image.Pt(r.size.X, 0)
}
