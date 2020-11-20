// Package ui contains code for the view in the MVC pattern.
package ui

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"runtime"
	"unicode"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"golang.org/x/image/font/gofont/goregular"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view"
	"github.com/btmura/ponzi2/internal/app/view/chart"
	"github.com/btmura/ponzi2/internal/app/view/rect"
	"github.com/btmura/ponzi2/internal/app/view/text"
	"github.com/btmura/ponzi2/internal/logger"
	"github.com/btmura/ponzi2/internal/matrix"
)

// Embed resources into the application. Get esc from github.com/mjibson/esc.
//go:generate esc -o bindata.go -pkg ui -include ".*(ply|png)" -modtime 1337 -private data

// Add a Windows icon by generating a SYSO file in the root that will be picked
// up by the Go build tools. Get rsrc from github.com/akavel/rsrc.
//go:generate rsrc -ico data/icon.ico -arch amd64 -o ../../../../cmd/ponzi2/ponzi2_windows.syso

// Application name for the window title.
const appName = "ponzi2"

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

// Constants used by Run for the "game loop".
const (
	updateSec  = 1.0 / view.FPS
	minUpdates = 1
	maxUpdates = 1000
)

// draggingMinimumPressCount is how many input cycles the mouse button must be pressed
// till the UI considers it a drag and drop event.
const draggingMinimumPressCount = 10

// viewPadding is padding between visual elements in the UI.
const viewPadding = 10

// inputSymbolBubbleRounding is the rounding of the input symbol bubble.
const inputSymbolBubbleRounding = 10

var inputSymbolTextRenderer = gfx.NewTextRenderer(goregular.TTF, 48)

func init() {
	// This is needed to arrange that main() runs on main thread for GLFW.
	// See documentation for functions that are only allowed to be called
	// from the main thread.
	runtime.LockOSThread()
}

// The UI renders the UI to view and edit the model's stocks that it observes.
type UI struct {
	// symbolToChartMap maps symbol to chart. Only one entry right now.
	symbolToChartMap map[string]*chart.Chart

	// titleBar renders the window titleBar.
	titleBar *titleBar

	// charts renders the charts in the main area.
	charts []*uiChart

	// sidebar is the sidebar of chart thumbnails on the side.
	sidebar *sidebar

	// instructionsTextBox renders instructional text when no chart is shown.
	instructionsTextBox *text.Box

	// inputSymbolTextBox stores and renders the symbol being entered by the user.
	inputSymbolTextBox *text.Box

	// inputSymbolSubmittedCallback is called when a new symbol is entered.
	inputSymbolSubmittedCallback func(symbol string)

	// sidebarChangeCallback is called when the sidebar changes.
	sidebarChangeCallback func(symbols []string)

	// chartZoomChangeCallback is called when the chart is zoomed in or out.
	chartZoomChangeCallback func(zoomChange chart.ZoomChange)

	// chartPriceStyleButtonClickCallback is called when the bar or candlestick buttons are clicked.
	chartPriceStyleButtonClickCallback func(priceStyle chart.PriceStyle)

	// chartRefreshButtonClickCallback is called when the main chart's refresh button is clicked.
	chartRefreshButtonClickCallback func(symbol string)

	// chartAddButtonClickCallback is called when the main chart's add button is clicked.
	chartAddButtonClickCallback func(symbol string)

	// thumbRemoveButtonClickCallback is called when a thumb's remove button is clicked.
	thumbRemoveButtonClickCallback func(symbol string)

	// thumbClickCallback is called when a thumb is clicked.
	thumbClickCallback func(symbol string)

	// win is the handle to the GLFW window.
	win *glfw.Window

	// winSize is the current window's size used to measure and draw the UI.
	winSize image.Point

	// mousePos is the next mouse position to report in global coordinates.
	mousePos view.MousePosition

	// mousePreviousPos is the previous mouse position reported in global coordinates.
	mousePreviousPos view.MousePosition

	// mouseLeftButtonPressed is whether the left mouse button was pressed.
	mouseLeftButtonPressed bool

	// mouseLeftButtonPressedPos is the is the position when the left mouse button was pressed.
	mouseLeftButtonPressedPos *view.MousePosition

	// mouseLeftButtonPressedCount is the number of loop iterations the left
	// mouse button has been pressed. Used to determine dragging.
	mouseLeftButtonPressedCount int

	// mouseLeftButtonReleased is whether the left mouse button was released.
	mouseLeftButtonReleased bool

	// mouseScrollDirection is the next mouse scroll event to report. Nil if no scroll has happened.
	mouseScrollDirection view.ScrollDirection

	// keyReleased is the key released. Unspecified if no key was released.
	keyReleased *view.KeyReleaseEvent
}

