package view

import "testing"

func TestAnimation_Start_Stop_Update_Value_NoLoop(t *testing.T) {
	a := &animationTest{t, newAnimation(3, false)}

	a.checkCurrFrame(0)
	a.checkNumFrames(3)
	a.checkRunning(false)

	a.Start()
	a.callValue(0)
	a.checkCurrFrame(0)
	a.checkNumFrames(3)
	a.checkRunning(true)

	a.callUpdate(true)
	a.callValue(0.5)
	a.checkCurrFrame(1)
	a.checkNumFrames(3)
	a.checkRunning(true)

	a.Stop()
	a.callValue(0.5)
	a.checkCurrFrame(1)
	a.checkNumFrames(3)
	a.checkRunning(false)

	// Animation should finish and stop after.
	a.callUpdate(true)
	a.callValue(1.0)
	a.checkCurrFrame(2)
	a.checkNumFrames(3)
	a.checkRunning(false)

	// Animation is finished. Update should have no affect.
	a.callUpdate(false)
	a.callValue(1.0)
	a.checkCurrFrame(2)
	a.checkNumFrames(3)
	a.checkRunning(false)
}

func TestAnimation_Start_Stop_Update_Value_Loop(t *testing.T) {
	a := &animationTest{t, newAnimation(3, true)}

	a.checkCurrFrame(0)
	a.checkNumFrames(3)
	a.checkRunning(false)

	a.Start()
	a.callValue(0)
	a.checkCurrFrame(0)
	a.checkNumFrames(3)
	a.checkRunning(true)

	a.callUpdate(true)
	a.callValue(0.5)
	a.checkCurrFrame(1)
	a.checkNumFrames(3)
	a.checkRunning(true)

	a.callUpdate(true)
	a.callValue(1.0)
	a.checkCurrFrame(2)
	a.checkNumFrames(3)
	a.checkRunning(true)

	// Animation should loop around.
	a.callUpdate(true)
	a.callValue(0)
	a.checkCurrFrame(0)
	a.checkNumFrames(3)
	a.checkRunning(true)

	a.callUpdate(true)
	a.callValue(0.5)
	a.checkCurrFrame(1)
	a.checkNumFrames(3)
	a.checkRunning(true)

	a.Stop()
	a.callValue(0.5)
	a.checkCurrFrame(1)
	a.checkNumFrames(3)
	a.checkRunning(false)

	// Animation should finish and stop after.
	a.callUpdate(true)
	a.callValue(1.0)
	a.checkCurrFrame(2)
	a.checkNumFrames(3)
	a.checkRunning(false)

	// Animation is finished. Update should have no affect.
	a.callUpdate(false)
	a.callValue(1.0)
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

func (a *animationTest) callUpdate(want bool) {
	a.Helper()
	if got := a.Update(); got != want {
		a.Errorf("a.Update() = %t, want %t", got, want)
	}
}

func (a *animationTest) callValue(want float32) {
	a.Helper()
	if got := a.Value(0); got != want {
		a.Errorf("a.Value() = %f, want %f", got, want)
	}
}
