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

	vs, err := shaderVertBytes()
	checkErr(err)

	fs, err := shaderFragBytes()
	checkErr(err)

	p, err := glCreateProgram(string(vs), string(fs))
	checkErr(err)

	gl.UseProgram(p)

	vertices := []float32{
		-1, -1, 0,
		0, 1, 0,
		1, -1, 0,
	}

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	{
		gl.BindBuffer(gl.ARRAY_BUFFER, glCreateArrayBuffer(vertices))
		gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, gl.PtrOffset(0))
		gl.EnableVertexAttribArray(0)
	}
	gl.BindVertexArray(0)

	gl.ClearColor(0, 0, 0, 0)

	for !win.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT)

		gl.BindVertexArray(vao)
		gl.DrawArrays(gl.TRIANGLES, 0, 3)
		gl.BindVertexArray(0)

		// Do OpenGL stuff.
		win.SwapBuffers()
		glfw.PollEvents()
	}
}