type uiChart struct {
	*chart.Chart
	*view.Fader
}

func newUIChart(c *chart.Chart) *uiChart {
	return &uiChart{
		Chart: c,
		Fader: view.NewStartedFader(1 * view.FPS),
	}
}

func (u *uiChart) Update() (dirty bool) {
	if u.Chart.Update() {
		dirty = true
	}
	if u.Fader.Update() {
		dirty = true
	}
	return dirty
}

func (u *uiChart) Render(fudge float32) {
	u.Fader.Render(fudge, func() {
		u.Chart.Render(fudge)
	})
}

// New creates a new UI.
func New() *UI {
	return &UI{
		symbolToChartMap:    map[string]*chart.Chart{},
		sidebar:             newSidebar(),
		instructionsTextBox: text.NewBox(gfx.NewTextRenderer(goregular.TTF, 24), "Type in symbol and press ENTER..."),
		inputSymbolTextBox: text.NewBox(inputSymbolTextRenderer, "",
			text.Bubble(rect.NewBubble(inputSymbolBubbleRounding)),
			text.Padding(viewPadding)),
	}
}

// Init initializes the View and returns a cleanup function.
func (u *UI) Init(_ context.Context) (cleanup func(), err error) {
	if err := glfw.Init(); err != nil {
		return nil, err
	}

	// Set the following hints for Linux compatibility.
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 5)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	win, err := glfw.CreateWindow(1024, 768, appName, nil, nil)
	if err != nil {
		return nil, err
	}
	u.titleBar = newTitleBar(win)
	u.win = win

	win.MakeContextCurrent()

	// Set the window icon. Windows additionally uses the embedded SYSO file.
	icon, err := png.Decode(bytes.NewReader(_escFSMustByte(false, "/data/icon.png")))
	if err != nil {
		return nil, err
	}
	win.SetIcon([]image.Image{icon})

	if err := gl.Init(); err != nil {
		return nil, err
	}
	logger.Infof("OpenGL version: %s", gl.GoStr(gl.GetString(gl.VERSION)))

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.ClearColor(0, 0, 0, 0)

	if err := gfx.InitProgram(); err != nil {
		return nil, err
	}

	gfx.SetAlpha(1.0)

	// Call the size callback to set the initial viewport.
	w, h := win.GetSize()
	u.handleSizeEvent(w, h)
	win.SetSizeCallback(func(win *glfw.Window, width, height int) {
		u.handleSizeEvent(width, height)
	})

	win.SetCharCallback(func(win *glfw.Window, char rune) {
		u.handleCharEvent(char)
	})

	win.SetKeyCallback(func(win *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		u.handleKeyEvent(key, action)
	})

	win.SetCursorPosCallback(func(win *glfw.Window, xpos, ypos float64) {
		u.handleCursorPosEvent(xpos, ypos)
	})

	win.SetMouseButtonCallback(func(win *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
		u.handleMouseButtonEvent(button, action)
	})

	win.SetScrollCallback(func(win *glfw.Window, xoff, yoff float64) {
		u.handleScrollEvent(yoff)
	})

	u.sidebar.SetChangeCallback(func(sidebar *Sidebar) {
		u.handleSidebarChangeEvent(sidebar)
	})

	u.sidebar.SetThumbRemoveButtonClickCallback(func(symbol string) {
		if u.thumbRemoveButtonClickCallback != nil {
			u.thumbRemoveButtonClickCallback(symbol)
		}
	})

	u.sidebar.SetThumbClickCallback(func(symbol string) {
		if u.thumbClickCallback != nil {
			u.thumbClickCallback(symbol)
		}
	})

	return func() { glfw.Terminate() }, nil
}

func (u *UI) handleSizeEvent(width, height int) {
	s := image.Pt(width, height)
	if s == u.winSize {
		return
	}

	defer u.WakeLoop()

	gl.Viewport(0, 0, int32(s.X), int32(s.Y))

	// Calculate the new ortho projection view matrix.
	fw, fh := float32(s.X), float32(s.Y)
	gfx.SetProjectionViewMatrix(matrix.Ortho(fw, fh, fw /* use width as depth */))

	u.winSize = s
}

