package view

import (
	"image"
	math "math"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/view/animation"
)

type button struct {
	icon          *gfx.VAO
	clickCallback func()
	spinning      *animation.Animation

	// bounds is the rectangle with global coords that should be drawn within.
	bounds image.Rectangle
}

func newButton(icon *gfx.VAO) *button {
	return &button{
		icon:     icon,
		spinning: animation.New(0.5*fps, animation.Loop()),
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

func (b *button) ProcessInput(
	bounds image.Rectangle,
	mousePos image.Point,
	mouseLeftButtonReleased bool,
	scheduledCallbacks *[]func(),
) (clicked bool) {
	b.bounds = bounds
	if mouseLeftButtonReleased && mousePos.In(bounds) {
		*scheduledCallbacks = append(*scheduledCallbacks, b.clickCallback)
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

func (b *button) Close() {
	b.clickCallback = nil
}
