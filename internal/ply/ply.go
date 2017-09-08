// Package ply provides a way to decode files in the Polygon File Format (PLY).
package ply

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/golang/glog"
)

func init() {
	flag.Parse() // Avoid glog errors about logging before flag.Parse.
}

// PLY has the elements parsed from a file in the Polygon File Format (PLY).
// More details: http://paulbourke.net/dataformats/ply/
type PLY struct {
	// Elements maps element name like "vertex" to the elements as they appear in the file.
	Elements map[string][]*Element
}

// Element is an element with property values.
type Element struct {
	// Float32s maps property name like "x" to a float32.
	Float32s map[string]float32

	// Int32s maps property name like "vertex1" to an int.
	Int32s map[string]int32

	// Uint8s maps property name like "red" to a uint8.
	Uint8s map[string]uint8

	// Uint32Lists maps property name like "vertex_indices" to a uint list.
	Uint32Lists map[string][]uint32
}

// header is used by Decode to parse the "element" and "property" lines.
type header struct {
	elementDescriptors []*elementDescriptor
}

// elementDescriptor describes an element parsed from a line starting with "element".
type elementDescriptor struct {
	name                string
	instances           int
	propertyDescriptors []*propertyDescriptor
}

// propertyDescriptor describes a property parsed from a line starting with "property".
type propertyDescriptor struct {
	name         string
	list         bool
	listSizeType string
	valueType    string
}

// Decode decodes the PLY file given as a io.Reader to a PLY struct or an error.
func Decode(r io.Reader) (*PLY, error) {
	sc := bufio.NewScanner(r)

	h := &header{}
	var ed *elementDescriptor // Current descriptor to append properties to.
processHeader:
	for sc.Scan() {
		line := sc.Text()
		switch { // Don't care about checking for ply and format lines.
		case strings.HasPrefix(line, "comment "):
			glog.Infof("ply.Decode: %s", line)

		case strings.HasPrefix(line, "element "):
			ed = &elementDescriptor{}
			h.elementDescriptors = append(h.elementDescriptors, ed)
			if _, err := fmt.Sscanf(line, "element %s %d", &ed.name, &ed.instances); err != nil {
				return nil, fmt.Errorf("ply: parsing element failed: %s", line)
			}

		case strings.HasPrefix(line, "property list "): // Keep above scalar property match below.
			pd := &propertyDescriptor{list: true}
			ed.propertyDescriptors = append(ed.propertyDescriptors, pd)
			if _, err := fmt.Sscanf(line, "property list %s %s %s", &pd.listSizeType, &pd.valueType, &pd.name); err != nil {
				return nil, fmt.Errorf("ply: parsing property list failed: %s", line)
			}

		case strings.HasPrefix(line, "property "):
			pd := &propertyDescriptor{}
			ed.propertyDescriptors = append(ed.propertyDescriptors, pd)
			if _, err := fmt.Sscanf(line, "property %s %s", &pd.valueType, &pd.name); err != nil {
				return nil, fmt.Errorf("ply: parsing property failed: %s", line)
			}

		case line == "end_header":
			break processHeader
		}
	}

	p := &PLY{}
	for _, ed := range h.elementDescriptors {
		for i := 0; i < ed.instances; i++ {
			if !sc.Scan() {
				return nil, fmt.Errorf("ply: expected %d rows for element %s", ed.instances, ed.name)
			}

			e := &Element{}
			if p.Elements == nil {
				p.Elements = map[string][]*Element{}
			}
			p.Elements[ed.name] = append(p.Elements[ed.name], e)

			pi := 0
			values := strings.Split(sc.Text(), " ")
			for i := 0; i < len(values); i++ {
				pd := ed.propertyDescriptors[pi]
				if pd.list {
					size, _ := strconv.Atoi(values[i])
					i++
					var ss []string
					for j := 0; j < size; j++ {
						ss = append(ss, values[i])
						i++
					}

					switch pd.valueType {
					case "uint":
						var us []uint32
						for _, s := range ss {
							u, err := strconv.ParseUint(s, 10, 32)
							if err != nil {
								return nil, err
							}
							us = append(us, uint32(u))
						}

						if e.Uint32Lists == nil {
							e.Uint32Lists = map[string][]uint32{}
						}
						e.Uint32Lists[pd.name] = us

					default:
						return nil, fmt.Errorf("ply: unsupported list value type: %s", pd.valueType)
					}
				} else {
					switch pd.valueType {
					case "float":
						f, err := strconv.ParseFloat(values[i], 32)
						if err != nil {
							return nil, err
						}

						if e.Float32s == nil {
							e.Float32s = map[string]float32{}
						}
						e.Float32s[pd.name] = float32(f)

					case "int":
						i, err := strconv.ParseInt(values[i], 10, 32)
						if err != nil {
							return nil, err
						}

						if e.Int32s == nil {
							e.Int32s = map[string]int32{}
						}
						e.Int32s[pd.name] = int32(i)

					case "uchar":
						u, err := strconv.ParseUint(values[i], 10, 8)
						if err != nil {
							return nil, err
						}

						if e.Uint8s == nil {
							e.Uint8s = map[string]uint8{}
						}
						e.Uint8s[pd.name] = uint8(u)

					default:
						return nil, fmt.Errorf("ply: unsupported scalar value type: %s", pd.valueType)
					}
				}
				pi++
			}
		}
	}

	return p, nil
}
