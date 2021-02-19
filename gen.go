package twowaysql

import "bytes"

// 抽象構文木から目標文字列を生成
// 欲しいサブ関数: 空白調整、if,elifの評価、
// バインド抽出は別のパスにする
// 左部分木、右部分木と辿る
// この設計は適切なのだろうか?
// 右部分木を持つのはif, elif, elseだけ?
func gen(trees *Tree) (string, error) {
	res, err := genInner(trees, "")
	if err != nil {
		return "", err
	}
	return arrageWhiteSpace(res), nil
}

func genInner(node *Tree, unsettled string) (string, error) {
	if node == nil {
		return "", nil
	}

	//行きがけ

	//左部分木に行く
	leftStr, err := genInner(node.Left, unsettled)
	if err != nil {
		return "", err
	}

	// 戻ってきた

	//右部分木に行く
	rightStr, err := genInner(node.Right, unsettled)
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
		if evalCondition(removeCommentSymbol(node.Token.str)) {
			return leftStr, nil
		}
		return rightStr, nil
	default:
		return leftStr, nil
	}
}

// /*value*/1000 -> ?/*value*/ みたいに変換する
func bindConvert(str string) string {
	return str
}

// /* If ... */ /* Elif ... */の条件を評価する
func evalCondition(str string) bool {
	return true
}

// 空白が二つ以上続いていたら一つにする。=1 -> = 1のような変換はできないる
// 単純な空白を想定。 -> issue よりロバストな実装
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
	return buff.String()
}

// /* */記号の削除
func removeCommentSymbol(str string) string {
	return str
}
