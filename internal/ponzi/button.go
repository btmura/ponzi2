package ponzi

import (
	"github.com/golang/glog"

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

func (b *button) render(vc viewContext) {
	if vc.leftClickedInBounds() {
		glog.Infof("bounds: %v pos: %v", vc.bounds, vc.mousePos)
	}

	gfx.SetColorMixAmount(1)
	gfx.SetModelMatrixRect(vc.bounds)
	b.iconVAO.Render()
}
