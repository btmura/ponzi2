package ponzi

import (
	"bufio"
	"image"
	"image/draw"
	"image/png"
	"log"
	"os"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

func newFace(f *truetype.Font) font.Face {
	return truetype.NewFace(f, &truetype.Options{
		Size:    24,
		DPI:     72,
		Hinting: font.HintingFull,
	})
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
