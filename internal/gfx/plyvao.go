package gfx

import (
	"io"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/golang/glog"

	gl2 "github.com/btmura/ponzi2/internal/gl"
	"github.com/btmura/ponzi2/internal/ply"
)

type plyVAO struct {
	vao   uint32 // array is the VAO name for gl.BindVertexArray.
	count int32  // count is the number of elements for gl.DrawElements.
}

func newPLYVAO(r io.Reader) *plyVAO {
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

	vbo := gl2.ArrayBuffer(vertices)
	nbo := gl2.ArrayBuffer(normals)
	tbo := gl2.ArrayBuffer(texCoords)
	ibo := gl2.ElementArrayBuffer(indices)

	var vao uint32

	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.EnableVertexAttribArray(positionLocation)
	gl.VertexAttribPointer(positionLocation, 3, gl.FLOAT, false, 0, gl.PtrOffset(0))

	gl.BindBuffer(gl.ARRAY_BUFFER, nbo)
	gl.EnableVertexAttribArray(normalLocation)
	gl.VertexAttribPointer(normalLocation, 3, gl.FLOAT, false, 0, gl.PtrOffset(0))

	gl.BindBuffer(gl.ARRAY_BUFFER, tbo)
	gl.EnableVertexAttribArray(texCoordLocation)
	gl.VertexAttribPointer(texCoordLocation, 2, gl.FLOAT, false, 0, gl.PtrOffset(0))

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ibo)
	gl.BindVertexArray(0)

	return &plyVAO{
		vao:   vao,
		count: int32(len(indices)),
	}
}

func (p *plyVAO) Render() {
	gl.BindVertexArray(p.vao)
	gl.DrawElements(gl.TRIANGLES, p.count, gl.UNSIGNED_SHORT, gl.Ptr(nil))
	gl.BindVertexArray(0)
}

func (p *plyVAO) Delete() {
	gl.DeleteVertexArrays(1, &p.vao)
}
