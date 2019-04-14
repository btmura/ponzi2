package ply

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDecode(t *testing.T) {
	for _, tt := range []struct {
		desc    string
		input   string
		want    *PLY
		wantErr bool
	}{
		{
			desc: "vertices and faces",
			input: `ply
format ascii 1.0
comment Created by Blender 2.78 (sub 0) - www.blender.org, source file: ''
element vertex 4
property float x
property float y
property float z
property float nx
property float ny
property float nz
element face 1
property list uchar uint vertex_indices
end_header
-1.000000 0.000000 0.000000 0.000000 0.000000 1.000000
0.000000 0.000000 0.000000 0.000000 0.000000 1.000000
0.000000 1.000000 0.000000 0.000000 0.000000 1.000000
-1.000000 1.000000 0.000000 0.000000 0.000000 1.000000
4 0 1 2 3`,
			want: &PLY{
				Elements: map[string][]*Element{
					"vertex": {
						{
							Float32s: map[string]float32{
								"x":  -1,
								"y":  0,
								"z":  0,
								"nx": 0,
								"ny": 0,
								"nz": 1,
							},
						},
						{
							Float32s: map[string]float32{
								"x":  0,
								"y":  0,
								"z":  0,
								"nx": 0,
								"ny": 0,
								"nz": 1,
							},
						},
						{
							Float32s: map[string]float32{
								"x":  0,
								"y":  1,
								"z":  0,
								"nx": 0,
								"ny": 0,
								"nz": 1,
							},
						},
						{
							Float32s: map[string]float32{
								"x":  -1,
								"y":  1,
								"z":  0,
								"nx": 0,
								"ny": 0,
								"nz": 1,
							},
						},
					},
					"face": {
						{
							Uint32Lists: map[string][]uint32{
								"vertex_indices": {0, 1, 2, 3},
							},
						},
					},
				},
			},
		},
		{
			desc: "colored edges",
			input: `ply
format ascii 1.0
element vertex 9
property float x
property float y
property float z
property float nx
property float ny
property float nz
property uchar red
property uchar green
property uchar blue
element edge 8
property int vertex1
property int vertex2
property uchar red
property uchar green
property uchar blue
end_header
0.000000 0.000000 -2.000000 -0.000000 0.000000 1.000000 255 255 255
-0.390181 0.000000 -1.961571 -0.000000 0.000000 1.000000 255 255 255
-0.765367 0.000000 -1.847759 -0.000000 0.000000 1.000000 255 255 255
-1.111140 0.000000 -1.662939 -0.000000 0.000000 1.000000 255 255 255
-1.414214 0.000000 -1.414214 -0.000000 0.000000 1.000000 255 255 255
-1.662939 0.000000 -1.111140 -0.000000 0.000000 1.000000 255 255 255
-1.847759 0.000000 -0.765367 -0.000000 0.000000 1.000000 255 255 255
-1.961571 0.000000 -0.390181 -0.000000 0.000000 1.000000 255 255 255
-2.000000 0.000000 -0.000000 -0.000000 0.000000 1.000000 255 255 255
2 1 255 255 255
3 2 255 255 255
4 3 255 255 255
5 4 255 255 255
6 5 255 255 255
7 6 255 255 255
8 7 255 255 255
9 8 255 255 255`,
			want: &PLY{
				Elements: map[string][]*Element{
					"vertex": {
						{
							Float32s: map[string]float32{
								"x":  0,
								"y":  0,
								"z":  -2,
								"nx": 0,
								"ny": 0,
								"nz": 1,
							},
							Uint8s: map[string]uint8{
								"red":   255,
								"green": 255,
								"blue":  255,
							},
						},
						{
							Float32s: map[string]float32{
								"x":  -0.390181,
								"y":  0,
								"z":  -1.961571,
								"nx": 0,
								"ny": 0,
								"nz": 1,
							},
							Uint8s: map[string]uint8{
								"red":   255,
								"green": 255,
								"blue":  255,
							},
						},
						{
							Float32s: map[string]float32{
								"x":  -0.765367,
								"y":  0,
								"z":  -1.847759,
								"nx": 0,
								"ny": 0,
								"nz": 1,
							},
							Uint8s: map[string]uint8{
								"red":   255,
								"green": 255,
								"blue":  255,
							},
						},
						{
							Float32s: map[string]float32{
								"x":  -1.111140,
								"y":  0,
								"z":  -1.662939,
								"nx": 0,
								"ny": 0,
								"nz": 1,
							},
							Uint8s: map[string]uint8{
								"red":   255,
								"green": 255,
								"blue":  255,
							},
						},
						{
							Float32s: map[string]float32{
								"x":  -1.414214,
								"y":  0,
								"z":  -1.414214,
								"nx": 0,
								"ny": 0,
								"nz": 1,
							},
							Uint8s: map[string]uint8{
								"red":   255,
								"green": 255,
								"blue":  255,
							},
						},
						{
							Float32s: map[string]float32{
								"x":  -1.662939,
								"y":  0,
								"z":  -1.111140,
								"nx": 0,
								"ny": 0,
								"nz": 1,
							},
							Uint8s: map[string]uint8{
								"red":   255,
								"green": 255,
								"blue":  255,
							},
						},
						{
							Float32s: map[string]float32{
								"x":  -1.847759,
								"y":  0,
								"z":  -0.765367,
								"nx": 0,
								"ny": 0,
								"nz": 1,
							},
							Uint8s: map[string]uint8{
								"red":   255,
								"green": 255,
								"blue":  255,
							},
						},
						{
							Float32s: map[string]float32{
								"x":  -1.961571,
								"y":  0,
								"z":  -0.390181,
								"nx": 0,
								"ny": 0,
								"nz": 1,
							},
							Uint8s: map[string]uint8{
								"red":   255,
								"green": 255,
								"blue":  255,
							},
						},
						{
							Float32s: map[string]float32{
								"x":  -2,
								"y":  0,
								"z":  0,
								"nx": 0,
								"ny": 0,
								"nz": 1,
							},
							Uint8s: map[string]uint8{
								"red":   255,
								"green": 255,
								"blue":  255,
							},
						},
					},
					"edge": {
						{
							Int32s: map[string]int32{
								"vertex1": 2,
								"vertex2": 1,
							},
							Uint8s: map[string]uint8{
								"red":   255,
								"green": 255,
								"blue":  255,
							},
						},
						{
							Int32s: map[string]int32{
								"vertex1": 3,
								"vertex2": 2,
							},
							Uint8s: map[string]uint8{
								"red":   255,
								"green": 255,
								"blue":  255,
							},
						},
						{
							Int32s: map[string]int32{
								"vertex1": 4,
								"vertex2": 3,
							},
							Uint8s: map[string]uint8{
								"red":   255,
								"green": 255,
								"blue":  255,
							},
						},
						{
							Int32s: map[string]int32{
								"vertex1": 5,
								"vertex2": 4,
							},
							Uint8s: map[string]uint8{
								"red":   255,
								"green": 255,
								"blue":  255,
							},
						},
						{
							Int32s: map[string]int32{
								"vertex1": 6,
								"vertex2": 5,
							},
							Uint8s: map[string]uint8{
								"red":   255,
								"green": 255,
								"blue":  255,
							},
						},
						{
							Int32s: map[string]int32{
								"vertex1": 7,
								"vertex2": 6,
							},
							Uint8s: map[string]uint8{
								"red":   255,
								"green": 255,
								"blue":  255,
							},
						},
						{
							Int32s: map[string]int32{
								"vertex1": 8,
								"vertex2": 7,
							},
							Uint8s: map[string]uint8{
								"red":   255,
								"green": 255,
								"blue":  255,
							},
						},
						{
							Int32s: map[string]int32{
								"vertex1": 9,
								"vertex2": 8,
							},
							Uint8s: map[string]uint8{
								"red":   255,
								"green": 255,
								"blue":  255,
							},
						},
					},
				},
			},
		},
	} {
		t.Run(tt.desc, func(t *testing.T) {
			got, gotErr := Decode(strings.NewReader(tt.input))

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("diff (-want, +got)\n%s", diff)
			}

			if (gotErr != nil) != tt.wantErr {
				t.Errorf("got error: %v, wanted err: %t", gotErr, tt.wantErr)
			}
		})
	}
}