func (u *UI) handleCharEvent(char rune) {
	u.keyReleased = &view.KeyReleaseEvent{Char: char}
	u.WakeLoop()
}

func (u *UI) handleKeyEvent(key glfw.Key, action glfw.Action) {
	if action != glfw.Release {
		return
	}

	switch key {
	case glfw.KeyEscape:
		u.keyReleased = &view.KeyReleaseEvent{Key: view.KeyEscape}
		u.WakeLoop()

	case glfw.KeyBackspace:
		u.keyReleased = &view.KeyReleaseEvent{Key: view.KeyBackspace}
		u.WakeLoop()

	case glfw.KeyEnter:
		u.keyReleased = &view.KeyReleaseEvent{Key: view.KeyEnter}
		u.WakeLoop()
	}
}

func (u *UI) handleCursorPosEvent(x, y float64) {
	// Flip Y-axis since the OpenGL coordinate system makes lower left the origin.
	pos := view.MousePosition{Point: image.Pt(int(x), u.winSize.Y-int(y))}
	if pos == u.mousePos {
		return
	}

	u.mousePreviousPos = u.mousePos
	u.mousePos = pos
	u.WakeLoop()
}

func (u *UI) handleMouseButtonEvent(button glfw.MouseButton, action glfw.Action) {
	if button != glfw.MouseButtonLeft {
		return
	}

	defer u.WakeLoop()

	switch action {
	case glfw.Press:
		u.mouseLeftButtonPressedCount = 1
	case glfw.Release:
		u.mouseLeftButtonPressedCount = 0
	}

	u.mouseLeftButtonPressed = action == glfw.Press
	u.mouseLeftButtonReleased = action == glfw.Release
}

func (u *UI) handleScrollEvent(yoff float64) {
	if yoff != -1 && yoff != +1 {
		return
	}

	defer u.WakeLoop()

	switch {
	case yoff < 0: // Scroll wheel down
		u.mouseScrollDirection = view.ScrollDown
	case yoff > 0: // Scroll wheel up.
		u.mouseScrollDirection = view.ScrollUp
	default:
		u.mouseScrollDirection = view.ScrollDirectionUnspecified
	}
}

func (u *UI) handleSidebarChangeEvent(sidebar *Sidebar) {
	if sidebar == nil {
		logger.Error("sidebar is nil")
		return
	}

	if u.sidebarChangeCallback == nil {
		logger.Error("sidebar change callback is nil")
		return
	}

	var symbols []string

	for _, slot := range sidebar.Slots {
		if slot.Symbol != "" {
			symbols = append(symbols, slot.Symbol)
		}
	}

	if u.sidebarChangeCallback != nil {
		u.sidebarChangeCallback(symbols)
	}
}

func (u *UI) handleChartZoomChangeEvent(zoomChange chart.ZoomChange) {
	if zoomChange == chart.ZoomChangeUnspecified {
		logger.Error("unspecified chart zoom change")
		return
	}

	if u.chartZoomChangeCallback != nil {
		u.chartZoomChangeCallback(zoomChange)
	}
}

// RunLoop runs the "game loop".
func (u *UI) RunLoop(ctx context.Context, runLoopHook func(context.Context) error) error {
start:
	var lag float64
	prevTime := glfw.GetTime()
	for !u.win.ShouldClose() {
		dirty := false

		currTime := glfw.GetTime() /* seconds */
		elapsed := currTime - prevTime
		prevTime = currTime
		lag += elapsed

		if err := runLoopHook(ctx); err != nil {
			return err
		}

		input := u.prepareInput()
		if u.processInput(input) {
			dirty = true
		}

		i := 0
		for ; i < minUpdates || i < maxUpdates && lag >= updateSec; i++ {
			if u.update() {
				dirty = true
			}
			lag -= updateSec
		}
		if lag < 0 {
			lag = 0
		}

		fudge := float32(lag / updateSec)
		if fudge < 0 || fudge > 1 {
			fudge = 0
		}

		u.render(fudge)
		u.win.SwapBuffers()

		glfw.PollEvents()
		if !dirty {
			glfw.WaitEvents()
			goto start
		}
	}

	return nil
}

// WakeLoop wakes up the loop if it is asleep.
func (u *UI) WakeLoop() {
	glfw.PostEmptyEvent()
}

