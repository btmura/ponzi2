package ponzi

import (
	"bytes"
	"image"

	"github.com/btmura/ponzi2/internal/gfx"
)

var addButtonVAO = gfx.ReadPLYVAO(bytes.NewReader(MustAsset("addButton.ply")))

type button struct{}

func (b *button) render(r image.Rectangle) {
	gfx.SetColorMixAmount(1)
	gfx.SetModelMatrixRect(r)
	addButtonVAO.Render()
}
