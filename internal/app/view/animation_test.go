package view

import "testing"

func TestAnimation_Start_Stop_Update_Value_NoLoop(t *testing.T) {
	a := newAnimation(3, false)
	checkState(t, a, 0, 3, false)

	a.Start()
	checkValue(t, a, 0)
	checkState(t, a, 0, 3, true)

	checkUpdate(t, a, true)
	checkValue(t, a, 0.5)
	checkState(t, a, 1, 3, true)

	a.Stop()
	checkValue(t, a, 0.5)
	checkState(t, a, 1, 3, false)

	// Animation should finish and stop after.
	checkUpdate(t, a, true)
	checkValue(t, a, 1.0)
	checkState(t, a, 2, 3, false)

	// Animation is finished. Update should have no affect.
	checkUpdate(t, a, false)
	checkValue(t, a, 1.0)
	checkState(t, a, 2, 3, false)
}

func TestAnimation_Start_Stop_Update_Value_Loop(t *testing.T) {
	a := newAnimation(3, true)
	checkState(t, a, 0, 3, false)

	a.Start()
	checkValue(t, a, 0)
	checkState(t, a, 0, 3, true)

	checkUpdate(t, a, true)
	checkValue(t, a, 0.5)
	checkState(t, a, 1, 3, true)

	checkUpdate(t, a, true)
	checkValue(t, a, 1.0)
	checkState(t, a, 2, 3, true)

	// Animation should loop around.
	checkUpdate(t, a, true)
	checkValue(t, a, 0)
	checkState(t, a, 0, 3, true)

	checkUpdate(t, a, true)
	checkValue(t, a, 0.5)
	checkState(t, a, 1, 3, true)

	a.Stop()
	checkValue(t, a, 0.5)
	checkState(t, a, 1, 3, false)

	// Animation should finish and stop after.
	checkUpdate(t, a, true)
	checkValue(t, a, 1.0)
	checkState(t, a, 2, 3, false)

	// Animation is finished. Update should have no affect.
	checkUpdate(t, a, false)
	checkValue(t, a, 1.0)
	checkState(t, a, 2, 3, false)
}

func checkState(t *testing.T, a *animation, wantCurrFrame, wantNumFrames int, wantRunning bool) {
	t.Helper()
	if a.currFrame != wantCurrFrame {
		t.Errorf("a.currFrame = %d, want %d", a.currFrame, wantCurrFrame)
	}
	if a.numFrames != wantNumFrames {
		t.Errorf("a.numFrames = %d, want %d", a.numFrames, wantNumFrames)
	}
	if a.running != wantRunning {
		t.Errorf("a.running = %t, want %t", a.running, wantRunning)
	}
}

func checkUpdate(t *testing.T, a *animation, want bool) {
	t.Helper()
	if got := a.Update(); got != want {
		t.Errorf("a.Update() = %t, want %t", got, want)
	}
}

func checkValue(t *testing.T, a *animation, want float32) {
	t.Helper()
	if got := a.Value(0); got != want {
		t.Errorf("a.Value() = %f, want %f", got, want)
	}
}
