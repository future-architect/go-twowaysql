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

	if destMap, ok := dest.(*[]map[string]interface{}); ok {
		rows, err := t.db.QueryxContext(ctx, q, bindParams...)
		if err != nil {
			return err
		}
		return convertResultToMap(destMap, rows)
	}

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
func (t *Twowaysql) Close() error {

	if err := t.db.Close(); err != nil {
		return fmt.Errorf("close db: %w", err)
	}

	return nil
}

// DB returns `*sqlx.DB`
func (t *Twowaysql) DB() *sqlx.DB {
	return t.db
}

// Transaction starts a transaction as a block.
// arguments function is return error will rollback, otherwise to commit.
func (t *Twowaysql) Transaction(ctx context.Context, fn func(tx *TwowaysqlTx) error) error {
	tx, err := t.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			if rerr := tx.Rollback(); rerr != nil {
				panic(fmt.Sprintf("panic occured %v and failed rollback %v", p, rerr))
			}
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			return fmt.Errorf("failed rollback %v: %w", rerr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
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

	if destMap, ok := dest.(*[]map[string]interface{}); ok {
		rows, err := t.tx.QueryxContext(ctx, q, bindParams...)
		if err != nil {
			return err
		}
		return convertResultToMap(destMap, rows)
	}

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

func convertResultToMap(dest *[]map[string]interface{}, rows *sqlx.Rows) error {
	defer rows.Close()
	for rows.Next() {
		row := map[string]interface{}{}
		if err := rows.MapScan(row); err != nil {
			return err
		}
		*dest = append(*dest, row)
	}
	return nil
}

// Tx returns `*sqlx.Tx`
func (t *TwowaysqlTx) Tx() *sqlx.Tx {
	return t.tx
}
