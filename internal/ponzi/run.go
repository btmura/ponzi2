package ponzi

import (
	"flag"
	"image"
	"runtime"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang/glog"
)

// Get go-bindata from github.com/jteeuwen/go-bindata. It's used to embed resources into the binary.
//go:generate go-bindata -pkg ponzi -prefix data -ignore ".*blend.*" data
//go:generate go generate github.com/btmura/ponzi2/internal/gfx

func init() {
	flag.Parse() // Avoid glog errors about logging before flag.Parse.

	// This is needed to arrange that main() runs on main thread.
	// See documentation for functions that are only allowed to be called from the main thread.
	runtime.LockOSThread()
}

// Run runs the stock chart viewer in a window.
func Run() {
	if err := glfw.Init(); err != nil {
		glog.Fatalf("Run: failed to init glfw: %v", err)
	}
	defer glfw.Terminate()

	// Set the following hints for Linux compatibility.
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 5)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	win, err := glfw.CreateWindow(800, 600, "ponzi", nil, nil)
	if err != nil {
		glog.Fatalf("Run: failed to create window: %v", err)
	}

	win.MakeContextCurrent()

	cfg, err := LoadConfig()
	if err != nil {
		glog.Fatalf("Run: failed to load config: %v", err)
	}

	m := &Model{}
	v := NewView(m)

	if s := cfg.CurrentStock.Symbol; s != "" {
		st := m.SetCurrentStock(s)
		v.SetChart(st)
		v.GoRefreshStock(st)
	}

	for _, cs := range cfg.Stocks {
		if cs.Symbol == "" {
			continue
		}

		st, added := m.AddSavedStock(cs.Symbol)
		if !added {
			continue
		}

		v.AddChartThumb(st)
		v.GoRefreshStock(st)
	}

	// Call the size callback to set the initial viewport.
	w, h := win.GetSize()
	v.Resize(image.Pt(w, h))
	win.SetSizeCallback(func(_ *glfw.Window, width, height int) {
		v.Resize(image.Pt(width, height))
	})

	win.SetCharCallback(func(_ *glfw.Window, char rune) {
		v.PushInputSymbolChar(char)
	})

	win.SetKeyCallback(func(_ *glfw.Window, key glfw.Key, _ int, action glfw.Action, _ glfw.ModifierKey) {
		if action != glfw.Release {
			return
		}

		switch key {
		case glfw.KeyEscape:
			v.ClearInputSymbol()

		case glfw.KeyBackspace:
			v.PopInputSymbolChar()

		case glfw.KeyEnter:
			v.SubmitSymbol()
		}
	})

	win.SetCursorPosCallback(func(_ *glfw.Window, x, y float64) {
		v.HandleCursorPos(x, y)
	})

	win.SetMouseButtonCallback(func(_ *glfw.Window, button glfw.MouseButton, action glfw.Action, _ glfw.ModifierKey) {
		v.HandleMouseButton(button, action)
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
			// TODO(btmura): update without rendering here
			lag -= secPerUpdate
		}

		fudge := float32(lag / secPerUpdate)
		v.Render(fudge)

		win.SwapBuffers()
		glfw.PollEvents()
	}
}
