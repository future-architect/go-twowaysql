package cli

import (
	"testing"

	"github.com/shibukawa/acquire-go"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func Test_findFiles(t *testing.T) {
	type args struct {
		filesOrDirs []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "empty file",
			args: args{
				filesOrDirs: []string{},
			},
			want: []string{},
		},
		{
			name: "single file",
			args: args{
				filesOrDirs: []string{
					"testdata/postgres/markdown/select_person.sql.md",
				},
			},
			want: []string{
				acquire.MustAcquire(acquire.File, "testdata/postgres/markdown/select_person.sql.md")[0],
			},
		},
		{
			name: "search dir",
			args: args{
				filesOrDirs: []string{
					"testdata/postgres/markdown",
				},
			},
			want: []string{
				acquire.MustAcquire(acquire.File, "testdata/postgres/markdown/select_person.sql.md")[0],
				acquire.MustAcquire(acquire.File, "testdata/postgres/markdown/select_person_notest.sql.md")[0],
				acquire.MustAcquire(acquire.File, "testdata/postgres/markdown/select_person_with_param.sql.md")[0],
			},
		},
		{
			name: "search dir (remove duplication)",
			args: args{
				filesOrDirs: []string{
					"testdata/postgres/markdown",
					"testdata/postgres",
				},
			},
			want: []string{
				acquire.MustAcquire(acquire.File, "testdata/postgres/markdown/select_person.sql.md")[0],
				acquire.MustAcquire(acquire.File, "testdata/postgres/markdown/select_person_notest.sql.md")[0],
				acquire.MustAcquire(acquire.File, "testdata/postgres/markdown/select_person_with_param.sql.md")[0],
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var files []string
			for _, f := range tt.args.filesOrDirs {
				files = append(files, acquire.MustAcquire(acquire.All, f)...)
			}
			got := findFiles(files)
			assert.Check(t, cmp.DeepEqual(tt.want, got))
		})
	}
}
