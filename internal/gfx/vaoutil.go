package gfx

import (
	"image"
	"image/draw"
	"io"

	"github.com/golang/glog"

	"github.com/btmura/ponzi2/internal/ply"
)

// PLYVAO returns a VAO decoded from a PLY Reader.
func PLYVAO(plyReader io.Reader) *VAO {
	return NewVAOLoadData(func() *VAOVertexData {
		return readTexturedPLYVAO(plyReader, nil)
	})
}

// TexturedPLYVAO returns a VAO decoded from a PLY Reader
// and a texture from another Reader.
func TexturedPLYVAO(plyReader, texturedReader io.Reader) *VAO {
	return NewVAOLoadData(func() *VAOVertexData {
		return readTexturedPLYVAO(plyReader, texturedReader)
	})
}

// readTexturedPLYVAO returns a VAO decoded from PLY reader.
func readTexturedPLYVAO(r, textureReader io.Reader) *VAOVertexData {
	p, err := ply.Decode(r)
	if err != nil {
		glog.Fatalf("readTexturedPLYVAO: decoding PLY failed: %v", err)
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
			glog.Fatalf("readTexturedPLYVAO: index list has %d elements, want 3", len(list))
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
			glog.Fatalf("readTexturedPLYVAO: decoding texture failed: %v", err)
		}

		rgba := image.NewRGBA(img.Bounds())
		draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)
		data.TextureRGBA = rgba
	}

	switch {
	case len(triangleIndices) > 0 && len(lineIndices) > 0:
		glog.Fatalf("readTexturedPLYVAO: both triangles and lines is unsupported")

	case len(triangleIndices) > 0:
		data.Mode = Triangles
		data.Indices = triangleIndices
		return data

	case len(lineIndices) > 0:
		data.Mode = Lines
		data.Indices = lineIndices
		return data
	}

	glog.Info("readTexturedPLYVAO: missing indices")
	return nil
}
