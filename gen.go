package twowaysql

import (
	"bytes"
	"strings"
	"unicode"
)

// 抽象構文木から目標文字列を生成
// バインド抽出は別のパスにする
// 左部分木、右部分木と辿る
// 現状右部分木を持つのはif, elif, elseだけ?
func gen(trees *Tree, params map[string]interface{}) (string, error) {
	res, err := genInner(trees, params)
	if err != nil {
		return "", err
	}
	return arrageWhiteSpace(res), nil
}

func genInner(node *Tree, params map[string]interface{}) (string, error) {
	if node == nil {
		return "", nil
	}

	//行きがけ

	//左部分木に行く
	leftStr, err := genInner(node.Left, params)
	if err != nil {
		return "", err
	}

	// 戻ってきた

	//右部分木に行く
	rightStr, err := genInner(node.Right, params)
	if err != nil {
		return "", err
	}

	// 何を返すか
	// 基本的に左部分木
	// If Elifの場合は条件次第
	switch kind := node.Kind; kind {
	case NdSQLStmt:
		return node.Token.str + leftStr, nil
	case NdBind:
		return bindConvert(node.Token.str) + leftStr, nil
	case NdIf, NdElif:
		if evalCondition(removeCommentSymbol(node.Token.str), params) {
			return leftStr, nil
		}
		return rightStr, nil
	default:
		return leftStr, nil
	}
}

// /*value*/1000 -> ?/*value*/ みたいに変換する
// ここではやらない
func bindConvert(str string) string {
	str = strings.TrimRightFunc(str, func(r rune) bool {
		return r != unicode.SimpleFold('/')
	})
	str = "?" + str
	return str
}

// /* If ... */ /* Elif ... */の条件を評価する
// 取り敢えずgenの動作を見るための仮実装
// TODO: 式言語?に対応する
// if exsits(deptNo)などはdepthNoにアクセスできなくてはならない。
// 将来的には構造体を作る必要がある。tokenize, ast, genはそのメソッドとなる。
func evalCondition(str string, params map[string]interface{}) bool {
	return strings.Contains(str, "true")
}

// 空白が二つ以上続いていたら一つにする。=1 -> = 1のような変換はできない
// 単純な空白を想定。 -> issue: よりロバストな実装
func arrageWhiteSpace(str string) string {
	ret := ""
	buff := bytes.NewBufferString(ret)
	for i := 0; i < len(str); i++ {
		if i < len(str)-1 && str[i] == ' ' && str[i+1] == ' ' {
			//do nothing
		} else {
			buff.WriteByte(str[i])
		}
	}
	ret = buff.String()
	ret = strings.TrimLeft(ret, " ")
	ret = strings.TrimRight(ret, " ")
	return ret
}

// /* */記号の削除
func removeCommentSymbol(str string) string {
	str = strings.TrimPrefix(str, "/*")
	str = strings.TrimSuffix(str, "*/")
	return str
}
