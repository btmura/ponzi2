package ponzi

import (
	"image"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang/glog"

	"github.com/btmura/ponzi2/internal/gfx"
)

type Controller struct {
	model *Model
	view  *View
}

func (p *Controller) Run() {
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

	if err := gl.Init(); err != nil {
		glog.Fatalf("newView: failed to init OpenGL: %v", err)
	}
	glog.Infof("OpenGL version: %s", gl.GoStr(gl.GetString(gl.VERSION)))

	gl.Enable(gl.CULL_FACE)
	gl.CullFace(gl.BACK)

	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.ClearColor(0, 0, 0, 0)

	if err := gfx.InitProgram(); err != nil {
		glog.Fatalf("newView: failed to init gfx: %v", err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		glog.Fatalf("Run: failed to load config: %v", err)
	}

	if s := cfg.CurrentStock.Symbol; s != "" {
		p.setChart(s)
	}

	for _, cs := range cfg.Stocks {
		p.addChartThumb(cs.Symbol)
	}

	// Call the size callback to set the initial viewport.
	w, h := win.GetSize()
	p.view.Resize(image.Pt(w, h))
	win.SetSizeCallback(func(_ *glfw.Window, width, height int) {
		p.view.Resize(image.Pt(width, height))
	})

	win.SetCharCallback(func(_ *glfw.Window, char rune) {
		p.view.PushInputSymbolChar(char)
	})

	win.SetKeyCallback(func(_ *glfw.Window, key glfw.Key, _ int, action glfw.Action, _ glfw.ModifierKey) {
		if action != glfw.Release {
			return
		}

		switch key {
		case glfw.KeyEscape:
			p.view.ClearInputSymbol()

		case glfw.KeyBackspace:
			p.view.PopInputSymbolChar()

		case glfw.KeyEnter:
			p.view.SubmitSymbol()
		}
	})

	win.SetCursorPosCallback(func(_ *glfw.Window, x, y float64) {
		p.view.HandleCursorPos(x, y)
	})

	win.SetMouseButtonCallback(func(_ *glfw.Window, button glfw.MouseButton, action glfw.Action, _ glfw.ModifierKey) {
		p.view.HandleMouseButton(button, action)
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
		p.view.Render(fudge)

		win.SwapBuffers()
		glfw.PollEvents()
	}
}

func (p *Controller) setChart(symbol string) {
	if symbol == "" {
		return
	}
	st := p.model.SetCurrentStock(symbol)
	p.view.SetChart(st)
	p.view.GoRefreshStock(st)
}

func (p *Controller) addChartThumb(symbol string) {
	if symbol == "" {
		return
	}

	st, added := p.model.AddSavedStock(symbol)
	if !added {
		return
	}

	p.view.AddChartThumb(st)
	p.view.GoRefreshStock(st)
}
