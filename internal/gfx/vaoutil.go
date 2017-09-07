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

// VertColoredLineVAO returns a horizontal colored line segment from (0, -1) to (0, 1).
func VertColoredLineVAO(topColor, botColor VAOColor) *VAO {
	return NewVAO(
		Lines,
		&VAOVertexData{
			Vertices: []float32{
				0, -1, 0,
				0, +1, 0,
			},
			Colors: []float32{
				topColor[0], topColor[1], topColor[2],
				botColor[0], botColor[1], botColor[2],
			},
			Indices: []uint16{
				0, 1,
			},
		},
	)
}

// ReadPLYVAO returns a VAO decoded from PLY reader.
func ReadPLYVAO(r io.Reader) *VAO {
	p, err := ply.Decode(r)
	if err != nil {
		glog.Fatalf("gfx.ReadPLYVAO: decoding PLY failed: %v", err)
	}

	data := &VAOVertexData{}

	for _, e := range p.Elements["vertex"] {
		x := e.Float32s["x"]
		y := e.Float32s["y"]
		z := e.Float32s["z"]
		data.Vertices = append(data.Vertices, x, y, z)

		if nx, ok := e.Float32s["nx"]; ok {
			ny := e.Float32s["ny"]
			nz := e.Float32s["nz"]
			data.Normals = append(data.Normals, nx, ny, nz)
		}

		if s, ok := e.Float32s["s"]; ok {
			t := e.Float32s["t"]
			data.TexCoords = append(data.TexCoords, s, 1-t)
		}

		if r, ok := e.Uint8s["red"]; ok {
			g := e.Uint8s["green"]
			b := e.Uint8s["blue"]
			data.Colors = append(data.Colors, float32(r)/0xff, float32(g)/0xff, float32(b)/0xff)
		}
	}

	// TODO(btmura): fail if its not a triangle
	for _, e := range p.Elements["face"] {
		for _, idx := range e.Uint32Lists["vertex_indices"] {
			data.Indices = append(data.Indices, uint16(idx))
		}
	}

	// TODO(btmura): support multiple drawings
	if len(data.Indices) > 0 {
		return NewVAO(Triangles, data)
	}

	for _, e := range p.Elements["edge"] {
		v1 := e.Int32s["vertex1"]
		v2 := e.Int32s["vertex2"]
		data.Indices = append(data.Indices, uint16(v1), uint16(v2))
	}

	return NewVAO(Lines, data)
}
