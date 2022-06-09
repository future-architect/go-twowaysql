package twowaysql

import (
	"bytes"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"gitlab.com/osaki-lab/tagscanner/runtimescan"
)

// Eval returns converted query and bind value.
// inputParams takes a tagged struct. Tags must be in the form `map:"tag_name"`.
// The return value is expected to be used to issue queries to the database
func Eval(inputQuery string, inputParams interface{}) (string, []interface{}, error) {
	mapParams := map[string]interface{}{}

	if inputParams != nil {
		if err := encode(mapParams, inputParams); err != nil {
			return "", nil, err
		}
	} else {
		mapParams = nil
	}

	tokens, err := tokenize(inputQuery)
	if err != nil {
		return "", nil, err
	}

	generatedTokens, err := parseCondition(tokens, mapParams)
	if err != nil {
		return "", nil, err
	}

	convertedQuery, params, err := build(generatedTokens, mapParams)
	if err != nil {
		return "", nil, err
	}

	return arrangeWhiteSpace(convertedQuery), params, nil
}

func build(tokens []token, inputParams map[string]interface{}) (string, []interface{}, error) {
	var b strings.Builder
	params := make([]interface{}, 0, len(tokens))

	for _, token := range tokens {
		if token.kind == tkBind {
			if elem, ok := inputParams[token.value]; ok {
				switch elemTyp := elem.(type) {
				case []string:
					token.str = bindLiterals(token.str, len(elemTyp))
					for _, value := range elemTyp {
						params = append(params, value)
					}
				case []int:
					token.str = bindLiterals(token.str, len(elemTyp))
					for _, value := range elemTyp {
						params = append(params, value)
					}
				case [][]interface{}:
					token.str = bindTable(token.str, len(elemTyp), len(elemTyp[0]))
					for _, rows := range elemTyp {
						for _, columns := range rows {
							params = append(params, columns)
						}
					}
				default:
					params = append(params, elem)
				}
			} else {
				return "", nil, fmt.Errorf("no parameter that matches the bind value: %s", token.value)
			}
		}
		b.WriteString(token.str)
	}
	return b.String(), params, nil
}

// ?/* ... */ -> (?, ?, ?)/* ... */みたいにする
func bindLiterals(str string, number int) string {
	str = strings.TrimLeftFunc(str, func(r rune) bool {
		return r != unicode.SimpleFold('/')
	})
	var b strings.Builder
	b.WriteRune('(')
	for i := 0; i < number; i++ {
		b.WriteRune('?')
		if i != number-1 {
			b.WriteString(", ")
		}
	}
	b.WriteRune(')')

	return fmt.Sprint(b.String(), str)
}

func bindTable(str string, rowNumber, columnNumber int) string {
	str = strings.TrimLeftFunc(str, func(r rune) bool {
		return r != unicode.SimpleFold('/')
	})

	var column strings.Builder
	column.WriteRune('(')
	for i := 0; i < columnNumber; i++ {
		column.WriteRune('?')
		if i != columnNumber-1 {
			column.WriteString(", ")
		}
	}
	column.WriteRune(')')

	var row strings.Builder
	row.WriteRune('(')
	for i := 0; i < rowNumber; i++ {
		row.WriteString(column.String())
		if i != rowNumber-1 {
			row.WriteString(", ")
		}
	}
	row.WriteRune(')')

	return fmt.Sprint(row.String(), str)
}

// 空白が二つ以上続いていたら一つにする。=1 -> = 1のような変換はできない
// 単純な空白を想定。 -> issue: よりロバストな実装
func arrangeWhiteSpace(str string) string {
	ret := ""
	buff := bytes.NewBufferString(ret)
	for i := 0; i < len(str); i++ {
		if i < len(str)-1 && str[i] == ' ' && str[i+1] == ' ' {
			continue
		}
		buff.WriteByte(str[i])
	}
	ret = buff.String()
	return strings.Trim(ret, " ")
}

type encoder struct {
	dest map[string]interface{}
}

func (m encoder) ParseTag(name, tagKey, tagStr, pathStr string, elemType reflect.Type) (tag interface{}, err error) {
	return runtimescan.BasicParseTag(name, tagKey, tagStr, pathStr, elemType)
}

func (m *encoder) VisitField(tag, value interface{}) (err error) {
	t := tag.(*runtimescan.BasicTag)
	m.dest[t.Tag] = value
	return nil
}

func (m encoder) EnterChild(tag interface{}) (err error) {
	return nil
}

func (m encoder) LeaveChild(tag interface{}) (err error) {
	return nil
}

func encode(dest map[string]interface{}, src interface{}) error {
	v := reflect.ValueOf(src)

	if v.Kind() == reflect.Pointer && v.Elem().Kind() == reflect.Map {
		if convertToMapStringAny(v.Elem(), dest) {
			return nil
		}
	}
	if v.Kind() == reflect.Map {
		if convertToMapStringAny(v, dest) {
			return nil
		}
	}

	tags := []string{"twowaysql", "db"}
	if err := runtimescan.Encode(src, tags, &encoder{
		dest: dest,
	}); err != nil {
		return err
	}

	// tagscanner does not support sql.NullXXX type.
	encodeSQLNullTyp(src, dest, tags)

	return nil
}

func convertToMapStringAny(mp reflect.Value, dest map[string]interface{}) bool {
	if mp.Type().Key().Kind() != reflect.String {
		return false
	}
	for _, k := range mp.MapKeys() {
		dest[k.String()] = mp.MapIndex(k).Interface()
	}
	return true
}

func encodeSQLNullTyp(src interface{}, dest map[string]interface{}, tags []string) {
	const targetPkg = "database/sql"
	srcFieldTyps := reflect.ValueOf(src).Type().Elem()
	srcFieldValues := reflect.ValueOf(src).Elem()
	for i := 0; i < srcFieldTyps.NumField(); i++ {
		srcFieldTyp := srcFieldTyps.Field(i)
		if srcFieldTyp.Type.PkgPath() != targetPkg {
			continue
		}
		tagValue := getTagValue(srcFieldTyp.Tag, tags)
		if tagValue == "" {
			continue
		}
		switch v := srcFieldValues.Field(i).Interface().(type) {
		case sql.NullBool:
			if v.Valid {
				dest[tagValue] = v.Bool
			} else {
				dest[tagValue] = nil
			}
		case sql.NullByte:
			// not support
			continue
		case sql.NullFloat64:
			if v.Valid {
				dest[tagValue] = v.Float64
			} else {
				dest[tagValue] = nil
			}
		case sql.NullInt16:
			if v.Valid {
				dest[tagValue] = v.Int16
			} else {
				dest[tagValue] = nil
			}
		case sql.NullInt32:
			if v.Valid {
				dest[tagValue] = v.Int32
			} else {
				dest[tagValue] = nil
			}
		case sql.NullInt64:
			if v.Valid {
				dest[tagValue] = v.Int64
			} else {
				dest[tagValue] = nil
			}
		case sql.NullString:
			if v.Valid {
				dest[tagValue] = v.String
			} else {
				dest[tagValue] = nil
			}
		case sql.NullTime:
			if v.Valid {
				dest[tagValue] = v.Time
			} else {
				dest[tagValue] = nil
			}
		}
	}
}

func getTagValue(structTag reflect.StructTag, targetTags []string) string {
	for _, t := range targetTags {
		tag := structTag.Get(t)
		if tag != "" {
			return tag
		}
	}
	return ""
}
