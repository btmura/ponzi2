package view

type animation struct {
	currFrame int
	numFrames int
	started   bool
}

func newAnimation(numFrames int) *animation {
	return &animation{numFrames: numFrames}
}

func (a *animation) Start() {
	a.started = true
	a.currFrame = 0
}

func (a *animation) Update() (animating bool) {
	if !a.started {
		return false
	}

	if a.currFrame+1 < a.numFrames {
		a.currFrame++
		return true
	}

	return false
}

func (a *animation) Value(fudge float32) float32 {
	return (float32(a.currFrame) + fudge) / float32(a.numFrames-1)
}
