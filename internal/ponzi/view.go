package ponzi

import (
	"fmt"
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
	majorIndexTextRenderer       = gfx.NewTextRenderer(goregular.TTF, 12)
	symbolQuoteTextRenderer      = gfx.NewTextRenderer(goregular.TTF, 24)
	axisLabelTextRenderer        = majorIndexTextRenderer
	thumbSymbolQuoteTextRenderer = majorIndexTextRenderer
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

	const pad = 5 // padding

	// Start in upper left. (0, 0) is lower left.
	// Move down below the major indices line.
	pt := image.Pt(pad, v.winSize.Y-pad-majorIndexTextRenderer.LineHeight())

	// Render major indices on one line.
	{
		pt := pt
		render := func(index string, q *modelQuote) {
			pt.X += majorIndexTextRenderer.Render(index, pt, white)
			pt.X += pad
			pt.X += majorIndexTextRenderer.Render(formatQuote(q), pt, quoteColor(q))
			pt.X += pad
		}

		render("DOW", v.model.dow)
		render("S&P", v.model.sap)
		render("NASDAQ", v.model.nasdaq)
	}

	pt.Y -= pad

	// Render the current symbol below the indices.
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
		v.chart.render(image.Rect(pt.X+ms.X+pad, pad, v.winSize.X-pad, pt.Y))
	}

	if v.chartThumbnail != nil {
		v.chartThumbnail.render(image.Rect(pt.X, pt.Y-ms.Y, pt.X+ms.X, pt.Y))
	}
}

func formatQuote(q *modelQuote) string {
	if q.price != 0 {
		return fmt.Sprintf("%.2f %+5.2f %+5.2f%%", q.price, q.change, q.percentChange*100.0)
	}
	return ""
}

func shortFormatQuote(q *modelQuote) string {
	if q.price != 0 {
		return fmt.Sprintf(" %.2f %+5.2f%% ", q.price, q.percentChange*100.0)
	}
	return ""
}

func quoteColor(q *modelQuote) [3]float32 {
	switch {
	case q.percentChange > 0:
		return green

	case q.percentChange < 0:
		return red
	}
	return white
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

// sliceRectangle horizontally cuts a rectangle from the bottom at the given percentages.
// It returns n+1 rectangles given n percentages.
func sliceRectangle(r image.Rectangle, percentages ...float32) []image.Rectangle {
	var rs []image.Rectangle
	addRect := func(minY, maxY int) {
		rs = append(rs, image.Rect(r.Min.X, minY, r.Max.X, maxY))
	}

	ry := r.Dy()  // Remaining Y to distribute.
	cy := r.Min.Y // Start at the bottom and cut horizontally up.
	for _, p := range percentages {
		dy := int(float32(r.Dy()) * p)
		addRect(cy, cy+dy)
		cy += dy // Bump upwards.
		ry -= dy // Subtract from remaining.
	}
	addRect(cy, cy+ry) // Use remaining Y for last rect.

	return rs
}
