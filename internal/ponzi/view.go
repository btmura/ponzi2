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

var (
	inputSymbolTextRenderer      = gfx.NewTextRenderer(goregular.TTF, 48)
	symbolQuoteTextRenderer      = gfx.NewTextRenderer(goregular.TTF, 24)
	axisLabelTextRenderer        = gfx.NewTextRenderer(goregular.TTF, 12)
	thumbSymbolQuoteTextRenderer = gfx.NewTextRenderer(goregular.TTF, 14)
)

const (
	mainChartRounding = 10
	mainChartPadding  = 5
	mainOuterPadding  = 5

	thumbChartRounding = 6
	thumbChartPadding  = 2
)

type view struct {
	model             *model // model is the model that will be rendered.
	chart             *chart
	chartThumbnail    *chartThumbnail
	viewMatrix        math2.Matrix4
	perspectiveMatrix math2.Matrix4
	orthoMatrix       math2.Matrix4
	winSize           image.Point
}

func createView(model *model) (*view, error) {

	// Initialize OpenGL and enable features.

	if err := gl.Init(); err != nil {
		return nil, err
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
		return nil, err
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
	}, nil
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
	if v.model.inputSymbol != "" {
		s := inputSymbolTextRenderer.Measure(v.model.inputSymbol)
		c := v.winSize.Sub(s).Div(2)
		inputSymbolTextRenderer.Render(v.model.inputSymbol, c, white)
	}

	// Start in upper left. (0, 0) is lower left.
	pt := image.Pt(mainOuterPadding, v.winSize.Y-mainOuterPadding)

	// Render the current symbol.
	if v.chart == nil || v.chart.stock != v.model.currentStock {
		if v.chart != nil {
			v.chart.close()
		}
		v.chart = createChart(v.model.currentStock)
	}

	if v.chartThumbnail == nil || v.chartThumbnail.stock != v.model.currentStock {
		if v.chartThumbnail != nil {
			v.chartThumbnail.close()
		}
		v.chartThumbnail = createChartThumbnail(v.model.currentStock)
	}

	ms := image.Pt(150, 100)

	if v.chart != nil {
		v.chart.render(image.Rect(pt.X+ms.X+mainOuterPadding, mainOuterPadding, v.winSize.X-mainOuterPadding, pt.Y))
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
		v.model.submitSymbol()
		go func() {
			if err := v.model.refresh(); err != nil {
				glog.Errorf("refresh failed: %v", err)
			}
		}()

	case glfw.KeyBackspace:
		v.model.popSymbolChar()
	}
}

func (v *view) handleChar(ch rune) {
	ch = unicode.ToUpper(ch)
	if _, ok := acceptedChars[ch]; ok {
		v.model.pushSymbolChar(ch)
	}
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
