package ponzi

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"math"
	"unicode"

	"github.com/btmura/ponzi2/internal/obj"
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang/glog"
	"golang.org/x/image/font/gofont/goregular"
)

// Locations for the vertex and fragment shaders.
const (
	projectionViewMatrixLocation = iota
	modelMatrixLocation
	normalMatrixLocation

	ambientLightColorLocation
	directionalLightColorLocation
	directionalLightVectorLocation

	positionLocation
	normalLocation
	texCoordLocation
	colorLocation

	textureLocation
	colorMixAmountLocation
	textColorLocation
)

var (
	cameraPosition = vector3{0, 5, 10}
	targetPosition = vector3{}
	up             = vector3{0, 1, 0}

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

type view struct {
	// model is the model that will be rendered.
	model *model

	// program is the OpenGL program created by createProgram.
	program uint32

	// orthoPlaneMesh is a plane with bounds from (0, 0) to (1, 1)
	// which in convenient for positioning text.
	orthoPlaneMesh *mesh

	texture uint32

	dowText    *staticText
	sapText    *staticText
	nasdaqText *staticText

	chart          *chart
	chartThumbnail *chartThumbnail

	smallText       *dynamicText
	symbolQuoteText *dynamicText
	priceText       *dynamicText
	inputSymbolText *dynamicText

	buttonRenderer *buttonRenderer

	viewMatrix        matrix4
	perspectiveMatrix matrix4
	orthoMatrix       matrix4

	winSize image.Point
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

	// Create shaders and link them into a program.

	p, err := createProgram(string(MustAsset("shader.vert")), string(MustAsset("shader.frag")))
	if err != nil {
		return nil, err
	}

	gl.UseProgram(p)

	// Setup the vertex shader uniforms.

	mm := newScaleMatrix(1, 1, 1)
	gl.UniformMatrix4fv(modelMatrixLocation, 1, false, &mm[0])

	vm := newViewMatrix(cameraPosition, targetPosition, up)
	nm := vm.inverse().transpose()
	gl.UniformMatrix4fv(normalMatrixLocation, 1, false, &nm[0])

	gl.Uniform3fv(ambientLightColorLocation, 1, &ambientLightColor[0])
	gl.Uniform3fv(directionalLightColorLocation, 1, &directionalLightColor[0])
	gl.Uniform3fv(directionalLightVectorLocation, 1, &directionalVector[0])

	// Setup the fragment shader uniforms.

	textureImage, err := createImage(MustAsset("texture.png"))
	if err != nil {
		return nil, err
	}

	texture := createTexture(textureImage)

	// Load meshes and create vertex array objects.

	objs, err := obj.Decode(bytes.NewReader(MustAsset("meshes.obj")))
	if err != nil {
		return nil, err
	}

	var orthoPlaneMesh *mesh
	for _, m := range createMeshes(objs) {
		switch m.id {
		case "orthoPlane":
			orthoPlaneMesh = m
		}
	}

	small, err := newTextFactory(orthoPlaneMesh, goregular.TTF, 12)
	if err != nil {
		return nil, err
	}

	medium, err := newTextFactory(orthoPlaneMesh, goregular.TTF, 24)
	if err != nil {
		return nil, err
	}

	large, err := newTextFactory(orthoPlaneMesh, goregular.TTF, 48)
	if err != nil {
		return nil, err
	}

	smallDynamicText := small.createDynamicText()

	ir, err := createButtonRenderer()
	if err != nil {
		return nil, err
	}

	return &view{
		model:           model,
		program:         p,
		orthoPlaneMesh:  orthoPlaneMesh,
		texture:         texture,
		dowText:         small.createStaticText("DOW"),
		sapText:         small.createStaticText("S&P"),
		nasdaqText:      small.createStaticText("NASDAQ"),
		smallText:       smallDynamicText,
		symbolQuoteText: medium.createDynamicText(),
		priceText:       smallDynamicText,
		inputSymbolText: large.createDynamicText(),
		buttonRenderer:  ir,
		viewMatrix:      vm,
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
	gl.UniformMatrix4fv(projectionViewMatrixLocation, 1, false, &v.orthoMatrix[0])

	// Render input symbol being typed in the center.
	if v.model.inputSymbol != "" {
		s := v.inputSymbolText.measure(v.model.inputSymbol)
		c := v.winSize.Sub(s).Div(2)
		v.inputSymbolText.render(v.model.inputSymbol, c, white)
	}

	const p = 10 // padding

	// Start in theistforner. (0, 0) is lower left.
	// Move down below the major indices line.
	c := image.Pt(p, v.winSize.Y-p-v.dowText.size.Y)

	// Render major indices on one line.
	{
		render := func(c image.Point, q *modelQuote) image.Point {
			return v.smallText.render(formatQuote(q), c, quoteColor(q))
		}

		c := c
		c = c.Add(v.dowText.render(c, white))
		c = c.Add(render(c, v.model.dow))
		c.X += p

		c = c.Add(v.sapText.render(c, white))
		c = c.Add(render(c, v.model.sap))
		c.X += p

		c = c.Add(v.nasdaqText.render(c, white))
		c = c.Add(render(c, v.model.nasdaq))
		c.X += p
	}

	// Move down a bit after the major indices.
	c.Y -= p

	// Render the current symbol below the indices.
	if v.chart == nil || v.chart.stock != v.model.currentStock {
		if v.chart != nil {
			v.chart.close()
		}
		v.chart = createChart(v.model.currentStock, v.symbolQuoteText, v.priceText, v.buttonRenderer)
	}

	if v.chartThumbnail == nil || v.chartThumbnail.stock != v.model.currentStock {
		if v.chartThumbnail != nil {
			v.chartThumbnail.close()
		}
		v.chartThumbnail = createchartThumbnail(v.model.currentStock, v.priceText)
	}

	ms := image.Pt(150, 100)

	if v.chart != nil {
		v.chart.render(image.Rect(c.X+ms.X+p, p, v.winSize.X-p, c.Y))
	}

	if v.chartThumbnail != nil {
		v.chartThumbnail.render(image.Rect(c.X, c.Y-ms.Y, c.X+ms.X, c.Y))
	}
}

func formatQuote(q *modelQuote) string {
	if q.price != 0 {
		return fmt.Sprintf(" %.2f %+5.2f %+5.2f%% ", q.price, q.change, q.percentChange*100.0)
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
			v.model.refresh()
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
	v.perspectiveMatrix = v.viewMatrix.mult(newPerspectiveMatrix(fovRadians, aspect, 1, 2000))

	// Calculate the new ortho projection view matrix.
	v.orthoMatrix = newOrthoMatrix(fw, fh, fw /* use width as depth */)
}

func createImage(data []byte) (*image.RGBA, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("createImage: %v", err)
	}

	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)
	return rgba, nil
}

func setModelMatrixRectangle(r image.Rectangle) {
	m := newScaleMatrix(float32(r.Dx()/2), float32(r.Dy()/2), 1)
	m = m.mult(newTranslationMatrix(float32(r.Min.X+r.Dx()/2), float32(r.Min.Y+r.Dy()/2), 0))
	gl.UniformMatrix4fv(modelMatrixLocation, 1, false, &m[0])
}

func sliceRectangle(r image.Rectangle, percentages ...float32) []image.Rectangle {
	var rects []image.Rectangle
	y := r.Min.Y
	for _, p := range percentages {
		dy := int(float32(r.Dy()) * p)
		rects = append(rects, image.Rect(r.Min.X, y, r.Max.X, y+dy))
		y += dy
	}
	return rects
}
