package cli

import (
	"io"
	"math"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gotest.tools/v3/assert"
)

func Test_parseParams(t *testing.T) {
	type args struct {
		params []string
		stdin  io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]any
		wantErr string
	}{
		{
			name: "empty values",
			args: args{
				params: []string{},
			},
			want: map[string]any{},
		},
		{
			name: "single raw values",
			args: args{
				params: []string{"name=tokyo", "utcOffset=9", "lat=35.6", "lon=139.6"},
			},
			want: map[string]any{
				"name":      "tokyo",
				"utcOffset": float64(9),
				"lat":       35.6,
				"lon":       139.6,
			},
		},
		{
			name: "JSON values",
			args: args{
				params: []string{`{"name": "tokyo", "utcOffset": 9, "lat": 35.6, "lon": 139.6}`},
			},
			want: map[string]any{
				"name":      "tokyo",
				"utcOffset": float64(9),
				"lat":       35.6,
				"lon":       139.6,
			},
		},
		{
			name: "JSON from stdin",
			args: args{
				stdin: strings.NewReader(`{"name": "tokyo", "utcOffset": 9, "lat": 35.6, "lon": 139.6}`),
			},
			want: map[string]any{
				"name":      "tokyo",
				"utcOffset": float64(9),
				"lat":       35.6,
				"lon":       139.6,
			},
		},
		{
			name: "invalid error (1): key only",
			args: args{
				params: []string{"name"},
			},
			want:    map[string]any{},
			wantErr: "1 error occurred:\n\t* invalid format: 'name' key=value or JSON is supported\n\n",
		},
		{
			name: "invalid error (2): JSON parse error",
			args: args{
				params: []string{`{"name": "tokyo",}`},
			},
			want:    map[string]any{},
			wantErr: "1 error occurred:\n\t* JSON parse error: invalid character '}' looking for beginning of object key string\n\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseParams(tt.args.params, tt.args.stdin)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
			} else {
				assert.NilError(t, err)
				assert.DeepEqual(t, got, tt.want, cmp.Comparer(func(x, y float64) bool {
					return math.Abs(x-y) < 0.01
				}))
			}
		})
	}
}
