package twowaysql

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
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

// SelectContext is a thin wrapper around db.SelectContext in the sqlx package.
// 事前条件: inputStructのフィールドとqueryで返ってくる要素の長さと並びは一致していなければならない。
func (t *Twowaysql) SelectContext(ctx context.Context, inputStructs interface{}, query string, params map[string]interface{}) error {
	t = t.withParams(params).withQuery(query)

	convertedQuery, bindParams, err := t.generateQueryAndBindValue()
	if err != nil {
		return err
	}

	rv := reflect.ValueOf(inputStructs).Interface()
	return t.db.SelectContext(ctx, rv, convertedQuery, bindParams...)

}

// Select is a thin wrapper around db.Select in the sqlx package.
// 事前条件: inputStructのフィールドとqueryで返ってくる要素の長さと並びは一致していなければならない。
func (t *Twowaysql) Select(inputStructs interface{}, query string, params map[string]interface{}) error {
	t = t.withParams(params).withQuery(query)

	convertedQuery, bindParams, err := t.generateQueryAndBindValue()
	if err != nil {
		return err
	}

	rv := reflect.ValueOf(inputStructs).Interface()
	return t.db.Select(rv, convertedQuery, bindParams...)
	//fmt.Println("rv", rv)

}

// Exec is a thin wrapper around db.Exec in the sqlx package.
func (t *Twowaysql) Exec(inputStructs interface{}, query string, params map[string]interface{}) (sql.Result, error) {
	t = t.withParams(params).withQuery(query)

	convertedQuery, bindParams, err := t.generateQueryAndBindValue()
	if err != nil {
		return nil, err
	}

	return t.db.Exec(convertedQuery, bindParams...)
}

// ExecContext is a thin wrapper around db.ExecContext in the sqlx package.
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
