package animation

//go:generate stringer -type=State
type State int

const (
	Stopped State = iota
	Running
	Finishing
)

type Animation struct {
	start     float32
	end       float32
	currFrame int
	numFrames int
	loop      bool
	state     State
}

type Opt func(a *Animation)

func New(numFrames int, opts ...Opt) *Animation {
	a := &Animation{
		end:       1,
		numFrames: numFrames,
	}
	for _, o := range opts {
		o(a)
	}
	return a
}

func Loop() Opt {
	return func(a *Animation) {
		a.loop = true
	}
}

func Started() Opt {
	return func(a *Animation) {
		a.Start()
	}
}

func (a *Animation) Reverse() *Animation {
	return &Animation{
		start:     a.end,
		end:       a.start,
		currFrame: a.numFrames - a.currFrame - 1,
		numFrames: a.numFrames,
		loop:      a.loop,
		state:     a.state,
	}
}

func (a *Animation) Start() {
	a.state = Running
}

func (a *Animation) Stop() {
	a.state = Finishing
}

func (a *Animation) Animating() bool {
	return a.state != Stopped
}

func (a *Animation) Update() (dirty bool) {
	switch a.state {
	case Running:
		if a.loop {
			a.currFrame = (a.currFrame + 1) % a.numFrames
			return true
		}

		if a.currFrame < a.numFrames-1 {
			a.currFrame++
			return true
		}
		a.state = Stopped
		return false

	case Finishing:
		if a.currFrame < a.numFrames-1 {
			a.currFrame++
			return true
		}
		a.state = Stopped
		return false

	default:
		return false
	}
}

func (a *Animation) Value(fudge float32) float32 {
	return a.start + (a.end-a.start)*a.percent(fudge)
}

func (a *Animation) percent(fudge float32) float32 {
	switch a.currFrame {
	case 0:
		return 0

	case a.numFrames - 1:
		return 1

	default:
		return (float32(a.currFrame) + fudge) / float32(a.numFrames-1)
	}
}
