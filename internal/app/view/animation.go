package view

import (
	"time"
)

type animation struct {
	numFrames int
	currFrame int
	animating bool
}

func newAnimation(dur time.Duration) *animation {
	return &animation{
		numFrames: int(dur.Seconds() * fps),
	}
}

func (a *animation) Start() {
	a.animating = true
	a.currFrame = 0
}

func (a *animation) Update() (animating bool) {
	if !a.animating {
		return false
	}

	if a.currFrame+1 < a.numFrames {
		a.currFrame++
		return true
	}

	if a.currFrame+1 == a.numFrames {
		a.animating = false
		return false
	}

	return false
}

func (a *animation) Value(fudge float32) float32 {
	return (float32(a.currFrame) + 0) / float32(a.numFrames-1)
}
