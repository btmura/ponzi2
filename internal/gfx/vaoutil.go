package gfx

import (
	"io"

	"github.com/golang/glog"

	"github.com/btmura/ponzi2/internal/ply"
)

func newPLYVAO(r io.Reader) *VAO2 {
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
		&VAOBufferData{
			Vertices:  vertices,
			Normals:   normals,
			TexCoords: texCoords,
			Indices:   indices,
		},
	)
}
