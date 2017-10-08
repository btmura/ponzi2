package gfx

import (
	"bytes"
	"image"
	"image/draw"
	"io"

	"github.com/golang/glog"

	"github.com/btmura/ponzi2/internal/ply"
)

// VAOColor is an RGB color with values from 0.0 to 1.0.
type VAOColor [3]float32

// HorizColoredLineVAO returns a horizontal colored line segment
// from (-1, 0) to (1, 0).
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

// VertColoredLineVAO returns a horizontal colored line segment
// from (0, -1) to (0, 1).
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

// ReadPLYVAO returns a VAO decoded from a PLY reader.
func ReadPLYVAO(r io.Reader) *VAO {
	return readTexturedPLYVAO(r, nil)
}

// SquareImageVAO returns a VAO that renders a square image.
func SquareImageVAO(r io.Reader) *VAO {
	return readTexturedPLYVAO(bytes.NewReader(MustAsset("squarePlane.ply")), r)
}

// readTexturedPLYVAO returns a VAO decoded from PLY reader.
func readTexturedPLYVAO(r, textureReader io.Reader) *VAO {
	p, err := ply.Decode(r)
	if err != nil {
		glog.Fatalf("ReadPLYVAO: decoding PLY failed: %v", err)
	}

	data := &VAOVertexData{}

	for _, e := range p.Elements["vertex"] {
		x := e.Float32s["x"]
		y := e.Float32s["y"]
		z := e.Float32s["z"]
		data.Vertices = append(data.Vertices, x, y, z)

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

	var triangleIndices []uint16
	for _, e := range p.Elements["face"] {
		list := e.Uint32Lists["vertex_indices"]
		if len(list) != 3 {
			glog.Fatalf("ReadPLYVAO: index list has %d elements, want 3", len(list))
		}
		for _, idx := range list {
			triangleIndices = append(triangleIndices, uint16(idx))
		}
	}

	var lineIndices []uint16
	for _, e := range p.Elements["edge"] {
		v1 := e.Int32s["vertex1"]
		v2 := e.Int32s["vertex2"]
		lineIndices = append(lineIndices, uint16(v1), uint16(v2))
	}

	if textureReader != nil {
		img, _, err := image.Decode(textureReader)
		if err != nil {
			glog.Fatalf("ReadPLYVAO: decoding texture failed: %v", err)
		}

		rgba := image.NewRGBA(img.Bounds())
		draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)
		data.TextureRGBA = rgba
	}

	switch {
	case len(triangleIndices) > 0 && len(lineIndices) > 0:
		glog.Fatalf("ReadPLYVAO: both triangles and lines is unsupported")

	case len(triangleIndices) > 0:
		data.Indices = triangleIndices
		return NewVAO(Triangles, data)

	case len(lineIndices) > 0:
		data.Indices = lineIndices
		return NewVAO(Lines, data)
	}

	glog.Fatal("ReadPLYVAO: missing indices")
	return nil
}
