package twowaysql

import (
	"context"
	"database/sql"
	"errors"
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

func (t *Twowaysql) generateQueryAndBindValue() (string, []interface{}, error) {

	st, err := t.parse()
	if err != nil {
		return "", nil, err
	}

	//ユーザがどんなクエリに変更されたかが見えるようにするために代入する
	t.convertedQuery = st.query

	var bindParams []interface{}
	for _, bind := range st.bindsValue {
		if elem, ok := t.params[bind]; ok {
			bindParams = append(bindParams, elem)
		} else {
			return "", nil, errors.New("no parameter that matches the bind value")
		}
	}

	//適したplace holderに変換
	st.query = t.db.Rebind(st.query)

	return st.query, bindParams, nil

}

// 事前条件: inputStructのフィールドとqueryで返ってくる要素の長さと並びは一致していなければならない。
func (t *Twowaysql) SelectContext(ctx context.Context, inputStructs interface{}, query string, params map[string]interface{}) error {
	t = t.withParams(params).withQuery(query)

	convertedQuery, bindParams, err := t.generateQueryAndBindValue()
	if err != nil {
		return err
	}

	//log.Println("query", convertedQuery)
	//log.Println(bindParams...)
	rows, err := t.db.QueryxContext(ctx, convertedQuery, bindParams...)
	if err != nil {
		return err
	}

	//fmt.Println("RV", reflect.ValueOf(inputStructs).Elem())
	rv := reflect.ValueOf(inputStructs).Elem()

	for rows.Next() {
		structPtr, err := runtimescan.NewStructInstance(inputStructs)
		if err != nil {
			return err
		}
		rows.StructScan(structPtr)
		//fmt.Println("Person", structPtr)
		rv.Set(reflect.Append(rv, reflect.ValueOf(structPtr).Elem()))
	}

	return nil

}

// 事前条件: inputStructのフィールドとqueryで返ってくる要素の長さと並びは一致していなければならない。
func (t *Twowaysql) Select(inputStructs interface{}, query string, params map[string]interface{}) error {
	t = t.withParams(params).withQuery(query)

	convertedQuery, bindParams, err := t.generateQueryAndBindValue()
	if err != nil {
		return err
	}

	//log.Println("query", convertedQuery)
	//log.Println(bindParams...)
	rows, err := t.db.Queryx(convertedQuery, bindParams...)
	if err != nil {
		return err
	}

	//fmt.Println("RV", reflect.ValueOf(inputStructs).Elem())
	rv := reflect.ValueOf(inputStructs).Elem()

	for rows.Next() {

		structPtr, err := runtimescan.NewStructInstance(inputStructs)
		if err != nil {
			return err
		}
		rows.StructScan(structPtr)
		//fmt.Println("Person", structPtr)
		rv.Set(reflect.Append(rv, reflect.ValueOf(structPtr).Elem()))
	}

	return nil

}

func (t *Twowaysql) Exec(inputStructs interface{}, query string, params map[string]interface{}) (sql.Result, error) {
	t = t.withParams(params).withQuery(query)

	convertedQuery, bindParams, err := t.generateQueryAndBindValue()
	if err != nil {
		return nil, err
	}

	return t.db.Exec(convertedQuery, bindParams...)
}

func (t *Twowaysql) ExecContext(ctx context.Context, inputStructs interface{}, query string, params map[string]interface{}) (sql.Result, error) {
	t = t.withParams(params).withQuery(query)

	convertedQuery, bindParams, err := t.generateQueryAndBindValue()
	if err != nil {
		return nil, err
	}

	return t.db.ExecContext(ctx, convertedQuery, bindParams...)
}

// ?/* ... */ を $1/* ... */のような形に変換する。
func convertPlaceHolder(str string) string {
	count := strings.Count(str, "?")
	for i := 0; i < count; i++ {
		str = strings.Replace(str, "?", "$"+strconv.Itoa(i+1), 1)
	}
	return str
}
