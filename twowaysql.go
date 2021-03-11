// Package twowaysql provides an implementation of 2WaySQL.
package twowaysql

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// Twowaysql is a struct for issuing 2WaySQL query
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
// params takes a tagged struct. The tags format must be `twowaysql:"tag_name"`.
// inputStruct takes a pointer to a slice of a struct. The struct tag format must be `db:"tag_name"`.
func (t *Twowaysql) Select(ctx context.Context, inputStructs interface{}, query string, params interface{}) error {

	convertedQuery, bindParams, err := Eval(query, params)
	if err != nil {
		return err
	}

	convertedQuery = t.db.Rebind(convertedQuery)

	return t.db.SelectContext(ctx, inputStructs, convertedQuery, bindParams...)

}

// Exec is a thin wrapper around db.Exec in the sqlx package.
// params takes a tagged struct. The tags format must be `twowaysql:"tag_name"`.
func (t *Twowaysql) Exec(ctx context.Context, query string, params interface{}) (sql.Result, error) {

	convertedQuery, bindParams, err := Eval(query, params)
	if err != nil {
		return nil, err
	}

	convertedQuery = t.db.Rebind(convertedQuery)

	return t.db.ExecContext(ctx, convertedQuery, bindParams...)
}
