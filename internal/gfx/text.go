package gfx

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

// TextFactory is a factory that creates either static or dynamic text
// that can be rendered on a given orthographic plane.
type TextFactory struct {
	mesh *Mesh
	face font.Face
}

// NewTextFactory creates a factory from an orthographic plane mesh and TTF bytes.
func NewTextFactory(mesh *Mesh, fontBytes []byte, size int) (*TextFactory, error) {
	f, err := truetype.Parse(fontBytes)
	if err != nil {
		return nil, err
	}

	face := truetype.NewFace(f, &truetype.Options{
		Size:    float64(size),
		DPI:     72,
		Hinting: font.HintingFull,
	})

	return &TextFactory{
		mesh: mesh,
		face: face,
	}, nil
}

// CreateStaticText creates static text which cannot be changed later.
func (f *TextFactory) CreateStaticText(text string) *StaticText {
	rgba := createTextImage(f.face, text)
	return &StaticText{
		mesh:    f.mesh,
		texture: gl2.CreateTexture(rgba),
		Size:    rgba.Bounds().Size(),
	}
}

// CreateDynamicText creates dynamic text which is rendered at runtime.
func (f *TextFactory) CreateDynamicText() *DynamicText {
	runeTextMap := make(map[rune]*StaticText)
	for _, r := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789.+-% " {
		runeTextMap[r] = f.CreateStaticText(string(r))
	}
	return &DynamicText{
		runeTextMap: runeTextMap,
	}
}

type StaticText struct {
	mesh    *Mesh
	texture uint32
	Size    image.Point
}

func (t *StaticText) Render(c image.Point, color [3]float32) image.Point {
	m := math2.NewScaleMatrix(float32(t.Size.X), float32(t.Size.Y), 1)
	m = m.Mult(math2.NewTranslationMatrix(float32(c.X), float32(c.Y), 0))
	gl.UniformMatrix4fv(modelMatrixLocation, 1, false, &m[0])

	gl.BindTexture(gl.TEXTURE_2D, t.texture)
	gl.Uniform3fv(textColorLocation, 1, &color[0])
	gl.Uniform1f(colorMixAmountLocation, 0)

	t.mesh.DrawElements()
	return image.Pt(t.Size.X, 0)
}

type DynamicText struct {
	runeTextMap map[rune]*StaticText
}

func (t *DynamicText) Measure(text string) image.Point {
	s := image.ZP
	for _, r := range text {
		if st := t.runeTextMap[r]; st != nil {
			s.X += st.Size.X
			if st.Size.Y > s.Y {
				s.Y = st.Size.Y
			}
		}
	}
	return s
}

func (t *DynamicText) Render(text string, c image.Point, color [3]float32) image.Point {
	w := 0
	for _, r := range text {
		if st := t.runeTextMap[r]; st != nil {
			st.Render(c, color)
			c.X += st.Size.X
			w += st.Size.X
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
