package view

import (
	math "math"

	"github.com/btmura/ponzi2/internal/app/gfx"
)

type button struct {
	icon          *gfx.VAO
	clickCallback func()
	spinning      *animation
}

func newButton(icon *gfx.VAO) *button {
	return &button{
		icon:     icon,
		spinning: newAnimation(0.5*fps, animationLoop()),
	}
}

func (b *button) StartSpinning() {
	b.spinning.Start()
}

func (b *button) Spinning() bool {
	return b.spinning.Animating()
}

func (b *button) StopSpinning() {
	b.spinning.Stop()
}

func (b *button) ProcessInput(ic inputContext) (clicked bool) {
	if ic.LeftClickInBounds() {
		*ic.ScheduledCallbacks = append(*ic.ScheduledCallbacks, b.clickCallback)
		return true
	}
	return false
}

func (b *button) Update() (dirty bool) {
	return b.spinning.Update()
}

func (b *button) Render(rc renderContext) {
	spinRadians := b.spinning.Value(rc.Fudge) * -2 * math.Pi
	gfx.SetModelMatrixRotatedRect(rc.Bounds, spinRadians)
	b.icon.Render()
}

func (b *button) SetClickCallback(cb func()) {
	b.clickCallback = cb
}
