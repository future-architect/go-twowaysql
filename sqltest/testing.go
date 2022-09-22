package sqltest

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/future-architect/go-exceltesting"
	"github.com/future-architect/go-twowaysql"
	"github.com/google/go-cmp/cmp"
	"github.com/jmoiron/sqlx"
)

type callbackForTest struct {
	t *testing.T
}

func (c callbackForTest) StartTest(doc *twowaysql.Document, tc twowaysql.TestCase) {
	c.t.Log("  Executing Query for Fixture")
}

func (c callbackForTest) ExecFixture(doc *twowaysql.Document, tc twowaysql.TestCase) {
	c.t.Log("  Executing Query for Fixture")
}

func (c callbackForTest) InsertFixtureTable(doc *twowaysql.Document, tc twowaysql.TestCase, table twowaysql.Table) {
	c.t.Logf("  Inserting Fixture Table: %s", table.Name)
}

func (c callbackForTest) Exec(doc *twowaysql.Document, tc twowaysql.TestCase) {
	c.t.Logf("Exec Query with: %v", tc.Params)
}

func (c callbackForTest) ExecTestQuery(doc *twowaysql.Document, tc twowaysql.TestCase) {
	c.t.Log("Exec Query for getting result")
}

func (c callbackForTest) EndTest(doc *twowaysql.Document, tc twowaysql.TestCase, failure, err error) {
	if failure != nil {
		c.t.Errorf("Test Result: Failure: %v", err)
	} else if err != nil {
		c.t.Errorf("Test Result: Error: %v", err)
	} else {
		c.t.Log("Test Result: OK")
	}
}

var _ Callback = &callbackForTest{}

func RunInTest(ctx context.Context, t *testing.T, db *sqlx.DB, doc *twowaysql.Document) {
	t.Helper()

	Run(ctx, db, doc, &callbackForTest{t: t})
}

func RunInTests(ctx context.Context, t *testing.T, db *sqlx.DB, docs map[string]*twowaysql.Document) {
	t.Helper()
	for k, doc := range docs {
		t.Run(k, func(t *testing.T) {
			RunInTest(ctx, t, db, doc)
		})
	}
}

type Callback interface {
	StartTest(doc *twowaysql.Document, tc twowaysql.TestCase)
	ExecFixture(doc *twowaysql.Document, tc twowaysql.TestCase)
	InsertFixtureTable(doc *twowaysql.Document, tc twowaysql.TestCase, tb twowaysql.Table)
	Exec(doc *twowaysql.Document, tc twowaysql.TestCase)
	ExecTestQuery(doc *twowaysql.Document, tc twowaysql.TestCase)
	EndTest(doc *twowaysql.Document, tc twowaysql.TestCase, failure, err error)
}

func Run(ctx context.Context, db *sqlx.DB, doc *twowaysql.Document, cb Callback) (failureCount, errCount int, err error) {
	err = func() error {
		tws := twowaysql.New(db)
		tx, err := tws.Begin(ctx)
		if err != nil {
			return fmt.Errorf("database connection test error: %w", err)
		}
		tx.Rollback()
		return nil
	}()
	if err != nil {
		return 0, 0, err
	}
	for _, tc := range doc.TestCases {
		func() {
			cb.StartTest(doc, tc)

			tws := twowaysql.New(db)
			tx, err := tws.Begin(ctx)
			if err != nil {
				errCount++
				cb.EndTest(doc, tc, nil, err)
				return
			}
			defer tx.Rollback()
			switch doc.CommonTestFixture.Lang {
			case "sql":
				cb.ExecFixture(doc, tc)
				_, err := tx.Tx().ExecContext(ctx, doc.CommonTestFixture.Code)
				if err != nil {
					errCount++
					cb.EndTest(doc, tc, nil, fmt.Errorf("common fixture exec error in %s: %w", tc.Name, err))
					return
				}
			case "yaml":
				for _, t := range doc.CommonTestFixture.Tables {
					cb.InsertFixtureTable(doc, tc, t)
					err := exceltesting.LoadRaw(tx.Tx().Tx, exceltesting.LoadRawRequest{
						TableName: t.Name,
						Columns:   t.Cells[0],
						Values:    t.Cells[1:],
					})
					if err != nil {
						errCount++
						cb.EndTest(doc, tc, nil, fmt.Errorf("common fixture error for %s table in %s: %w", t.Name, tc.Name, err))
						return
					}
				}
			}
			for _, t := range tc.Fixtures {
				cb.InsertFixtureTable(doc, tc, t)
				err := exceltesting.LoadRaw(tx.Tx().Tx, exceltesting.LoadRawRequest{
					TableName: t.Name,
					Columns:   t.Cells[0],
					Values:    t.Cells[1:],
				})
				if err != nil {
					errCount++
					cb.EndTest(doc, tc, nil, fmt.Errorf("fixture error for %s table in %s: %w", t.Name, tc.Name, err))
					return
				}
			}
			var result []map[string]any
			if tc.TestQuery == "" {
				cb.Exec(doc, tc)
				err := tx.Select(ctx, &result, doc.SQL, tc.Params)
				if err != nil {
					errCount++
					cb.EndTest(doc, tc, nil, fmt.Errorf("exec SQL error in %s: %w", tc.Name, err))
					return
				}
			} else {
				cb.Exec(doc, tc)
				_, err := tx.Exec(ctx, doc.SQL, tc.Params)
				if err != nil {
					errCount++
					cb.EndTest(doc, tc, nil, fmt.Errorf("exec SQL error in %s: %w", tc.Name, err))
					return
				}
				cb.ExecTestQuery(doc, tc)
				err = tx.Select(ctx, &result, tc.TestQuery, nil)
				if err != nil {
					errCount++
					cb.EndTest(doc, tc, nil, fmt.Errorf("exec SQL error for result in %s: %w", tc.Name, err))
					return
				}
			}
			fail := compare(tc.Expect, result)
			if fail != nil {
				failureCount++
			}
			cb.EndTest(doc, tc, fail, nil)
		}()
	}
	return failureCount, errCount, nil
}

func compare(expectedCells [][]string, actual []map[string]any) error {
	var expected []map[string]any
	if len(expectedCells) == 0 {
		if len(actual) == 0 {
			return nil
		}
		if diff := cmp.Diff(expected, actual); diff != "" {
			return fmt.Errorf("result mismatch: %s", diff)
		}
	}
	header := expectedCells[0]
	expectRows := expectedCells[1:]
	if len(actual) > 0 {
		for _, r := range expectRows {
			row := make(map[string]any)
			for i, h := range header {
				switch actual[0][h].(type) {
				case int:
					integer, err := strconv.ParseInt(r[i], 10, 64)
					if err != nil {
						row[h] = r[i]
					} else {
						row[h] = int(integer)
					}
				case int64:
					integer, err := strconv.ParseInt(r[i], 10, 64)
					if err != nil {
						row[h] = r[i]
					} else {
						row[h] = integer
					}
				case float64:
					f, err := strconv.ParseFloat(r[i], 64)
					if err != nil {
						row[h] = r[i]
					} else {
						row[h] = f
					}
				case bool:
					row[h] = r[i] == "true"
				default:
					row[h] = r[i]
				}
			}
			expected = append(expected, row)
		}
	} else {
		for _, r := range expectRows {
			row := make(map[string]any)
			for i, h := range header {
				row[h] = r[i]
			}
			expected = append(expected, row)
		}
	}
	if diff := cmp.Diff(expected, actual); diff != "" {
		return fmt.Errorf("result mismatch: %s", diff)
	}
	return nil
}
