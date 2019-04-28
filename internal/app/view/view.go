// Package view contains code for the view in the MVC pattern.
// TODO(btmura): package should not export mutable types like Chart that could interrupt the game loop
package view

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
	"github.com/btmura/ponzi2/internal/app/view/animation"
	"github.com/btmura/ponzi2/internal/app/view/centeredtext"
	"github.com/btmura/ponzi2/internal/app/view/chart"
	"github.com/btmura/ponzi2/internal/matrix"
)

// Embed resources into the application. Get esc from github.com/mjibson/esc.
//go:generate esc -o bindata.go -pkg view -include ".*(ply|png)" -modtime 1337 -private data

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

// ZoomChange specifies whether the user zooms in or out.
type ZoomChange int

// ZoomChange values.
//go:generate stringer -type=ZoomChange
const (
	ZoomChangeUnspecified ZoomChange = iota
	ZoomIn
	ZoomOut
)

// The View renders the UI to view and edit the model's stocks that it observes.
type View struct {
	// title renders the window title.
	title *Title

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
	chartZoomChangeCallback func(zoomChange ZoomChange)

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
	Render(fudge float32) error
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
func New() *View {
	return &View{
		title:                        NewTitle(),
		sidebar:                      new(sidebar),
		instructionsText:             centeredtext.New(gfx.NewTextRenderer(goregular.TTF, 24), "Type in symbol and press ENTER..."),
		inputSymbolText:              centeredtext.New(inputSymbolTextRenderer, "", centeredtext.Bubble(chartRounding, chartPadding)),
		inputSymbolSubmittedCallback: func(symbol string) {},
		chartZoomChangeCallback:      func(zoomChange ZoomChange) {},
	}
}

// Init initializes the View and returns a cleanup function.
func (v *View) Init(ctx context.Context) (cleanup func(), err error) {
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
	v.win = win

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
	v.handleSizeEvent(w, h)
	win.SetSizeCallback(func(win *glfw.Window, width, height int) {
		v.handleSizeEvent(width, height)
	})

	win.SetCharCallback(func(win *glfw.Window, char rune) {
		v.handleCharEvent(char)
	})

	win.SetKeyCallback(func(win *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		v.handleKeyEvent(key, action)
	})

	win.SetCursorPosCallback(func(win *glfw.Window, xpos, ypos float64) {
		v.handleCursorPosEvent(xpos, ypos)
	})

	win.SetMouseButtonCallback(func(win *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
		v.handleMouseButtonEvent(button, action)
	})

	win.SetScrollCallback(func(win *glfw.Window, xoff, yoff float64) {
		v.handleScrollEvent(yoff)
	})

	return func() { glfw.Terminate() }, nil
}

func (v *View) handleSizeEvent(width, height int) {
	glog.V(2).Infof("width:%o height:%o", width, height)
	defer v.WakeLoop()

	s := image.Pt(width, height)
	if v.winSize == s {
		return
	}

	gl.Viewport(0, 0, int32(s.X), int32(s.Y))

	// Calculate the new ortho projection view matrix.
	fw, fh := float32(s.X), float32(s.Y)
	gfx.SetProjectionViewMatrix(matrix.Ortho(fw, fh, fw /* use width as depth */))

	v.winSize = s

	// Reset the sidebar scroll offset if the sidebar is shorter than the window.
	m := v.metrics()
	if m.sidebarBounds.Dy() < m.sidebarRegion.Dy() {
		v.sidebar.sidebarScrollOffset = image.ZP
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

func (v *View) metrics() viewMetrics {
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

	if len(v.sidebar.thumbs) == 0 {
		cb := image.Rect(0, 0, v.winSize.X, v.winSize.Y)
		cb = cb.Inset(viewPadding)
		return viewMetrics{chartBounds: cb, chartRegion: cb}
	}

	cb := image.Rect(viewPadding+chartThumbSize.X, 0, v.winSize.X, v.winSize.Y)
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

	sh := (viewPadding+chartThumbSize.Y)*len(v.sidebar.thumbs) + viewPadding

	sb := image.Rect(
		viewPadding, v.winSize.Y-sh,
		viewPadding+chartThumbSize.X, v.winSize.Y,
	)
	sb = sb.Add(v.sidebar.sidebarScrollOffset)

	fb := image.Rect(
		sb.Min.X, sb.Max.Y-viewPadding-chartThumbSize.Y,
		sb.Max.X, sb.Max.Y-viewPadding,
	)

	// Side bar region.
	sr := image.Rect(
		viewPadding, 0,
		viewPadding+chartThumbSize.X, v.winSize.Y,
	)

	return viewMetrics{
		chartBounds:      cb,
		sidebarBounds:    sb,
		firstThumbBounds: fb,
		sidebarRegion:    sr,
		chartRegion:      cb,
	}
}

func (v *View) handleCharEvent(char rune) {
	glog.V(2).Infof("char:%c", char)
	defer v.WakeLoop()

	char = unicode.ToUpper(char)
	if _, ok := acceptedChars[char]; ok {
		v.inputSymbolText.Text += string(char)
	}
}

func (v *View) handleKeyEvent(key glfw.Key, action glfw.Action) {
	glog.V(2).Infof("key:%v action:%v", key, action)
	defer v.WakeLoop()

	if action != glfw.Release {
		return
	}

	switch key {
	case glfw.KeyEscape:
		v.inputSymbolText.Text = ""

	case glfw.KeyBackspace:
		if l := len(v.inputSymbolText.Text); l > 0 {
			v.inputSymbolText.Text = v.inputSymbolText.Text[:l-1]
		}

	case glfw.KeyEnter:
		v.inputSymbolSubmittedCallback(v.inputSymbolText.Text)
		v.inputSymbolText.Text = ""
	}
}

func (v *View) handleCursorPosEvent(x, y float64) {
	glog.V(2).Infof("x:%f y:%f", x, y)
	defer v.WakeLoop()

	// Flip Y-axis since the OpenGL coordinate system makes lower left the origin.
	v.mousePos = image.Pt(int(x), v.winSize.Y-int(y))
}

func (v *View) handleMouseButtonEvent(button glfw.MouseButton, action glfw.Action) {
	glog.V(2).Infof("button:%v action:%v", button, action)
	defer v.WakeLoop()

	if button != glfw.MouseButtonLeft {
		return
	}

	switch action {
	case glfw.Press:
		v.mouseLeftButtonPressedCount = 1
	case glfw.Release:
		v.mouseLeftButtonPressedCount = 0
	}

	v.mouseLeftButtonReleased = action == glfw.Release
}

func (v *View) handleScrollEvent(yoff float64) {
	glog.V(2).Infof("yoff:%f", yoff)
	defer v.WakeLoop()

	if yoff != -1 && yoff != +1 {
		return
	}

	if len(v.sidebar.thumbs) == 0 {
		return
	}

	m := v.metrics()

	switch {
	case v.mousePos.In(m.sidebarRegion):
		// Don't scroll if sidebar is shorter than the window.
		if m.sidebarBounds.Dy() < v.winSize.Y {
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

		v.sidebar.sidebarScrollOffset = v.sidebar.sidebarScrollOffset.Add(off)

	case v.mousePos.In(m.chartRegion):
		switch {
		case yoff < 0: // Scroll wheel down
			v.chartZoomChangeCallback(ZoomOut)
		case yoff > 0: // Scroll wheel up
			v.chartZoomChangeCallback(ZoomIn)
		}
	}
}

// RunLoop runs the "game loop".
func (v *View) RunLoop(ctx context.Context, runLoopHook func(context.Context) error) error {
start:
	var lag float64
	dirty := false
	prevTime := glfw.GetTime()
	for !v.win.ShouldClose() {
		currTime := glfw.GetTime() /* seconds */
		elapsed := currTime - prevTime
		prevTime = currTime
		lag += elapsed

		if err := runLoopHook(ctx); err != nil {
			return err
		}

		callbacks := v.processInput()

		i := 0
		for ; i < minUpdates || i < maxUpdates && lag >= updateSec; i++ {
			dirty = v.update()
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
		v.render(fudge)

		v.win.SwapBuffers()
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
func (v *View) WakeLoop() {
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

func (v *View) processInput() []func() {
	m := v.metrics()

	ic := inputContext{
		Bounds:                  m.chartBounds,
		MousePos:                v.mousePos,
		MouseLeftButtonDragging: v.mouseLeftButtonPressedCount > fps/2,
		MouseLeftButtonReleased: v.mouseLeftButtonReleased,
		ScheduledCallbacks:      new([]func()),
	}

	if ic.MouseLeftButtonDragging {
		fmt.Printf("drag: %t count: %d\n", ic.MouseLeftButtonDragging, v.mouseLeftButtonPressedCount)
	}

	v.instructionsText.ProcessInput(ic.Bounds)
	v.inputSymbolText.ProcessInput(ic.Bounds)

	for i := 0; i < len(v.charts); i++ {
		ch := v.charts[i]
		ch.chart.ProcessInput(ic.Bounds, ic.MousePos, ic.MouseLeftButtonReleased, ic.ScheduledCallbacks)
	}

	v.sidebar.ProcessInput(ic, m)

	// Reset any flags for the next inputContext.
	if v.mouseLeftButtonPressedCount > 0 {
		v.mouseLeftButtonPressedCount++
	}
	v.mouseLeftButtonReleased = false

	return *ic.ScheduledCallbacks
}

func (v *View) update() (dirty bool) {
	for i := 0; i < len(v.charts); i++ {
		ch := v.charts[i]
		if ch.Update() {
			dirty = true
		}
		if ch.DoneExiting() {
			v.charts = append(v.charts[:i], v.charts[i+1:]...)
			ch.Close()
			i--
		}
	}

	if v.sidebar.Update() {
		dirty = true
	}

	return dirty
}

func (v *View) render(fudge float32) {
	v.title.Render(v.win)

	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	// Render the main chart.
	for _, ch := range v.charts {
		ch.Render(fudge)
	}

	// Render instructions if there are no charts to show.
	if len(v.charts) == 0 {
		v.instructionsText.Render(fudge)
	}

	// Render the input symbol over the chart.
	v.inputSymbolText.Render(fudge)

	// Render the sidebar thumbnails.
	v.sidebar.Render(fudge)
}

// SetInputSymbolSubmittedCallback sets the callback for when a new symbol is entered.
func (v *View) SetInputSymbolSubmittedCallback(cb func(symbol string)) {
	v.inputSymbolSubmittedCallback = cb
}

// SetChartZoomChangeCallback sets the callback for when the chart is zoomed in or out.
func (v *View) SetChartZoomChangeCallback(cb func(zoomzoomChange ZoomChange)) {
	v.chartZoomChangeCallback = cb
}

// SetTitle sets the View's title.
func (v *View) SetTitle(title *Title) {
	defer v.WakeLoop()
	v.title = title
}

// NewChart returns a new chart.
func (v *View) NewChart() *chart.Chart {
	return chart.NewChart(fps)
}

// NewChartThumb returns a new chart thumbnail.
func (v *View) NewChartThumb() *chart.Thumb {
	return chart.NewChartThumb(fps)
}

// SetChart sets the View's main chart.
func (v *View) SetChart(ch *chart.Chart) {
	defer v.WakeLoop()
	for _, ch := range v.charts {
		ch.Exit()
	}
	v.charts = append([]*viewChart{newViewChart(ch)}, v.charts...)
}

// AddChartThumb adds the ChartThumbnail to the side bar.
func (v *View) AddChartThumb(th *chart.Thumb) {
	defer v.WakeLoop()
	v.sidebar.AddChartThumb(th)
}

// RemoveChartThumb removes the ChartThumbnail from the side bar.
func (v *View) RemoveChartThumb(th *chart.Thumb) {
	defer v.WakeLoop()
	v.sidebar.RemoveChartThumb(th)
}
