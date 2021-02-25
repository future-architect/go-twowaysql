package twowaysql

import (
	"testing"
)

func TestGen(t *testing.T) {
	tests := []struct {
		name  string
		input *Tree
		want  string
	}{
		{
			name:  "",
			input: makeEmpty(),
			want:  "",
		},
		{
			name:  "no comment",
			input: makeNoComment(),
			want:  "SELECT * FROM person WHERE employee_no < 1000 AND dept_no = 1",
		},
		{
			name:  "if",
			input: makeTreeIf(),
			want:  "SELECT * FROM person WHERE employee_no < 1000 AND dept_no = 1",
		},
		{
			name:  "if and bind",
			input: makeTreeIfBind(),
			want:  "SELECT * FROM person WHERE employee_no < ?/*maxEmpNo*/",
		},
		{
			name:  "if elif else",
			input: makeIfElifElse(),
			want:  "SELECT * FROM person WHERE employee_no < 1000 AND dept_no =1",
		},
		{
			name:  "if nest",
			input: makeIfNest(),
			want:  "SELECT * FROM person WHERE employee_no < 1000 AND id=3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := gen(tt.input, map[string]interface{}{}); err != nil || got != tt.want {
				if err != nil {
					t.Error(err)
				}
				if got != tt.want {
					t.Errorf("Doesn't Match:\nexpected: \n%v\n but got: \n%v\n", tt.want, got)
				}
			}
		})
	}
}
