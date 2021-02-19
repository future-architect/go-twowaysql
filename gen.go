package twowaysql

// 抽象構文木から目標文字列を生成
// 欲しいサブ関数: 空白調整、if,elifの評価、
// バインド抽出は別のパスにする
// 左部分木、右部分木と辿る
// この設計は適切なのだろうか?
// skipがtrueならappendしない。
func gen(trees *Tree, skip bool) (string, error) {
	if trees == nil {
		return "", nil
	}
	res := ""
	node := trees

	str, err := gen(node.Left, skip)
	if err != nil {
		return "", err
	}
	switch kind := node.Kind; kind {
	case NdSQLStmt:
		res = appendIfNotSkip(res, str, skip)
	case NdBind:
		res = appendIfNotSkip(res, bindConvert(str), skip)
	case NdIf, NdElif:
		if evalCondition(node.Token.str) {
			res = appendIfNotSkip(res, str, skip)
			//ENDが来るまで右部分木をSKIP
			skip = true
		}
	case NdElse:
	case NdEnd:
		skip = false
	case NdEndOfProgram:
	}

	str, err = gen(node.Right, skip)
	if err != nil {
		return "", err
	}

	return res, nil
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
