package gfx

import (
	"io"

	"github.com/golang/glog"

	"github.com/btmura/ponzi2/internal/ply"
)

// VAOColor is an RGB color with values from 0.0 to 1.0.
type VAOColor [3]float32

// HorizColoredLineVAO returns a horizontal colored line segment from (-1, 0) to (1, 0).
func HorizColoredLineVAO(leftColor, rightColor VAOColor) *VAO {
	return NewVAO(
		Lines,
		&VAOVertexData{
			Vertices: []float32{
				-1, 0, 0,
				+1, 0, 0,
			},
			Colors: []float32{
				leftColor[0], leftColor[1], leftColor[2],
				rightColor[0], rightColor[1], rightColor[2],
			},
			Indices: []uint16{
				0, 1,
			},
		},
	)
}

func newPLYVAO(r io.Reader) *VAO {
	p, err := ply.Decode(r)
	if err != nil {
		glog.Fatalf("gfx.newPLYVAO: decoding PLY failed: %v", err)
	}

	var vertices []float32
	var normals []float32
	var texCoords []float32
	var indices []uint16

	for _, e := range p.Elements["vertex"] {
		x := e.Floats["x"]
		y := e.Floats["y"]
		z := e.Floats["z"]
		vertices = append(vertices, x, y, z)

		nx := e.Floats["nx"]
		ny := e.Floats["ny"]
		nz := e.Floats["nz"]
		normals = append(normals, nx, ny, nz)

		s := e.Floats["s"]
		t := e.Floats["t"]
		texCoords = append(texCoords, s, 1-t)
	}

	for _, e := range p.Elements["face"] {
		for _, idx := range e.UintLists["vertex_indices"] {
			indices = append(indices, uint16(idx))
		}
	}

	return NewVAO(
		Triangles,
		&VAOVertexData{
			Vertices:  vertices,
			Normals:   normals,
			TexCoords: texCoords,
			Indices:   indices,
		},
	)
}
