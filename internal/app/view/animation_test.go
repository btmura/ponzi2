package view

import "testing"

func TestAnimation_Start_Stop_Update_Value_NoLoop(t *testing.T) {
	at := &animationTester{t}

	a := newAnimation(3)

	at.callValueReturns(a, 0.2, 0) // Fudge has no effect on first frame.
	at.checkCurrFrame(a, 0)
	at.checkNumFrames(a, 3)
	at.checkState(a, aStopped)

	a.Start()
	at.callValueReturns(a, 0.1, 0) // Fudge still has no effect on first frame.
	at.checkCurrFrame(a, 0)
	at.checkNumFrames(a, 3)
	at.checkState(a, aRunning)

	at.callUpdateReturns(a, true)
	at.callValueReturns(a, 0.1, 0.55) // Fudge takes affect.
	at.checkCurrFrame(a, 1)
	at.checkNumFrames(a, 3)
	at.checkState(a, aRunning)

	a.Stop()
	at.callValueReturns(a, 0, 0.5)
	at.checkCurrFrame(a, 1)
	at.checkNumFrames(a, 3)
	at.checkState(a, aFinishing)

	// Animation should finish and stop after.
	at.callUpdateReturns(a, true)
	at.callValueReturns(a, 0, 1.0)
	at.checkCurrFrame(a, 2)
	at.checkNumFrames(a, 3)
	at.checkState(a, aFinishing)

	// Animation is finished. Update should have no affect.
	at.callUpdateReturns(a, false)
	at.callValueReturns(a, 0.5, 1.0) // Fudge has no effect on last frame.
	at.checkCurrFrame(a, 2)
	at.checkNumFrames(a, 3)
	at.checkState(a, aStopped)
}

func TestAnimation_Start_Stop_Update_Value_Loop(t *testing.T) {
	at := &animationTester{t}
	a := newAnimation(3, animationLoop())

	at.callValueReturns(a, 0.2, 0) // Fudge has no effect on first frame.
	at.checkCurrFrame(a, 0)
	at.checkNumFrames(a, 3)
	at.checkState(a, aStopped)

	a.Start()
	at.callValueReturns(a, 0.1, 0) // Fudge still has no effect on first frame.
	at.checkCurrFrame(a, 0)
	at.checkNumFrames(a, 3)
	at.checkState(a, aRunning)

	at.callUpdateReturns(a, true)
	at.callValueReturns(a, 0.1, 0.55) // Fudge takes affect.
	at.checkCurrFrame(a, 1)
	at.checkNumFrames(a, 3)
	at.checkState(a, aRunning)

	at.callUpdateReturns(a, true)
	at.callValueReturns(a, 0, 1.0)
	at.checkCurrFrame(a, 2)
	at.checkNumFrames(a, 3)
	at.checkState(a, aRunning)

	// Animation should loop around.
	at.callUpdateReturns(a, true)
	at.callValueReturns(a, 0, 0)
	at.checkCurrFrame(a, 0)
	at.checkNumFrames(a, 3)
	at.checkState(a, aRunning)

	at.callUpdateReturns(a, true)
	at.callValueReturns(a, 0, 0.5)
	at.checkCurrFrame(a, 1)
	at.checkNumFrames(a, 3)
	at.checkState(a, aRunning)

	a.Stop()
	at.callValueReturns(a, 0, 0.5)
	at.checkCurrFrame(a, 1)
	at.checkNumFrames(a, 3)
	at.checkState(a, aFinishing)

	// Animation should finish and stop after.
	at.callUpdateReturns(a, true)
	at.callValueReturns(a, 0, 1.0)
	at.checkCurrFrame(a, 2)
	at.checkNumFrames(a, 3)
	at.checkState(a, aFinishing)

	// Animation is finished. Update should have no affect.
	at.callUpdateReturns(a, false)
	at.callValueReturns(a, 0, 1.0)
	at.checkCurrFrame(a, 2)
	at.checkNumFrames(a, 3)
	at.checkState(a, aStopped)
}

func TestAnimation_Rewinded(t *testing.T) {
	at := &animationTester{t}

	a := newAnimation(3)

	b := a.Rewinded()
	at.callValueReturns(b, 0.2, 0) // Fudge has no effect on first frame.
	at.checkCurrFrame(b, 0)
	at.checkNumFrames(b, 1)
	at.checkState(b, aStopped)
}

type animationTester struct {
	*testing.T
}

func (at *animationTester) checkCurrFrame(a *animation, want int) {
	at.Helper()
	if a.currFrame != want {
		at.Errorf("a.currFrame = %d, want %d", a.currFrame, want)
	}
}

func (at *animationTester) checkNumFrames(a *animation, want int) {
	at.Helper()
	if a.numFrames != want {
		at.Errorf("a.numFrames = %d, want %d", a.numFrames, want)
	}
}

func (at *animationTester) checkState(a *animation, want animationState) {
	at.Helper()
	if a.state != want {
		at.Errorf("a.state = %v, want %v", a.state, want)
	}
}

func (at *animationTester) callUpdateReturns(a *animation, want bool) {
	at.Helper()
	if got := a.Update(); got != want {
		at.Errorf("a.Update() = %t, want %t", got, want)
	}
}

func (at *animationTester) callValueReturns(a *animation, fudge, want float32) {
	at.Helper()
	if got := a.Value(fudge); got != want {
		at.Errorf("a.Value(%f) = %f, want %f", fudge, got, want)
	}
}
