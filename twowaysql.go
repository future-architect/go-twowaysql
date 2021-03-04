package twowaysql

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"gitlab.com/osaki-lab/tagscanner/runtimescan"
)

type Person struct {
	FirstName string `db:"first_name"`
	LastName  string `db:"last_name"`
	Email     string `db:"email"`
}

// inputStructは本来structだが現時点ではpeople決め打ちで書く
type Twowaysql struct {
	db             *sqlx.DB
	query          string
	convertedQuery string
	params         map[string]interface{}
}

func New(db *sqlx.DB) *Twowaysql {
	return &Twowaysql{
		db: db,
	}
}

// ConvertedQuery returns convertedQuery property
func (t *Twowaysql) ConvertedQuery() string {
	return t.convertedQuery
}

func (t *Twowaysql) withQuery(query string) *Twowaysql {
	return &Twowaysql{
		db:             t.db,
		query:          query,
		convertedQuery: t.convertedQuery,
		params:         t.params,
	}
}

func (t *Twowaysql) withParams(params map[string]interface{}) *Twowaysql {
	return &Twowaysql{
		db:             t.db,
		query:          t.query,
		convertedQuery: t.convertedQuery,
		params:         params,
	}
}

// 事前条件: inputStructのフィールドとqueryで返ってくる要素の長さと並びは一致していなければならない。
func (t *Twowaysql) SelectContext(ctx context.Context, inputStructs interface{}, query string, params map[string]interface{}) error {
	t = t.withParams(params).withQuery(query)

	st, err := t.parse()
	if err != nil {
		return err
	}

	//ユーザがどんなクエリに変更されたかが見えるようにするために代入する
	t.convertedQuery = st.query

	var bindParams []interface{}
	for _, bind := range st.bindsValue {
		if elem, ok := t.params[bind]; ok {
			bindParams = append(bindParams, elem)
		} else {
			return errors.New("no parameter that matches the bind value")
		}
	}

	//一時的な措置、本当はどこかでdatabaseのtypeを知る必要がある。
	postgres := true
	if postgres {
		st.query = convertPlaceHolder(st.query)
	}

	//log.Println("query", convertedQuery)
	//log.Println(params...)
	rows, err := t.db.QueryxContext(ctx, st.query, bindParams...)
	if err != nil {
		return err
	}

	fmt.Println("TYPE", reflect.TypeOf(inputStructs).Elem())
	fmt.Println("RV", reflect.ValueOf(inputStructs).Elem())
	rv := reflect.ValueOf(inputStructs).Elem()
	//rv := reflect.MakeSlice(reflect.TypeOf(inputStructs).Elem(), 0, 0)
	for rows.Next() {
		results := make(map[string]interface{})
		err = rows.MapScan(results)
		if err != nil {
			return err
		}
		for key, value := range results {
			fmt.Println("KEY:", key)
			//データベースから返ってくる値の表現が分からない
			//どんな感じにinterfaceを処理するか?
			switch value.(type) {
			case []uint8:
				fmt.Printf("Value: %s\n", value)
				str := fmt.Sprintf("%s", value)
				results[key] = strings.TrimRight(str, " ")
			default:
				fmt.Printf("Value: %v\n", value)
			}
		}

		/*
			structType := reflect.TypeOf(t.inputStructs).Elem()
			fmt.Println("Type", structType)
			structValue := reflect.New(structType).Elem()
			fmt.Println("Value", structValue)
		*/
		//structPtr := reflect.New(structType).Elem().Interface()
		//p := t.inputStructs.(Person)
		/*
			rt := structValue.Type()
			for i := 0; i < rt.NumField(); i++ {
				// フィールドの取得
				f := rt.Field(i)
				// フィールド名
				fmt.Println(f.Name)
				// 型
				fmt.Println(f.Type)
				// タグ
				fmt.Println(f.Tag)
			}
		*/
		/*
			var p Person
			var pp interface{}
			pp = p
		*/
		//pp := []Person{}
		structValue, err := runtimescan.NewStructInstance(inputStructs)
		if err != nil {
			return err
		}
		err = Decode(structValue, results)
		if err != nil {
			return err
		}
		fmt.Println("Person", structValue)
		rv.Set(reflect.Append(rv, reflect.ValueOf(structValue).Elem()))
	}

	fmt.Println("RV", rv)
	inputStructs = rv

	return nil

}

// User should implement runtimescan.Decoder interface
// This instance is created in user code before runtimescan.Decode() function call
type decoder struct {
	src map[string]interface{}
}

func (m decoder) ParseTag(name, tagStr, pathStr string, elemType reflect.Type) (tag interface{}, err error) {
	return runtimescan.BasicParseTag(name, tagStr, pathStr, elemType)
}

func (m *decoder) ExtractValue(tag interface{}) (value interface{}, err error) {
	t := tag.(*runtimescan.BasicTag)
	v, ok := m.src[t.Tag]
	if !ok {
		return nil, runtimescan.Skip
	}
	return v, nil
}

func Decode(dest interface{}, src map[string]interface{}) error {
	dec := &decoder{
		src: src,
	}
	return runtimescan.Decode(dest, "db", dec)
}

// ?/* ... */ を $1/* ... */のような形に変換する。
func convertPlaceHolder(str string) string {
	count := strings.Count(str, "?")
	for i := 0; i < count; i++ {
		str = strings.Replace(str, "?", "$"+strconv.Itoa(i+1), 1)
	}
	return str
}
