package ponzi

import (
	"bufio"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/golang/freetype/truetype"
	"github.com/golang/glog"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"

	"github.com/btmura/ponzi2/internal/gl2"
	"github.com/btmura/ponzi2/internal/math2"
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
		texture: gl2.CreateTexture(rgba),
		size:    rgba.Bounds().Size(),
	}
}

// createDynamicText creates dynamic text which is rendered at runtime.
func (f *textFactory) createDynamicText() *dynamicText {
	runeTextMap := make(map[rune]*staticText)
	for _, r := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789.+-% " {
		runeTextMap[r] = f.createStaticText(string(r))
	}
	return &dynamicText{
		runeTextMap: runeTextMap,
	}
}

type staticText struct {
	mesh    *mesh
	texture uint32
	size    image.Point
}

func (t *staticText) render(c image.Point, color [3]float32) image.Point {
	m := math2.NewScaleMatrix(float32(t.size.X), float32(t.size.Y), 1)
	m = m.Mult(math2.NewTranslationMatrix(float32(c.X), float32(c.Y), 0))
	gl.UniformMatrix4fv(modelMatrixLocation, 1, false, &m[0])

	gl.BindTexture(gl.TEXTURE_2D, t.texture)
	gl.Uniform3fv(textColorLocation, 1, &color[0])
	gl.Uniform1f(colorMixAmountLocation, 0)

	t.mesh.drawElements()
	return image.Pt(t.size.X, 0)
}

type dynamicText struct {
	runeTextMap map[rune]*staticText
}

func (t *dynamicText) measure(text string) image.Point {
	s := image.ZP
	for _, r := range text {
		if st := t.runeTextMap[r]; st != nil {
			s.X += st.size.X
			if st.size.Y > s.Y {
				s.Y = st.size.Y
			}
		}
	}
	return s
}

func (t *dynamicText) render(text string, c image.Point, color [3]float32) image.Point {
	w := 0
	for _, r := range text {
		if st := t.runeTextMap[r]; st != nil {
			st.render(c, color)
			c.X += st.size.X
			w += st.size.X
		}
	}
	return image.Pt(w, 0)
}

func createTextImage(face font.Face, text string) *image.RGBA {
	w := font.MeasureString(face, text)
	m := face.Metrics() // Used for height and descent.

	fg, bg := image.NewUniform(color.RGBA{255, 0, 0, 255}), image.Transparent

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
	glog.Infof("Wrote PNG file: %s", out.Name())

	return nil
}
