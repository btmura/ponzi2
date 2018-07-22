package view

import "testing"

func TestAnimation_Start_Stop_Update_Value_NoLoop(t *testing.T) {
	a := &animationTest{t, newAnimation(3, false)}

	a.callValueReturns(0.2, 0) // Fudge has no effect on first frame.
	a.checkCurrFrame(0)
	a.checkNumFrames(3)
	a.checkRunning(false)

	a.Start()
	a.callValueReturns(0.1, 0) // Fudge still has no effect on first frame.
	a.checkCurrFrame(0)
	a.checkNumFrames(3)
	a.checkRunning(true)

	a.callUpdateReturns(true)
	a.callValueReturns(0.1, 0.55) // Fudge takes affect.
	a.checkCurrFrame(1)
	a.checkNumFrames(3)
	a.checkRunning(true)

	a.Stop()
	a.callValueReturns(0, 0.5)
	a.checkCurrFrame(1)
	a.checkNumFrames(3)
	a.checkRunning(false)

	// Animation should finish and stop after.
	a.callUpdateReturns(true)
	a.callValueReturns(0, 1.0)
	a.checkCurrFrame(2)
	a.checkNumFrames(3)
	a.checkRunning(false)

	// Animation is finished. Update should have no affect.
	a.callUpdateReturns(false)
	a.callValueReturns(0.5, 1.0) // Fudge has no effect on last frame.
	a.checkCurrFrame(2)
	a.checkNumFrames(3)
	a.checkRunning(false)
}

func TestAnimation_Start_Stop_Update_Value_Loop(t *testing.T) {
	a := &animationTest{t, newAnimation(3, true)}

	a.callValueReturns(0.2, 0) // Fudge has no effect on first frame.
	a.checkCurrFrame(0)
	a.checkNumFrames(3)
	a.checkRunning(false)

	a.Start()
	a.callValueReturns(0.1, 0) // Fudge still has no effect on first frame.
	a.checkCurrFrame(0)
	a.checkNumFrames(3)
	a.checkRunning(true)

	a.callUpdateReturns(true)
	a.callValueReturns(0.1, 0.55) // Fudge takes affect.
	a.checkCurrFrame(1)
	a.checkNumFrames(3)
	a.checkRunning(true)

	a.callUpdateReturns(true)
	a.callValueReturns(0, 1.0)
	a.checkCurrFrame(2)
	a.checkNumFrames(3)
	a.checkRunning(true)

	// Animation should loop around.
	a.callUpdateReturns(true)
	a.callValueReturns(0, 0)
	a.checkCurrFrame(0)
	a.checkNumFrames(3)
	a.checkRunning(true)

	a.callUpdateReturns(true)
	a.callValueReturns(0, 0.5)
	a.checkCurrFrame(1)
	a.checkNumFrames(3)
	a.checkRunning(true)

	a.Stop()
	a.callValueReturns(0, 0.5)
	a.checkCurrFrame(1)
	a.checkNumFrames(3)
	a.checkRunning(false)

	// Animation should finish and stop after.
	a.callUpdateReturns(true)
	a.callValueReturns(0, 1.0)
	a.checkCurrFrame(2)
	a.checkNumFrames(3)
	a.checkRunning(false)

	// Animation is finished. Update should have no affect.
	a.callUpdateReturns(false)
	a.callValueReturns(0, 1.0)
	a.checkCurrFrame(2)
	a.checkNumFrames(3)
	a.checkRunning(false)
}

type animationTest struct {
	*testing.T
	*animation
}

func (a *animationTest) checkCurrFrame(wantCurrFrame int) {
	a.Helper()
	if a.currFrame != wantCurrFrame {
		a.Errorf("a.currFrame = %d, want %d", a.currFrame, wantCurrFrame)
	}
}

func (a *animationTest) checkNumFrames(wantNumFrames int) {
	a.Helper()
	if a.numFrames != wantNumFrames {
		a.Errorf("a.numFrames = %d, want %d", a.numFrames, wantNumFrames)
	}
}

func (a *animationTest) checkRunning(wantRunning bool) {
	if a.running != wantRunning {
		a.Errorf("a.running = %t, want %t", a.running, wantRunning)
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
