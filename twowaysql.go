// Package twowaysql provides an implementation of 2WaySQL.
package twowaysql

import (
	"context"
	"database/sql"
	"fmt"

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
// dest takes a pointer to a slice of a struct. The struct tag format must be `db:"tag_name"`.
func (t *Twowaysql) Select(ctx context.Context, dest interface{}, query string, params interface{}) error {

	eval, bindParams, err := Eval(query, params)
	if err != nil {
		return err
	}

	q := t.db.Rebind(eval)

	return t.db.SelectContext(ctx, dest, q, bindParams...)

}

// Exec is a thin wrapper around db.Exec in the sqlx package.
// params takes a tagged struct. The tags format must be `twowaysql:"tag_name"`.
func (t *Twowaysql) Exec(ctx context.Context, query string, params interface{}) (sql.Result, error) {

	eval, bindParams, err := Eval(query, params)
	if err != nil {
		return nil, err
	}

	q := t.db.Rebind(eval)

	return t.db.ExecContext(ctx, q, bindParams...)
}

// Begin is a thin wrapper around db.BeginTxx in the sqlx package.
func (t *Twowaysql) Begin(ctx context.Context) (*TwowaysqlTx, error) {

	tx, err := t.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &TwowaysqlTx{tx: tx}, nil
}

// Close is a thin wrapper around db.Close in the sqlx package.
func (t *Twowaysql) Close(ctx context.Context) error {

	if err := t.db.Close(); err != nil {
		return fmt.Errorf("close db: %w", err)
	}

	return nil
}

// TwowaysqlTx is a structure for issuing 2WaySQL queries within a transaction.
type TwowaysqlTx struct {
	tx *sqlx.Tx
}

// Commit is a thin wrapper around tx.Commit in the sqlx package.
func (t *TwowaysqlTx) Commit() error {

	if err := t.tx.Commit(); err != nil {
		return err
	}

	return nil
}

// Rollback is a thin wrapper around tx.Rollback in the sqlx package.
func (t *TwowaysqlTx) Rollback() error {

	if err := t.tx.Rollback(); err != nil {
		return err
	}

	return nil
}

// Select is a thin wrapper around db.Select in the sqlx package.
// params takes a tagged struct. The tags format must be `twowaysql:"tag_name"`.
// dest takes a pointer to a slice of a struct. The struct tag format must be `db:"tag_name"`.
// It is an equivalent implementation of Twowaysql.Select
func (t *TwowaysqlTx) Select(ctx context.Context, dest interface{}, query string, params interface{}) error {

	eval, bindParams, err := Eval(query, params)
	if err != nil {
		return err
	}

	q := t.tx.Rebind(eval)

	return t.tx.SelectContext(ctx, dest, q, bindParams...)

}

// Exec is a thin wrapper around db.Exec in the sqlx package.
// params takes a tagged struct. The tags format must be `twowaysql:"tag_name"`.
// It is an equivalent implementation of Twowaysql.Exec
func (t *TwowaysqlTx) Exec(ctx context.Context, query string, params interface{}) (sql.Result, error) {

	eval, bindParams, err := Eval(query, params)
	if err != nil {
		return nil, err
	}

	q := t.tx.Rebind(eval)

	return t.tx.ExecContext(ctx, q, bindParams...)
}
