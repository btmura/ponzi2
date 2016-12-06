package ponzi

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestDecodeObjs(t *testing.T) {
	for _, tt := range []struct {
		desc    string
		input   string
		want    []*obj
		wantErr error
	}{
		{
			desc: "missing object ID",
			input: `
				# Blender v2.76 (sub 0) OBJ File: ''
				# www.blender.org
				v -0.032120 -0.290752 -0.832947
				v 1.967880 -0.290752 -0.832947
				v -0.032120 -0.290752 -2.832947
				s off
				f 1 2 4
			`,
			wantErr: errors.New("missing object ID"),
		},
		{
			desc: "valid cube",
			input: `
				# Blender v2.76 (sub 0) OBJ File: ''
				# www.blender.org
				o Cube
				v 1.000000 -1.000000 -1.000000
				v 1.000000 -1.000000 1.000000
				v -1.000000 -1.000000 1.000000
				v -1.000000 -1.000000 -1.000000
				v 1.000000 1.000000 -0.999999
				v 0.999999 1.000000 1.000001
				v -1.000000 1.000000 1.000000
				v -1.000000 1.000000 -1.000000
				vt 0.000000 0.000000
				vt 1.000000 0.000000
				vt 1.000000 1.000000
				vt 0.000000 1.000000
				vn 0.000000 -1.000000 0.000000
				vn 0.000000 1.000000 0.000000
				vn 1.000000 0.000000 0.000000
				vn -0.000000 -0.000000 1.000000
				vn -1.000000 -0.000000 -0.000000
				vn 0.000000 0.000000 -1.000000
				s off
				f 2/1/1 3/2/1 4/3/1
				f 8/1/2 7/2/2 6/3/2
				f 5/1/3 6/2/3 2/3/3
				f 6/1/4 7/2/4 3/3/4
				f 3/4/5 7/1/5 8/2/5
				f 1/1/6 4/2/6 8/3/6
				f 1/4/1 2/1/1 4/3/1
				f 5/4/2 8/1/2 6/3/2
				f 1/4/3 5/1/3 2/3/3
				f 2/4/4 6/1/4 3/3/4
				f 4/3/5 3/4/5 8/2/5
				f 5/4/6 1/1/6 8/3/6
			`,
			want: []*obj{
				{
					id: "Cube",
					vertices: []*objVertex{
						{1, -1, -1},
						{1, -1, 1},
						{-1, -1, 1},
						{-1, -1, -1},
						{1, 1, -0.999999},
						{0.999999, 1, 1.000001},
						{-1, 1, 1},
						{-1, 1, -1},
					},
					texCoords: []*objTexCoord{
						{0, 0},
						{1, 0},
						{1, 1},
						{0, 1},
					},
					normals: []*objNormal{
						{0, -1, 0},
						{0, 1, 0},
						{1, 0, 0},
						{0, 0, 1},
						{-1, 0, 0},
						{0, 0, -1},
					},
					faces: []*objFace{
						{{2, 1, 1}, {3, 2, 1}, {4, 3, 1}},
						{{8, 1, 2}, {7, 2, 2}, {6, 3, 2}},
						{{5, 1, 3}, {6, 2, 3}, {2, 3, 3}},
						{{6, 1, 4}, {7, 2, 4}, {3, 3, 4}},
						{{3, 4, 5}, {7, 1, 5}, {8, 2, 5}},
						{{1, 1, 6}, {4, 2, 6}, {8, 3, 6}},
						{{1, 4, 1}, {2, 1, 1}, {4, 3, 1}},
						{{5, 4, 2}, {8, 1, 2}, {6, 3, 2}},
						{{1, 4, 3}, {5, 1, 3}, {2, 3, 3}},
						{{2, 4, 4}, {6, 1, 4}, {3, 3, 4}},
						{{4, 3, 5}, {3, 4, 5}, {8, 2, 5}},
						{{5, 4, 6}, {1, 1, 6}, {8, 3, 6}},
					},
				},
			},
		},
		{
			desc: "multiple objects",
			input: `
				# Blender v2.76 (sub 0) OBJ File: ''
				# www.blender.org
				o Plane.001
				v 0.652447 0.140019 -0.450452
				v 2.652447 0.140019 -0.450452
				v 0.652447 0.140019 -2.450452
				v 2.652447 0.140019 -2.450452
				s off
				f 2 4 3
				f 1 2 3
				o Plane
				v -1.079860 0.672774 2.814899
				v 0.920140 0.672774 2.814899
				v -1.079860 0.672774 0.814900
				v 0.920140 0.672774 0.814900
				s off
				f 6 8 7
				f 5 6 7
			`,
			want: []*obj{
				{
					id: "Plane.001",
					vertices: []*objVertex{
						{0.652447, 0.140019, -0.450452},
						{2.652447, 0.140019, -0.450452},
						{0.652447, 0.140019, -2.450452},
						{2.652447, 0.140019, -2.450452},
					},
					faces: []*objFace{
						{{2, 0, 0}, {4, 0, 0}, {3, 0, 0}},
						{{1, 0, 0}, {2, 0, 0}, {3, 0, 0}},
					},
				},
				{
					id: "Plane",
					vertices: []*objVertex{
						{-1.079860, 0.672774, 2.814899},
						{0.920140, 0.672774, 2.814899},
						{-1.079860, 0.672774, 0.814900},
						{0.920140, 0.672774, 0.814900},
					},
					faces: []*objFace{
						{{6, 0, 0}, {8, 0, 0}, {7, 0, 0}},
						{{5, 0, 0}, {6, 0, 0}, {7, 0, 0}},
					},
				},
			},
		},
	} {
		got, gotErr := decodeObjs(strings.NewReader(tt.input))
		if !reflect.DeepEqual(got, tt.want) || !errorContains(gotErr, tt.wantErr) {
			t.Errorf("[%s] decodeObjs(%q) = (%v, %v), want (%v, %v)", tt.desc, tt.input, pp(got), gotErr, pp(tt.want), tt.wantErr)
		}
	}
}

var pp = pretty.Sprint

func errorContains(gotErr, wantErr error) bool {
	return strings.Contains(fmt.Sprint(gotErr), fmt.Sprint(wantErr))
}
