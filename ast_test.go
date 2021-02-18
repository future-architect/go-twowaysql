package twowaysql

import (
	"fmt"
	"testing"
)

func TestAst(t *testing.T) {
	tests := []struct {
		name  string
		input []Token
		want  *Tree
	}{
		{
			name: "if",
			input: []Token{
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
			},
			want: makeTreeif(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := ast(tt.input); err != nil || !treeEqual(tt.want, got) {
				if err != nil {
					t.Error(err)
				}
				t.Errorf("Doesn't Match expected: %v, but got: %v\n", tt.want, got)
				printTree(tt.want)
				printTree(got)
			}
		})
	}

}

func walk(t *Tree, ch chan *Token) {
	if t == nil {
		return
	}

	walkInner(t.Left, ch)
	ch <- t.Token
	walkInner(t.Right, ch)

	close(ch)
}

func walkInner(t *Tree, ch chan *Token) {
	if t == nil {
		return
	}

	walkInner(t.Left, ch)
	ch <- t.Token
	walkInner(t.Right, ch)
}

func treeEqual(t1, t2 *Tree) bool {
	ch1 := make(chan *Token)
	ch2 := make(chan *Token)

	go walk(t1, ch1)
	go walk(t2, ch2)

	var s1, s2 []*Token

	for n := range ch1 {
		s1 = append(s1, n)
	}
	for n := range ch2 {
		s2 = append(s2, n)
	}

	if len(s1) == len(s2) {
		for i := 0; i < len(s1); i++ {
			if s1[i] != s2[i] {
				return false
			}
		}
		return true
	}

	return false
}

func printTree(t *Tree) {
	if t == nil {
		return
	}

	printWalkInner(t.Left)
	fmt.Println(t.Token)
	printWalkInner(t.Right)

}

func printWalkInner(t *Tree) {
	if t == nil {
		return
	}

	printWalkInner(t.Left)
	fmt.Println(t.Token)
	printWalkInner(t.Right)
}

func makeTreeif() *Tree {
	NdEnd1 := Tree{
		Kind: NdEnd,
		Token: &Token{
			kind: TkEnd,
			str:  "/* END */",
		},
	}

	NdSQLstmt1 := Tree{
		Kind: NdSQLstmt,
		Left: &NdEnd1,
		Token: &Token{
			kind: TkSQLStmt,
			str:  " AND dept_no = 1",
		},
	}

	NdIf1 := Tree{
		Kind: NdIf,
		Left: &NdSQLstmt1,
		Token: &Token{
			kind: TkIf,
			str:  "/* IF true */",
		},
	}

	NdSQLstmt2 := Tree{
		Kind: NdSQLstmt,
		Left: &NdIf1,
		Token: &Token{
			kind: TkSQLStmt,
			str:  "SELECT * FROM person WHERE employee_no < 1000 ",
		},
	}

	return &NdSQLstmt2
}
