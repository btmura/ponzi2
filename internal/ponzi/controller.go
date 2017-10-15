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

// Frames per second.
const fps = 60.0

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

// Controller runs the program in a "game loop".
type Controller struct {
	// model is the data that the Controller connects to the View.
	model *Model

	// view is the UI that the Controller updates.
	view *View

	// symbolToChartMap maps symbol to Chart. Only one entry right now.
	symbolToChartMap map[string]*Chart

	// symbolToChartThumbMap maps symbol to ChartThumbnail.
	symbolToChartThumbMap map[string]*ChartThumb

	// pendingStockUpdates has stock updates ready to apply to the model.
	pendingStockUpdates chan controllerStockUpdate

	// enableSavingConfigs enables saving config changes.
	enableSavingConfigs bool

	// pendingConfigSaves is a channel with configs to save.
	pendingConfigSaves chan *Config

	// doneSavingConfigs indicates saving is done and the program may quit.
	doneSavingConfigs chan bool

	// mousePos is the current global mouse position.
	mousePos image.Point

	// mouseLeftButtonClicked is whether the left mouse button was clicked.
	mouseLeftButtonClicked bool

	// winSize is the current window's size used to measure and draw the UI.
	winSize image.Point
}

// controllerStockUpdate bundles a stock and new data for that stock.
type controllerStockUpdate struct {
	// symbol refers to the stock to update.
	symbol string

	// tradingHistory is the new data to update the stock with.
	tradingHistory *stock.TradingHistory
}

// NewController creates a new Controller.
func NewController() *Controller {
	return &Controller{
		model:                 NewModel(),
		view:                  NewView(),
		symbolToChartMap:      map[string]*Chart{},
		symbolToChartThumbMap: map[string]*ChartThumb{},
		pendingStockUpdates:   make(chan controllerStockUpdate),
		pendingConfigSaves:    make(chan *Config),
		doneSavingConfigs:     make(chan bool),
	}
}

// Run initializes and runs the "game loop".
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
		glog.Fatalf("Run: failed to init OpenGL: %v", err)
	}
	glog.Infof("OpenGL version: %s", gl.GoStr(gl.GetString(gl.VERSION)))

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.ClearColor(0, 0, 0, 0)

	if err := gfx.InitProgram(); err != nil {
		glog.Fatalf("Run: failed to init gfx: %v", err)
	}

	// Load the config and setup the initial UI.
	cfg, err := LoadConfig()
	if err != nil {
		glog.Fatalf("Run: failed to load config: %v", err)
	}

	if s := cfg.GetCurrentStock().GetSymbol(); s != "" {
		c.setChart(s)
	}

	for _, cs := range cfg.GetStocks() {
		if s := cs.GetSymbol(); s != "" {
			c.addChartThumb(s)
		}
	}

	// Process config changes in the background until the program ends.
	go func() {
		for cfg := range c.pendingConfigSaves {
			if err := SaveConfig(cfg); err != nil {
				glog.Warningf("Run: failed to save config: %v", err)
			}
		}
		c.doneSavingConfigs <- true
	}()

	// Enable saving configs after UI is setup and change processor started.
	c.enableSavingConfigs = true

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

	const secPerUpdate = 1.0 / fps

	var lag float64
	prevTime := glfw.GetTime()
	for !win.ShouldClose() {
		currTime := glfw.GetTime()
		elapsed := currTime - prevTime
		prevTime = currTime
		lag += elapsed

		for lag >= secPerUpdate {
			c.update()
			lag -= secPerUpdate
		}

		fudge := float32(lag / secPerUpdate)
		c.render(fudge)

		win.SwapBuffers()
		glfw.PollEvents()
	}

	// Disable config changes to start shutting down save processor.
	c.enableSavingConfigs = false
	close(c.pendingConfigSaves)
	<-c.doneSavingConfigs
}

func (c *Controller) update() {
	// Process any stock updates.
loop:
	for {
		select {
		case u := <-c.pendingStockUpdates:
			st, updated := c.model.UpdateStock(u.symbol, u.tradingHistory)
			if !updated {
				break loop
			}
			if ch, ok := c.symbolToChartMap[u.symbol]; ok {
				ch.SetLoading(false)
				ch.SetStock(st)
			}
			if th, ok := c.symbolToChartThumbMap[u.symbol]; ok {
				th.SetLoading(false)
				th.SetStock(st)
			}

		default:
			break loop
		}
	}

	c.view.Update()
}

func (c *Controller) render(fudge float32) {
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

	st, changed := c.model.SetCurrentStock(symbol)
	if !changed {
		return
	}

	for symbol, ch := range c.symbolToChartMap {
		delete(c.symbolToChartMap, symbol)
		ch.Close()
	}

	ch := NewChart()
	c.symbolToChartMap[symbol] = ch

	ch.SetStock(st)
	ch.SetRefreshButtonClickCallback(func() {
		c.refreshStock(symbol)
	})
	ch.SetAddButtonClickCallback(func() {
		c.addChartThumb(symbol)
	})

	c.view.SetChart(ch)
	c.refreshStock(symbol)
	c.saveConfig()
}

func (c *Controller) addChartThumb(symbol string) {
	if symbol == "" {
		return
	}

	st, added := c.model.AddSavedStock(symbol)
	if !added {
		return
	}

	th := NewChartThumb()
	c.symbolToChartThumbMap[symbol] = th

	th.SetStock(st)
	th.SetRemoveButtonClickCallback(func() {
		c.removeChartThumb(symbol)
	})
	th.SetThumbClickCallback(func() {
		c.setChart(symbol)
	})

	c.view.AddChartThumb(th)
	c.refreshStock(symbol)
	c.saveConfig()
}

func (c *Controller) removeChartThumb(symbol string) {
	if symbol == "" {
		return
	}

	if !c.model.RemoveSavedStock(symbol) {
		return
	}

	th := c.symbolToChartThumbMap[symbol]
	delete(c.symbolToChartThumbMap, symbol)
	th.Close()

	c.view.RemoveChartThumb(th)
	c.saveConfig()
}

func (c *Controller) refreshStock(symbol string) {
	if ch, ok := c.symbolToChartMap[symbol]; ok {
		ch.SetLoading(true)
	}
	if th, ok := c.symbolToChartThumbMap[symbol]; ok {
		th.SetLoading(true)
	}
	go func() {
		end := time2.Midnight(time.Now().In(time2.NewYorkLoc))
		start := end.Add(-6 * 30 * 24 * time.Hour)
		hist, err := stock.GetTradingHistory(&stock.GetTradingHistoryRequest{
			Symbol:    symbol,
			StartDate: start,
			EndDate:   end,
		})
		if err != nil {
			glog.Warningf("refreshStock: failed to get trading history for %s: %v", symbol, err)
			return
		}

		c.pendingStockUpdates <- controllerStockUpdate{
			symbol:         symbol,
			tradingHistory: hist,
		}
	}()
}

func (c *Controller) saveConfig() {
	if !c.enableSavingConfigs {
		glog.Warning("saveConfig: ignoring save request, saving disabled")
		return
	}

	// Make the config on the main thread to save the exact config at the time.
	cfg := &Config{}
	if st := c.model.CurrentStock; st != nil {
		cfg.CurrentStock = &Stock{Symbol: st.Symbol}
	}
	for _, st := range c.model.SavedStocks {
		cfg.Stocks = append(cfg.Stocks, &Stock{st.Symbol})
	}

	// Queue the config for saving.
	go func() {
		c.pendingConfigSaves <- cfg
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
