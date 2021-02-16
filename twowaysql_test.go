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
			input: `SELECT * FROM person WHERE employee_no < 1000 /* IF true */ AND dept_no = 1`,
			want:  `SELECT * FROM person WHERE employee_no < 1000 AND dept_no = 1`,
		},
		{
			name:  "if false",
			input: `SELECT * FROM person WHERE employee_no < 1000 /* IF false */ AND dept_no = 1`,
			want:  `SELECT * FROM person WHERE employee_no < 1000`,
		},
		{
			name:  "if true else",
			input: `SELECT * FROM person WHERE employee_no < 1000  AND dept_no = /* IF true */ 1 /* ELSE */ 2`,
			want:  `SELECT * FROM person WHERE employee_no < 1000 AND dept_no = 1`,
		},
		{
			name:  "if false else",
			input: `SELECT * FROM person WHERE employee_no < 1000  AND dept_no = /* IF false */ 1 /* ELSE */ 2`,
			want:  `SELECT * FROM person WHERE employee_no < 1000 AND dept_no = 2`,
		},
		{
			name:  "if true elif true else",
			input: `SELECT * FROM person WHERE employee_no < 1000  AND dept_no = /* IF true */ 1 /* ELIF true */ 2 /* ELSE */ 3`,
			want:  `SELECT * FROM person WHERE employee_no < 1000 AND dept_no = 1`,
		},
		{
			name:  "if false elif true else",
			input: `SELECT * FROM person WHERE employee_no < 1000  AND dept_no = /* IF false */ 1 /* ELIF true */ 2 /* ELSE */ 3`,
			want:  `SELECT * FROM person WHERE employee_no < 1000 AND dept_no = 2`,
		},
		{
			name:  "if false elif false else",
			input: `SELECT * FROM person WHERE employee_no < 1000  AND dept_no = /* IF false */ 1 /* ELIF true */ 2 /* ELSE */ 3`,
			want:  `SELECT * FROM person WHERE employee_no < 1000 AND dept_no = 3`,
		},
		{
			name:  "bind parameter",
			input: `SELECT * FROM person WHERE employee_no < /*maxEmpNo*/1000`,
			want:  `SELECT * FROM person WHERE employee_no < ?/*maxEmpNo*/`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convert(tt.input); got != tt.want {
				t.Errorf("Doesn't Match\n expected: %s, but got: %s\n", tt.want, got)
			}
		})
	}
}
