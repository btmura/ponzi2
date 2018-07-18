package view

type animation struct {
	numFrames int
	loop      bool
	currFrame int
	running   bool
}

func newAnimation(numFrames int, loop bool) *animation {
	return &animation{numFrames: numFrames, loop: loop}
}

func (a *animation) Start() {
	a.running = true
}

func (a *animation) Stop() {
	a.running = false
}

func (a *animation) Update() (animating bool) {
	if a.loop {
		if a.running || a.currFrame != 0 {
			a.currFrame = (a.currFrame + 1) % a.numFrames
			return true
		}
		return false
	}

	if !a.running && a.currFrame == 0 {
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
