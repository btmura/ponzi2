package view

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAnimation_Start_Stop_Update_Value_NoLoop(t *testing.T) {
	a := newAnimation(3)

	want := &animation{
		start:     0,
		end:       1,
		currFrame: 0,
		numFrames: 3,
		state:     aStopped,
	}
	checkFields(t, want, a)
	checkValue(t, 0, a, 0.2) // Fudge has no effect on first frame.

	a.Start()

	want = &animation{
		start:     0,
		end:       1,
		currFrame: 0,
		numFrames: 3,
		state:     aRunning,
	}
	checkFields(t, want, a)
	checkValue(t, 0, a, 0.1) // Fudge still has no effect on first frame.

	checkUpdate(t, true, a)

	want = &animation{
		start:     0,
		end:       1,
		currFrame: 1,
		numFrames: 3,
		state:     aRunning,
	}
	checkFields(t, want, a)
	checkValue(t, 0.55, a, 0.1) // Fudge takes affect.

	a.Stop()

	want = &animation{
		start:     0,
		end:       1,
		currFrame: 1,
		numFrames: 3,
		state:     aFinishing,
	}
	checkFields(t, want, a)
	checkValue(t, 0.5, a, 0)

	// Animation should finish and stop after.
	checkUpdate(t, true, a)

	want = &animation{
		start:     0,
		end:       1,
		currFrame: 2,
		numFrames: 3,
		state:     aFinishing,
	}
	checkFields(t, want, a)
	checkValue(t, 1.0, a, 0)

	// Animation is finished. Update should have no affect.
	checkUpdate(t, false, a)

	want = &animation{
		start:     0,
		end:       1,
		currFrame: 2,
		numFrames: 3,
		state:     aStopped,
	}
	checkFields(t, want, a)
	checkValue(t, 1.0, a, 0.5) // Fudge has no effect on last frame.
}

func TestAnimation_Start_Stop_Update_Value_Loop(t *testing.T) {
	a := newAnimation(3, animationLoop())

	want := &animation{
		start:     0,
		end:       1,
		currFrame: 0,
		numFrames: 3,
		state:     aStopped,
		loop:      true,
	}
	checkFields(t, want, a)
	checkValue(t, 0, a, 0.2) // Fudge has no effect on first frame.

	a.Start()

	want = &animation{
		start:     0,
		end:       1,
		currFrame: 0,
		numFrames: 3,
		state:     aRunning,
		loop:      true,
	}
	checkFields(t, want, a)
	checkValue(t, 0, a, 0.1) // Fudge still has no effect on first frame.

	checkUpdate(t, true, a)

	want = &animation{
		start:     0,
		end:       1,
		currFrame: 1,
		numFrames: 3,
		state:     aRunning,
		loop:      true,
	}
	checkFields(t, want, a)
	checkValue(t, 0.55, a, 0.1) // Fudge takes affect.

	checkUpdate(t, true, a)

	want = &animation{
		start:     0,
		end:       1,
		currFrame: 2,
		numFrames: 3,
		state:     aRunning,
		loop:      true,
	}
	checkFields(t, want, a)
	checkValue(t, 1.0, a, 0)

	// Animation should loop around.
	checkUpdate(t, true, a)

	want = &animation{
		start:     0,
		end:       1,
		currFrame: 0,
		numFrames: 3,
		state:     aRunning,
		loop:      true,
	}
	checkFields(t, want, a)
	checkValue(t, 0, a, 0)

	checkUpdate(t, true, a)

	want = &animation{
		start:     0,
		end:       1,
		currFrame: 1,
		numFrames: 3,
		state:     aRunning,
		loop:      true,
	}
	checkFields(t, want, a)
	checkValue(t, 0.5, a, 0)

	a.Stop()

	want = &animation{
		start:     0,
		end:       1,
		currFrame: 1,
		numFrames: 3,
		state:     aFinishing,
		loop:      true,
	}
	checkFields(t, want, a)
	checkValue(t, 0.5, a, 0)

	// Animation should finish and stop after.
	checkUpdate(t, true, a)

	want = &animation{
		start:     0,
		end:       1,
		currFrame: 2,
		numFrames: 3,
		state:     aFinishing,
		loop:      true,
	}
	checkFields(t, want, a)
	checkValue(t, 1.0, a, 0)

	// Animation is finished. Update should have no affect.
	checkUpdate(t, false, a)

	want = &animation{
		start:     0,
		end:       1,
		currFrame: 2,
		numFrames: 3,
		state:     aStopped,
		loop:      true,
	}
	checkFields(t, want, a)
	checkValue(t, 1.0, a, 0)
}

func TestAnimation_Rewinded(t *testing.T) {
	a := newAnimation(3)
	b := a.Rewinded()

	want := &animation{
		start:     0,
		end:       0,
		currFrame: 0,
		numFrames: 1,
		state:     aStopped,
	}
	checkFields(t, want, b)
	checkValue(t, 0, b, 0)

	a.Start()
	b = a.Rewinded()

	want = &animation{
		start:     0,
		end:       0,
		currFrame: 0,
		numFrames: 1,
		state:     aRunning,
	}
	checkFields(t, want, b)
	checkValue(t, 0, b, 0)

	checkUpdate(t, true, a)
	b = a.Rewinded()

	want = &animation{
		start:     0.5,
		end:       0,
		currFrame: 0,
		numFrames: 2,
		state:     aRunning,
	}
	checkFields(t, want, b)
	checkValue(t, 0.5, b, 0)

	checkUpdate(t, true, a)
	b = a.Rewinded()

	want = &animation{
		start:     1,
		end:       0,
		currFrame: 0,
		numFrames: 3,
		state:     aRunning,
	}
	checkValue(t, 1.0, b, 0)
}

func checkFields(t *testing.T, a, b *animation) {
	t.Helper()
	if diff := cmp.Diff(a, b, cmp.AllowUnexported(animation{})); diff != "" {
		t.Errorf("differs (-want, +got):\n%s", diff)
	}
}

func checkUpdate(t *testing.T, want bool, a *animation) {
	t.Helper()
	if got := a.Update(); got != want {
		t.Errorf("Update() = %t, want %t", got, want)
	}
}

func checkValue(t *testing.T, want float32, a *animation, fudge float32) {
	t.Helper()
	if got := a.Value(fudge); got != want {
		t.Errorf("Value(%f) = %f, want %f", fudge, got, want)
	}
}
