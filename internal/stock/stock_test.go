package stock

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestParseDate(t *testing.T) {
	for _, tt := range []struct {
		desc  string
		input string
		want  time.Time
	}{
		{
			desc:  "date with time",
			input: "2018-06-29 10:48:47",
			want:  time.Date(2018, 6, 29, 10, 48, 47, 0, time.UTC),
		},
		{
			desc:  "date without time",
			input: "2018-06-28",
			want:  time.Date(2018, 6, 28, 0, 0, 0, 0, time.UTC),
		},
	} {
		t.Run(tt.desc, func(t *testing.T) {
			got, gotErr := parseDate(tt.input)
			if gotErr != nil {
				t.Fatalf("parseDate returned an error (%v), want success", gotErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("differs:\n%s", diff)
			}
		})
	}
}
