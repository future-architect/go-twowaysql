package twowaysql

import (
	"reflect"

	"github.com/robertkrimen/otto"
)

// 抽象構文木からトークン列を生成
// 左部分木、右部分木と辿る
// 現状右部分木を持つのはif, elif, elseだけ?
func gen(trees *tree, params map[string]interface{}) ([]token, error) {
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
// 現状は/* If condition */のconditionがtruthyかどうか判別している。
// notに対応した方がいいだろうか?
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
