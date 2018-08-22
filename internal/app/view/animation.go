package view

//go:generate stringer -type=animationState
type animationState int

const (
	aStopped animationState = iota
	aRunning
	aFinishing
)

type animation struct {
	start     float32
	end       float32
	currFrame int
	numFrames int
	loop      bool
	state     animationState
}

type animationOpt func(a *animation)

func newAnimation(numFrames int, opts ...animationOpt) *animation {
	a := &animation{
		end:       1,
		numFrames: numFrames,
	}
	for _, o := range opts {
		o(a)
	}
	return a
}

func animationLoop() animationOpt {
	return func(a *animation) {
		a.loop = true
	}
}

// TODO(btmura): add test for Rewinded and start and end values
func (a *animation) Rewinded() *animation {
	return &animation{
		start:     a.Value(0),
		end:       a.start,
		numFrames: a.currFrame + 1,
		loop:      a.loop,
		state:     a.state,
	}
}

func (a *animation) Start() {
	a.state = aRunning
}

func (a *animation) Stop() {
	a.state = aFinishing
}

func (a *animation) Animating() bool {
	return a.state != aStopped
}

func (a *animation) Update() (dirty bool) {
	switch a.state {
	case aRunning:
		if a.loop {
			a.currFrame = (a.currFrame + 1) % a.numFrames
			return true
		}

		if a.currFrame < a.numFrames-1 {
			a.currFrame++
			return true
		}
		a.state = aStopped
		return false

	case aFinishing:
		if a.currFrame < a.numFrames-1 {
			a.currFrame++
			return true
		}
		a.state = aStopped
		return false

	default:
		return false
	}
}

func (a *animation) Value(fudge float32) float32 {
	return a.start + (a.end-a.start)*a.percent(fudge)
}

func (a *animation) percent(fudge float32) float32 {
	switch a.currFrame {
	case 0:
		return 0

	case a.numFrames - 1:
		return 1

	default:
		return (float32(a.currFrame) + fudge) / float32(a.numFrames-1)
	}
}
