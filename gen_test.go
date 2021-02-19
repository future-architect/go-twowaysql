package twowaysql

import "testing"

func TestGen(t *testing.T) {
	tests := []struct {
		name     string
		input    *Tree
		want     string
		wantBind []string
	}{
		{
			name:     "",
			input:    makeEmpty(),
			want:     "",
			wantBind: nil,
		},
		{
			name:     "no comment",
			input:    makeNoComment(),
			want:     "SELECT * FROM person WHERE employee_no < 1000  AND dept_no = 1",
			wantBind: nil,
		},
		{
			name:     "if",
			input:    makeTreeIf(),
			want:     "SELECT * FROM person WHERE employee_no < 1000  AND dept_no = 1",
			wantBind: nil,
		},
		{
			name:     "if and bind",
			input:    makeTreeIfBind(),
			want:     "SELECT * FROM person WHERE employee_no < ?/*maxEmpNo*/",
			wantBind: []string{"maxEmpNo"},
		},
		{
			name:     "if elif else",
			input:    makeIfElifElse(),
			want:     "SELECT * FROM person WHERE employee_no < 1000 AND dept_no =1",
			wantBind: nil,
		},
		{
			name:     "if nest",
			input:    makeIfNest(),
			want:     "SELECT * FROM person WHERE employee_no < 1000 /* IF true */  AND id=3",
			wantBind: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, gotBind, err := gen(tt.input); err != nil || got != tt.want || sliceEqual(gotBind, tt.wantBind) {
				if err != nil {
					t.Error(err)
				}
				if got != tt.want {
					t.Errorf("Doesn't Match expected: %v, but got: %v\n", tt.want, got)
				}
				if !sliceEqual(gotBind, tt.wantBind) {
					t.Errorf("Bind: Doesn't Match expected: %v, but got: %v\n", tt.wantBind, gotBind)
				}
			}
		})
	}
}

func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}
