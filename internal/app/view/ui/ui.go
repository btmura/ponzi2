// Package ui contains code for the view in the MVC pattern.
// TODO(btmura): package should not export mutable types like Chart that could interrupt the game loop
package ui

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"runtime"
	"time"
	"unicode"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang/glog"
	"golang.org/x/image/font/gofont/goregular"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view"
	"github.com/btmura/ponzi2/internal/app/view/animation"
	"github.com/btmura/ponzi2/internal/app/view/centeredtext"
	"github.com/btmura/ponzi2/internal/app/view/chart"
	"github.com/btmura/ponzi2/internal/errors"
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

const viewPadding = 10

const (
	chartRounding = 10
	chartPadding  = 5
)

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

	// chartRange is the current data range to use for Charts.
	chartRange model.Range

	// chartThumbRange is the current data range to use for ChartThumbnails.
	chartThumbRange model.Range

	// titleBar renders the window titleBar.
	titleBar *titleBar

	// charts renders the charts in the main area.
	charts []*viewChart

	// sidebar is the sidebar of chart thumbnails on the side.
	sidebar *sidebar

	// instructionsText is instructional text show when no chart is shown.
	instructionsText *centeredtext.CenteredText

	// inputSymbolText stores and renders the symbol being entered by the user.
	inputSymbolText *centeredtext.CenteredText

	// inputSymbolSubmittedCallback is called when a new symbol is entered.
	inputSymbolSubmittedCallback func(symbol string)

	// chartZoomChangeCallback is called when the chart is zoomed in or out.
	chartZoomChangeCallback func(zoomChange view.ZoomChange)

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

	// mousePos is the current global mouse position.
	mousePos image.Point

	// mouseLeftButtonPressedCount is the number of loop iterations the left
	// mouse button has been pressed. Used to determine dragging.
	mouseLeftButtonPressedCount int

	// mouseLeftButtonReleased is whether the left mouse button was clicked.
	mouseLeftButtonReleased bool
}

type viewChart struct {
	chart *chart.Chart
	*viewAnimator
}

func newViewChart(ch *chart.Chart) *viewChart {
	return &viewChart{
		chart:        ch,
		viewAnimator: newViewAnimator(ch),
	}
}

type viewUpdateRenderCloser interface {
	Update() (dirty bool)
	Render(fudge float32)
	Close()
}

type viewAnimator struct {
	updateRenderCloser viewUpdateRenderCloser
	exiting            bool
	fade               *animation.Animation
}

func newViewAnimator(updateRenderer viewUpdateRenderCloser) *viewAnimator {
	return &viewAnimator{
		updateRenderCloser: updateRenderer,
		fade:               animation.New(1*fps, animation.Started()),
	}
}

func (v *viewAnimator) Exit() {
	v.exiting = true
	v.fade = v.fade.Rewinded()
	v.fade.Start()
}

func (v *viewAnimator) DoneExiting() bool {
	return v.exiting && !v.fade.Animating()
}

func (v *viewAnimator) Update() (dirty bool) {
	if v.updateRenderCloser.Update() {
		dirty = true
	}
	if v.fade.Update() {
		dirty = true
	}
	return dirty
}

func (v *viewAnimator) Render(fudge float32) {
	old := gfx.Alpha()
	defer gfx.SetAlpha(old)
	gfx.SetAlpha(v.fade.Value(fudge))
	v.updateRenderCloser.Render(fudge)
}

func (v *viewAnimator) Close() {
	v.updateRenderCloser.Close()
}

// New creates a new View.
func New() *UI {
	return &UI{
		symbolToChartMap:                map[string]*chart.Chart{},
		symbolToChartThumbMap:           map[string]*chart.Thumb{},
		chartRange:                      model.OneYear,
		chartThumbRange:                 model.OneYear,
		sidebar:                         new(sidebar),
		instructionsText:                centeredtext.New(gfx.NewTextRenderer(goregular.TTF, 24), "Type in symbol and press ENTER..."),
		inputSymbolText:                 centeredtext.New(inputSymbolTextRenderer, "", centeredtext.Bubble(chartRounding, chartPadding)),
		inputSymbolSubmittedCallback:    func(symbol string) {},
		chartZoomChangeCallback:         func(zoomChange view.ZoomChange) {},
		chartRefreshButtonClickCallback: func(symbol string) {},
		chartAddButtonClickCallback:     func(symbol string) {},
		thumbRemoveButtonClickCallback:  func(symbol string) {},
		thumbClickCallback:              func(symbol string) {},
	}
}

