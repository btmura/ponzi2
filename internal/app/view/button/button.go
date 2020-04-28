package button

import (
	"image"
	math "math"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/view"
	"github.com/btmura/ponzi2/internal/app/view/animation"
)

type Button struct {
	icon          *gfx.VAO
	clickCallback func()
	spinning      *animation.Animation

	// bounds is the rectangle with global coords that should be drawn within.
	bounds image.Rectangle
}

func New(icon *gfx.VAO, fps int) *Button {
	return &Button{
		icon:     icon,
		spinning: animation.New(int(0.5*float32(fps)), animation.Loop()),
	}
}

func (b *Button) StartSpinning() {
	b.spinning.Start()
}

func (b *Button) Spinning() bool {
	return b.spinning.Animating()
}

func (b *Button) StopSpinning() {
	b.spinning.Stop()
}

func (b *Button) SetBounds(bounds image.Rectangle) {
	b.bounds = bounds
}

func (b *Button) ProcessInput(input *view.Input) (clicked bool) {
	if input.MouseLeftButtonClicked.In(b.bounds) {
		input.AddFiredCallback(b.clickCallback)
		return true
	}
	return false
}

func (b *Button) Update() (dirty bool) {
	return b.spinning.Update()
}

func (b *Button) Render(fudge float32) {
	spinRadians := b.spinning.Value(fudge) * -2 * math.Pi
	gfx.SetModelMatrixRotatedRect(b.bounds, spinRadians)
	b.icon.Render()
}

func (b *Button) SetClickCallback(cb func()) {
	b.clickCallback = cb
}

func (b *Button) Close() {
	b.clickCallback = nil
}
