package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/future-architect/go-twowaysql/private/testhelper"
	"github.com/shibukawa/acquire-go"
	"gotest.tools/v3/assert"

	_ "github.com/jackc/pgx/v4/stdlib"
)

func Test_run(t *testing.T) {
	driver := "pgx"
	dbSrc := testhelper.SourceStr(t)
	type args struct {
		srcPath      string
		params       []string
		explain      bool
		rollback     bool
		outputFormat string
	}
	tests := []struct {
		name      string
		args      args
		wantOut   string
		wantError string
	}{
		{
			name: "simple get: json out",
			args: args{
				srcPath:      "testdata/postgres/sql/select_person.sql",
				params:       []string{"first_name=Evan"},
				outputFormat: "json",
			},
			wantOut: testhelper.TrimIndent(t, `
			[
			  {
			    "email": "evanmacmans@example.com",
			    "first_name": "Evan",
			    "last_name": "MacMans"
			  }
			]`),
		},
		{
			name: "simple get: yaml out",
			args: args{
				srcPath:      "testdata/postgres/sql/select_person.sql",
				params:       []string{"first_name=Evan"},
				outputFormat: "yaml",
			},
			wantOut: testhelper.TrimIndent(t, `
				- email: evanmacmans@example.com
				  first_name: Evan
				  last_name: MacMans`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.args.srcPath)
			files := acquire.MustAcquire(acquire.File, tt.args.srcPath)
			out := &bytes.Buffer{}
			err := run(driver, dbSrc, files[0], tt.args.params, tt.args.explain, tt.args.rollback, tt.args.outputFormat, out)
			if tt.wantError != "" {
				assert.Error(t, err, tt.wantError)
			} else {
				assert.NilError(t, err)
				assert.Equal(t, tt.wantOut, strings.TrimSpace(out.String()))
			}
		})
	}
}