// Init initializes the View and returns a cleanup function.
func (u *UI) Init(ctx context.Context) (cleanup func(), err error) {
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
	glog.V(2).Infof("OpenGL version: %s", gl.GoStr(gl.GetString(gl.VERSION)))

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

	return func() { glfw.Terminate() }, nil
}

func (u *UI) handleSizeEvent(width, height int) {
	glog.V(2).Infof("width:%o height:%o", width, height)
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

	// Reset the sidebar scroll offset if the sidebar is shorter than the window.
	m := u.metrics()
	if m.sidebarBounds.Dy() < m.sidebarRegion.Dy() {
		u.sidebar.sidebarScrollOffset = image.ZP
	}
}

// viewMetrics has dynamic metrics used to render the view.
type viewMetrics struct {
	// chartBounds is where to draw the main chart.
	chartBounds image.Rectangle

	// sidebarBounds is where to draw the sidebar that can move up or down.
	sidebarBounds image.Rectangle

	// firstThumbBounds is where to draw the first thumbnail in the sidebar.
	firstThumbBounds image.Rectangle

	// sidebarRegion is where to detect scroll events for the sidebar.
	sidebarRegion image.Rectangle

	// chartRegion is where to detect scroll events for the chart.
	chartRegion image.Rectangle
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

	if len(u.sidebar.thumbs) == 0 {
		cb := image.Rect(0, 0, u.winSize.X, u.winSize.Y)
		cb = cb.Inset(viewPadding)
		return viewMetrics{chartBounds: cb, chartRegion: cb}
	}

	cb := image.Rect(viewPadding+chartThumbSize.X, 0, u.winSize.X, u.winSize.Y)
	cb = cb.Inset(viewPadding)

	// +---+---------+---+---------+---+
	// |   | padding |   | padding |   |
	// |   +---------+   +---------+   |
	// |   | thumb   |   |         |   |
	// |   +---------+   |         |   |
	// | p | padding | p | chart   | p |
	// |   +---------+   |         |   |
	// |   | thumb   |   |         |   |
	// |   +---------+   +---------+   |
	// |   | padding |   | padding |   |
	// +---+---------+---+---------+---+

	sh := (viewPadding+chartThumbSize.Y)*len(u.sidebar.thumbs) + viewPadding

	sb := image.Rect(
		viewPadding, u.winSize.Y-sh,
		viewPadding+chartThumbSize.X, u.winSize.Y,
	)
	sb = sb.Add(u.sidebar.sidebarScrollOffset)

	fb := image.Rect(
		sb.Min.X, sb.Max.Y-viewPadding-chartThumbSize.Y,
		sb.Max.X, sb.Max.Y-viewPadding,
	)

	// Side bar region.
	sr := image.Rect(
		viewPadding, 0,
		viewPadding+chartThumbSize.X, u.winSize.Y,
	)

	return viewMetrics{
		chartBounds:      cb,
		sidebarBounds:    sb,
		firstThumbBounds: fb,
		sidebarRegion:    sr,
		chartRegion:      cb,
	}
}

func (u *UI) handleCharEvent(char rune) {
	glog.V(2).Infof("char:%c", char)
	defer u.WakeLoop()

	char = unicode.ToUpper(char)
	if _, ok := acceptedChars[char]; ok {
		u.inputSymbolText.Text += string(char)
	}
}

func (u *UI) handleKeyEvent(key glfw.Key, action glfw.Action) {
	glog.V(2).Infof("key:%v action:%v", key, action)
	defer u.WakeLoop()

	if action != glfw.Release {
		return
	}

	switch key {
	case glfw.KeyEscape:
		u.inputSymbolText.Text = ""

	case glfw.KeyBackspace:
		if l := len(u.inputSymbolText.Text); l > 0 {
			u.inputSymbolText.Text = u.inputSymbolText.Text[:l-1]
		}

	case glfw.KeyEnter:
		u.inputSymbolSubmittedCallback(u.inputSymbolText.Text)
		u.inputSymbolText.Text = ""
	}
}

func (u *UI) handleCursorPosEvent(x, y float64) {
	glog.V(2).Infof("x:%f y:%f", x, y)
	defer u.WakeLoop()

	// Flip Y-axis since the OpenGL coordinate system makes lower left the origin.
	u.mousePos = image.Pt(int(x), u.winSize.Y-int(y))
}

func (u *UI) handleMouseButtonEvent(button glfw.MouseButton, action glfw.Action) {
	glog.V(2).Infof("button:%v action:%v", button, action)
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

	u.mouseLeftButtonReleased = action == glfw.Release
}

