// Package ply provides a way to decode files in the Polygon File Format (PLY).
package ply

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/golang/glog"
)

// PLY has the elements parsed from a file in the Polygon File Format (PLY).
// More details: http://paulbourke.net/dataformats/ply/
type PLY struct {
	// Elements maps element name like "vertex" to the elements as they appear in the file.
	Elements map[string][]*Element
}

// Element is an element with property values.
type Element struct {
	// Floats maps property name like "x" to a float32.
	Floats map[string]float32

	// UintLists maps property name like "vertex_indices" to a uint list.
	UintLists map[string][]uint
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
		switch { // Don't care about checking for ply, format, or comment lines.
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

		default:
			glog.Infof("skipping: %q", line)
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
						var us []uint
						for _, s := range ss {
							i, err := strconv.Atoi(s)
							if err != nil {
								return nil, err
							}
							us = append(us, uint(i))
						}

						if e.UintLists == nil {
							e.UintLists = map[string][]uint{}
						}
						e.UintLists[pd.name] = us

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

						if e.Floats == nil {
							e.Floats = map[string]float32{}
						}
						e.Floats[pd.name] = float32(f)

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
