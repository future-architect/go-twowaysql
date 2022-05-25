package twowaysql

import (
	"errors"
	"fmt"
)

type nodeKind int

const (
	ndSQLStmt nodeKind = iota + 1
	ndBind
	ndIf
	ndElif
	ndElse
	ndEnd
	ndEndOfProgram
)

// tree is a component of an abstract syntax tree
type tree struct {
	Kind  nodeKind
	Left  *tree
	Right *tree
	Token *token
}

// astはトークン列から抽象構文木を生成する。
// 生成規則:
// program = stmt
// stmt = 	SQLStmt stmt |
//			BIND	stmt |
//		  	"IF" stmt ("ELLF" stmt)* ("ELSE" stmt)? "END" stmt |
//			EndOfProgram
//
func ast(tokens []token) (*tree, error) {
	node, err := program(tokens)
	if err != nil {
		return nil, err
	}
	if node.nodeCount() != len(tokens) {
		return nil, errors.New("can not generate abstract syntax tree")
	}

	return node, nil
}

func program(tokens []token) (*tree, error) {
	index := 0
	return stmt(tokens, &index)
}

// token index token[index]を見ている
func stmt(tokens []token, index *int) (*tree, error) {
	var node *tree
	var err error

	if consume(tokens, index, tkSQLStmt) {
		// SQLStmt stmt
		node = &tree{
			Kind:  ndSQLStmt,
			Token: &tokens[*index-1],
		}

		node.Left, err = stmt(tokens, index)
		if err != nil {
			return nil, err
		}

	} else if consume(tokens, index, tkBind) {
		// Bind stmt
		node = &tree{
			Kind:  ndBind,
			Token: &tokens[*index-1],
		}

		node.Left, err = stmt(tokens, index)
		if err != nil {
			return nil, err
		}
	} else if consume(tokens, index, tkEndOfProgram) {
		// EndOfProgram
		node = &tree{
			Kind: nodeKind(tkEndOfProgram),
			// consumeはTkEndOfProgramの時はインクリメントしないから1を引かない
			// かなりよくない設計、一貫性がない。
			Token: &tokens[*index],
		}
		return node, nil

	} else if consume(tokens, index, tkIf) {
		// "IF" stmt ("ELLF" stmt)* ("ELSE" stmt)? "END" stmt
		node = &tree{
			Kind:  ndIf,
			Token: &tokens[*index-1],
		}
		node.Left, err = stmt(tokens, index)
		if err != nil {
			return nil, err
		}
		tmpNode := node
		for {
			// ("ELLF" stmt)*
			if consume(tokens, index, tkElif) {
				child := &tree{
					Kind:  ndElif,
					Token: &tokens[*index-1],
				}
				tmpNode.Right = child
				tmpNode = child

				child.Left, err = stmt(tokens, index)
				if err != nil {
					return nil, err
				}
				continue
			}
			break
		}

		if consume(tokens, index, tkElse) {
			// ("ELSE" stmt)?
			child := &tree{
				Kind:  ndElse,
				Token: &tokens[*index-1],
			}
			tmpNode.Right = child
			tmpNode = child

			child.Left, err = stmt(tokens, index)
			if err != nil {
				return nil, err
			}
		}

		if consume(tokens, index, tkEnd) {
			// "END"
			child := &tree{
				Kind:  ndEnd,
				Token: &tokens[*index-1],
			}
			tmpNode.Right = child

			child.Left, err = stmt(tokens, index)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("can not parse: expected /* END */, but got %v", tokens[*index].kind)
		}

		// どれも一致しなかった
		return node, nil
	}
	return node, nil
}

// tokenが所望のものか調べる。一致していればインデックスを一つ進める
func consume(tokens []token, index *int, kind tokenKind) bool {
	//println("str: ", tokens[*index].str, "kind: ", tokens[*index].kind, "want kind: ", kind)
	if tokens[*index].kind == kind {
		// TkEndOfProgramでインクリメントしてしまうと
		// その後のconsume呼び出しでIndex Out Of Bounds例外が発生してしまう
		if kind != tkEndOfProgram {
			*index++
		}
		return true
	}
	return false
}

func (t *tree) nodeCount() int {
	count := 1
	if t.Left != nil {
		t.Left.countInner(&count)
	}
	if t.Right != nil {
		t.Right.countInner(&count)
	}
	return count
}

func (t *tree) countInner(count *int) {
	*count++
	if t.Left != nil {
		t.Left.countInner(count)
	}
	if t.Right != nil {
		t.Right.countInner(count)
	}
}
