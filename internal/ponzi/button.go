package ponzi

import (
	"github.com/btmura/ponzi2/internal/gfx"
)

type Button struct {
	iconVAO        *gfx.VAO
	clickCallbacks []func()
}

func NewButton(iconVAO *gfx.VAO) *Button {
	return &Button{
		iconVAO: iconVAO,
	}
}

func (b *Button) Render(vc ViewContext) {
	if vc.LeftClickInBounds() {
		vc.scheduleCallbacks(b.clickCallbacks)
	}

	gfx.SetColorMixAmount(1)
	gfx.SetModelMatrixRect(vc.bounds)
	b.iconVAO.Render()
}

func (b *Button) AddClickCallback(cb func()) {
	b.clickCallbacks = append(b.clickCallbacks, cb)
}
