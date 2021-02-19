package twowaysql

// 抽象構文木から目標文字列を生成
// 欲しいサブ関数: 空白調整、if,elifの評価、
// バインド抽出は別のパスにする
// 左部分木、右部分木と辿る
// この設計は適切なのだろうか?
// 右部分木を持つのはif, elif, elseだけ?
func gen(trees *Tree) (string, error) {
	res, err := genInner(trees, "", false)
	if err != nil {
		return "", err
	}
	return arrageWhiteSpace(res), nil
}

func genInner(node *Tree, unsettled string, skip bool) (string, error) {
	if node == nil {
		return "", nil
	}

	//行きがけ
	/*
		switch kind := node.Kind; kind {
		case NdSQLStmt:
			unsettled += node.Token.str
		case NdBind:
			unsettled += bindConvert(node.Token.str)
		case NdEnd:
			skip = false
		default:
		}
	*/

	//左部分木に行く
	leftStr, err := genInner(node.Left, unsettled, skip)
	if err != nil {
		return "", err
	}
	//fmt.Println("str: ", node.Token.str, "leftStr:", leftStr)

	// 戻ってきた

	//右部分木に行く
	rightStr, err := genInner(node.Right, unsettled, skip)
	if err != nil {
		return "", err
	}

	// IF, ELIFのチェック
	switch kind := node.Kind; kind {
	case NdSQLStmt:
		return node.Token.str + leftStr, nil
	case NdBind:
		return bindConvert(node.Token.str) + leftStr, nil
	case NdIf, NdElif:
		if evalCondition(removeCommentSymbol(node.Token.str)) {
			//fmt.Println("leftStr:", leftStr)
			return leftStr, nil
		} else {
			return rightStr, nil
		}
	default:
	}

	return leftStr, nil
}

func appendIfNotSkip(s1, s2 string, skip bool) string {
	if !skip {
		return s1 + s2
	}
	return s1
}

// /*value*/1000 -> ?/*value*/ みたいに変換する
func bindConvert(str string) string {
	return str
}

// /* If ... */ /* Elif ... */の条件を評価する
func evalCondition(str string) bool {
	return true
}

//空白を調整する
func arrageWhiteSpace(str string) string {
	return str
}

// /* */記号の削除
func removeCommentSymbol(str string) string {
	return str
}
