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

const outerPadding = 5

type view struct {
	model             *model // model is the model that will be rendered.
	inputSymbol       string // inputSymbol is the symbol being entered by the user.
	chart             *chart
	chartThumbnail    *chartThumbnail
	viewMatrix        math2.Matrix4
	perspectiveMatrix math2.Matrix4
	orthoMatrix       math2.Matrix4
	winSize           image.Point
}

func newView(model *model) *view {

	// Initialize OpenGL and enable features.

	if err := gl.Init(); err != nil {
		glog.Fatalf("ponzi.newView: failed to init OpenGL: %v", err)
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
		glog.Fatalf("ponzi.newView: failed to init gfx: %v", err)
	}

	// Setup the vertex shader uniforms.

	gfx.SetModelMatrix(math2.ScaleMatrix(1, 1, 1))

	vm := math2.ViewMatrix(cameraPosition, targetPosition, up)
	gfx.SetNormalMatrix(vm.Inverse().Transpose())

	gfx.SetAmbientLightColor(ambientLightColor)
	gfx.SetDirectionalLightColor(directionalLightColor)
	gfx.SetDirectionalLightVector(directionalVector)

	return &view{
		model:      model,
		viewMatrix: vm,
	}
}

func (v *view) update() {
	v.model.Lock()
	defer v.model.Unlock()

	if v.chart != nil {
		v.chart.update()
	}
	if v.chartThumbnail != nil {
		v.chartThumbnail.update()
	}
}

func (v *view) render(fudge float32) {
	v.model.Lock()
	defer v.model.Unlock()

	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gfx.SetProjectionViewMatrix(v.orthoMatrix)

	// Render input symbol being typed in the center.
	if v.inputSymbol != "" {
		s := inputSymbolTextRenderer.Measure(v.inputSymbol)
		c := v.winSize.Sub(s).Div(2)
		inputSymbolTextRenderer.Render(v.inputSymbol, c, white)
	}

	// Start in upper left. (0, 0) is lower left.
	pt := image.Pt(outerPadding, v.winSize.Y-outerPadding)

	// Render the current symbol.
	if v.chart == nil || v.chart.stock != v.model.currentStock {
		if v.chart != nil {
			v.chart.close()
		}
		v.chart = newChart(v.model.currentStock)
	}

	if v.chartThumbnail == nil || v.chartThumbnail.stock != v.model.currentStock {
		if v.chartThumbnail != nil {
			v.chartThumbnail.close()
		}
		v.chartThumbnail = newChartThumbnail(v.model.currentStock)
	}

	ms := image.Pt(140, 90)

	if v.chart != nil {
		v.chart.render(image.Rect(pt.X+ms.X+outerPadding, outerPadding, v.winSize.X-outerPadding, pt.Y))
	}

	if v.chartThumbnail != nil {
		v.chartThumbnail.render(image.Rect(pt.X, pt.Y-ms.Y, pt.X+ms.X, pt.Y))
	}
}

func (v *view) handleKey(key glfw.Key, action glfw.Action) {
	if action != glfw.Release {
		return
	}

	switch key {
	case glfw.KeyEnter:
		v.submitSymbol()
		go func() {
			if err := v.model.refresh(); err != nil {
				glog.Errorf("refresh failed: %v", err)
			}
		}()

	case glfw.KeyBackspace:
		v.popSymbolChar()
	}
}

func (v *view) handleChar(ch rune) {
	ch = unicode.ToUpper(ch)
	if _, ok := acceptedChars[ch]; ok {
		v.pushSymbolChar(ch)
	}
}

func (v *view) handleMouseButton(button glfw.MouseButton, action glfw.Action) {
	glog.Infof("view.handleMouseButton: button: %v action: %v", button, action)
}

// resize responds to window size changes by updating internal matrices.
func (v *view) resize(newSize image.Point) {
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

func (v *view) pushSymbolChar(ch rune) {
	v.inputSymbol += string(ch)
}

func (v *view) popSymbolChar() {
	if l := len(v.inputSymbol); l > 0 {
		v.inputSymbol = v.inputSymbol[:l-1]
	}
}

func (v *view) submitSymbol() {
	v.model.Lock()
	defer v.model.Unlock()
	v.model.currentStock = newModelStock(v.inputSymbol)
	v.inputSymbol = ""
}