// prepareInput creates a view.Input for the current game loop iteration
// and cleans up for the next game loop iteration.
func (u *UI) prepareInput() *view.Input {
	mousePos := u.mousePos

	input := &view.Input{
		MousePos: &mousePos,
	}

	dragging := u.mouseLeftButtonPressedCount > draggingMinimumPressCount

	switch {
	case u.mouseLeftButtonPressed:
		u.mouseLeftButtonPressedPos = &mousePos

	case dragging && !u.mouseLeftButtonReleased:
		input.MouseLeftButtonDragging = &view.MouseDraggingEvent{
			CurrentPos:  mousePos,
			PreviousPos: u.mousePreviousPos,
			PressedPos:  *u.mouseLeftButtonPressedPos,
		}

	case u.mouseLeftButtonReleased:
		if dragging {
			input.MouseLeftButtonDragging = &view.MouseDraggingEvent{
				CurrentPos:  mousePos,
				PressedPos:  *u.mouseLeftButtonPressedPos,
				ReleasedPos: &mousePos,
			}
		} else {
			input.MouseLeftButtonClicked = &view.MouseClickEvent{
				PressedPos:  *u.mouseLeftButtonPressedPos,
				ReleasedPos: mousePos,
			}
		}
	}

	if u.mouseScrollDirection != view.ScrollDirectionUnspecified {
		input.MouseScrolled = &view.MouseScrollEvent{
			CurrentPos: mousePos,
			Direction:  u.mouseScrollDirection,
		}
	}

	if u.keyReleased != nil {
		input.KeyReleased = &view.KeyReleaseEvent{
			Char: u.keyReleased.Char,
			Key:  u.keyReleased.Key,
		}
	}

	// Reset any flags for the next input.
	if u.mouseLeftButtonPressedCount > 0 {
		u.mouseLeftButtonPressedCount++
	}
	u.mouseLeftButtonPressed = false
	u.mouseLeftButtonReleased = false
	u.mouseScrollDirection = view.ScrollDirectionUnspecified
	u.keyReleased = nil

	return input
}

func (u *UI) processInput(input *view.Input) (dirty bool) {
	m := u.metrics()

	u.instructionsTextBox.SetBounds(m.chartBounds)
	u.inputSymbolTextBox.SetBounds(m.winBounds)

	u.updateInputSymbolTextBox(input)

	u.sidebar.SetBounds(m.sidebarBounds)
	u.sidebar.ProcessInput(input)

	for i := 0; i < len(u.charts); i++ {
		c := u.charts[i]
		c.SetBounds(m.chartBounds)
		c.ProcessInput(input)
	}

	for _, cb := range input.FiredCallbacks() {
		cb()
	}

	return len(input.FiredCallbacks()) != 0
}

func (u *UI) updateInputSymbolTextBox(input *view.Input) {
	b := u.inputSymbolTextBox

	if char := input.KeyReleased.GetChar(); char != 0 {
		char = unicode.ToUpper(char)
		if _, ok := acceptedChars[char]; !ok {
			return
		}

		b.SetText(b.Text() + string(char))
		input.ClearKeyboardInput()
	}

	switch input.KeyReleased.GetKey() {
	case view.KeyEscape:
		b.SetText("")
		input.ClearKeyboardInput()

	case view.KeyBackspace:
		if l := len(b.Text()); l > 0 {
			b.SetText(b.Text()[:l-1])
			input.ClearKeyboardInput()
		}

	case view.KeyEnter:
		txt := b.Text()
		input.AddFiredCallback(func() {
			if u.inputSymbolSubmittedCallback != nil {
				u.inputSymbolSubmittedCallback(txt)
			}
		})
		b.SetText("")
		input.ClearKeyboardInput()
	}
}

func (u *UI) update() (dirty bool) {
	for i := 0; i < len(u.charts); i++ {
		c := u.charts[i]
		if c.Update() {
			dirty = true
		}
		if c.DoneFadingOut() {
			u.charts = append(u.charts[:i], u.charts[i+1:]...)
			c.Close()
			i--
		}
	}

	if u.sidebar.Update() {
		dirty = true
	}

	if u.instructionsTextBox.Update() {
		dirty = true
	}

	if u.inputSymbolTextBox.Update() {
		dirty = true
	}

	return dirty
}

