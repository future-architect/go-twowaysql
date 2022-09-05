package twowaysql

import (
	"log"
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func TestParseMarkdown(t *testing.T) {
	type args struct {
		src string
	}
	tests := []struct {
		name    string
		args    args
		want    *Document
		wantErr string
	}{
		{
			name: "title and sql",
			args: args{
				src: TrimIndent(t, `
				# Search User Query

				~~~sql
				SELECT email, name FROM persons WHERE first_name=/*first_name*/'bob';
				~~~
				`),
			},
			want: &Document{
				Title: "Search User Query",
				SQL:   `SELECT email, name FROM persons WHERE first_name=/*first_name*/'bob';`,
			},
		},
		{
			name: "with parameter",
			args: args{
				src: TrimIndent(t, `
				# Search User Query

				~~~sql
				SELECT email, name FROM persons WHERE first_name=/*first_name*/'bob';
				~~~

				## Parameter

				| Name       | Type   | Description |
				|------------|--------|-------------|
				| first_name | string | search key  |
				`),
			},
			want: &Document{
				Title: "Search User Query",
				SQL:   `SELECT email, name FROM persons WHERE first_name=/*first_name*/'bob';`,
				Params: []Param{
					{
						Name:        "first_name",
						Type:        TextType,
						Description: "search key",
					},
				},
			},
		},
		{
			name: "with CRUD matrix",
			args: args{
				src: TrimIndent(t, `
				# Search User Query

				~~~sql
				SELECT email, name FROM persons WHERE first_name=/*first_name*/'bob';
				~~~

				## CRUD Matrix

				| Table      | C | R | U | D | Description |
				|------------|---|---|---|---|-------------|
				| persons    | X |   |   |   |             |
				`),
			},
			want: &Document{
				Title: "Search User Query",
				SQL:   `SELECT email, name FROM persons WHERE first_name=/*first_name*/'bob';`,
				CRUDMatrix: []CRUDMatrix{
					{
						Table: "persons",
						C:     true,
						R:     false,
						U:     false,
						D:     false,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMarkdownString(tt.args.src)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
			} else {
				assert.NilError(t, err)
				assert.Check(t, cmp.DeepEqual(tt.want, got))
			}
		})
	}
}