func (u *UI) handleScrollEvent(yoff float64) {
	glog.V(2).Infof("yoff:%f", yoff)
	defer u.WakeLoop()

	if yoff != -1 && yoff != +1 {
		return
	}

	if len(u.sidebar.thumbs) == 0 {
		return
	}

	m := u.metrics()

	switch {
	case u.mousePos.In(m.sidebarRegion):
		// Don't scroll if sidebar is shorter than the window.
		if m.sidebarBounds.Dy() < u.winSize.Y {
			return
		}

		// Scroll wheel down: yoff = -1 up: yoff = +1
		off := sidebarScrollAmount.Mul(-int(yoff))
		tmpRect := m.sidebarBounds.Add(off)
		if botGap := tmpRect.Min.Y - m.sidebarRegion.Min.Y; botGap > 0 {
			off.Y -= botGap
		}
		if topGap := m.sidebarRegion.Max.Y - tmpRect.Max.Y; topGap > 0 {
			off.Y += topGap
		}

		u.sidebar.sidebarScrollOffset = u.sidebar.sidebarScrollOffset.Add(off)

	case u.mousePos.In(m.chartRegion):
		zoomChange := view.ZoomChangeUnspecified

		switch {
		case yoff < 0: // Scroll wheel down
			zoomChange = view.ZoomOut
		case yoff > 0: // Scroll wheel up
			zoomChange = view.ZoomIn
		}

		if zoomChange == view.ZoomChangeUnspecified {
			return
		}

		r := nextRange(u.chartRange, zoomChange)

		if u.chartRange == r {
			return
		}

		u.chartRange = r

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

		callbacks := u.processInput()

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
		glog.V(3).Infof("updates:%o lag(%f)/updateSec(%f)=fudge(%f) dirty:%t render:%v", i, lag, updateSec, fudge, dirty, time.Since(now).Seconds())

		// Call any callbacks scheduled by views.
		for _, cb := range callbacks {
			cb()
		}

		// Mark dirty since new charts or thumbs may have been added.
		if len(callbacks) != 0 {
			dirty = true
		}

		glfw.PollEvents()
		if !dirty {
			glog.V(3).Info("wait events")
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

type inputContext struct {
	// Bounds is the rectangle with global coords that should be drawn within.
	Bounds image.Rectangle

	// MousePos is the current global mouse position.
	MousePos image.Point

	// MouseLeftButtonDragging is whether the left mouse button is dragging.
	MouseLeftButtonDragging bool

	// MouseLeftButtonReleased is whether the left mouse button was released.
	MouseLeftButtonReleased bool

	// ScheduledCallbacks are callbacks to be called at the end of Render.
	ScheduledCallbacks *[]func()
}

// LeftClickInBounds returns true if the left mouse button was clicked within
// the context's bounds. Doesn't take into account overlapping view parts.
func (ic inputContext) LeftClickInBounds() bool {
	return ic.MouseLeftButtonReleased && ic.MousePos.In(ic.Bounds)
}

func (u *UI) processInput() []func() {
	m := u.metrics()

	ic := inputContext{
		Bounds:                  m.chartBounds,
		MousePos:                u.mousePos,
		MouseLeftButtonDragging: u.mouseLeftButtonPressedCount > fps/2,
		MouseLeftButtonReleased: u.mouseLeftButtonReleased,
		ScheduledCallbacks:      new([]func()),
	}

	if ic.MouseLeftButtonDragging {
		fmt.Printf("drag: %t count: %d\n", ic.MouseLeftButtonDragging, u.mouseLeftButtonPressedCount)
	}

	u.instructionsText.ProcessInput(ic.Bounds)
	u.inputSymbolText.ProcessInput(ic.Bounds)

	for i := 0; i < len(u.charts); i++ {
		ch := u.charts[i]
		ch.chart.ProcessInput(ic.Bounds, ic.MousePos, ic.MouseLeftButtonReleased, ic.ScheduledCallbacks)
	}

	u.sidebar.ProcessInput(ic, m)

	// Reset any flags for the next inputContext.
	if u.mouseLeftButtonPressedCount > 0 {
		u.mouseLeftButtonPressedCount++
	}
	u.mouseLeftButtonReleased = false

	return *ic.ScheduledCallbacks
}

func (u *UI) update() (dirty bool) {
	for i := 0; i < len(u.charts); i++ {
		ch := u.charts[i]
		if ch.Update() {
			dirty = true
		}
		if ch.DoneExiting() {
			u.charts = append(u.charts[:i], u.charts[i+1:]...)
			ch.Close()
			i--
		}
	}

	if u.sidebar.Update() {
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
		u.instructionsText.Render(fudge)
	}

	// Render the input symbol over the chart.
	u.inputSymbolText.Render(fudge)

	// Render the sidebar thumbnails.
	u.sidebar.Render(fudge)
}

// SetInputSymbolSubmittedCallback sets the callback for when a new symbol is entered.
func (u *UI) SetInputSymbolSubmittedCallback(cb func(symbol string)) error {
	if cb == nil {
		return errors.Errorf("missing callback")
	}
	u.inputSymbolSubmittedCallback = cb
	return nil
}

// SetChartZoomChangeCallback sets the callback for when the chart is zoomed in or out.
func (u *UI) SetChartZoomChangeCallback(cb func(zoomChange view.ZoomChange)) error {
	if cb == nil {
		return errors.Errorf("missing callback")
	}
	u.chartZoomChangeCallback = cb
	return nil
}

// SetChartRefreshButtonClickCallback sets the callback for when the main chart's refresh button is clicked.
func (u *UI) SetChartRefreshButtonClickCallback(cb func(symbol string)) error {
	if cb == nil {
		return errors.Errorf("missing callback")
	}
	u.chartRefreshButtonClickCallback = cb
	return nil
}

// SetChartAddButtonClickCallback sets the callback for when the main chart's add button is clicked.
func (u *UI) SetChartAddButtonClickCallback(cb func(symbol string)) error {
	if cb == nil {
		return errors.Errorf("missing callback")
	}
	u.chartAddButtonClickCallback = cb
	return nil
}

// SetThumbRemoveButtonClickCallback sets the callback for when a thumb's remove button is clicked.
func (u *UI) SetThumbRemoveButtonClickCallback(cb func(symbol string)) error {
	if cb == nil {
		return errors.Errorf("missing callback")
	}
	u.thumbRemoveButtonClickCallback = cb
	return nil
}

// SetThumbClickCallback sets the callback for when a thumb is clicked.
func (u *UI) SetThumbClickCallback(cb func(symbol string)) error {
	if cb == nil {
		return errors.Errorf("missing callback")
	}
	u.thumbClickCallback = cb
	return nil
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

	defer u.WakeLoop()
	for _, ch := range u.charts {
		ch.Exit()
	}
	u.charts = append([]*viewChart{newViewChart(ch)}, u.charts...)

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

// SetLoading sets the charts and thumbs matching the symbol and range to loading.
func (u *UI) SetLoading(symbol string, dataRange model.Range) error {
	if err := model.ValidateSymbol(symbol); err != nil {
		return err
	}

	if dataRange == model.RangeUnspecified {
		return errors.Errorf("range not set")
	}

	for s, ch := range u.symbolToChartMap {
		if s == symbol && u.chartRange == dataRange {
			ch.SetLoading(true)
			ch.SetError(false)
		}
	}

	for s, th := range u.symbolToChartThumbMap {
		if s == symbol && u.chartThumbRange == dataRange {
			th.SetLoading(true)
			th.SetError(false)
		}
	}

	return nil
}

// SetData loads the data to charts and thumbs matching the symbol and range.
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

// SetError sets an error on the charts and thumbs matching the symbol.
func (u *UI) SetError(symbol string, updateErr error) error {
	if err := model.ValidateSymbol(symbol); err != nil {
		return err
	}

	if ch, ok := u.symbolToChartMap[symbol]; ok {
		ch.SetLoading(false)
		ch.SetError(true)
	}

	if th, ok := u.symbolToChartThumbMap[symbol]; ok {
		th.SetLoading(false)
		th.SetError(true)
	}

	return nil
}

// ChartRange returns the range desired by the main chart.
func (u *UI) ChartRange() model.Range {
	return u.chartRange
}

// ChartThumbRange returns the range desired by the chart thumbnails.
func (u *UI) ChartThumbRange() model.Range {
	return u.chartThumbRange
}

func nextRange(r model.Range, zoomChange view.ZoomChange) model.Range {
	// zoomRanges are the ranges from most zoomed out to most zoomed in.
	var zoomRanges = []model.Range{
		model.OneYear,
		model.OneDay,
	}

	// Find the current zoom range.
	i := 0
	for j := range zoomRanges {
		if zoomRanges[j] == r {
			i = j
		}
	}

	// Adjust the zoom one increment.
	switch zoomChange {
	case view.ZoomIn:
		if i+1 < len(zoomRanges) {
			i++
		}
	case view.ZoomOut:
		if i-1 >= 0 {
			i--
		}
	}

	return zoomRanges[i]
}