func (u *UI) render(fudge float32) {
	u.titleBar.Render(fudge)

	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	// Render the main chart.
	for _, c := range u.charts {
		c.Render(fudge)
	}

	// Render instructions if there are no charts to show.
	if len(u.charts) == 0 {
		u.instructionsTextBox.Render(fudge)
	}

	// Render the input symbol over the chart.
	u.inputSymbolTextBox.Render(fudge)

	// Render the sidebar thumbnails.
	u.sidebar.Render(fudge)
}

// viewMetrics has dynamic metrics used to render the view.
type viewMetrics struct {
	// winBounds is main window.
	winBounds image.Rectangle

	// chartBounds is where to draw the main chart.
	chartBounds image.Rectangle

	// sidebarBounds is where to draw the sidebar that can move up or down.
	sidebarBounds image.Rectangle
}

func (u *UI) metrics() viewMetrics {
	// +---+---------+---+
	// |   | padding |   |
	// |   +---------+   |
	// |   |         |   |
	// |   |         |   |
	// | p | chart   | p |
	// |   |         |   |
	// |   |         |   |
	// |   +---------+   |
	// |   | padding |   |
	// +---+---------+---+

	m := viewMetrics{
		winBounds: image.Rect(0, 0, u.winSize.X, u.winSize.Y),
	}

	sidebarSize := u.sidebar.ContentSize()

	if sidebarSize.Y == 0 {
		m.chartBounds = m.winBounds.Inset(viewPadding)
		return m
	}

	m.chartBounds = image.Rect(viewPadding+sidebarSize.X, 0, u.winSize.X, u.winSize.Y)
	m.chartBounds = m.chartBounds.Inset(viewPadding)

	// +---+---------+---+---------+---+
	// |   |         |   | padding |   |
	// |   |         |   +---------+   |
	// |   |         |   |         |   |
	// |   |         |   |         |   |
	// | p | sidebar | p | chart   | p |
	// |   |         |   |         |   |
	// |   |         |   |         |   |
	// |   |         |   +---------+   |
	// |   |         |   | padding |   |
	// +---+---------+---+---------+---+

	m.sidebarBounds = image.Rect(
		viewPadding, 0,
		viewPadding+sidebarSize.X, u.winSize.Y,
	)

	return m
}

// SetInputSymbolSubmittedCallback sets the callback for when a new symbol is entered.
func (u *UI) SetInputSymbolSubmittedCallback(cb func(symbol string)) {
	u.inputSymbolSubmittedCallback = cb
}

// SetSidebarChangeCallback sets the callback for when the sidebar changes.
func (u *UI) SetSidebarChangeCallback(cb func(symbols []string)) {
	u.sidebarChangeCallback = cb
}

// SetChartZoomChangeCallback sets the callback for when the chart is zoomed in or out.
func (u *UI) SetChartZoomChangeCallback(cb func(chart.ZoomChange)) {
	u.chartZoomChangeCallback = cb
}

// SetChartPriceStyleButtonClickCallback sets the callback for when the bar or candle stick buttons are clicked.
func (u *UI) SetChartPriceStyleButtonClickCallback(cb func(newPriceStyle chart.PriceStyle)) {
	u.chartPriceStyleButtonClickCallback = cb
}

// SetChartRefreshButtonClickCallback sets the callback for when the main chart's refresh button is clicked.
func (u *UI) SetChartRefreshButtonClickCallback(cb func(symbol string)) {
	u.chartRefreshButtonClickCallback = cb
}

// SetChartAddButtonClickCallback sets the callback for when the main chart's add button is clicked.
func (u *UI) SetChartAddButtonClickCallback(cb func(symbol string)) {
	u.chartAddButtonClickCallback = cb
}

// SetThumbRemoveButtonClickCallback sets the callback for when a thumb's remove button is clicked.
func (u *UI) SetThumbRemoveButtonClickCallback(cb func(symbol string)) {
	u.thumbRemoveButtonClickCallback = cb
}

// SetThumbClickCallback sets the callback for when a thumb is clicked.
func (u *UI) SetThumbClickCallback(cb func(symbol string)) {
	u.thumbClickCallback = cb
}

