package ponzi

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type obj struct {
	id        string
	vertices  []*objVertex
	texCoords []*objTexCoord
	normals   []*objNormal
	faces     []*objFace
}

type objVertex struct {
	x float32
	y float32
	z float32
}

type objTexCoord struct {
	s float32
	t float32
}

type objNormal struct {
	x float32
	y float32
	z float32
}

// numFaceElements is the number of required face elements. Only triangles are supported.
const numFaceElements = 3

// objFace is a face described by ObjFaceElements.
type objFace [numFaceElements]objFaceElement

// objFaceElement describes one point of a face.
type objFaceElement struct {
	// vertexIndex specifies a required vertex by global index starting from 1.
	vertexIndex int

	// texCoordIndex specifies an optional texture coordinate by global index starting from 1.
	// It is 0 if no texture coordinate was specified.
	texCoordIndex int

	// normalIndex specifies an optional normal by global index starting from 1.
	normalIndex int
}

func decodeObjs(r io.Reader) ([]*obj, error) {
	var allObjs []*obj
	var currentObj *obj

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
			currentObj.vertices = append(currentObj.vertices, v)

		case strings.HasPrefix(line, "vt "):
			tc, err := decodeObjTexCoord(line)
			if err != nil {
				return nil, err
			}
			if currentObj == nil {
				return nil, errors.New("missing object ID")
			}
			currentObj.texCoords = append(currentObj.texCoords, tc)

		case strings.HasPrefix(line, "vn "):
			n, err := decodeObjNormal(line)
			if err != nil {
				return nil, err
			}
			if currentObj == nil {
				return nil, errors.New("missing object ID")
			}
			currentObj.normals = append(currentObj.normals, n)

		case strings.HasPrefix(line, "f "):
			f, err := decodeObjFace(line)
			if err != nil {
				return nil, err
			}
			if currentObj == nil {
				return nil, errors.New("missing object ID")
			}
			currentObj.faces = append(currentObj.faces, f)
		}
	}

	return allObjs, nil
}

func decodeObjObject(line string) (*obj, error) {
	o := &obj{}
	if _, err := fmt.Sscanf(line, "o %s", &o.id); err != nil {
		return nil, err
	}
	return o, nil
}

func decodeObjVertex(line string) (*objVertex, error) {
	v := &objVertex{}
	if _, err := fmt.Sscanf(line, "v %f %f %f", &v.x, &v.y, &v.z); err != nil {
		return nil, err
	}
	return v, nil
}

func decodeObjTexCoord(line string) (*objTexCoord, error) {
	tc := &objTexCoord{}
	if _, err := fmt.Sscanf(line, "vt %f %f", &tc.s, &tc.t); err != nil {
		return nil, err
	}
	return tc, nil
}

func decodeObjNormal(line string) (*objNormal, error) {
	n := &objNormal{}
	if _, err := fmt.Sscanf(line, "vn %f %f %f", &n.x, &n.y, &n.z); err != nil {
		return nil, err
	}
	return n, nil
}

func decodeObjFace(line string) (*objFace, error) {
	f := &objFace{}

	var specs [numFaceElements]string
	if _, err := fmt.Sscanf(line, "f %s %s %s", &specs[0], &specs[1], &specs[2]); err != nil {
		return nil, err
	}

	var err error
	makeElement := func(spec string) (objFaceElement, error) {
		tokens := strings.Split(spec, "/")
		if len(tokens) == 0 {
			return objFaceElement{}, errors.New("face has no elements")
		}

		e := objFaceElement{}

		e.vertexIndex, err = strconv.Atoi(tokens[0])
		if err != nil {
			return objFaceElement{}, err
		}

		if len(tokens) < 2 {
			return e, nil
		}

		e.texCoordIndex, err = strconv.Atoi(tokens[1])
		if err != nil {
			return objFaceElement{}, err
		}

		if len(tokens) < 3 {
			return e, nil
		}

		e.normalIndex, err = strconv.Atoi(tokens[2])
		if err != nil {
			return objFaceElement{}, err
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
