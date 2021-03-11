package twowaysql

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name  string
		input *tree
		want  []token
	}{
		{
			name:  "",
			input: wantEmpty,
			want:  []token{},
		},
		{
			name:  "no comment",
			input: wantNoComment,
			want: []token{
				{
					kind: tkSQLStmt,
					str:  "SELECT * FROM person WHERE employee_no < 1000  AND dept_no = 1",
				},
			},
		},
		{
			name:  "if",
			input: wantTreeIf,
			want: []token{
				{
					kind: tkSQLStmt,
					str:  "SELECT * FROM person WHERE employee_no < 1000 ",
				},
				{
					kind: tkSQLStmt,
					str:  " AND dept_no = 1",
				},
			},
		},
		{
			name:  "if and bind",
			input: wantTreeIfBind,
			want: []token{
				{
					kind: tkSQLStmt,
					str:  "SELECT * FROM person WHERE employee_no < ",
				},
				{
					kind:  tkBind,
					str:   "?/*maxEmpNo*/",
					value: "maxEmpNo",
				},
				{
					kind: tkSQLStmt,
					str:  " ",
				},
			},
		},
		{
			name:  "if elif else",
			input: wantIfElifElse,
			want: []token{
				{
					kind: tkSQLStmt,
					str:  "SELECT * FROM person WHERE employee_no < 1000 ",
				},
				{
					kind: tkSQLStmt,
					str:  "AND dept_no =1",
				},
			},
		},
		{
			name:  "if nest",
			input: wantIfNest,
			want: []token{
				{
					kind: tkSQLStmt,
					str:  "SELECT * FROM person WHERE employee_no < 1000 ",
				},
				{
					kind: tkSQLStmt,
					str:  " ",
				},
				{
					kind: tkSQLStmt,
					str:  " AND id=3 ",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := tt.input.parse(map[string]interface{}{}); err != nil || !tokensEqual(tt.want, got) {
				if err != nil {
					t.Error(err)
				}
				if !tokensEqual(tt.want, got) {
					t.Errorf("Doesn't Match:\nexpected: \n%v\n but got: \n%v\n", tt.want, got)
				}
			}
		})
	}
}
