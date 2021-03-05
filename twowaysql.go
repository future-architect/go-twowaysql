package twowaysql

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
)

// Twowaysql is a struct for issuing two way sql query
type Twowaysql struct {
	db *sqlx.DB
}

// New returns instance of Twowaysql
func New(db *sqlx.DB) *Twowaysql {
	return &Twowaysql{
		db: db,
	}
}

// Generate returns converted query and bind value
// The return value is expected to be used to issue queries to the database
func (t *Twowaysql) Generate(query string, params map[string]interface{}) (string, []interface{}, error) {

	st, err := t.parse(query, params)
	if err != nil {
		return "", nil, err
	}

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
func (t *Twowaysql) SelectContext(ctx context.Context, inputStructs interface{}, query string, params map[string]interface{}) error {

	convertedQuery, bindParams, err := t.Generate(query, params)
	if err != nil {
		return err
	}

	return t.db.SelectContext(ctx, inputStructs, convertedQuery, bindParams...)

}

// Select is a thin wrapper around db.Select in the sqlx package.
func (t *Twowaysql) Select(inputStructs interface{}, query string, params map[string]interface{}) error {

	convertedQuery, bindParams, err := t.Generate(query, params)
	if err != nil {
		return err
	}

	return t.db.Select(inputStructs, convertedQuery, bindParams...)
	//fmt.Println("rv", rv)

}

// Exec is a thin wrapper around db.Exec in the sqlx package.
func (t *Twowaysql) Exec(query string, params map[string]interface{}) (sql.Result, error) {

	convertedQuery, bindParams, err := t.Generate(query, params)
	if err != nil {
		return nil, err
	}

	return t.db.Exec(convertedQuery, bindParams...)
}

// ExecContext is a thin wrapper around db.ExecContext in the sqlx package.
func (t *Twowaysql) ExecContext(ctx context.Context, query string, params map[string]interface{}) (sql.Result, error) {

	convertedQuery, bindParams, err := t.Generate(query, params)
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
