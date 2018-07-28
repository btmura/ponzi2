package view

import (
	math "math"

	"github.com/btmura/ponzi2/internal/gfx"
)

type button struct {
	icon          *gfx.VAO
	clickCallback func()
	spinning      *animation
}

func newButton(icon *gfx.VAO) *button {
	return &button{
		icon:     icon,
		spinning: newAnimation(0.5*fps, true),
	}
}

func (b *button) StartSpinning() {
	b.spinning.Start()
}

func (b *button) StopSpinning() {
	b.spinning.Stop()
}

func (b *button) Update() (dirty bool) {
	return b.spinning.Update()
}

func (b *button) Render(vc viewContext) (clicked bool) {
	if vc.LeftClickInBounds() {
		*vc.ScheduledCallbacks = append(*vc.ScheduledCallbacks, b.clickCallback)
		clicked = true
	}

	spinRadians := b.spinning.Value(vc.Fudge) * -2 * math.Pi
	gfx.SetModelMatrixRotatedRect(vc.Bounds, spinRadians)
	b.icon.Render()

	return clicked
}

func (b *button) SetClickCallback(cb func()) {
	b.clickCallback = cb
}
