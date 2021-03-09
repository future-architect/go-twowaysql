package twowaysql

import (
	"context"
	"database/sql"

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

// Select is a thin wrapper around db.Select in the sqlx package.
func (t *Twowaysql) Select(ctx context.Context, inputStructs interface{}, query string, params map[string]interface{}) error {

	convertedQuery, bindParams, err := Eval(query, params)
	if err != nil {
		return err
	}

	//適したplace holderに変換
	convertedQuery = t.db.Rebind(convertedQuery)

	return t.db.SelectContext(ctx, inputStructs, convertedQuery, bindParams...)

}

// Exec is a thin wrapper around db.Exec in the sqlx package.
func (t *Twowaysql) Exec(ctx context.Context, query string, params map[string]interface{}) (sql.Result, error) {

	convertedQuery, bindParams, err := Eval(query, params)
	if err != nil {
		return nil, err
	}

	//適したplace holderに変換
	convertedQuery = t.db.Rebind(convertedQuery)

	return t.db.ExecContext(ctx, convertedQuery, bindParams...)
}
