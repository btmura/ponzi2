package ponzi

import (
	"image"
	"math"
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

	ambientLightColor     = [3]float32{1, 1, 1}
	directionalLightColor = [3]float32{1, 1, 1}
	directionalVector     = [3]float32{1, 1, 1}
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

	// inputSymbol is the symbol being entered by the user.
	inputSymbol string

	// nextViewContext is the next viewContext to pass down the view hierarchy.
	nextViewContext ViewContext

	// winSize is the current window's size used to measure and draw the UI.
	winSize image.Point

	viewMatrix        math2.Matrix4
	perspectiveMatrix math2.Matrix4
	orthoMatrix       math2.Matrix4
}

// ViewContext is passed down the view hierarchy providing drawing hints and event information.
// Meant to be passed around like a Rectangle or Point rather than a pointer to avoid mistakes.
type ViewContext struct {
	// bounds is the rectangle with global coordinates that the view part should draw within.
	bounds image.Rectangle

	// mousePos is the current global mouse position.
	mousePos image.Point

	// mouseLeftButtonClicked is whether the left mouse button was clicked.
	mouseLeftButtonClicked bool

	// scheduleCallback is a function to gather callbacks to be executed.
	scheduleCallbacks func(cbs []func())
}

// LeftClickInBounds returns true if the left mouse button was clicked within the context's bounds.
// Doesn't take into account overlapping view parts.
func (vc ViewContext) LeftClickInBounds() bool {
	return vc.mouseLeftButtonClicked && vc.mousePos.In(vc.bounds)
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

	// Setup the vertex shader uniforms.

	gfx.SetModelMatrix(math2.ScaleMatrix(1, 1, 1))

	vm := math2.ViewMatrix(cameraPosition, targetPosition, up)
	gfx.SetNormalMatrix(vm.Inverse().Transpose())

	gfx.SetAmbientLightColor(ambientLightColor)
	gfx.SetDirectionalLightColor(directionalLightColor)
	gfx.SetDirectionalLightVector(directionalVector)

	v := &View{
		model:      model,
		viewMatrix: vm,
	}

	if model.CurrentStock != nil {
		v.chart = v.newChart(model.CurrentStock)
	}

	for _, st := range model.Sidebar.Stocks {
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

	var callbacks []func()
	vc.scheduleCallbacks = func(cbs []func()) {
		callbacks = append(callbacks, cbs...)
	}

	// Render input symbol being typed in the center.
	if v.inputSymbol != "" {
		s := inputSymbolTextRenderer.Measure(v.inputSymbol)
		pt := v.winSize.Sub(s).Div(2)
		if len(v.model.Sidebar.Stocks) > 0 {
			pt.X += viewChartThumbSize.X / 2
		}
		inputSymbolTextRenderer.Render(v.inputSymbol, pt, white)
	}

	// Render the main chart.
	if v.chart != nil {
		vc.bounds = image.Rectangle{image.ZP, v.winSize}.Inset(viewOuterPadding)
		if len(v.model.Sidebar.Stocks) > 0 {
			vc.bounds.Min.X += viewOuterPadding + viewChartThumbSize.X
		}
		v.chart.Render(vc)
	}

	// Render the sidebar thumbnails.
	for i, th := range v.chartThumbs {
		min := image.Pt(viewOuterPadding, 0)
		max := image.Pt(min.X+viewChartThumbSize.X, 0)
		max.Y = v.winSize.Y - (viewOuterPadding+viewChartThumbSize.Y)*i - viewOuterPadding
		min.Y = max.Y - viewChartThumbSize.Y
		vc.bounds = image.Rectangle{min, max}
		th.Render(vc)
	}

	// Call any callbacks scheduled by views.
	for _, cb := range callbacks {
		cb()
	}

	// Reset any flags for the next viewContext.
	v.nextViewContext.mouseLeftButtonClicked = false
}

func (v *View) newChart(st *ModelStock) *Chart {
	ch := NewChart(st)
	ch.AddAddButtonClickCallback(func() {
		if !v.model.Sidebar.AddStock(st) {
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
		if !v.model.Sidebar.RemoveStock(st) {
			return
		}

		var newThumbs []*ChartThumbnail
		for _, thumb := range v.chartThumbs {
			if thumb == th {
				continue
			}
			newThumbs = append(newThumbs, thumb)
		}
		th.Close()
		v.chartThumbs = newThumbs

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
	for _, st := range v.model.Sidebar.Stocks {
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

	// Calculate the new perspective projection view matrix.
	fw, fh := float32(v.winSize.X), float32(v.winSize.Y)
	aspect := fw / fh
	fovRadians := float32(math.Pi) / 3
	v.perspectiveMatrix = v.viewMatrix.Mult(math2.PerspectiveMatrix(fovRadians, aspect, 1, 2000))

	// Calculate the new ortho projection view matrix.
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
	v.inputSymbol += string(ch)
}

func (v *View) popSymbolChar() {
	if l := len(v.inputSymbol); l > 0 {
		v.inputSymbol = v.inputSymbol[:l-1]
	}
}

func (v *View) submitSymbol() {
	v.model.Lock()
	defer v.model.Unlock()
	v.model.CurrentStock = newModelStock(v.inputSymbol)
	v.inputSymbol = ""
}

// HandleCursorPos is a callback registered with GLFW to track cursor movement.
func (v *View) HandleCursorPos(x, y float64) {
	// Flip Y-axis since the OpenGL coordinate system makes lower left the origin.
	v.nextViewContext.mousePos = image.Pt(int(x), v.winSize.Y-int(y))
}

// HandleMouseButton is a callback registered with GLFW to track mouse clicks.
func (v *View) HandleMouseButton(button glfw.MouseButton, action glfw.Action) {
	if button != glfw.MouseButtonLeft {
		glog.Infof("handleMouseButton: ignoring mouse button(%v) and action(%v)", button, action)
		return // Only interested in left clicks right now.
	}
	v.nextViewContext.mouseLeftButtonClicked = action == glfw.Release
}
