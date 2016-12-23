package ponzi

import (
	"bufio"
	"image"
	"image/draw"
	"image/png"
	"log"
	"os"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// textFactory is a factory that creates either static or dynamic text
// that can be rendered on a given orthographic plane.
type textFactory struct {
	mesh *mesh
	face font.Face
}

// newTextFactory creates a factory from an orthographic plane mesh and TTF bytes.
func newTextFactory(mesh *mesh, fontBytes []byte, size int) (*textFactory, error) {
	f, err := truetype.Parse(fontBytes)
	if err != nil {
		return nil, err
	}

	face := truetype.NewFace(f, &truetype.Options{
		Size:    float64(size),
		DPI:     72,
		Hinting: font.HintingFull,
	})

	return &textFactory{
		mesh: mesh,
		face: face,
	}, nil
}

// createStaticText creates static text which cannot be changed later.
func (f *textFactory) createStaticText(text string) *staticText {
	rgba := createTextImage(f.face, text)
	return &staticText{
		mesh:    f.mesh,
		texture: createTexture(rgba),
		size:    rgba.Bounds().Size(),
	}
}

// createDynamicText creates dynamic text which is rendered at runtime.
func (f *textFactory) createDynamicText() *dynamicText {
	staticTextMap := map[rune]*staticText{}
	for _, r := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789.+-% " {
		staticTextMap[r] = f.createStaticText(string(r))
	}
	return &dynamicText{
		staticTextMap: staticTextMap,
	}
}

type staticText struct {
	mesh    *mesh
	texture uint32
	size    image.Point
}

func (t *staticText) render(c image.Point) image.Point {
	m := newScaleMatrix(float32(t.size.X), float32(t.size.Y), 1)
	m = m.mult(newTranslationMatrix(float32(c.X), float32(c.Y), 0))
	gl.UniformMatrix4fv(modelMatrixLocation, 1, false, &m[0])
	gl.BindTexture(gl.TEXTURE_2D, t.texture)
	t.mesh.drawElements()
	return image.Pt(t.size.X, 0)
}

type dynamicText struct {
	staticTextMap map[rune]*staticText
}

func (t *dynamicText) render(text string, c image.Point) image.Point {
	w := 0
	for _, r := range text {
		if st := t.staticTextMap[r]; st != nil {
			st.render(c)
			c.X += st.size.X
			w += st.size.X
		}
	}
	return image.Pt(w, 0)
}

func createTextImage(face font.Face, text string) *image.RGBA {
	w := font.MeasureString(face, text)
	m := face.Metrics() // Used for height and descent.

	fg, bg := image.White, image.Transparent

	rgba := image.NewRGBA(image.Rect(0, 0, w.Round(), m.Height.Round())) // (MinX, MinY), (MaxX, MaxY)
	draw.Draw(rgba, rgba.Bounds(), bg, image.ZP, draw.Src)

	d := &font.Drawer{
		Dst:  rgba,
		Src:  fg,
		Face: face,
		Dot: fixed.Point26_6{
			Y: m.Height - m.Descent,
		},
	}
	d.DrawString(text)

	return rgba
}

func writePNGFile(rgba *image.RGBA) error {
	out, err := os.Create("text.png")
	if err != nil {
		return err
	}
	defer out.Close()

	b := bufio.NewWriter(out)
	if err := png.Encode(b, rgba); err != nil {
		return err
	}
	if err := b.Flush(); err != nil {
		return err
	}
	log.Printf("Wrote PNG file: %s", out.Name())

	return nil
}
