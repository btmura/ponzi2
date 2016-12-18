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

	win, err := glfw.CreateWindow(640, 480, "ponzi", nil, nil)
	checkErr(err)

	win.MakeContextCurrent()

	r, err := createRenderer()
	checkErr(err)

	// Call the size callback to set the initial viewport.
	w, h := win.GetSize()
	r.resize(image.Pt(w, h))
	win.SetSizeCallback(func(w *glfw.Window, width, height int) {
		r.resize(image.Pt(width, height))
	})

	for !win.ShouldClose() {
		r.render()
		win.SwapBuffers()
		glfw.PollEvents()
	}
}
