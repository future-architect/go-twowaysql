package twowaysql

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "if true",
			input: `SELECT * FROM person WHERE employee_no < 1000 /* IF true */ AND dept_no = 1 /* END */`,
			want:  `SELECT * FROM person WHERE employee_no < 1000 AND dept_no = 1`,
		},
		{
			name:  "if false",
			input: `SELECT * FROM person WHERE employee_no < 1000 /* IF false */ AND dept_no = 1 /* END */`,
			want:  `SELECT * FROM person WHERE employee_no < 1000`,
		},
		{
			name:  "if true else",
			input: `SELECT * FROM person WHERE employee_no < 1000  /* IF true */ AND dept_no = 1 /* ELSE */ boss_no = 2 /* END */`,
			want:  `SELECT * FROM person WHERE employee_no < 1000 AND dept_no = 1`,
		},
		{
			name:  "if false else",
			input: `SELECT * FROM person WHERE employee_no < 1000  /* IF false */ AND  dept_no =1 /* ELSE */AND boss_no = 2 /* END */`,
			want:  `SELECT * FROM person WHERE employee_no < 1000 AND boss_no = 2`,
		},
		{
			name:  "if true elif true else",
			input: `SELECT * FROM person WHERE employee_no < 1000 /* IF true */ AND dept_no = 1 /* ELIF true */ AND boss_no = 2 /* ELSE */ AND id = 3 /* END */`,
			want:  `SELECT * FROM person WHERE employee_no < 1000 AND dept_no = 1`,
		},
		{
			name:  "if false elif true else",
			input: `SELECT * FROM person WHERE employee_no < 1000 /* IF false */ AND dept_no = 1 /* ELIF true */ AND boss_no = 2 /* ELSE */ AND id = 3 /* END */`,
			want:  `SELECT * FROM person WHERE employee_no < 1000 AND boss_no = 2`,
		},
		{
			name:  "if false elif false else",
			input: `SELECT * FROM person WHERE employee_no < 1000 /* IF false */ AND dept_no = 1 /* ELIF false */ AND boss_no = 2 /* ELSE */ AND id = 3 /* END */`,
			want:  `SELECT * FROM person WHERE employee_no < 1000 AND id = 3`,
		},
		{
			name:  "bind parameter",
			input: `SELECT * FROM person WHERE employee_no < /*maxEmpNo*/1000`,
			want:  `SELECT * FROM person WHERE employee_no < ?/*maxEmpNo*/`,
		},
		{
			name:  "if false elif false else",
			input: `SELECT * FROM person WHERE employee_no < 1000 /* IF false */ AND dept_no = 1 /* ELIF false */ AND boss_no = 2 /* ELSE */ AND id = /*maxEmpNo*/3 /* END */`,
			want:  `SELECT * FROM person WHERE employee_no < 1000 AND id = ?/*maxEmpNo*/`,
		},
		{
			name:  "if nest",
			input: `SELECT * FROM person WHERE employee_no < 1000 /* IF true */ /* IF false */ AND dept_no =1 /* ELSE */ AND id=3 /* END */ /* ELSE*/ AND boss_id=4 /* END */`,
			want:  `SELECT * FROM person WHERE employee_no < 1000 AND id=3`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tw := New(nil).withQuery(tt.input)
			if got, err := tw.parse(); err != nil || got != tt.want {
				if err != nil {
					t.Log(err)
				}
				t.Errorf("Doesn't Match\nexpected: \n%s\n but got: \n%s\n", tt.want, got)
			}
		})
	}
}
func TestCondition(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "if true",
			input: `SELECT * FROM person WHERE employee_no < 1000 /* IF checked */ AND dept_no = 1 /* END */`,
			want:  `SELECT * FROM person WHERE employee_no < 1000 AND dept_no = 1`,
		},
		{
			name:  "if false",
			input: `SELECT * FROM person WHERE employee_no < 1000 /* IF uncheckd */ AND dept_no = 1 /* END */`,
			want:  `SELECT * FROM person WHERE employee_no < 1000`,
		},
		{
			name:  "if truthy 1",
			input: `SELECT * FROM person WHERE employee_no < 1000 /* IF name */ AND dept_no = 1 /* END */`,
			want:  `SELECT * FROM person WHERE employee_no < 1000 AND dept_no = 1`,
		},
		{
			name:  "if truthy 2",
			input: `SELECT * FROM person WHERE employee_no < 1000 /* IF deptNo */ AND dept_no = 1 /* END */`,
			want:  `SELECT * FROM person WHERE employee_no < 1000 AND dept_no = 1`,
		},
		{
			name:  "if truthy 3",
			input: `SELECT * FROM person WHERE employee_no < 1000 /* IF name */ AND dept_no = 1 /* END */`,
			want:  `SELECT * FROM person WHERE employee_no < 1000 AND dept_no = 1`,
		},
		{
			name:  "if falsy 1",
			input: `SELECT * FROM person WHERE employee_no < 1000 /* IF zero */ AND dept_no = 1 /* END */`,
			want:  `SELECT * FROM person WHERE employee_no < 1000`,
		},
		{
			name:  "if falsy 2",
			input: `SELECT * FROM person WHERE employee_no < 1000 /* IF nil */ AND dept_no = 1 /* END */`,
			want:  `SELECT * FROM person WHERE employee_no < 1000`,
		},
	}
	var params = map[string]interface{}{"name": "HR", "maxEmpNo": 2000, "deptNo": 15, "checked": true, "uncheckd": false, "zero": 0, "nil": nil}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tw := New(nil).withQuery(tt.input).withParams(params)
			if got, err := tw.parse(); err != nil || got != tt.want {
				if err != nil {
					t.Log(err)
				}
				t.Errorf("Doesn't Match\nexpected: \n%s\n but got: \n%s\n", tt.want, got)
			}
		})
	}
}

func TestParseAbnormal(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "no END",
			input: `SELECT * FROM person WHERE employee_no < 1000 /* IF true */ AND dept_no = 1`,
		},
		{
			name:  "extra END 1",
			input: "SELECT * FROM person WHERE employee_no < 1000  AND dept_no = 1 /* END */",
		},
		{
			name:  "extra END 2",
			input: "SELECT * FROM person WHERE employee_no < 1000  /* END */ AND dept_no = 1 ",
		},
		{
			name:  "invalid Elif pos",
			input: `SELECT * FROM person WHERE employee_no < 1000 /* ELIF true */ AND dept_no = 1`,
		},
		{
			name:  "not match if, elif and end",
			input: `SELECT * FROM person WHERE employee_no < 1000 /* IF true */ /* IF false */ AND dept_no =1 /* ELSE */ AND id=3 /* ELSE*/ AND boss_id=4 /* END */`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tw := New(nil).withQuery(tt.input)
			if got, err := tw.parse(); err == nil {
				t.Log("got", got)
				t.Errorf("should return error")
			} else {
				t.Log(err)
			}
		})
	}
}