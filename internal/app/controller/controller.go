package controller

import (
	"context"
	"fmt"
	"image"
	"runtime"
	"unicode"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang/glog"

	"github.com/btmura/ponzi2/internal/app/config"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view"
	"github.com/btmura/ponzi2/internal/gfx"
	math2 "github.com/btmura/ponzi2/internal/math"
	"github.com/btmura/ponzi2/internal/stock/iex"
)

// Application name for the window title.
const appName = "ponzi"

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

func init() {
	// This is needed to arrange that main() runs on main thread for GLFW.
	// See documentation for functions that are only allowed to be called
	// from the main thread.
	runtime.LockOSThread()
}

// Controller runs the program in a "game loop".
type Controller struct {
	// iexClient fetches stock data.
	iexClient *iex.Client

	// model is the data that the Controller connects to the View.
	model *model.Model

	// view is the UI that the Controller updates.
	view *view.View

	// symbolToChartMap maps symbol to Chart. Only one entry right now.
	symbolToChartMap map[string]*view.Chart

	// symbolToChartThumbMap maps symbol to ChartThumbnail.
	symbolToChartThumbMap map[string]*view.ChartThumb

	// pendingStockUpdates has stock updates ready to apply to the model.
	pendingStockUpdates chan controllerStockUpdate

	// enableSavingConfigs enables saving config changes.
	enableSavingConfigs bool

	// pendingConfigSaves is a channel with configs to save.
	pendingConfigSaves chan *config.Config

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
	// symbol is the stock's symbol.
	symbol string

	// update is the new data for the stock. Nil if an error happened.
	update *model.StockUpdate

	// updateErr is the error getting the update. Nil if no error happened.
	updateErr error
}

// NewController creates a new Controller.
func NewController(iexClient *iex.Client) *Controller {
	return &Controller{
		iexClient:             iexClient,
		model:                 model.NewModel(),
		view:                  view.NewView(),
		symbolToChartMap:      map[string]*view.Chart{},
		symbolToChartThumbMap: map[string]*view.ChartThumb{},
		pendingStockUpdates:   make(chan controllerStockUpdate),
		pendingConfigSaves:    make(chan *config.Config),
		doneSavingConfigs:     make(chan bool),
	}
}

// Run initializes and runs the "game loop".
func (c *Controller) Run() {
	ctx := context.Background()

	if err := glfw.Init(); err != nil {
		glog.Fatalf("Run: failed to init glfw: %v", err)
	}
	defer glfw.Terminate()

	// Set the following hints for Linux compatibility.
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 5)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	win, err := glfw.CreateWindow(800, 600, appName, nil, nil)
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
	cfg, err := config.Load()
	if err != nil {
		glog.Fatalf("Run: failed to load config: %v", err)
	}

	if s := cfg.GetCurrentStock().GetSymbol(); s != "" {
		c.setChart(ctx, s)
	}

	for _, cs := range cfg.GetStocks() {
		if s := cs.GetSymbol(); s != "" {
			c.addChartThumb(ctx, s)
		}
	}

	// Process config changes in the background until the program ends.
	go func() {
		for cfg := range c.pendingConfigSaves {
			if err := config.Save(cfg); err != nil {
				glog.Infof("Run: failed to save config: %v", err)
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
		c.setKey(ctx, key, action)
	})

	win.SetCursorPosCallback(func(_ *glfw.Window, x, y float64) {
		c.setCursorPos(x, y)
	})

	win.SetMouseButtonCallback(func(_ *glfw.Window, button glfw.MouseButton, action glfw.Action, _ glfw.ModifierKey) {
		c.setMouseButton(button, action)
	})

	const (
		secPerUpdate = 1.0 / fps
		maxUpdates   = 10
	)

	var lag float64
	prevTime := glfw.GetTime()
	for !win.ShouldClose() {
		currTime := glfw.GetTime()
		elapsed := currTime - prevTime
		prevTime = currTime
		lag += elapsed

		// Set animating to true to keep looping when update isn't called.
		animating := true
		for i := 0; lag >= secPerUpdate && i < maxUpdates; i++ {
			if !c.update() {
				animating = false
			}
			lag -= secPerUpdate
		}

		fudge := float32(lag / secPerUpdate)
		c.render(fudge)
		win.SwapBuffers()

		if animating {
			glog.V(2).Info("poll events")
			glfw.PollEvents()
		} else {
			// Wake up to handle non-GLFW events like network updates.
			glog.V(2).Infof("wait events")
			glfw.WaitEventsTimeout(1 /* second */)
		}
	}

	// Disable config changes to start shutting down save processor.
	c.enableSavingConfigs = false
	close(c.pendingConfigSaves)
	<-c.doneSavingConfigs
}

func (c *Controller) update() (animating bool) {
	// Process any stock updates.
loop:
	for {
		select {
		case u := <-c.pendingStockUpdates:
			switch {
			case u.update != nil:
				st, updated := c.model.UpdateStock(u.update)
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

			case u.updateErr != nil:
				if ch, ok := c.symbolToChartMap[u.symbol]; ok {
					ch.SetLoading(false)
					ch.SetError(true)
				}
				if th, ok := c.symbolToChartThumbMap[u.symbol]; ok {
					th.SetLoading(false)
					th.SetError(true)
				}
			}
		default:
			break loop
		}
	}

	c.refreshWindowTitle()
	return c.view.Update()
}

