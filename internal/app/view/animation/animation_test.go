package animation

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAnimation_Start_Stop_Update_Value_NoLoop(t *testing.T) {
	a := New(3)

	want := &Animation{
		start:     0,
		end:       1,
		currFrame: 0,
		numFrames: 3,
		state:     Stopped,
	}
	checkFields(t, want, a)
	checkValue(t, 0, a, 0.2) // Fudge has no effect on first frame.

	a.Start()

	want = &Animation{
		start:     0,
		end:       1,
		currFrame: 0,
		numFrames: 3,
		state:     Running,
	}
	checkFields(t, want, a)
	checkValue(t, 0, a, 0.1) // Fudge still has no effect on first frame.

	checkUpdate(t, true, a)

	want = &Animation{
		start:     0,
		end:       1,
		currFrame: 1,
		numFrames: 3,
		state:     Running,
	}
	checkFields(t, want, a)
	checkValue(t, 0.55, a, 0.1) // Fudge takes affect.

	a.Stop()

	want = &Animation{
		start:     0,
		end:       1,
		currFrame: 1,
		numFrames: 3,
		state:     Finishing,
	}
	checkFields(t, want, a)
	checkValue(t, 0.5, a, 0)

	// Animation should finish and stop after.
	checkUpdate(t, true, a)

	want = &Animation{
		start:     0,
		end:       1,
		currFrame: 2,
		numFrames: 3,
		state:     Finishing,
	}
	checkFields(t, want, a)
	checkValue(t, 1.0, a, 0)

	// Animation is finished. Update should have no affect.
	checkUpdate(t, false, a)

	want = &Animation{
		start:     0,
		end:       1,
		currFrame: 2,
		numFrames: 3,
		state:     Stopped,
	}
	checkFields(t, want, a)
	checkValue(t, 1.0, a, 0.5) // Fudge has no effect on last frame.
}

func TestAnimation_Start_Stop_Update_Value_Loop(t *testing.T) {
	a := New(3, Loop())

	want := &Animation{
		start:     0,
		end:       1,
		currFrame: 0,
		numFrames: 3,
		state:     Stopped,
		loop:      true,
	}
	checkFields(t, want, a)
	checkValue(t, 0, a, 0.2) // Fudge has no effect on first frame.

	a.Start()

	want = &Animation{
		start:     0,
		end:       1,
		currFrame: 0,
		numFrames: 3,
		state:     Running,
		loop:      true,
	}
	checkFields(t, want, a)
	checkValue(t, 0, a, 0.1) // Fudge still has no effect on first frame.

	checkUpdate(t, true, a)

	want = &Animation{
		start:     0,
		end:       1,
		currFrame: 1,
		numFrames: 3,
		state:     Running,
		loop:      true,
	}
	checkFields(t, want, a)
	checkValue(t, 0.55, a, 0.1) // Fudge takes affect.

	checkUpdate(t, true, a)

	want = &Animation{
		start:     0,
		end:       1,
		currFrame: 2,
		numFrames: 3,
		state:     Running,
		loop:      true,
	}
	checkFields(t, want, a)
	checkValue(t, 1.0, a, 0)

	// Animation should loop around.
	checkUpdate(t, true, a)

	want = &Animation{
		start:     0,
		end:       1,
		currFrame: 0,
		numFrames: 3,
		state:     Running,
		loop:      true,
	}
	checkFields(t, want, a)
	checkValue(t, 0, a, 0)

	checkUpdate(t, true, a)

	want = &Animation{
		start:     0,
		end:       1,
		currFrame: 1,
		numFrames: 3,
		state:     Running,
		loop:      true,
	}
	checkFields(t, want, a)
	checkValue(t, 0.5, a, 0)

	a.Stop()

	want = &Animation{
		start:     0,
		end:       1,
		currFrame: 1,
		numFrames: 3,
		state:     Finishing,
		loop:      true,
	}
	checkFields(t, want, a)
	checkValue(t, 0.5, a, 0)

	// Animation should finish and stop after.
	checkUpdate(t, true, a)

	want = &Animation{
		start:     0,
		end:       1,
		currFrame: 2,
		numFrames: 3,
		state:     Finishing,
		loop:      true,
	}
	checkFields(t, want, a)
	checkValue(t, 1.0, a, 0)

	// Animation is finished. Update should have no affect.
	checkUpdate(t, false, a)

	want = &Animation{
		start:     0,
		end:       1,
		currFrame: 2,
		numFrames: 3,
		state:     Stopped,
		loop:      true,
	}
	checkFields(t, want, a)
	checkValue(t, 1.0, a, 0)
}

func TestAnimation_Rewinded(t *testing.T) {
	a := New(3)
	b := a.Rewinded()

	want := &Animation{
		start:     0,
		end:       0,
		currFrame: 0,
		numFrames: 1,
		state:     Stopped,
	}
	checkFields(t, want, b)
	checkValue(t, 0, b, 0)

	a.Start()
	b = a.Rewinded()

	want = &Animation{
		start:     0,
		end:       0,
		currFrame: 0,
		numFrames: 1,
		state:     Running,
	}
	checkFields(t, want, b)
	checkValue(t, 0, b, 0)

	checkUpdate(t, true, a)
	b = a.Rewinded()

	want = &Animation{
		start:     0.5,
		end:       0,
		currFrame: 0,
		numFrames: 2,
		state:     Running,
	}
	checkFields(t, want, b)
	checkValue(t, 0.5, b, 0)

	checkUpdate(t, true, a)
	b = a.Rewinded()

	want = &Animation{
		start:     1,
		end:       0,
		currFrame: 0,
		numFrames: 3,
		state:     Running,
	}
	checkValue(t, 1.0, b, 0)
}

func checkFields(t *testing.T, a, b *Animation) {
	t.Helper()
	if diff := cmp.Diff(a, b, cmp.AllowUnexported(Animation{})); diff != "" {
		t.Errorf("differs (-want, +got):\n%s", diff)
	}
}

func checkUpdate(t *testing.T, want bool, a *Animation) {
	t.Helper()
	if got := a.Update(); got != want {
		t.Errorf("Update() = %t, want %t", got, want)
	}
}

func checkValue(t *testing.T, want float32, a *Animation, fudge float32) {
	t.Helper()
	if got := a.Value(fudge); got != want {
		t.Errorf("Value(%f) = %f, want %f", fudge, got, want)
	}
}
