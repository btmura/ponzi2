package ponzi

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"log"
	"math"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
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

	textureLocation
)

var (
	cameraPosition = vector3{0, 5, 10}
	targetPosition = vector3{}
	up             = vector3{0, 1, 0}

	ambientLightColor     = [3]float32{0.5, 0.5, 0.5}
	directionalLightColor = [3]float32{0.5, 0.5, 0.5}
	directionalVector     = [3]float32{0.5, 0.5, 0.5}
)

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
	monoText   *dynamicText

	symbolText *staticText

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
	log.Printf("OpenGL version: %s", gl.GoStr(gl.GetString(gl.VERSION)))

	gl.Enable(gl.CULL_FACE)
	gl.CullFace(gl.BACK)

	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.ClearColor(0, 0, 0, 0)

	// Create shaders and link them into a program.

	vs, err := shaderVertBytes()
	if err != nil {
		return nil, err
	}

	fs, err := shaderFragBytes()
	if err != nil {
		return nil, err
	}

	p, err := createProgram(string(vs), string(fs))
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

	textureBytes, err := texturePngBytes()
	if err != nil {
		return nil, err
	}

	textureImage, err := createImage(textureBytes)
	if err != nil {
		return nil, err
	}

	texture := createTexture(textureImage)

	// Load meshes and create vertex array objects.

	objBytes, err := meshesObjBytes()
	if err != nil {
		return nil, err
	}

	objs, err := decodeObjs(bytes.NewReader(objBytes))
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

	monoBytes, err := inconsolataRegularTtfBytes()
	if err != nil {
		return nil, err
	}

	mono, err := newTextFactory(orthoPlaneMesh, monoBytes, 18)
	if err != nil {
		return nil, err
	}

	orbitronBytes, err := orbitronMediumTtfBytes()
	if err != nil {
		return nil, err
	}

	orbitron, err := newTextFactory(orthoPlaneMesh, orbitronBytes, 24)
	if err != nil {
		return nil, err
	}

	return &view{
		model:          model,
		program:        p,
		orthoPlaneMesh: orthoPlaneMesh,
		texture:        texture,
		dowText:        mono.createStaticText("DOW"),
		sapText:        mono.createStaticText("S&P"),
		nasdaqText:     mono.createStaticText("NASDAQ"),
		monoText:       mono.createDynamicText(),
		symbolText:     orbitron.createStaticText("SPY"),
		viewMatrix:     vm,
	}, nil
}

func (v *view) render() {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UniformMatrix4fv(projectionViewMatrixLocation, 1, false, &v.orthoMatrix[0])

	p := 10 // padding

	// Start in the upper left corner. (0, 0) is bottom left.
	c := image.Pt(p, v.winSize.Y-p-v.dowText.size.Y)

	// Render major indices on one line.
	{
		c := c
		c = c.Add(v.dowText.render(c))
		c = c.Add(v.monoText.render(v.dowPriceText(), c))

		c = c.Add(v.sapText.render(c))
		c = c.Add(v.monoText.render(v.sapPriceText(), c))

		c = c.Add(v.nasdaqText.render(c))
		c = c.Add(v.monoText.render(v.nasdaqPriceText(), c))
	}

	c.Y -= p + v.symbolText.size.Y
	v.symbolText.render(c)
}

func (v *view) dowPriceText() string {
	return formatQuote(v.model.dow)
}

func (v *view) sapPriceText() string {
	return formatQuote(v.model.sap)
}

func (v *view) nasdaqPriceText() string {
	return formatQuote(v.model.nasdaq)
}

func formatQuote(q *quote) string {
	if q != nil {
		return fmt.Sprintf(" %.2f %+5.2f %+5.2f%% ", q.price, q.change, q.percentChange*100.0)
	}
	return "..."
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

func (v *view) handleKey(key glfw.Key, action glfw.Action) {
	log.Printf("key: %v", key)
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
