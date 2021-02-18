package twowaysql

import (
	"errors"
	"fmt"
)

type NodeKind int

const (
	NdSQLstmt NodeKind = iota + 1
	NdBind
	NdIf
	NdElif
	NdElse
	NdEnd
	NdEndOfProgram
)

type Tree struct {
	Kind  NodeKind
	Left  *Tree
	Right *Tree
	Token *Token
}

// astはトークン列から抽象構文木を生成する。
// 生成規則: （自信はない)
// program = stmt*	EndOfpragram
// stmt = 	SQLStmt |
//			BIND	|
//		  	"IF" stmt ("ELLF" stmt)* ("ELSE" stmt)? "END"
//
func ast(tokens []Token) ([]*Tree, error) {
	node, err := program(tokens)
	if err != nil {
		return nil, err
	}

	return node, nil
}

// channelで送るようにする?
func program(tokens []Token) ([]*Tree, error) {
	index := 0
	var trees []*Tree
	var node *Tree
	var err error

	for {
		if consume(tokens, &index, TokenKind(NdEndOfProgram)) {
			node = &Tree{
				Kind:  NdEndOfProgram,
				Token: &tokens[index-1],
			}
			trees = append(trees, node)
			return trees, nil
		}
		node, err = stmt(tokens, &index)
		if err != nil {
			return nil, err
		}
		trees = append(trees, node)
	}
}

// token index token[index]を見ている
// indexの操作が非常に煩雑でerror prone
func stmt(tokens []Token, index *int) (*Tree, error) {
	var node *Tree
	if consume(tokens, index, TkSQLStmt) {
		node = &Tree{
			Kind:  NdSQLstmt,
			Token: &tokens[*index-1],
		}
		return node, nil
	} else if consume(tokens, index, TkBind) {
		node = &Tree{
			Kind:  NdBind,
			Token: &tokens[*index-1],
		}
		return node, nil
	} else if consume(tokens, index, TkIf) {
		var err error
		node = &Tree{
			Kind:  NdIf,
			Token: &tokens[*index-1],
		}
		node.Left, err = stmt(tokens, index)
		if err != nil {
			return nil, err
		}
		tmpNode := node
		for {
			if consume(tokens, index, TkElif) {
				child := &Tree{
					Kind:  NdElif,
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
		if consume(tokens, index, TkElse) {
			child := &Tree{
				Kind:  NdElse,
				Token: &tokens[*index-1],
			}
			tmpNode.Right = child
			tmpNode = child

			child.Left, err = stmt(tokens, index)
			if err != nil {
				return nil, err
			}
		}

		if consume(tokens, index, TkEnd) {
			child := &Tree{
				Kind:  NdEnd,
				Token: &tokens[*index-1],
			}
			tmpNode.Right = child
		} else {
			return nil, fmt.Errorf("can not parse: expected /* END */, but got %v", tokens[*index].kind)
		}
		return node, nil
	} else {
		return nil, errors.New("can not parse")
	}
}

func consume(tokens []Token, index *int, kind TokenKind) bool {
	if tokens[*index].kind == kind {
		*index++
		return true
	}
	return false
}
