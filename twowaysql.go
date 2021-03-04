package twowaysql

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
)

type Twowaysql struct {
	db             *sqlx.DB
	convertedQuery string
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

// GenerateQueryAndBindValue returns converted query and bind value
// The return value is expected to be used to issue queries to the database
func (t *Twowaysql) GenerateQueryAndBindValue(query string, params map[string]interface{}) (string, []interface{}, error) {

	st, err := t.parse(query, params)
	if err != nil {
		return "", nil, err
	}

	//ユーザがどんなクエリに変更されたかが見えるようにするために代入する
	t.convertedQuery = st.query

	var bindParams []interface{}
	for _, bind := range st.bindsValue {
		if elem, ok := params[bind]; ok {
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

	convertedQuery, bindParams, err := t.GenerateQueryAndBindValue(query, params)
	if err != nil {
		return err
	}

	return t.db.SelectContext(ctx, inputStructs, convertedQuery, bindParams...)

}

// Select is a thin wrapper around db.Select in the sqlx package.
// 事前条件: inputStructのフィールドとqueryで返ってくる要素の長さと並びは一致していなければならない。
func (t *Twowaysql) Select(inputStructs interface{}, query string, params map[string]interface{}) error {

	convertedQuery, bindParams, err := t.GenerateQueryAndBindValue(query, params)
	if err != nil {
		return err
	}

	return t.db.Select(inputStructs, convertedQuery, bindParams...)
	//fmt.Println("rv", rv)

}

// Exec is a thin wrapper around db.Exec in the sqlx package.
func (t *Twowaysql) Exec(inputStructs interface{}, query string, params map[string]interface{}) (sql.Result, error) {

	convertedQuery, bindParams, err := t.GenerateQueryAndBindValue(query, params)
	if err != nil {
		return nil, err
	}

	return t.db.Exec(convertedQuery, bindParams...)
}

// ExecContext is a thin wrapper around db.ExecContext in the sqlx package.
func (t *Twowaysql) ExecContext(ctx context.Context, inputStructs interface{}, query string, params map[string]interface{}) (sql.Result, error) {

	convertedQuery, bindParams, err := t.GenerateQueryAndBindValue(query, params)
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
