package ponzi

import (
	"bytes"
	"errors"
	"image"

	"github.com/btmura/ponzi2/internal/gl2"
	"github.com/btmura/ponzi2/internal/obj"
	"github.com/go-gl/gl/v4.5-core/gl"
)

// buttonIcon is the icon to show on the button.
type buttonIcon int

// buttonIcon values.
const (
	addButtonIcon buttonIcon = iota
)

// buttonRenderer renders a button.
type buttonRenderer struct {
	mesh    *mesh
	texture uint32
}

func createButtonRenderer() (*buttonRenderer, error) {
	objs, err := obj.Decode(bytes.NewReader(MustAsset("meshes.obj")))
	if err != nil {
		return nil, err
	}

	var bm *mesh
	for _, m := range createMeshes(objs) {
		switch m.id {
		case "orthoPlane": // TODO(btmura): use separate mesh
			bm = m
		}
	}
	if bm == nil {
		return nil, errors.New("missing button mesh")
	}

	img, err := createImage(MustAsset("buttonTexture.png"))
	if err != nil {
		return nil, err
	}

	tex := gl2.CreateTexture(img)

	return &buttonRenderer{
		mesh:    bm,
		texture: tex,
	}, nil
}

func (b *buttonRenderer) render(pt, sz image.Point, ic buttonIcon) {
	m := newScaleMatrix(float32(sz.X), float32(sz.Y), 1)
	m = m.mult(newTranslationMatrix(float32(pt.X), float32(pt.Y), 0))
	gl.UniformMatrix4fv(modelMatrixLocation, 1, false, &m[0])

	gl.BindTexture(gl.TEXTURE_2D, b.texture)
	gl.Uniform1f(colorMixAmountLocation, 0)

	b.mesh.drawElements()
}
