package ponzi

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"log"
	"math"

	"golang.org/x/image/font"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/golang/freetype/truetype"
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

type renderer struct {
	program uint32

	planeMesh *mesh
	cubeMesh  *mesh

	texture uint32

	loadingText *renderableText

	viewMatrix        matrix4
	perspectiveMatrix matrix4
	orthoMatrix       matrix4

	winWidth  int
	winHeight int
}

func createRenderer() (*renderer, error) {

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

	var planeMesh, cubeMesh *mesh
	for _, m := range createMeshes(objs) {
		switch m.id {
		case "Plane":
			planeMesh = m
		case "Cube":
			cubeMesh = m
		}
	}

	fontBytes, err := orbitronMediumTtfBytes()
	if err != nil {
		return nil, err
	}

	f, err := truetype.Parse(fontBytes)
	if err != nil {
		return nil, err
	}

	face := newFace(f)
	loadingText := createRenderableText(planeMesh, face, "Loading DATA...")

	return &renderer{
		program:     p,
		planeMesh:   planeMesh,
		cubeMesh:    cubeMesh,
		texture:     texture,
		loadingText: loadingText,
		viewMatrix:  vm,
	}, nil
}

func (r *renderer) render() {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UniformMatrix4fv(projectionViewMatrixLocation, 1, false, &r.perspectiveMatrix[0])

	mm := newScaleMatrix(1, 1, 1)
	gl.UniformMatrix4fv(modelMatrixLocation, 1, false, &mm[0])

	gl.BindTexture(gl.TEXTURE_2D, r.texture)
	r.cubeMesh.drawElements()

	gl.UniformMatrix4fv(projectionViewMatrixLocation, 1, false, &r.orthoMatrix[0])

	r.loadingText.render(float32(r.winWidth)/2, float32(r.winHeight)/2)
}

func (r *renderer) resize(width, height int) {
	// Return if the window has not changed size.
	if r.winWidth == width && r.winHeight == height {
		return
	}

	gl.Viewport(0, 0, int32(width), int32(height))

	r.winWidth, r.winHeight = width, height

	// Calculate the new perspective projection view matrix.
	fw, fh := float32(width), float32(height)
	aspect := fw / fh
	fovRadians := float32(math.Pi) / 3
	r.perspectiveMatrix = r.viewMatrix.mult(newPerspectiveMatrix(fovRadians, aspect, 1, 2000))

	// Calculate the new ortho projection view matrix.
	r.orthoMatrix = newOrthoMatrix(fw, fh, fw /* use width as depth */)
}

type renderableText struct {
	mesh    *mesh
	texture uint32
	width   float32
	height  float32
}

func createRenderableText(mesh *mesh, face font.Face, text string) *renderableText {
	rgba := createTextImage(face, text)
	return &renderableText{
		mesh:    mesh,
		texture: createTexture(rgba),
		width:   float32(rgba.Bounds().Size().X),
		height:  float32(rgba.Bounds().Size().Y),
	}
}

func (rt *renderableText) render(x, y float32) {
	m := newScaleMatrix(rt.width, rt.height, 1)
	m = m.mult(newTranslationMatrix(x, y, 0))
	gl.UniformMatrix4fv(modelMatrixLocation, 1, false, &m[0])
	gl.BindTexture(gl.TEXTURE_2D, rt.texture)
	rt.mesh.drawElements()
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
