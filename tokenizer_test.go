package twowaysql

import (
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []Token
	}{
		{
			name:  "empty",
			input: "",
			want: []Token{
				{
					kind: TkEndOfProgram,
				},
			},
		},
		{
			name:  "no comment",
			input: "SELECT * FROM person WHERE employee_no < 1000  AND dept_no = 1",
			want: []Token{
				{
					kind: TkSQLStmt,
					str:  "SELECT * FROM person WHERE employee_no < 1000  AND dept_no = 1",
				},
				{
					kind: TkEndOfProgram,
				},
			},
		},
		{
			name:  "if",
			input: "SELECT * FROM person WHERE employee_no < 1000 /* IF true */ AND dept_no = 1/* END */",
			want: []Token{
				{
					kind: TkSQLStmt,
					str:  "SELECT * FROM person WHERE employee_no < 1000 ",
				},
				{
					kind: TkIf,
					str:  "/* IF true */",
				},
				{
					kind: TkSQLStmt,
					str:  " AND dept_no = 1",
				},
				{
					kind: TkEnd,
					str:  "/* END */",
				},
				{
					kind: TkEndOfProgram,
				},
			},
		},
		{
			name:  "if and bind",
			input: "SELECT * FROM person WHERE employee_no < /*maxEmpNo*/1000 /* IF false */ AND dept_no = /*deptNo*/1 /* END */",
			want: []Token{
				{
					kind: TkSQLStmt,
					str:  "SELECT * FROM person WHERE employee_no < ",
				},
				{
					kind: TkBind,
					str:  "/*maxEmpNo*/1000",
				},
				{
					kind: TkSQLStmt,
					str:  " ",
				},
				{
					kind: TkIf,
					str:  "/* IF false */",
				},
				{
					kind: TkSQLStmt,
					str:  " AND dept_no = ",
				},
				{
					kind: TkBind,
					str:  "/*deptNo*/1",
				},
				{
					kind: TkSQLStmt,
					str:  " ",
				},
				{
					kind: TkEnd,
					str:  "/* END */",
				},
				{
					kind: TkEndOfProgram,
				},
			},
		},
		{
			name:  "if elif else",
			input: "SELECT * FROM person WHERE employee_no < 1000 /* IF true */AND dept_no =1/* ELIF true*/ AND boss_no = 2 /*ELSE */ AND id=3/* END */",
			want: []Token{
				{
					kind: TkSQLStmt,
					str:  "SELECT * FROM person WHERE employee_no < 1000 ",
				},
				{
					kind: TkIf,
					str:  "/* IF true */",
				},
				{
					kind: TkSQLStmt,
					str:  "AND dept_no =1",
				},
				{
					kind: TkElif,
					str:  "/* ELIF true*/",
				},
				{
					kind: TkSQLStmt,
					str:  " AND boss_no = 2 ",
				},
				{
					kind: TkElse,
					str:  "/*ELSE */",
				},
				{
					kind: TkSQLStmt,
					str:  " AND id=3",
				},
				{
					kind: TkEnd,
					str:  "/* END */",
				},
				{
					kind: TkEndOfProgram,
				},
			},
		},
		{
			name:  "if nest",
			input: "SELECT * FROM person WHERE employee_no < 1000 /* IF true */ /* IF false */ AND dept_no =1 /* ELSE */ AND id=3 /* END */ /* ELSE*/ AND boss_id=4 /* END */",
			want: []Token{
				{
					kind: TkSQLStmt,
					str:  "SELECT * FROM person WHERE employee_no < 1000 ",
				},
				{
					kind: TkIf,
					str:  "/* IF true */",
				},
				{
					kind: TkSQLStmt,
					str:  " ",
				},
				{
					kind: TkIf,
					str:  "/* IF false */",
				},
				{
					kind: TkSQLStmt,
					str:  " AND dept_no =1 ",
				},
				{
					kind: TkElse,
					str:  "/* ELSE */",
				},
				{
					kind: TkSQLStmt,
					str:  " AND id=3 ",
				},
				{
					kind: TkEnd,
					str:  "/* END */",
				},
				{
					kind: TkSQLStmt,
					str:  " ",
				},
				{
					kind: TkElse,
					str:  "/* ELSE*/",
				},
				{
					kind: TkSQLStmt,
					str:  " AND boss_id=4 ",
				},
				{
					kind: TkEnd,
					str:  "/* END */",
				},
				{
					kind: TkEndOfProgram,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := tokinize(tt.input); err != nil || !tokensEqual(tt.want, got) {
				if err != nil {
					t.Error(err)
				}
				t.Errorf("Doesn't Match expected: %v, but got: %v\n", tt.want, got)
			}
		})
	}
}
func TestTokenizeShouldReturnError(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "bad comment format",
			input: "SELECT * FROM person WHERE employee_no < 1000 /* IF true / AND dept_no = 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tokinize(tt.input); err == nil {
				t.Error("Should Error")
			}
		})
	}
}

func tokensEqual(want, got []Token) bool {
	if len(want) != len(got) {
		return false
	}
	for i := 0; i < len(want); i++ {
		if want[i] != got[i] {
			return false
		}
	}
	return true
}
