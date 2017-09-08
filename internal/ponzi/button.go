package ponzi

import (
	"image"

	"github.com/btmura/ponzi2/internal/gfx"
)

type button struct {
	iconVAO *gfx.VAO
}

func newButton(iconVAO *gfx.VAO) *button {
	return &button{
		iconVAO: iconVAO,
	}
}

func (b *button) render(r image.Rectangle) {
	gfx.SetColorMixAmount(1)
	gfx.SetModelMatrixRect(r)
	b.iconVAO.Render()
}
