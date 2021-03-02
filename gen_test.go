package twowaysql

import (
	"testing"
)

func TestGen(t *testing.T) {
	tests := []struct {
		name  string
		input *Tree
		want  []Token
	}{
		{
			name:  "",
			input: makeEmpty(),
			want:  []Token{},
		},
		{
			name:  "no comment",
			input: makeNoComment(),
			want: []Token{
				{
					kind: TkSQLStmt,
					str:  "SELECT * FROM person WHERE employee_no < 1000  AND dept_no = 1",
				},
			},
		},
		{
			name:  "if",
			input: makeTreeIf(),
			want: []Token{
				{
					kind: TkSQLStmt,
					str:  "SELECT * FROM person WHERE employee_no < 1000 ",
				},
				{
					kind: TkSQLStmt,
					str:  " AND dept_no = 1",
				},
			},
		},
		{
			name:  "if and bind",
			input: makeTreeIfBind(),
			want: []Token{
				{
					kind: TkSQLStmt,
					str:  "SELECT * FROM person WHERE employee_no < ",
				},
				{
					kind:  TkBind,
					str:   "?/*maxEmpNo*/",
					value: "maxEmpNo",
				},
				{
					kind: TkSQLStmt,
					str:  " ",
				},
			},
		},
		{
			name:  "if elif else",
			input: makeIfElifElse(),
			want: []Token{
				{
					kind: TkSQLStmt,
					str:  "SELECT * FROM person WHERE employee_no < 1000 ",
				},
				{
					kind: TkSQLStmt,
					str:  "AND dept_no =1",
				},
			},
		},
		{
			name:  "if nest",
			input: makeIfNest(),
			want: []Token{
				{
					kind: TkSQLStmt,
					str:  "SELECT * FROM person WHERE employee_no < 1000 ",
				},
				{
					kind: TkSQLStmt,
					str:  " ",
				},
				{
					kind: TkSQLStmt,
					str:  " AND id=3 ",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := gen(tt.input, map[string]interface{}{}); err != nil || !tokensEqual(tt.want, got) {
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
