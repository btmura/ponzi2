package view

import (
	"context"
	"fmt"
	"image"
	"runtime"
	"unicode"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"golang.org/x/image/font/gofont/goregular"

	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/gfx"
	"github.com/btmura/ponzi2/internal/logger"
	"github.com/btmura/ponzi2/internal/matrix"
)

// Get esc from github.com/mjibson/esc. It's used to embed resources into the binary.
//go:generate esc -o bindata.go -pkg view -include ".*(ply|png)" -modtime 1337 -private data

// Application name for the window title.
const appName = "ponzi"

// Constants used by Run for the "game loop".
const (
	fps          = 120.0
	secPerUpdate = 1.0 / fps
	maxUpdates   = 10
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

const viewOuterPadding = 10

var viewChartThumbSize = image.Pt(155, 105)

var (
	viewInputSymbolTextRenderer = gfx.NewTextRenderer(goregular.TTF, 48)
	viewInstructionsText        = NewCenteredText(gfx.NewTextRenderer(goregular.TTF, 24), "Type in symbol and press ENTER...")
)

func init() {
	// This is needed to arrange that main() runs on main thread for GLFW.
	// See documentation for functions that are only allowed to be called
	// from the main thread.
	runtime.LockOSThread()
}

// The View renders the UI to view and edit the model's stocks that it observes.
type View struct {
	// win is the handle to the GLFW window.
	win *glfw.Window

	// winSize is the current window's size used to measure and draw the UI.
	winSize image.Point

	// chart renders the currently viewed stock.
	chart *Chart

	// chartThumbs renders the stocks in the sidebar.
	chartThumbs []*ChartThumb

	// inputSymbol stores and renders the symbol being entered by the user.
	inputSymbol *CenteredText

	// inputSymbolSubmittedCallback is called when a new symbol is entered.
	inputSymbolSubmittedCallback func(symbol string)

	// mousePos is the current global mouse position.
	mousePos image.Point

	// mouseLeftButtonClicked is whether the left mouse button was clicked.
	mouseLeftButtonClicked bool
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
		inputSymbol:                  NewCenteredText(viewInputSymbolTextRenderer, "", CenteredTextBubble(chartRounding, chartPadding)),
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

	// Call the size callback to set the initial viewport.
	w, h := win.GetSize()
	v.setSize(w, h)
	win.SetSizeCallback(func(win *glfw.Window, width, height int) {
		v.setSize(width, height)
	})

	win.SetCharCallback(func(win *glfw.Window, char rune) {
		v.setChar(char)
	})

	win.SetKeyCallback(func(win *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		v.setKey(ctx, key, action)
	})

	win.SetCursorPosCallback(func(win *glfw.Window, xpos, ypos float64) {
		v.setCursorPos(xpos, ypos)
	})

	win.SetMouseButtonCallback(func(win *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
		v.setMouseButton(button, action)
	})

	return func() { glfw.Terminate() }, nil
}

func (v *View) setSize(width, height int) {
	s := image.Pt(width, height)
	if v.winSize == s {
		return
	}

	gl.Viewport(0, 0, int32(s.X), int32(s.Y))

	// Calculate the new ortho projection view matrix.
	fw, fh := float32(s.X), float32(s.Y)
	gfx.SetProjectionViewMatrix(matrix.Ortho(fw, fh, fw /* use width as depth */))

	v.winSize = s
}

func (v *View) setChar(char rune) {
	char = unicode.ToUpper(char)
	if _, ok := acceptedChars[char]; ok {
		v.inputSymbol.Text += string(char)
	}
}

func (v *View) setKey(ctx context.Context, key glfw.Key, action glfw.Action) {
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

func (v *View) setCursorPos(x, y float64) {
	// Flip Y-axis since the OpenGL coordinate system makes lower left the origin.
	v.mousePos = image.Pt(int(x), v.winSize.Y-int(y))
}

func (v *View) setMouseButton(button glfw.MouseButton, action glfw.Action) {
	if button != glfw.MouseButtonLeft {
		logger.Infof("ignoring mouse button(%v) and action(%v)", button, action)
		return // Only interested in left clicks right now.
	}
	v.mouseLeftButtonClicked = action == glfw.Release
}

// Run runs the "game loop".
func (v *View) Run(preupdate func()) {
	var lag float64
	animating := false
	prevTime := glfw.GetTime()
	for !v.win.ShouldClose() {
		currTime := glfw.GetTime()
		elapsed := currTime - prevTime
		prevTime = currTime
		lag += elapsed

		i := 0
		for ; lag >= secPerUpdate && i < maxUpdates; i++ {
			preupdate()
			animating = v.update()
			lag -= secPerUpdate
		}

		logger.Infof("updates: %d animating: %t", i, animating)

		v.render(float32(lag / secPerUpdate))
		v.win.SwapBuffers()

		glfw.PollEvents()
		if !animating {
			logger.Info("wait events")
			glfw.WaitEventsTimeout(1 /* seconds */)
		}
	}
}

func (v *View) update() (animating bool) {
	if v.chart != nil {
		if v.chart.Update() {
			animating = true
		}
	}
	for _, th := range v.chartThumbs {
		if th.Update() {
			animating = true
		}
	}
	return animating
}

func (v *View) render(fudge float32) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	vc := viewContext{
		Bounds:                 image.Rectangle{image.ZP, v.winSize},
		MousePos:               v.mousePos,
		MouseLeftButtonClicked: v.mouseLeftButtonClicked,
		Fudge:              fudge,
		ScheduledCallbacks: new([]func()),
	}

	ogBnds := vc.Bounds.Inset(viewOuterPadding)

	// Calculate bounds for main area.
	vc.Bounds = ogBnds
	if len(v.chartThumbs) > 0 {
		vc.Bounds.Min.X += viewOuterPadding + viewChartThumbSize.X
	}

	// Render the the main chart or instructions.
	if v.chart != nil {
		v.chart.Render(vc)
	} else {
		viewInstructionsText.Render(vc.Bounds)
	}

	// Render the input symbol over the chart.
	v.inputSymbol.Render(vc.Bounds)

	// Render the sidebar thumbnails.
	vc.Bounds = image.Rect(
		viewOuterPadding, ogBnds.Max.Y-viewChartThumbSize.Y,
		viewOuterPadding+viewChartThumbSize.X, ogBnds.Max.Y,
	)
	for _, th := range v.chartThumbs {
		th.Render(vc)
		vc.Bounds = vc.Bounds.Sub(image.Pt(0, viewChartThumbSize.Y+viewOuterPadding))
	}

	// Call any callbacks scheduled by views.
	for _, cb := range *vc.ScheduledCallbacks {
		cb()
	}

	// Reset any flags for the next viewContext.
	v.mouseLeftButtonClicked = false
}

// SetInputSymbolSubmittedCallback sets the callback for when a new symbol is entered.
func (v *View) SetInputSymbolSubmittedCallback(cb func(symbol string)) {
	v.inputSymbolSubmittedCallback = cb
}

// SetChart sets the View's main chart.
func (v *View) SetChart(ch *Chart) {
	v.chart = ch
}

// AddChartThumb adds the ChartThumbnail to the side bar.
func (v *View) AddChartThumb(th *ChartThumb) {
	v.chartThumbs = append(v.chartThumbs, th)
}

// RemoveChartThumb removes the ChartThumbnail from the side bar.
func (v *View) RemoveChartThumb(th *ChartThumb) {
	for i, thumb := range v.chartThumbs {
		if thumb == th {
			v.chartThumbs = append(v.chartThumbs[:i], v.chartThumbs[i+1:]...)
			break
		}
	}
}

// SetTitle sets the title to the given stock.
func (v *View) SetTitle(st *model.Stock) {
	glfw.GetCurrentContext().SetTitle(v.windowTitle(st))
}

func (v *View) windowTitle(st *model.Stock) string {
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
