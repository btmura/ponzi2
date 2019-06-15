package view

import (
	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/view/animation"
)

// Fader fades in and out by changing the current alpha value.
type Fader struct {
	fadingOut bool
	fade      *animation.Animation
}

// NewFader returns a Fader that starts fading in.
func NewFader(numFrames int) *Fader {
	return &Fader{
		fade: animation.New(numFrames, animation.Started()),
	}
}

// FadeOut starts fading out by reversing the current fade in animation.
func (f *Fader) FadeOut() {
	f.fadingOut = true
	f.fade = f.fade.Rewinded()
	f.fade.Start()
}

// DoneFadingOut returns true if the fade out animation is done.
func (f *Fader) DoneFadingOut() bool {
	return f.fadingOut && !f.fade.Animating()
}

// Update moves the animation one step forward.
func (f *Fader) Update() (dirty bool) {
	if f.fade.Update() {
		dirty = true
	}
	return dirty
}

// Render adjusts the alpha, calls the given inner render function, and then restores the alpha.
func (f *Fader) Render(innerRender func(fudge float32), fudge float32) {
	old := gfx.Alpha()
	defer gfx.SetAlpha(old)

	gfx.SetAlpha(old * f.fade.Value(fudge))
	innerRender(fudge)
}
