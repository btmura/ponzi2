package iex

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestValidTokenRegexp(t *testing.T) {
	for _, tt := range []struct {
		desc  string
		input string
		want  bool
	}{
		{
			desc:  "empty",
			input: "",
			want:  false,
		},
		{
			desc:  "whitespace",
			input: " ",
			want:  false,
		},
		{
			desc:  "trailing whitespace",
			input: " abc ",
			want:  false,
		},
		{
			desc:  "letters",
			input: "abc",
			want:  true,
		},
		{
			desc:  "numbers",
			input: "0123",
			want:  true,
		},
	} {
		t.Run(tt.desc, func(t *testing.T) {
			got := validTokenRegexp.MatchString(tt.input)

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("diff (-want, +got)\n%s", diff)
			}
		})
	}
}
