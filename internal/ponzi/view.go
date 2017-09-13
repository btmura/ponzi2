package ponzi

import (
	"image"
	"unicode"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang/glog"
	"golang.org/x/image/font/gofont/goregular"

	"github.com/btmura/ponzi2/internal/gfx"
	math2 "github.com/btmura/ponzi2/internal/math"
)

// Colors used throughout the UI.
var (
	green  = [3]float32{0.25, 1, 0}
	red    = [3]float32{1, 0.3, 0}
	yellow = [3]float32{1, 1, 0}
	purple = [3]float32{0.5, 0, 1}
	white  = [3]float32{1, 1, 1}
	gray   = [3]float32{0.15, 0.15, 0.15}
)

var (
	cameraPosition = math2.Vector3{X: 0, Y: 5, Z: 10}
	targetPosition = math2.Vector3{}
	up             = math2.Vector3{X: 0, Y: 1, Z: 0}
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

var inputSymbolTextRenderer = gfx.NewTextRenderer(goregular.TTF, 48)

const viewOuterPadding = 10

var viewChartThumbSize = image.Pt(140, 90)

// The View renders the UI to view and edit the model's stocks that it observes.
type View struct {
	// model has the stock data to be rendered.
	model *Model

	// chart renders the model's current stock.
	chart *Chart

	// chartThumbs renders the model's other stocks.
	chartThumbs []*ChartThumbnail

	// inputSymbol stores and renders the symbol being entered by the user.
	inputSymbol *CenteredText

	// nextViewContext is the next viewContext to pass down the view hierarchy.
	nextViewContext ViewContext

	// winSize is the current window's size used to measure and draw the UI.
	winSize image.Point

	orthoMatrix math2.Matrix4
}

// ViewContext is passed down the view hierarchy providing drawing hints and event information.
// Meant to be passed around like a Rectangle or Point rather than a pointer to avoid mistakes.
type ViewContext struct {
	// Bounds is the rectangle with global coordinates that the view part should draw within.
	Bounds image.Rectangle

	// MousePos is the current global mouse position.
	MousePos image.Point

	// MouseLeftButtonClicked is whether the left mouse button was clicked.
	MouseLeftButtonClicked bool

	// values stores values collected throughout the Render pass.
	values *viewContextValues
}

// viewContextValues stores values collected throughout the Render pass.
type viewContextValues struct {
	// scheduledCallbacks are callbacks that should be called at the end of Render.
	scheduledCallbacks []func()
}

// LeftClickInBounds returns true if the left mouse button was clicked within the context's bounds.
// Doesn't take into account overlapping view parts.
func (vc ViewContext) LeftClickInBounds() bool {
	return vc.MouseLeftButtonClicked && vc.MousePos.In(vc.Bounds)
}

// ScheduleCallbacks schedules a slice of callbacks that will be called after Render is done.
func (vc ViewContext) ScheduleCallbacks(cbs []func()) {
	vc.values.scheduledCallbacks = append(vc.values.scheduledCallbacks, cbs...)
}

// NewView creates a new View that observes the given Model.
func NewView(model *Model) *View {

	// Initialize OpenGL and enable features.

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

	if err := gfx.Init(); err != nil {
		glog.Fatalf("newView: failed to init gfx: %v", err)
	}

	v := &View{
		model:       model,
		inputSymbol: NewCenteredText(inputSymbolTextRenderer, ""),
	}

	if model.CurrentStock != nil {
		v.chart = v.newChart(model.CurrentStock)
	}

	for _, st := range model.Stocks {
		v.addSidebarChartThumb(st)
	}

	return v
}

// Update updates the view.
func (v *View) Update() {
	v.model.Lock()
	defer v.model.Unlock()

	if v.chart != nil {
		v.chart.Update()
	}

	for _, th := range v.chartThumbs {
		th.Update()
	}
}

// Render renders the view.
func (v *View) Render(fudge float32) {
	v.model.Lock()
	defer v.model.Unlock()

	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gfx.SetProjectionViewMatrix(v.orthoMatrix)

	if v.chart == nil || v.chart.stock != v.model.CurrentStock {
		if v.chart != nil {
			v.chart.Close()
		}
		v.chart = v.newChart(v.model.CurrentStock)
	}

	vc := v.nextViewContext
	vc.values = &viewContextValues{}

	// Render the input symbol and the main chart.
	if v.chart != nil {
		vc.Bounds = image.Rectangle{image.ZP, v.winSize}.Inset(viewOuterPadding)
		if len(v.model.Stocks) > 0 {
			vc.Bounds.Min.X += viewOuterPadding + viewChartThumbSize.X
		}
		v.inputSymbol.Render(vc)
		v.chart.Render(vc)
	}

	// Render the sidebar thumbnails.
	for i, th := range v.chartThumbs {
		min := image.Pt(viewOuterPadding, 0)
		max := image.Pt(min.X+viewChartThumbSize.X, 0)
		max.Y = v.winSize.Y - (viewOuterPadding+viewChartThumbSize.Y)*i - viewOuterPadding
		min.Y = max.Y - viewChartThumbSize.Y
		vc.Bounds = image.Rectangle{min, max}
		th.Render(vc)
	}

	// Call any callbacks scheduled by views.
	for _, cb := range vc.values.scheduledCallbacks {
		cb()
	}

	// Reset any flags for the next viewContext.
	v.nextViewContext.MouseLeftButtonClicked = false
}

func (v *View) newChart(st *ModelStock) *Chart {
	ch := NewChart(st)
	ch.AddAddButtonClickCallback(func() {
		if !v.model.AddStock(st) {
			return
		}
		v.addSidebarChartThumb(st)
		go func() {
			if err := v.saveConfig(); err != nil {
				glog.Warningf("addButtonClickCallback: failed to save config: %v", err)
			}
		}()
	})
	return ch
}

func (v *View) addSidebarChartThumb(st *ModelStock) {
	th := NewChartThumbnail(st)
	v.chartThumbs = append(v.chartThumbs, th)

	th.AddRemoveButtonClickCallback(func() {
		if !v.model.RemoveStock(st) {
			return
		}

		for i, thumb := range v.chartThumbs {
			if thumb == th {
				v.chartThumbs = append(v.chartThumbs[:i], v.chartThumbs[i+1:]...)
				break
			}
		}
		th.Close()

		go func() {
			if err := v.saveConfig(); err != nil {
				glog.Warningf("removeButtonClickCallback: failed to save config: %v", err)
			}
		}()
	})

	th.AddThumbClickCallback(func() {
		v.model.CurrentStock = st
	})
}

func (v *View) saveConfig() error {
	cfg := &Config{}
	for _, st := range v.model.Stocks {
		cfg.Stocks = append(cfg.Stocks, ConfigStock{st.symbol})
	}
	return SaveConfig(cfg)
}

// Resize responds to window size changes by updating internal matrices.
func (v *View) Resize(newSize image.Point) {
	// Return if the window has not changed size.
	if v.winSize == newSize {
		return
	}

	gl.Viewport(0, 0, int32(newSize.X), int32(newSize.Y))

	v.winSize = newSize

	// Calculate the new ortho projection view matrix.
	fw, fh := float32(v.winSize.X), float32(v.winSize.Y)
	v.orthoMatrix = math2.OrthoMatrix(fw, fh, fw /* use width as depth */)
}

// HandleKey is a callback registered with GLFW to receive key presses.
func (v *View) HandleKey(key glfw.Key, action glfw.Action) {
	if action != glfw.Release {
		return
	}

	switch key {
	case glfw.KeyEnter:
		v.submitSymbol()
		go func() {
			if err := v.model.Refresh(); err != nil {
				glog.Errorf("refresh failed: %v", err)
			}
		}()

	case glfw.KeyBackspace:
		v.popSymbolChar()
	}
}

// HandleChar is a callback registered with GLFW to receive character input.
func (v *View) HandleChar(ch rune) {
	ch = unicode.ToUpper(ch)
	if _, ok := acceptedChars[ch]; ok {
		v.pushSymbolChar(ch)
	}
}

func (v *View) pushSymbolChar(ch rune) {
	v.inputSymbol.Text += string(ch)
}

func (v *View) popSymbolChar() {
	if l := len(v.inputSymbol.Text); l > 0 {
		v.inputSymbol.Text = v.inputSymbol.Text[:l-1]
	}
}

func (v *View) submitSymbol() {
	v.model.Lock()
	defer v.model.Unlock()
	v.model.CurrentStock = newModelStock(v.inputSymbol.Text)
	v.inputSymbol.Text = ""
}

// HandleCursorPos is a callback registered with GLFW to track cursor movement.
func (v *View) HandleCursorPos(x, y float64) {
	// Flip Y-axis since the OpenGL coordinate system makes lower left the origin.
	v.nextViewContext.MousePos = image.Pt(int(x), v.winSize.Y-int(y))
}

// HandleMouseButton is a callback registered with GLFW to track mouse clicks.
func (v *View) HandleMouseButton(button glfw.MouseButton, action glfw.Action) {
	if button != glfw.MouseButtonLeft {
		glog.Infof("handleMouseButton: ignoring mouse button(%v) and action(%v)", button, action)
		return // Only interested in left clicks right now.
	}
	v.nextViewContext.MouseLeftButtonClicked = action == glfw.Release
}
