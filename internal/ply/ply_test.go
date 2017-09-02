package ply

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDecode(t *testing.T) {
	for _, tt := range []struct {
		desc  string
		input string
		want  *PLY
	}{
		{
			desc: "orthoPlane",
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
							Floats: map[string]float32{
								"x":  -1,
								"y":  0,
								"z":  0,
								"nx": 0,
								"ny": 0,
								"nz": 1,
							},
						},
						{
							Floats: map[string]float32{
								"x":  0,
								"y":  0,
								"z":  0,
								"nx": 0,
								"ny": 0,
								"nz": 1,
							},
						},
						{
							Floats: map[string]float32{
								"x":  0,
								"y":  1,
								"z":  0,
								"nx": 0,
								"ny": 0,
								"nz": 1,
							},
						},
						{
							Floats: map[string]float32{
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
							UintLists: map[string][]uint{
								"vertex_indices": {0, 1, 2, 3},
							},
						},
					},
				},
			},
		},
	} {
		t.Run(tt.desc, func(t *testing.T) {
			got, gotErr := Decode(strings.NewReader(tt.input))
			if gotErr != nil {
				t.Fatalf("Decode returned an error (%v), want success", gotErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("differs:\n%s", diff)
			}
		})
	}
}
