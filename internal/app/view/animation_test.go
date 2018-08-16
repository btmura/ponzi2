package view

import "testing"

func TestAnimation_Start_Stop_Update_Value_NoLoop(t *testing.T) {
	a := &animationTest{t, newAnimation(3)}

	a.callValueReturns(0.2, 0) // Fudge has no effect on first frame.
	a.checkCurrFrame(0)
	a.checkNumFrames(3)
	a.checkState(aStopped)

	a.Start()
	a.callValueReturns(0.1, 0) // Fudge still has no effect on first frame.
	a.checkCurrFrame(0)
	a.checkNumFrames(3)
	a.checkState(aRunning)

	a.callUpdateReturns(true)
	a.callValueReturns(0.1, 0.55) // Fudge takes affect.
	a.checkCurrFrame(1)
	a.checkNumFrames(3)
	a.checkState(aRunning)

	a.Stop()
	a.callValueReturns(0, 0.5)
	a.checkCurrFrame(1)
	a.checkNumFrames(3)
	a.checkState(aFinishing)

	// Animation should finish and stop after.
	a.callUpdateReturns(true)
	a.callValueReturns(0, 1.0)
	a.checkCurrFrame(2)
	a.checkNumFrames(3)
	a.checkState(aFinishing)

	// Animation is finished. Update should have no affect.
	a.callUpdateReturns(false)
	a.callValueReturns(0.5, 1.0) // Fudge has no effect on last frame.
	a.checkCurrFrame(2)
	a.checkNumFrames(3)
	a.checkState(aStopped)
}

func TestAnimation_Start_Stop_Update_Value_Loop(t *testing.T) {
	a := &animationTest{t, newAnimation(3, animationLoop())}

	a.callValueReturns(0.2, 0) // Fudge has no effect on first frame.
	a.checkCurrFrame(0)
	a.checkNumFrames(3)
	// a.checkState(false)

	a.Start()
	a.callValueReturns(0.1, 0) // Fudge still has no effect on first frame.
	a.checkCurrFrame(0)
	a.checkNumFrames(3)
	// a.checkState(true)

	a.callUpdateReturns(true)
	a.callValueReturns(0.1, 0.55) // Fudge takes affect.
	a.checkCurrFrame(1)
	a.checkNumFrames(3)
	// a.checkState(true)

	a.callUpdateReturns(true)
	a.callValueReturns(0, 1.0)
	a.checkCurrFrame(2)
	a.checkNumFrames(3)
	// a.checkState(true)

	// Animation should loop around.
	a.callUpdateReturns(true)
	a.callValueReturns(0, 0)
	a.checkCurrFrame(0)
	a.checkNumFrames(3)
	// a.checkState(true)

	a.callUpdateReturns(true)
	a.callValueReturns(0, 0.5)
	a.checkCurrFrame(1)
	a.checkNumFrames(3)
	// a.checkState(true)

	a.Stop()
	a.callValueReturns(0, 0.5)
	a.checkCurrFrame(1)
	a.checkNumFrames(3)
	// a.checkState(false)

	// Animation should finish and stop after.
	a.callUpdateReturns(true)
	a.callValueReturns(0, 1.0)
	a.checkCurrFrame(2)
	a.checkNumFrames(3)
	// a.checkState(false)

	// Animation is finished. Update should have no affect.
	a.callUpdateReturns(false)
	a.callValueReturns(0, 1.0)
	a.checkCurrFrame(2)
	a.checkNumFrames(3)
	// a.checkState(false)
}

type animationTest struct {
	*testing.T
	*animation
}

func (a *animationTest) checkCurrFrame(want int) {
	a.Helper()
	if a.currFrame != want {
		a.Errorf("a.currFrame = %d, want %d", a.currFrame, want)
	}
}

func (a *animationTest) checkNumFrames(want int) {
	a.Helper()
	if a.numFrames != want {
		a.Errorf("a.numFrames = %d, want %d", a.numFrames, want)
	}
}

func (a *animationTest) checkState(want animationState) {
	a.Helper()
	if a.state != want {
		a.Errorf("a.state = %v, want %v", a.state, want)
	}
}

func (a *animationTest) callUpdateReturns(want bool) {
	a.Helper()
	if got := a.Update(); got != want {
		a.Errorf("a.Update() = %t, want %t", got, want)
	}
}

func (a *animationTest) callValueReturns(fudge, want float32) {
	a.Helper()
	if got := a.Value(fudge); got != want {
		a.Errorf("a.Value(%f) = %f, want %f", fudge, got, want)
	}
}