func (c *Controller) render(fudge float32) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	vc := view.ViewContext{
		Bounds:                 image.Rectangle{image.ZP, c.winSize},
		MousePos:               c.mousePos,
		MouseLeftButtonClicked: c.mouseLeftButtonClicked,
		Fudge:              fudge,
		ScheduledCallbacks: new([]func()),
	}
	c.view.Render(vc)

	// Call any callbacks scheduled by views.
	for _, cb := range *vc.ScheduledCallbacks {
		cb()
	}

	// Reset any flags for the next viewContext.
	c.mouseLeftButtonClicked = false
}

func (c *Controller) setChart(ctx context.Context, symbol string) {
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

	ch := view.NewChart()
	c.symbolToChartMap[symbol] = ch

	ch.SetStock(st)
	ch.SetRefreshButtonClickCallback(func() {
		c.refreshStock(ctx, symbol)
	})
	ch.SetAddButtonClickCallback(func() {
		c.addChartThumb(ctx, symbol)
	})

	c.view.SetChart(ch)
	c.refreshStock(ctx, symbol)
	c.saveConfig()
}

func (c *Controller) addChartThumb(ctx context.Context, symbol string) {
	if symbol == "" {
		return
	}

	st, added := c.model.AddSavedStock(symbol)
	if !added {
		return
	}

	th := view.NewChartThumb()
	c.symbolToChartThumbMap[symbol] = th

	th.SetStock(st)
	th.SetRemoveButtonClickCallback(func() {
		c.removeChartThumb(symbol)
	})
	th.SetThumbClickCallback(func() {
		c.setChart(ctx, symbol)
	})

	c.view.AddChartThumb(th)
	c.refreshStock(ctx, symbol)
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

func (c *Controller) refreshStock(ctx context.Context, symbol string) {
	if ch, ok := c.symbolToChartMap[symbol]; ok {
		ch.SetLoading(true)
		ch.SetError(false)
	}
	if th, ok := c.symbolToChartThumbMap[symbol]; ok {
		th.SetLoading(true)
		th.SetError(false)
	}
	go func() {
		req := &iex.GetTradingSessionSeriesRequest{Symbol: symbol}
		sr, err := c.iexClient.GetTradingSessionSeries(ctx, req)
		if err != nil {
			c.pendingStockUpdates <- controllerStockUpdate{
				symbol:    symbol,
				updateErr: err,
			}
			return
		}
		c.pendingStockUpdates <- controllerStockUpdate{
			symbol: symbol,
			update: modelStockUpdate(sr),
		}
	}()
}

func (c *Controller) saveConfig() {
	if !c.enableSavingConfigs {
		glog.Infof("saveConfig: ignoring save request, saving disabled")
		return
	}

	// Make the config on the main thread to save the exact config at the time.
	cfg := &config.Config{}
	if st := c.model.CurrentStock; st != nil {
		cfg.CurrentStock = &config.Stock{Symbol: st.Symbol}
	}
	for _, st := range c.model.SavedStocks {
		cfg.Stocks = append(cfg.Stocks, &config.Stock{Symbol: st.Symbol})
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

func (c *Controller) setKey(ctx context.Context, key glfw.Key, action glfw.Action) {
	if action != glfw.Release {
		return
	}

	switch key {
	case glfw.KeyEscape:
		c.view.ClearInputSymbol()

	case glfw.KeyBackspace:
		c.view.PopInputSymbolChar()

	case glfw.KeyEnter:
		c.setChart(ctx, c.view.InputSymbol())
		c.view.ClearInputSymbol()
	}
}

func (c *Controller) setCursorPos(x, y float64) {
	// Flip Y-axis since the OpenGL coordinate system makes lower left the origin.
	c.mousePos = image.Pt(int(x), c.winSize.Y-int(y))
}

func (c *Controller) setMouseButton(button glfw.MouseButton, action glfw.Action) {
	if button != glfw.MouseButtonLeft {
		glog.Infof("setMouseButton: ignoring mouse button(%v) and action(%v)", button, action)
		return // Only interested in left clicks right now.
	}
	c.mouseLeftButtonClicked = action == glfw.Release
}

func (c *Controller) refreshWindowTitle() {
	glfw.GetCurrentContext().SetTitle(c.windowTitle())
}

func (c *Controller) windowTitle() string {
	st := c.model.CurrentStock
	if st == nil {
		return appName
	}

	if st.Price() == 0 {
		return fmt.Sprintf("%s - %s", st.Symbol, appName)
	}

	return fmt.Sprintf("%s %.2f %+5.2f %+5.2f%% %s (Updated: %s) - %s",
		st.Symbol,
		st.Price(),
		st.Change(),
		st.PercentChange(),
		st.Date().Format("1/2/06"),
		st.LastUpdateTime.Format("1/2/06 15:04"),
		appName)
}
