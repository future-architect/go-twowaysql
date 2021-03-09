package twowaysql

import (
	"github.com/robertkrimen/otto"
)

// 抽象構文木からトークン列を生成
// 左部分木、右部分木と辿る
// 現状右部分木を持つのはif, elif, elseだけ?
func parse(trees *tree, params map[string]interface{}) ([]token, error) {
	res, err := genInner(trees, params)
	if err != nil {
		return []token{}, err
	}
	return res, nil
}

func genInner(node *tree, params map[string]interface{}) ([]token, error) {
	if node == nil {
		return []token{}, nil
	}

	//行きがけ

	//左部分木に行く
	leftStr, err := genInner(node.Left, params)
	if err != nil {
		return []token{}, err
	}

	//左部分木から戻ってきた

	//右部分木に行く
	rightStr, err := genInner(node.Right, params)
	if err != nil {
		return []token{}, err
	}

	//右部分木から戻ってきた
	// 何を返すか
	// 基本的に左部分木
	// If Elifの場合は条件次第
	switch kind := node.Kind; kind {
	case ndSQLStmt, ndBind:
		//めちゃめちゃ実行効率悪い気が...
		return append([]token{*node.Token}, leftStr...), nil
	case ndIf, ndElif:
		truth, err := evalCondition(node.Token.condition, params)
		if err != nil {
			return []token{}, err
		}
		if truth {
			return leftStr, nil
		}
		return rightStr, nil
	default:
		return leftStr, nil
	}
}

// /* If ... */ /* Elif ... */の条件を評価する
// TODO: 式言語?に対応する
func evalCondition(condition string, params map[string]interface{}) (bool, error) {
	vm := otto.New()
	for key, value := range params {
		err := vm.Set(key, value)
		if err != nil {
			return false, err
		}
	}
	result, err := vm.Run(condition)
	if err != nil {
		return false, err
	}
	truth, err := result.ToBoolean()
	if err != nil {
		return false, err
	}
	return truth, nil
}
