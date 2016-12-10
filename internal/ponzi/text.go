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

func writeText() error {
	text := "Loading DATA..."

	fontBytes, err := ptm55ftTtfBytes()
	if err != nil {
		return err
	}

	f, err := truetype.Parse(fontBytes)
	if err != nil {
		return err
	}

	face := truetype.NewFace(f, &truetype.Options{
		Size:    12,
		DPI:     72,
		Hinting: font.HintingFull,
	})

	w := font.MeasureString(face, text)
	m := face.Metrics() // Used for height and descent.

	fg, bg := image.White, image.Black

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

	// Write the text image to a file.

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
	log.Printf("Wrote text image: %s", out.Name())

	return nil
}
