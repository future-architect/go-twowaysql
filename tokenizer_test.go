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
			want:  []Token{},
		},
		{
			name:  "no comment",
			input: "SELECT * FROM person WHERE employee_no < 1000  AND dept_no = 1",
			want: []Token{
				{
					kind: TkSQLStmt,
					str:  "SELECT * FROM person WHERE employee_no < 1000  AND dept_no = 1",
				},
			},
		},
		{
			name:  "if",
			input: "SELECT * FROM person WHERE employee_no < 1000 /* IF true */ AND dept_no = 1",
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
			},
		},
		{
			name:  "if and bind",
			input: "SELECT * FROM person WHERE employee_no < /*maxEmpNo*/1000 /* IF exists(deptNo)*/ AND dept_no = /*deptNo*/1",
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
					str:  "/* IF exists(deptNo)*/",
				},
				{
					kind: TkSQLStmt,
					str:  " AND dept_no = ",
				},
				{
					kind: TkBind,
					str:  "/*deptNo*/1",
				},
			},
		},
		{
			name:  "if elif else",
			input: "SELECT * FROM person WHERE employee_no < 1000 AND dept_no = /* IF true */1/* ELIF true*/ 2 /*ELSE */ 3",
			want: []Token{
				{
					kind: TkSQLStmt,
					str:  "SELECT * FROM person WHERE employee_no < 1000 AND dept_no = ",
				},
				{
					kind: TkIf,
					str:  "/* IF true */",
				},
				{
					kind: TkSQLStmt,
					str:  "1",
				},
				{
					kind: TkElif,
					str:  "/* ELIF true*/",
				},
				{
					kind: TkSQLStmt,
					str:  " 2 ",
				},
				{
					kind: TkElse,
					str:  "/*ELSE */",
				},
				{
					kind: TkSQLStmt,
					str:  " 3",
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
