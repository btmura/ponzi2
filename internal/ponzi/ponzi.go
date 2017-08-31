package ponzi

import (
	"image"
	"runtime"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang/glog"
)

//go:generate go generate github.com/btmura/ponzi2/internal/gfx

func init() {
	// This is needed to arrange that main() runs on main thread.
	// See documentation for functions that are only allowed to be called from the main thread.
	runtime.LockOSThread()
}

// Run runs the stock chart viewer in a window.
func Run() {
	checkErr := func(err error) {
		if err != nil {
			panic(err)
		}
	}

	checkErr(glfw.Init())
	defer glfw.Terminate()

	// Set the following hints for Linux compatibility.
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 5)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	win, err := glfw.CreateWindow(800, 600, "ponzi", nil, nil)
	checkErr(err)

	win.MakeContextCurrent()

	m := newModel("SPY")
	v, err := createView(m)
	checkErr(err)

	// GLFW, GL, and shaders OK! Go fetch data for the model.
	go func() {
		if err := m.refresh(); err != nil {
			glog.Errorf("refresh failed: %v", err)
		}
	}()

	// Call the size callback to set the initial viewport.
	w, h := win.GetSize()
	v.resize(image.Pt(w, h))
	win.SetSizeCallback(func(w *glfw.Window, width, height int) {
		v.resize(image.Pt(width, height))
	})

	win.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		v.handleKey(key, action)
	})

	win.SetCharCallback(func(w *glfw.Window, char rune) {
		v.handleChar(char)
	})

	const secPerUpdate = 1.0 / 60

	var lag float64
	prevTime := glfw.GetTime()
	for !win.ShouldClose() {
		currTime := glfw.GetTime()
		elapsed := currTime - prevTime
		prevTime = currTime
		lag += elapsed

		for lag >= secPerUpdate {
			v.update()
			lag -= secPerUpdate
		}

		fudge := float32(lag / secPerUpdate)
		v.render(fudge)

		win.SwapBuffers()
		glfw.PollEvents()
	}
}
