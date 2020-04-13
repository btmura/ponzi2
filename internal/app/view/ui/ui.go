// Package ui contains code for the view in the MVC pattern.
package ui

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"runtime"
	"time"
	"unicode"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"golang.org/x/image/font/gofont/goregular"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view"
	"github.com/btmura/ponzi2/internal/app/view/chart"
	"github.com/btmura/ponzi2/internal/app/view/text"
	"github.com/btmura/ponzi2/internal/errors"
	"github.com/btmura/ponzi2/internal/log"
	"github.com/btmura/ponzi2/internal/matrix"
)

// Embed resources into the application. Get esc from github.com/mjibson/esc.
//go:generate esc -o bindata.go -pkg ui -include ".*(ply|png)" -modtime 1337 -private data

// Add a Windows icon by generating a SYSO file in the root that will be picked
// up by the Go build tools. Get rsrc from github.com/akavel/rsrc.
//go:generate rsrc -ico data/icon.ico -arch amd64 -o ../../../ponzi2_windows.syso

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
	fps        = 120.0
	updateSec  = 1.0 / fps
	minUpdates = 1
	maxUpdates = 1000
)

// draggingMinimumPressCount is how many input cycles the mouse button must be pressed
// till the UI considers it a drag and drop event.
const draggingMinimumPressCount = 5

// viewPadding is padding between visual elements in the UI.
const viewPadding = 10

var inputSymbolTextRenderer = gfx.NewTextRenderer(goregular.TTF, 48)

func init() {
	// This is needed to arrange that main() runs on main thread for GLFW.
	// See documentation for functions that are only allowed to be called
	// from the main thread.
	runtime.LockOSThread()
}

// The UI renders the UI to view and edit the model's stocks that it observes.
type UI struct {
	// symbolToChartMap maps symbol to Chart. Only one entry right now.
	symbolToChartMap map[string]*chart.Chart

	// symbolToChartThumbMap maps symbol to ChartThumbnail.
	symbolToChartThumbMap map[string]*chart.Thumb

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
}

type uiChart struct {
	*chart.Chart
	*view.Fader
}

func newUIChart(ch *chart.Chart) *uiChart {
	return &uiChart{
		Chart: ch,
		Fader: view.NewFader(1 * fps),
	}
}

func (c *uiChart) Update() (dirty bool) {
	if c.Chart.Update() {
		dirty = true
	}
	if c.Fader.Update() {
		dirty = true
	}
	return dirty
}

func (c *uiChart) Render(fudge float32) {
	c.Fader.Render(c.Chart.Render, fudge)
}

// New creates a new View.
func New() *UI {
	return &UI{
		symbolToChartMap:      map[string]*chart.Chart{},
		symbolToChartThumbMap: map[string]*chart.Thumb{},
		sidebar:               newSidebar(),
		instructionsTextBox:   text.NewBox(gfx.NewTextRenderer(goregular.TTF, 24), "Type in symbol and press ENTER..."),
		inputSymbolTextBox:    text.NewBox(inputSymbolTextRenderer, ""),
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
	log.Infof("OpenGL version: %s", gl.GoStr(gl.GetString(gl.VERSION)))

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

	return func() { glfw.Terminate() }, nil
}

func (u *UI) handleSizeEvent(width, height int) {
	log.Debugf("width: %d height: %d", width, height)
	defer u.WakeLoop()

	s := image.Pt(width, height)
	if u.winSize == s {
		return
	}

	gl.Viewport(0, 0, int32(s.X), int32(s.Y))

	// Calculate the new ortho projection view matrix.
	fw, fh := float32(s.X), float32(s.Y)
	gfx.SetProjectionViewMatrix(matrix.Ortho(fw, fh, fw /* use width as depth */))

	u.winSize = s
}

func (u *UI) handleCharEvent(char rune) {
	log.Debugf("char: %c", char)
	defer u.WakeLoop()

	char = unicode.ToUpper(char)
	if _, ok := acceptedChars[char]; ok {
		u.inputSymbolTextBox.SetText(u.inputSymbolTextBox.Text() + string(char))
	}
}

func (u *UI) handleKeyEvent(key glfw.Key, action glfw.Action) {
	log.Debugf("key: %v action: %v", key, action)
	defer u.WakeLoop()

	if action != glfw.Release {
		return
	}

	switch key {
	case glfw.KeyEscape:
		u.inputSymbolTextBox.SetText("")

	case glfw.KeyBackspace:
		if l := len(u.inputSymbolTextBox.Text()); l > 0 {
			u.inputSymbolTextBox.SetText(u.inputSymbolTextBox.Text()[:l-1])
		}

	case glfw.KeyEnter:
		if u.inputSymbolSubmittedCallback != nil {
			u.inputSymbolSubmittedCallback(u.inputSymbolTextBox.Text())
		}
		u.inputSymbolTextBox.SetText("")
	}
}

func (u *UI) handleCursorPosEvent(x, y float64) {
	log.Debugf("x: %f y: %f", x, y)
	defer u.WakeLoop()

	// Flip Y-axis since the OpenGL coordinate system makes lower left the origin.
	u.mousePos = view.MousePosition{
		Point: image.Pt(int(x), u.winSize.Y-int(y)),
	}
}

func (u *UI) handleMouseButtonEvent(button glfw.MouseButton, action glfw.Action) {
	log.Debugf("button: %v action: %v", button, action)
	defer u.WakeLoop()

	if button != glfw.MouseButtonLeft {
		return
	}

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
		log.Error("sidebar is nil")
		return
	}

	if u.sidebarChangeCallback == nil {
		log.Error("sidebar change callback is nil")
		return
	}

	var symbols []string

	for _, slot := range sidebar.Slots {
		if slot.Thumbnail == nil {
			log.Error("sidebar reporting slot with nil thumbnail")
			continue
		}

		for s, th := range u.symbolToChartThumbMap {
			if th == slot.Thumbnail {
				symbols = append(symbols, s)
			}
		}
	}

	if u.sidebarChangeCallback != nil {
		u.sidebarChangeCallback(symbols)
	}
}