// SetChart sets the main chart to the given symbol and data.
func (u *UI) SetChart(symbol string, data chart.Data, priceStyle chart.PriceStyle) bool {
	if err := model.ValidateSymbol(symbol); err != nil {
		logger.Errorf("invalid symbol: %v", err)
		return false
	}

	for symbol, c := range u.symbolToChartMap {
		delete(u.symbolToChartMap, symbol)
		c.Close()
	}

	c := chart.NewChart(priceStyle)
	u.symbolToChartMap[symbol] = c

	u.titleBar.SetData(data)
	c.SetData(data)

	c.SetBarButtonClickCallback(func() {
		if u.chartPriceStyleButtonClickCallback != nil {
			u.chartPriceStyleButtonClickCallback(chart.Bar)
		}
	})

	c.SetCandlestickButtonClickCallback(func() {
		if u.chartPriceStyleButtonClickCallback != nil {
			u.chartPriceStyleButtonClickCallback(chart.Candlestick)
		}
	})

	c.SetRefreshButtonClickCallback(func() {
		if u.chartRefreshButtonClickCallback != nil {
			u.chartRefreshButtonClickCallback(symbol)
		}
	})

	c.SetAddButtonClickCallback(func() {
		if u.chartAddButtonClickCallback != nil {
			u.chartAddButtonClickCallback(symbol)
		}
	})

	c.SetZoomChangeCallback(func(zoomChange chart.ZoomChange) {
		u.handleChartZoomChangeEvent(zoomChange)
	})

	defer u.WakeLoop()
	for _, c := range u.charts {
		c.FadeOut()
	}
	u.charts = append([]*uiChart{newUIChart(c)}, u.charts...)

	return true
}

func (u *UI) SetChartPriceStyle(newPriceStyle chart.PriceStyle) {
	if newPriceStyle == chart.PriceStyleUnspecified {
		logger.Error("unspecified price style")
		return
	}

	for _, c := range u.symbolToChartMap {
		c.SetPriceStyle(newPriceStyle)
	}

	u.sidebar.SetPriceStyle(newPriceStyle)
	u.WakeLoop()
}

// AddChartThumb adds a thumbnail with the given symbol and data.
func (u *UI) AddChartThumb(symbol string, data chart.Data) (changed bool) {
	if err := model.ValidateSymbol(symbol); err != nil {
		logger.Errorf("invalid symbol: %v", err)
		return false
	}

	if u.sidebar.AddChartThumb(symbol) {
		defer u.WakeLoop()
		u.sidebar.SetData(symbol, data)
		return true
	}

	return false
}

// RemoveChartThumb removes the thumbnail with given symbol.
func (u *UI) RemoveChartThumb(symbol string) (changed bool) {
	if err := model.ValidateSymbol(symbol); err != nil {
		logger.Errorf("invalid symbol: %v", err)
		return false
	}

	if u.sidebar.RemoveChartThumb(symbol) {
		defer u.WakeLoop()
		return true
	}

	return false
}

// SetLoading sets the charts and slots matching the symbol and interval to loading.
func (u *UI) SetLoading(symbol string) (changed bool) {
	if err := model.ValidateSymbol(symbol); err != nil {
		logger.Errorf("invalid symbol: %v", err)
		return false
	}

	if c, ok := u.symbolToChartMap[symbol]; ok {
		changed = true
		c.SetLoading(true)
		c.SetErrorMessage("")
	}

	if u.sidebar.SetLoading(symbol) {
		changed = true
	}

	if changed {
		defer u.WakeLoop()
		return true
	}

	return false
}

// SetData loads the data to charts and thumbnails matching the symbol and interval.
func (u *UI) SetData(symbol string, data chart.Data) (changed bool) {
	if err := model.ValidateSymbol(symbol); err != nil {
		logger.Errorf("invalid symbol: %v", err)
		return false
	}

	if c, ok := u.symbolToChartMap[symbol]; ok {
		changed = true
		c.SetLoading(false)
		u.titleBar.SetData(data)
		c.SetData(data)
	}

	if u.sidebar.SetData(symbol, data) {
		changed = true
	}

	if changed {
		defer u.WakeLoop()
		return true
	}

	return false
}

// SetErrorMessage sets or resets an error message on charts and thumbnails that match the symbol.
// An empty error message clears any previously set error messages.
func (u *UI) SetErrorMessage(symbol string, errorMessage string) (changed bool) {
	if err := model.ValidateSymbol(symbol); err != nil {
		logger.Errorf("invalid symbol: %v", err)
		return
	}

	if c, ok := u.symbolToChartMap[symbol]; ok {
		changed = true
		c.SetLoading(false)
		c.SetErrorMessage(errorMessage)
	}

	if u.sidebar.SetErrorMessage(symbol, errorMessage) {
		changed = true
	}

	if changed {
		defer u.WakeLoop()
		return true
	}

	return false
}
