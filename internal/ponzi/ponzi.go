package ponzi

import (
	"image"
	"runtime"

	"github.com/go-gl/glfw/v3.2/glfw"
)

// Get go-bindata from github.com/jteeuwen/go-bindata. It's used to embed resources into the binary.
//go:generate go-bindata -pkg ponzi -prefix data -ignore ".*blend.*" data

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

	m := &model{}
	v, err := createView(m)
	checkErr(err)

	// GLFW, GL, and shaders OK! Go fetch data for the model.
	m.startRefresh()

	// Call the size callback to set the initial viewport.
	w, h := win.GetSize()
	v.resize(image.Pt(w, h))
	win.SetSizeCallback(func(w *glfw.Window, width, height int) {
		v.resize(image.Pt(width, height))
	})

	// Register the key callback.
	win.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		v.handleKey(key, action)
	})

	// Register the char callback.
	win.SetCharCallback(func(w *glfw.Window, char rune) {
		v.handleChar(char)
	})

	for !win.ShouldClose() {
		v.render()
		win.SwapBuffers()
		glfw.PollEvents()
	}
}
