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
			name:  "",
			input: "SELECT * FROM person WHERE employee_no < 1000 /* IF true */ AND dept_no = 1",
			want: []Token{
				{
					kind: TkSQLStmt,
					str:  "SELECT * FROM person WHERE employee_no ",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tokinize(tt.input); !tokensEqual(tt.want, got) {
				t.Errorf("Doesn't Match\n expected: %v, but got: %v\n", tt.want, got)
			}
		})
	}
}

func TestTry(t *testing.T) {
	//t.Errorf("%v", TkIf)
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
