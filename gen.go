package twowaysql

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
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
		truth, err := evalCondition(removeCommentSymbol(node.Token.str), params, kind)
		if err != nil {
			return "", err
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
// kindはNdIfかNdElifでなくてはならない(呼び出し側の制約)
// 現状は/* If condition */のconditionがtruthyかどうか判別している。
// notに対応した方がいいだろうか?
func evalCondition(str string, params map[string]interface{}, kind NodeKind) (bool, error) {
	//テスト用
	if strings.Contains(str, "true") {
		return true, nil
	}
	if strings.Contains(str, "false") {
		return false, nil
	}
	var val string
	switch kind {
	case NdIf:
		val = retrieveValueFromIf(str)
	case NdElif:
		val = retrieveValueFromElif(str)
	default:
		panic("kind must be NdIf or NdElif")
	}
	//log.Println("val:", val)
	if elem, ok := params[val]; ok {
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
