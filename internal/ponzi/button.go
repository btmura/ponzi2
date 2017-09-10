package ponzi

import (
	"github.com/btmura/ponzi2/internal/gfx"
)

// Button is a button that can be rendered and clicked.
type Button struct {
	iconVAO        *gfx.VAO
	clickCallbacks []func()
}

// NewButton creates a new button.
func NewButton(iconVAO *gfx.VAO) *Button {
	return &Button{
		iconVAO: iconVAO,
	}
}

// Render renders the button and detects button clicks.
func (b *Button) Render(vc ViewContext) {
	if vc.LeftClickInBounds() {
		vc.scheduleCallbacks(b.clickCallbacks)
	}

	gfx.SetColorMixAmount(1)
	gfx.SetModelMatrixRect(vc.bounds)
	b.iconVAO.Render()
}

// AddClickCallback adds a callback for when the button is clicked.
func (b *Button) AddClickCallback(cb func()) {
	b.clickCallbacks = append(b.clickCallbacks, cb)
}
