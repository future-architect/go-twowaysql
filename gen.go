package twowaysql

import (
	"fmt"
	"reflect"
)

// 抽象構文木からトークン列を生成
// 左部分木、右部分木と辿る
// 現状右部分木を持つのはif, elif, elseだけ?
func gen(trees *Tree, params map[string]interface{}) ([]Token, error) {
	res, err := genInner(trees, params)
	if err != nil {
		return []Token{}, err
	}
	return res, nil
}

func genInner(node *Tree, params map[string]interface{}) ([]Token, error) {
	if node == nil {
		return []Token{}, nil
	}

	//行きがけ

	//左部分木に行く
	leftStr, err := genInner(node.Left, params)
	if err != nil {
		return []Token{}, err
	}

	//左部分木から戻ってきた

	//右部分木に行く
	rightStr, err := genInner(node.Right, params)
	if err != nil {
		return []Token{}, err
	}

	//右部分木から戻ってきた
	// 何を返すか
	// 基本的に左部分木
	// If Elifの場合は条件次第
	switch kind := node.Kind; kind {
	case ndSQLStmt, ndBind:
		//めちゃめちゃ実行効率悪い気が...
		return append([]Token{*node.Token}, leftStr...), nil
	case ndIf, ndElif:
		truth, err := evalCondition(node.Token.condition, params)
		if err != nil {
			return []Token{}, err
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
// 現状は/* If condition */のconditionがtruthyかどうか判別している。
// notに対応した方がいいだろうか?
func evalCondition(value string, params map[string]interface{}) (bool, error) {
	//テスト用
	if value == "true" {
		return true, nil
	}
	if value == "false" {
		return false, nil
	}
	var val string
	//log.Println("val:", val)
	if elem, ok := params[value]; ok {
		if truth, ok := isTrue(elem); ok {
			return truth, nil
		}
		return false, fmt.Errorf("IF/ELIF can not use %v", elem)
	}
	return false, fmt.Errorf("invalid condition %v", val)
	//return strings.Contains(str, "true")
}

// IsTrue reports whether the value is 'true', in the sense of not the zero of its type,
// and whether the value has a meaningful truth value. This is the definition of
// truth used by if and other such actions.
// text/templateのexec.goから拝借(いいのか?)
func isTrue(val interface{}) (truth, ok bool) {
	return isTrueInner(reflect.ValueOf(val))
}

func isTrueInner(val reflect.Value) (truth, ok bool) {
	if !val.IsValid() {
		// Something like var x interface{}, never set. It's a form of nil.
		return false, true
	}
	switch val.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		truth = val.Len() > 0
	case reflect.Bool:
		truth = val.Bool()
	case reflect.Complex64, reflect.Complex128:
		truth = val.Complex() != 0
	case reflect.Chan, reflect.Func, reflect.Ptr, reflect.Interface:
		truth = !val.IsNil()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		truth = val.Int() != 0
	case reflect.Float32, reflect.Float64:
		truth = val.Float() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		truth = val.Uint() != 0
	case reflect.Struct:
		truth = true // Struct values are always true.
	default:
		return
	}
	return truth, true
}
