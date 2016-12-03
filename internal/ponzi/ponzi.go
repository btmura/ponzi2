package ponzi

import (
	"log"
	"runtime"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

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

	win, err := glfw.CreateWindow(640, 480, "ponzi", nil, nil)
	checkErr(err)

	win.MakeContextCurrent()

	checkErr(gl.Init()) // Must be run after MakeContextCurrent.
	log.Printf("OpenGL version: %s", gl.GoStr(gl.GetString(gl.VERSION)))

	for !win.ShouldClose() {
		// Do OpenGL stuff.
		win.SwapBuffers()
		glfw.PollEvents()
	}
}
