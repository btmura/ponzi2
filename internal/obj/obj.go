// Package obj provides a way to decode OBJ files.
package obj

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Object is an object from an OBJ file.
type Object struct {
	ID        string
	Vertices  []*Vertex
	TexCoords []*TexCoord
	Normals   []*Normal
	Faces     []*Face
}

// Vertex is a vertex parsed from a line starting with "v".
type Vertex struct {
	X float32
	Y float32
	Z float32
}

// TexCoord is a texture coordinate parsed from a line starting with "vt".
type TexCoord struct {
	S float32
	T float32
}

// Normal is a normal parsed from a line starting with "vn".
type Normal struct {
	X float32
	Y float32
	Z float32
}

// numFaceElements is the number of required face elements. Only triangles are supported.
const numFaceElements = 3

// Face is a face described by ObjFaceElements.
type Face [numFaceElements]FaceElement

// FaceElement describes one point of a face.
type FaceElement struct {
	// VertexIndex specifies a required vertex by global index starting from 1.
	VertexIndex int

	// TexCoordIndex specifies an optional texture coordinate by global index starting from 1.
	// It is 0 if no texture coordinate was specified.
	TexCoordIndex int

	// NormalIndex specifies an optional normal by global index starting from 1.
	NormalIndex int
}

// Decode returns a slice of Objs decoded from a Reader.
func Decode(r io.Reader) ([]*Object, error) {
	var allObjs []*Object
	var currentObj *Object

	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		switch {
		case strings.HasPrefix(line, "o "):
			o, err := decodeObjObject(line)
			if err != nil {
				return nil, err
			}
			currentObj = o
			allObjs = append(allObjs, o)

		case strings.HasPrefix(line, "v "):
			v, err := decodeObjVertex(line)
			if err != nil {
				return nil, err
			}
			if currentObj == nil {
				return nil, errors.New("missing object ID")
			}
			currentObj.Vertices = append(currentObj.Vertices, v)

		case strings.HasPrefix(line, "vt "):
			tc, err := decodeObjTexCoord(line)
			if err != nil {
				return nil, err
			}
			if currentObj == nil {
				return nil, errors.New("missing object ID")
			}
			currentObj.TexCoords = append(currentObj.TexCoords, tc)

		case strings.HasPrefix(line, "vn "):
			n, err := decodeObjNormal(line)
			if err != nil {
				return nil, err
			}
			if currentObj == nil {
				return nil, errors.New("missing object ID")
			}
			currentObj.Normals = append(currentObj.Normals, n)

		case strings.HasPrefix(line, "f "):
			f, err := decodeObjFace(line)
			if err != nil {
				return nil, err
			}
			if currentObj == nil {
				return nil, errors.New("missing object ID")
			}
			currentObj.Faces = append(currentObj.Faces, f)
		}
	}

	return allObjs, nil
}

func decodeObjObject(line string) (*Object, error) {
	o := &Object{}
	if _, err := fmt.Sscanf(line, "o %s", &o.ID); err != nil {
		return nil, err
	}
	return o, nil
}

func decodeObjVertex(line string) (*Vertex, error) {
	v := &Vertex{}
	if _, err := fmt.Sscanf(line, "v %f %f %f", &v.X, &v.Y, &v.Z); err != nil {
		return nil, err
	}
	return v, nil
}

func decodeObjTexCoord(line string) (*TexCoord, error) {
	tc := &TexCoord{}
	if _, err := fmt.Sscanf(line, "vt %f %f", &tc.S, &tc.T); err != nil {
		return nil, err
	}
	return tc, nil
}

func decodeObjNormal(line string) (*Normal, error) {
	n := &Normal{}
	if _, err := fmt.Sscanf(line, "vn %f %f %f", &n.X, &n.Y, &n.Z); err != nil {
		return nil, err
	}
	return n, nil
}

func decodeObjFace(line string) (*Face, error) {
	f := &Face{}

	var specs [numFaceElements]string
	if _, err := fmt.Sscanf(line, "f %s %s %s", &specs[0], &specs[1], &specs[2]); err != nil {
		return nil, err
	}

	var err error
	makeElement := func(spec string) (FaceElement, error) {
		tokens := strings.Split(spec, "/")
		if len(tokens) == 0 {
			return FaceElement{}, errors.New("face has no elements")
		}

		e := FaceElement{}

		e.VertexIndex, err = strconv.Atoi(tokens[0])
		if err != nil {
			return FaceElement{}, err
		}

		if len(tokens) < 2 {
			return e, nil
		}

		e.TexCoordIndex, err = strconv.Atoi(tokens[1])
		if err != nil {
			return FaceElement{}, err
		}

		if len(tokens) < 3 {
			return e, nil
		}

		e.NormalIndex, err = strconv.Atoi(tokens[2])
		if err != nil {
			return FaceElement{}, err
		}

		return e, nil
	}

	for i, s := range specs {
		if f[i], err = makeElement(s); err != nil {
			return nil, err
		}
	}

	return f, nil
}