func (u *UI) handleChartZoomChangeEvent(zoomChange chart.ZoomChange) {
	if zoomChange == chart.ZoomChangeUnspecified {
		log.Error("chart zoom change is unspecified")
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
	dirty := false
	prevTime := glfw.GetTime()
	for !u.win.ShouldClose() {
		currTime := glfw.GetTime() /* seconds */
		elapsed := currTime - prevTime
		prevTime = currTime
		lag += elapsed

		if err := runLoopHook(ctx); err != nil {
			return err
		}

		u.processInput()

		i := 0
		for ; i < minUpdates || i < maxUpdates && lag >= updateSec; i++ {
			dirty = u.update()
			lag -= updateSec
		}
		if lag < 0 {
			lag = 0
		}

		fudge := float32(lag / updateSec)
		if fudge < 0 || fudge > 1 {
			fudge = 0
		}

		now := time.Now()
		u.render(fudge)

		u.win.SwapBuffers()
		log.Debugf("updates: %o lag(%f)/updateSec(%f)=fudge(%f) dirty: %t render: %v", i, lag, updateSec, fudge, dirty, time.Since(now).Seconds())

		glfw.PollEvents()
		if !dirty {
			log.Debugf("wait events")
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

func (u *UI) processInput() {
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
			CurrentPos: mousePos,
			PressedPos: *u.mouseLeftButtonPressedPos,
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

	// Reset any flags for the next inputContext.
	defer func() {
		if u.mouseLeftButtonPressedCount > 0 {
			u.mouseLeftButtonPressedCount++
		}
		u.mouseLeftButtonPressed = false
		u.mouseLeftButtonReleased = false
		u.mouseScrollDirection = view.ScrollDirectionUnspecified
	}()

	m := u.metrics()

	bounds := m.chartBounds
	u.instructionsTextBox.SetBounds(bounds)
	u.inputSymbolTextBox.SetBounds(bounds)

	u.sidebar.SetBounds(m.sidebarBounds)
	u.sidebar.ProcessInput(input)

	for i := 0; i < len(u.charts); i++ {
		ch := u.charts[i]
		ch.SetBounds(bounds)
		ch.ProcessInput(input)
	}

	for _, cb := range input.ScheduledCallbacks {
		cb()
	}
}

func (u *UI) update() (dirty bool) {
	for i := 0; i < len(u.charts); i++ {
		ch := u.charts[i]
		if ch.Update() {
			dirty = true
		}
		if ch.DoneFadingOut() {
			u.charts = append(u.charts[:i], u.charts[i+1:]...)
			ch.Close()
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
	for _, ch := range u.charts {
		ch.Render(fudge)
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

	sidebarSize := u.sidebar.ContentSize()

	if sidebarSize.Y == 0 {
		cb := image.Rect(0, 0, u.winSize.X, u.winSize.Y)
		cb = cb.Inset(viewPadding)
		return viewMetrics{chartBounds: cb}
	}

	cb := image.Rect(viewPadding+sidebarSize.X, 0, u.winSize.X, u.winSize.Y)
	cb = cb.Inset(viewPadding)

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

	sb := image.Rect(
		viewPadding, 0,
		viewPadding+sidebarSize.X, u.winSize.Y,
	)

	return viewMetrics{
		chartBounds:   cb,
		sidebarBounds: sb,
	}
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
func (u *UI) SetChart(symbol string, data *chart.Data) error {
	if err := model.ValidateSymbol(symbol); err != nil {
		return err
	}

	for symbol, ch := range u.symbolToChartMap {
		delete(u.symbolToChartMap, symbol)
		ch.Close()
	}

	ch := chart.NewChart(fps)
	u.symbolToChartMap[symbol] = ch

	if err := u.titleBar.SetData(data); err != nil {
		return err
	}

	if err := ch.SetData(data); err != nil {
		return err
	}

	ch.SetRefreshButtonClickCallback(func() {
		if u.chartRefreshButtonClickCallback != nil {
			u.chartRefreshButtonClickCallback(symbol)
		}
	})

	ch.SetAddButtonClickCallback(func() {
		if u.chartAddButtonClickCallback != nil {
			u.chartAddButtonClickCallback(symbol)
		}
	})

	ch.SetZoomChangeCallback(func(zoomChange chart.ZoomChange) {
		u.handleChartZoomChangeEvent(zoomChange)
	})

	defer u.WakeLoop()
	for _, ch := range u.charts {
		ch.FadeOut()
	}
	u.charts = append([]*uiChart{newUIChart(ch)}, u.charts...)

	return nil
}

// AddChartThumb adds a thumbnail with the given symbol and data.
func (u *UI) AddChartThumb(symbol string, data *chart.Data) error {
	if err := model.ValidateSymbol(symbol); err != nil {
		return err
	}

	th := chart.NewThumb(fps)
	u.symbolToChartThumbMap[symbol] = th

	if err := th.SetData(data); err != nil {
		return err
	}

	th.SetRemoveButtonClickCallback(func() {
		if u.thumbRemoveButtonClickCallback != nil {
			u.thumbRemoveButtonClickCallback(symbol)
		}
	})

	th.SetThumbClickCallback(func() {
		if u.thumbClickCallback != nil {
			u.thumbClickCallback(symbol)
		}
	})

	defer u.WakeLoop()
	u.sidebar.AddChartThumb(th)

	return nil
}

// RemoveChartThumb removes the ChartThumbnail from the side bar.
func (u *UI) RemoveChartThumb(symbol string) error {
	if err := model.ValidateSymbol(symbol); err != nil {
		return err
	}

	th := u.symbolToChartThumbMap[symbol]
	delete(u.symbolToChartThumbMap, symbol)
	th.Close()

	defer u.WakeLoop()
	u.sidebar.RemoveChartThumb(th)

	return nil
}

// SetLoading sets the charts and slots matching the symbol and range to loading.
func (u *UI) SetLoading(symbol string, dataRange model.Range) error {
	if err := model.ValidateSymbol(symbol); err != nil {
		return err
	}

	if dataRange == model.RangeUnspecified {
		return errors.Errorf("range not set")
	}

	for s, ch := range u.symbolToChartMap {
		if s == symbol {
			ch.SetLoading(true)
			ch.SetError(nil)
		}
	}

	for s, th := range u.symbolToChartThumbMap {
		if s == symbol {
			th.SetLoading(true)
			th.SetError(nil)
		}
	}

	return nil
}

// SetData loads the data to charts and thumbnails matching the symbol and range.
func (u *UI) SetData(symbol string, data *chart.Data) error {
	if err := model.ValidateSymbol(symbol); err != nil {
		return err
	}

	if ch, ok := u.symbolToChartMap[symbol]; ok {
		ch.SetLoading(false)

		if err := u.titleBar.SetData(data); err != nil {
			return err
		}

		if err := ch.SetData(data); err != nil {
			return err
		}
	}

	if th, ok := u.symbolToChartThumbMap[symbol]; ok {
		th.SetLoading(false)

		if err := th.SetData(data); err != nil {
			return err
		}
	}

	return nil
}

// SetError sets an error on the charts and slots matching the symbol.
func (u *UI) SetError(symbol string, updateErr error) error {
	if err := model.ValidateSymbol(symbol); err != nil {
		return err
	}

	if ch, ok := u.symbolToChartMap[symbol]; ok {
		ch.SetLoading(false)
		ch.SetError(updateErr)
	}

	if th, ok := u.symbolToChartThumbMap[symbol]; ok {
		th.SetLoading(false)
		th.SetError(updateErr)
	}

	return nil
}
