package view

import (
	"image"
	math "math"

	"github.com/btmura/ponzi2/internal/app/gfx"
)

type button struct {
	icon          *gfx.VAO
	clickCallback func()
	spinning      *animation

	// bounds is the rectangle with global coords that should be drawn within.
	bounds image.Rectangle
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
	b.bounds = ic.Bounds
	if ic.LeftClickInBounds() {
		*ic.ScheduledCallbacks = append(*ic.ScheduledCallbacks, b.clickCallback)
		return true
	}
	return false
}

func (b *button) Update() (dirty bool) {
	return b.spinning.Update()
}

func (b *button) Render(fudge float32) {
	spinRadians := b.spinning.Value(fudge) * -2 * math.Pi
	gfx.SetModelMatrixRotatedRect(b.bounds, spinRadians)
	b.icon.Render()
}

func (b *button) SetClickCallback(cb func()) {
	b.clickCallback = cb
}
