package view

import (
	"testing"
)

func TestAnimation_Start_Update_Value(t *testing.T) {
	a := newAnimation(3, false)

	checkState := func(wantCurrFrame, wantNumFrames int, wantStarted bool) {
		t.Helper()
		if a.currFrame != wantCurrFrame {
			t.Errorf("a.currFrame = %d, want %d", a.currFrame, wantCurrFrame)
		}
		if a.numFrames != wantNumFrames {
			t.Errorf("a.numFrames = %d, want %d", a.numFrames, wantNumFrames)
		}
		if a.running != wantStarted {
			t.Errorf("a.started = %t, want %t", a.running, wantStarted)
		}
	}

	checkUpdate := func(want bool) {
		t.Helper()
		if got := a.Update(); got != want {
			t.Errorf("a.Update() = %t, want %t", got, want)
		}
	}

	checkValue := func(want float32) {
		t.Helper()
		if got := a.Value(0); got != want {
			t.Errorf("a.Value() = %f, want %f", got, want)
		}
	}

	checkState(0, 3, false)

	a.Start()
	checkValue(0)
	checkState(0, 3, true)

	checkUpdate(true)
	checkValue(0.5)
	checkState(1, 3, true)

	checkUpdate(true)
	checkValue(1.0)
	checkState(2, 3, true)

	// Animation is finished. Update should have no affect.
	checkUpdate(false)
	checkValue(1.0)
	checkState(2, 3, true)
}
