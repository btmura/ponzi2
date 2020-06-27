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

// NewStartedFader returns a Fader that starts fading in.
func NewStartedFader(numFrames int) *Fader {
	return &Fader{
		fade: animation.New(numFrames, animation.Started()),
	}
}

// NewStoppedFader returns a Fader that is stopped.
func NewStoppedFader(numFrames int) *Fader {
	return &Fader{
		fade: animation.New(numFrames),
	}
}

// FadeIn starts fading in by reversing the current fade if necessary.
func (f *Fader) FadeIn() {
	if f.fadingOut {
		f.fadingOut = false
		f.fade = f.fade.Rewinded()
	}
	f.fade.Start()
}

// FadeOut starts fading out by reversing the current fade if necessary.
func (f *Fader) FadeOut() {
	if !f.fadingOut {
		f.fadingOut = true
		f.fade = f.fade.Rewinded()
	}
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
func (f *Fader) Render(fudge float32, innerRender func()) {
	old := gfx.Alpha()
	defer gfx.SetAlpha(old)

	gfx.SetAlpha(old * f.fade.Value(fudge))
	innerRender()
}
