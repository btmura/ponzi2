package ponzi

import (
	"image"
	"time"
	"unicode"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang/glog"

	"github.com/btmura/ponzi2/internal/gfx"
	math2 "github.com/btmura/ponzi2/internal/math"
	"github.com/btmura/ponzi2/internal/stock"
	time2 "github.com/btmura/ponzi2/internal/time"
)

// acceptedChars are the chars the user can enter for a symbol.
var acceptedChars = map[rune]bool{
	'A': true, 'B': true, 'C': true,
	'D': true, 'E': true, 'F': true,
	'G': true, 'H': true, 'I': true,
	'J': true, 'K': true, 'L': true,
	'M': true, 'N': true, 'O': true,
	'P': true, 'Q': true, 'R': true,
	'S': true, 'T': true, 'U': true,
	'V': true, 'W': true, 'X': true,
	'Y': true, 'Z': true,
}

type Controller struct {
	model *Model

	view *View

	// pendingStockUpdates is a channel with stock updates ready to apply to the model.
	pendingStockUpdates chan controllerStockUpdate

	// mousePos is the current global mouse position.
	mousePos image.Point

	// mouseLeftButtonClicked is whether the left mouse button was clicked.
	mouseLeftButtonClicked bool

	// winSize is the current window's size used to measure and draw the UI.
	winSize image.Point
}

// controllerStockUpdate bundles a stock and new data for that stock.
type controllerStockUpdate struct {
	// stock is the stock to update.
	stock *ModelStock

	// tradingHistory is the new data to update the stock with.
	tradingHistory *stock.TradingHistory
}

// NewController creates a new Controller.
func NewController() *Controller {
	return &Controller{
		model:               NewModel(),
		view:                NewView(),
		pendingStockUpdates: make(chan controllerStockUpdate),
	}
}

func (c *Controller) Run() {
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
		c.setChart(s)
	}

	for _, cs := range cfg.Stocks {
		c.addChartThumb(cs.Symbol)
	}

	// Call the size callback to set the initial viewport.
	w, h := win.GetSize()
	c.setSize(w, h)
	win.SetSizeCallback(func(_ *glfw.Window, width, height int) {
		c.setSize(width, height)
	})

	win.SetCharCallback(func(_ *glfw.Window, char rune) {
		c.setChar(char)
	})

	win.SetKeyCallback(func(_ *glfw.Window, key glfw.Key, _ int, action glfw.Action, _ glfw.ModifierKey) {
		c.setKey(key, action)
	})

	win.SetCursorPosCallback(func(_ *glfw.Window, x, y float64) {
		c.setCursorPos(x, y)
	})

	win.SetMouseButtonCallback(func(_ *glfw.Window, button glfw.MouseButton, action glfw.Action, _ glfw.ModifierKey) {
		c.setMouseButton(button, action)
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
			// add update here
			lag -= secPerUpdate
		}

		fudge := float32(lag / secPerUpdate)
		c.render(fudge)

		win.SwapBuffers()
		glfw.PollEvents()
	}
}

func (c *Controller) render(fudge float32) {
	// Process any stock updates.
loop:
	for {
		select {
		case u := <-c.pendingStockUpdates:
			u.stock.Update(u.tradingHistory)

		default:
			break loop
		}
	}

	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	vc := ViewContext{
		Bounds:                 image.Rectangle{image.ZP, c.winSize},
		MousePos:               c.mousePos,
		MouseLeftButtonClicked: c.mouseLeftButtonClicked,
		Fudge:  fudge,
		values: &viewContextValues{},
	}
	c.view.Render(vc)

	// Call any callbacks scheduled by views.
	for _, cb := range vc.values.scheduledCallbacks {
		cb()
	}

	// Reset any flags for the next viewContext.
	c.mouseLeftButtonClicked = false
}

