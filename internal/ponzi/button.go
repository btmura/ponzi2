package ponzi

import (
	"bytes"
	"errors"
	"image"

	"github.com/go-gl/gl/v4.5-core/gl"

	"github.com/btmura/ponzi2/internal/gfx"
	gl2 "github.com/btmura/ponzi2/internal/gl"
	"github.com/btmura/ponzi2/internal/obj"
)

// buttonIcon is the icon to show on the button.
type buttonIcon int

// buttonIcon values.
const (
	addButtonIcon buttonIcon = iota
)

// buttonRenderer renders a button.
type buttonRenderer struct {
	mesh    *gfx.Mesh
	texture uint32
}

func createButtonRenderer() (*buttonRenderer, error) {
	objs, err := obj.Decode(bytes.NewReader(MustAsset("meshes.obj")))
	if err != nil {
		return nil, err
	}

	var bm *gfx.Mesh
	for _, m := range gfx.CreateMeshes(objs) {
		switch m.ID {
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
	gfx.SetModelMatrixOrtho(pt, sz)

	gl.BindTexture(gl.TEXTURE_2D, b.texture)
	gfx.SetColorMixAmount(0)

	b.mesh.DrawElements()
}
