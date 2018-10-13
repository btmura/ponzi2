// Package view contains code for the view in the MVC pattern.
package view

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
	"github.com/golang/glog"
	"golang.org/x/image/font/gofont/goregular"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/matrix"
)

// Embed resources into the application. Get esc from github.com/mjibson/esc.
//go:generate esc -o bindata.go -pkg view -include ".*(ply|png)" -modtime 1337 -private data

// Add a Windows icon by generating a SYSO file in the root that will be picked
// up by the Go build tools. Get rsrc from github.com/akavel/rsrc.
//go:generate rsrc -ico data/icon.ico -arch amd64 -o ../../../ponzi2.syso

// Application name for the window title.
const appName = "ponzi2"

// Colors used throughout the UI.
var (
	green     = [3]float32{0.25, 1, 0}
	red       = [3]float32{1, 0.3, 0}
	yellow    = [3]float32{1, 1, 0}
	purple    = [3]float32{0.5, 0, 1}
	white     = [3]float32{1, 1, 1}
	gray      = [3]float32{0.15, 0.15, 0.15}
	lightGray = [3]float32{0.35, 0.35, 0.35}
	orange    = [3]float32{1, 0.5, 0}
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

// Constants used by Run for the "game loop".
const (
	fps        = 120.0
	updateSec  = 1.0 / fps
	minUpdates = 1
	maxUpdates = 1000
)

const viewPadding = 10

var (
	chartThumbSize         = image.Pt(155, 105)
	chartThumbRenderOffset = image.Pt(0, viewPadding+chartThumbSize.Y)
	sidebarScrollAmount    = chartThumbRenderOffset
)

var (
	inputSymbolTextRenderer = gfx.NewTextRenderer(goregular.TTF, 48)
	instructionsText        = newCenteredText(gfx.NewTextRenderer(goregular.TTF, 24), "Type in symbol and press ENTER...")
)

func init() {
	// This is needed to arrange that main() runs on main thread for GLFW.
	// See documentation for functions that are only allowed to be called
	// from the main thread.
	runtime.LockOSThread()
}

// The View renders the UI to view and edit the model's stocks that it observes.
type View struct {
	// charts renders the charts in the main area.
	charts []*viewChart

	// chartThumbs renders the stocks in the sidebar.
	chartThumbs []*viewChartThumb

	// inputSymbol stores and renders the symbol being entered by the user.
	inputSymbol *centeredText

	// inputSymbolSubmittedCallback is called when a new symbol is entered.
	inputSymbolSubmittedCallback func(symbol string)

	// win is the handle to the GLFW window.
	win *glfw.Window

	// winSize is the current window's size used to measure and draw the UI.
	winSize image.Point

	// mousePos is the current global mouse position.
	mousePos image.Point

	// mouseLeftButtonClicked is whether the left mouse button was clicked.
	mouseLeftButtonClicked bool

	// sidebarScrollOffset stores the Y offset accumulated from scroll events
	// that should be used to calculate the sidebar's bounds.
	sidebarScrollOffset image.Point
}

type viewChart struct {
	chart *Chart
	*viewAnimator
}

func newViewChart(ch *Chart) *viewChart {
	return &viewChart{
		chart:        ch,
		viewAnimator: newViewAnimator(ch),
	}
}

type viewChartThumb struct {
	chartThumb *ChartThumb
	*viewAnimator
}

func newViewChartThumb(th *ChartThumb) *viewChartThumb {
	return &viewChartThumb{
		chartThumb:   th,
		viewAnimator: newViewAnimator(th),
	}
}

type viewUpdateRenderCloser interface {
	Update() (dirty bool)
	Render(viewContext)
	Close()
}

type viewAnimator struct {
	updateRenderCloser viewUpdateRenderCloser
	exiting            bool
	fade               *animation
	inset              *animation
}

func newViewAnimator(updateRenderer viewUpdateRenderCloser) *viewAnimator {
	return &viewAnimator{
		updateRenderCloser: updateRenderer,
		fade:               newAnimation(1*fps, animationStarted()),
	}
}

func (v *viewAnimator) Exit() {
	v.exiting = true
	v.fade = v.fade.Rewinded()
	v.fade.Start()
}

func (v *viewAnimator) Animating() bool {
	return v.fade.Animating()
}

func (v *viewAnimator) Remove() bool {
	return v.exiting && !v.Animating()
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

func (v *viewAnimator) Render(vc viewContext) {
	old := gfx.Alpha()
	defer gfx.SetAlpha(old)
	gfx.SetAlpha(v.fade.Value(vc.Fudge))
	v.updateRenderCloser.Render(vc)
}

func (v *viewAnimator) Close() {
	v.updateRenderCloser.Close()
}

// viewContext is passed down the view hierarchy providing drawing hints and
// event information. Meant to be passed around like a Rectangle or Point rather
// than a pointer to avoid mistakes.
type viewContext struct {
	// Bounds is the rectangle with global coords that should be drawn within.
	Bounds image.Rectangle

	// MousePos is the current global mouse position.
	MousePos image.Point

	// MouseLeftButtonClicked is whether the left mouse button was clicked.
	MouseLeftButtonClicked bool

	// Fudge is the position from 0 to 1 between the current and next frame.
	Fudge float32

	// ScheduledCallbacks are callbacks to be called at the end of Render.
	ScheduledCallbacks *[]func()
}

// LeftClickInBounds returns true if the left mouse button was clicked within
// the context's bounds. Doesn't take into account overlapping view parts.
func (vc viewContext) LeftClickInBounds() bool {
	return vc.MouseLeftButtonClicked && vc.MousePos.In(vc.Bounds)
}

// New creates a new View.
func New() *View {
	return &View{
		inputSymbol:                  newCenteredText(inputSymbolTextRenderer, "", centeredTextBubble(chartRounding, chartPadding)),
		inputSymbolSubmittedCallback: func(symbol string) {},
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

	win, err := glfw.CreateWindow(800, 600, appName, nil, nil)
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
		v.sidebarScrollOffset = image.ZP
	}
}

// viewMetrics has dynamic metrics used to render the view.
type viewMetrics struct {
	// chartBounds is where to draw the main chart.
	chartBounds image.Rectangle

	// sidebarBounds is where to draw the sidebar with thumbnails.
	sidebarBounds image.Rectangle

	// firstThumbBounds is where to draw the first thumbnail in the sidebar.
	firstThumbBounds image.Rectangle

	// sidebarRegion is where to detect scroll events for the sidebar.
	sidebarRegion image.Rectangle
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

	if len(v.chartThumbs) == 0 {
		cb := image.Rect(0, 0, v.winSize.X, v.winSize.Y)
		cb = cb.Inset(viewPadding)
		return viewMetrics{chartBounds: cb}
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

	sh := (viewPadding+chartThumbSize.Y)*len(v.chartThumbs) + viewPadding

	sb := image.Rect(
		viewPadding, v.winSize.Y-sh,
		viewPadding+chartThumbSize.X, v.winSize.Y,
	)
	sb = sb.Add(v.sidebarScrollOffset)

	fb := image.Rect(
		sb.Min.X, sb.Max.Y-viewPadding-chartThumbSize.Y,
		sb.Max.X, sb.Max.Y-viewPadding,
	)

	ssb := image.Rect(
		viewPadding, 0,
		viewPadding+chartThumbSize.X, v.winSize.Y,
	)

	return viewMetrics{
		chartBounds:      cb,
		sidebarBounds:    sb,
		firstThumbBounds: fb,
		sidebarRegion:    ssb,
	}
}

func (v *View) handleCharEvent(char rune) {
	glog.V(2).Infof("char:%c", char)
	defer v.WakeLoop()

	char = unicode.ToUpper(char)
	if _, ok := acceptedChars[char]; ok {
		v.inputSymbol.Text += string(char)
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
		v.inputSymbol.Text = ""

	case glfw.KeyBackspace:
		if l := len(v.inputSymbol.Text); l > 0 {
			v.inputSymbol.Text = v.inputSymbol.Text[:l-1]
		}

	case glfw.KeyEnter:
		v.inputSymbolSubmittedCallback(v.inputSymbol.Text)
		v.inputSymbol.Text = ""
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

	v.mouseLeftButtonClicked = action == glfw.Release
}

func (v *View) handleScrollEvent(yoff float64) {
	glog.V(2).Infof("yoff:%f", yoff)
	defer v.WakeLoop()

	if yoff != -1 && yoff != +1 {
		return
	}

	if len(v.chartThumbs) == 0 {
		return
	}

	m := v.metrics()

	if !v.mousePos.In(m.sidebarRegion) {
		return
	}

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

	v.sidebarScrollOffset = v.sidebarScrollOffset.Add(off)
}

// SetInputSymbolSubmittedCallback sets the callback for when a new symbol is entered.
func (v *View) SetInputSymbolSubmittedCallback(cb func(symbol string)) {
	v.inputSymbolSubmittedCallback = cb
}

// RunLoop runs the "game loop".
func (v *View) RunLoop(preupdate func()) {
start:
	var lag float64
	dirty := false
	prevTime := glfw.GetTime()
	for !v.win.ShouldClose() {
		currTime := glfw.GetTime() /* seconds */
		elapsed := currTime - prevTime
		prevTime = currTime
		lag += elapsed

		i := 0
		for ; i < minUpdates || i < maxUpdates && lag >= updateSec; i++ {
			preupdate()
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
		if v.render(fudge) {
			dirty = true
		}
		v.win.SwapBuffers()
		glog.V(2).Infof("updates:%o lag(%f)/updateSec(%f)=fudge(%f) dirty:%t render:%v", i, lag, updateSec, fudge, dirty, time.Since(now).Seconds())

		glfw.PollEvents()
		if !dirty {
			glog.V(2).Info("wait events")
			glfw.WaitEvents()
			goto start
		}
	}
}

// WakeLoop wakes up the loop if it is asleep.
func (v *View) WakeLoop() {
	glfw.PostEmptyEvent()
}

func (v *View) update() (dirty bool) {
	for i := 0; i < len(v.charts); i++ {
		ch := v.charts[i]
		if ch.Update() {
			dirty = true
		}
		if ch.Remove() {
			v.charts = append(v.charts[:i], v.charts[i+1:]...)
			ch.Close()
			i--
		}
	}

	for i := 0; i < len(v.chartThumbs); i++ {
		th := v.chartThumbs[i]
		if th.Update() {
			dirty = true
		}
		if th.Remove() {
			v.chartThumbs = append(v.chartThumbs[:i], v.chartThumbs[i+1:]...)
			th.Close()
			i--
		}
	}

	return dirty
}

func (v *View) render(fudge float32) (dirty bool) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	m := v.metrics()

	vc := viewContext{
		Bounds:                 m.chartBounds,
		MousePos:               v.mousePos,
		MouseLeftButtonClicked: v.mouseLeftButtonClicked,
		Fudge:              fudge,
		ScheduledCallbacks: new([]func()),
	}

	// Render the main chart.
	for _, ch := range v.charts {
		ch.Render(vc)
	}

	// Render instructions if there are no charts to show.
	if len(v.charts) == 0 {
		instructionsText.Render(vc.Bounds)
	}

	// Render the input symbol over the chart.
	v.inputSymbol.Render(vc.Bounds)

	// Render the sidebar thumbnails.
	if len(v.chartThumbs) != 0 {
		vc.Bounds = m.firstThumbBounds
		for _, th := range v.chartThumbs {
			th.Render(vc)
			vc.Bounds = vc.Bounds.Sub(chartThumbRenderOffset)
		}
	}

	// Call any callbacks scheduled by views.
	for _, cb := range *vc.ScheduledCallbacks {
		cb()
	}

	// Reset any flags for the next viewContext.
	v.mouseLeftButtonClicked = false

	// Return dirty if some callbacks were scheduled.
	return len(*vc.ScheduledCallbacks) != 0
}

// SetChart sets the View's main chart.
func (v *View) SetChart(ch *Chart) {
	defer v.WakeLoop()
	for _, ch := range v.charts {
		ch.Exit()
	}
	v.charts = append([]*viewChart{newViewChart(ch)}, v.charts...)
}

// AddChartThumb adds the ChartThumbnail to the side bar.
func (v *View) AddChartThumb(th *ChartThumb) {
	defer v.WakeLoop()
	v.chartThumbs = append(v.chartThumbs, newViewChartThumb(th))
}

// RemoveChartThumb removes the ChartThumbnail from the side bar.
func (v *View) RemoveChartThumb(th *ChartThumb) {
	defer v.WakeLoop()
	for _, vth := range v.chartThumbs {
		if vth.chartThumb == th {
			vth.Exit()
			break
		}
	}
}