func (c *Controller) setChart(symbol string) {
	if symbol == "" {
		return
	}

	st := c.model.SetCurrentStock(symbol)
	ch := c.view.SetChart(st)
	ch.SetAddButtonClickCallback(func() {
		st, added := c.model.AddSavedStock(symbol)
		if !added {
			return
		}

		c.view.AddChartThumb(st)
		c.goSaveConfig()
	})

	c.goRefreshStock(st)
}

func (c *Controller) addChartThumb(symbol string) {
	if symbol == "" {
		return
	}

	st, added := c.model.AddSavedStock(symbol)
	if !added {
		return
	}

	th := c.view.AddChartThumb(st)
	th.SetRemoveButtonClickCallback(func() {
		removed := c.model.RemoveSavedStock(symbol)
		if !removed {
			return
		}

		c.view.RemoveChartThumb(th)
		th.SetRemoveButtonClickCallback(nil)
		c.goSaveConfig()
	})
	th.SetThumbClickCallback(func() {
		c.setChart(symbol)
	})

	c.goRefreshStock(st)
}

func (c *Controller) goRefreshStock(st *ModelStock) {
	go func() {
		end := time2.Midnight(time.Now().In(time2.NewYorkLoc))
		start := end.Add(-6 * 30 * 24 * time.Hour)
		hist, err := stock.GetTradingHistory(&stock.GetTradingHistoryRequest{
			Symbol:    st.Symbol,
			StartDate: start,
			EndDate:   end,
		})
		if err != nil {
			glog.Warningf("goRefreshStock: failed to get trading history for %s: %v", st.Symbol, err)
			return
		}

		c.pendingStockUpdates <- controllerStockUpdate{
			stock:          st,
			tradingHistory: hist,
		}
	}()
}

func (c *Controller) goSaveConfig() {
	// Make the config on the main thread to save the exact config at the time.
	cfg := &Config{}
	if st := c.model.CurrentStock; st != nil {
		cfg.CurrentStock = ConfigStock{st.Symbol}
	}
	for _, st := range c.model.SavedStocks {
		cfg.Stocks = append(cfg.Stocks, ConfigStock{st.Symbol})
	}

	// Handle saving to disk in a separate go routine.
	go func() {
		if err := SaveConfig(cfg); err != nil {
			glog.Warningf("goSaveConfig: failed to save config: %v", err)
		}
	}()
}

func (c *Controller) setSize(width, height int) {
	s := image.Pt(width, height)
	if c.winSize == s {
		return
	}

	gl.Viewport(0, 0, int32(s.X), int32(s.Y))

	// Calculate the new ortho projection view matrix.
	fw, fh := float32(s.X), float32(s.Y)
	gfx.SetProjectionViewMatrix(math2.OrthoMatrix(fw, fh, fw /* use width as depth */))

	c.winSize = s
}

func (c *Controller) setChar(char rune) {
	char = unicode.ToUpper(char)
	if _, ok := acceptedChars[char]; ok {
		c.view.PushInputSymbolChar(char)
	}
}

func (c *Controller) setKey(key glfw.Key, action glfw.Action) {
	if action != glfw.Release {
		return
	}

	switch key {
	case glfw.KeyEscape:
		c.view.ClearInputSymbol()

	case glfw.KeyBackspace:
		c.view.PopInputSymbolChar()

	case glfw.KeyEnter:
		c.setChart(c.view.InputSymbol())
		c.view.ClearInputSymbol()
	}
}

func (c *Controller) setCursorPos(x, y float64) {
	// Flip Y-axis since the OpenGL coordinate system makes lower left the origin.
	c.mousePos = image.Pt(int(x), c.winSize.Y-int(y))
}

func (c *Controller) setMouseButton(button glfw.MouseButton, action glfw.Action) {
	if button != glfw.MouseButtonLeft {
		glog.Infof("handleMouseButton: ignoring mouse button(%v) and action(%v)", button, action)
		return // Only interested in left clicks right now.
	}
	c.mouseLeftButtonClicked = action == glfw.Release
}
