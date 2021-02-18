package twowaysql

import (
	"fmt"
	"testing"
)

//　課題: テストケースの準備(Treeの作成)に手間がかかる
func TestAst(t *testing.T) {
	tests := []struct {
		name  string
		input []Token
		want  []*Tree
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
				{
					kind: TkEndOfProgram,
				},
			},
			want: makeTreesif(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := ast(tt.input); err != nil || !treesEqual(tt.want, got) {
				if err != nil {
					t.Error(err)
				}
				t.Errorf("Doesn't Match expected: %v, but got: %v\n", tt.want, got)
				fmt.Println("want:")
				for _, tree := range tt.want {
					printTree(tree)
				}
				fmt.Println("got:")
				for _, tree := range got {
					printTree(tree)
				}
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

func treesEqual(ts1, ts2 []*Tree) bool {
	if len(ts1) != len(ts2) {
		return false
	}
	for i := 0; i < len(ts1); i++ {
		ch1 := make(chan *Token)
		ch2 := make(chan *Token)

		go walk(ts1[i], ch1)
		go walk(ts2[i], ch2)

		var s1, s2 []*Token

		for n := range ch1 {
			s1 = append(s1, n)
		}
		for n := range ch2 {
			s2 = append(s2, n)
		}

		if len(s1) == len(s2) {
			for i := 0; i < len(s1); i++ {
				if *s1[i] != *s2[i] {
					fmt.Printf("not macth %v != %v\n", s1[i], s2[i])
					return false
				}
			}
		}
	}
	return true
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

func makeTreesif() []*Tree {

	NdEndOfProgram1 := Tree{
		Kind: NdEndOfProgram,
		Token: &Token{
			kind: TkEndOfProgram,
		},
	}

	NdEnd1 := Tree{
		Kind: NdEnd,
		Token: &Token{
			kind: TkEnd,
			str:  "/* END */",
		},
	}

	NdSQLstmt2 := Tree{
		Kind: NdSQLstmt,
		Token: &Token{
			kind: TkSQLStmt,
			str:  " AND dept_no = 1",
		},
	}

	NdIf1 := Tree{
		Kind:  NdIf,
		Left:  &NdSQLstmt2,
		Right: &NdEnd1,
		Token: &Token{
			kind: TkIf,
			str:  "/* IF true */",
		},
	}

	NdSQLstmt1 := Tree{
		Kind: NdSQLstmt,
		Token: &Token{
			kind: TkSQLStmt,
			str:  "SELECT * FROM person WHERE employee_no < 1000 ",
		},
	}

	var trees []*Tree
	trees = append(trees, &NdSQLstmt1, &NdIf1, &NdEndOfProgram1)

	return trees
}
