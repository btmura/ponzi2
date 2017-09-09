package ponzi

import (
	"github.com/btmura/ponzi2/internal/gfx"
)

type button struct {
	iconVAO        *gfx.VAO
	clickCallbacks []func()
}

func newButton(iconVAO *gfx.VAO) *button {
	return &button{
		iconVAO: iconVAO,
	}
}

func (b *button) render(vc viewContext) {
	if vc.leftClickedInBounds() {
		vc.scheduleCallbacks(b.clickCallbacks)
	}

	gfx.SetColorMixAmount(1)
	gfx.SetModelMatrixRect(vc.bounds)
	b.iconVAO.Render()
}

func (b *button) addClickCallback(cb func()) {
	b.clickCallbacks = append(b.clickCallbacks, cb)
}
