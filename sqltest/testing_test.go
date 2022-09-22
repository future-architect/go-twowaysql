package sqltest

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/future-architect/go-twowaysql"
	"github.com/future-architect/go-twowaysql/private/testhelper"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	_ "github.com/jackc/pgx/v4/stdlib"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func rowCount(t *testing.T, db *sqlx.DB) int {
	var count int
	r, err := db.DB.Query("select count(*) from persons;")
	if err != nil {
		panic(err)
	}
	defer r.Close()
	for r.Next() {
		err := r.Scan(&count)
		if err != nil {
			panic(err)
		}
	}
	return count
}

type dummyCallback struct {
	t *testing.T
}

func (dc dummyCallback) StartTest(doc *twowaysql.Document, tc twowaysql.TestCase) {
}

func (dc dummyCallback) ExecFixture(doc *twowaysql.Document, tc twowaysql.TestCase) {
}

func (dc dummyCallback) InsertFixtureTable(doc *twowaysql.Document, tc twowaysql.TestCase, tb twowaysql.Table) {
}

func (dc dummyCallback) Exec(doc *twowaysql.Document, tc twowaysql.TestCase) {
}

func (dc dummyCallback) ExecTestQuery(doc *twowaysql.Document, tc twowaysql.TestCase) {
}

func (dc dummyCallback) EndTest(doc *twowaysql.Document, tc twowaysql.TestCase, failure error, err error) {
	if failure != nil {
		dc.t.Logf("Failure: %v", failure)
	}
	if err != nil {
		dc.t.Logf("Error: %v", err)
	}
}

var _ Callback = &dummyCallback{}

func TestRun(t *testing.T) {
	driver := "pgx"
	srcStr := testhelper.SourceStr(t)
	db, err := sqlx.Open(driver, srcStr)
	if err != nil {
		panic(err)
	}
	count := rowCount(t, db)
	type args struct {
		src string
	}
	tests := []struct {
		name             string
		args             args
		wantErr          string
		wantFailureCount int
		wantErrorCount   int
		wantTests        int
	}{
		{
			name: "query test",
			args: args{
				src: testhelper.TrimIndent(t, `
				# Select Query

				~~~sql
				SELECT email, first_name, last_name FROM persons WHERE first_name=/*first_name*/'bob';
				~~~

				## Tests

				### Case: Query Evan

				~~~yaml
				params: { first_name: Evan }
				expect:
				- { email: evanmacmans@example.com, first_name: Evan, last_name: MacMans }
				~~~
				`),
			},
			wantErr:   "",
			wantTests: 1,
		},
		{
			name: "exec test",
			args: args{
				src: testhelper.TrimIndent(t, `
				# Select Query

				~~~sql
				INSERT INTO persons (employee_no, dept_no, email, first_name, last_name, created_at) VALUES (/*en*/1, /*dn*/10, /*em*/'a@examplecom', /*fn*/'a', /*ln*/'b', CURRENT_TIMESTAMP);
				~~~

				## Tests

				### Case: Query Evan

				~~~yaml
				params: { en: 4, dn: 13, em: 'dan@example.com', fn: 'Dan', ln: 'Connor' }
				testQuery: SELECT count(employee_no) FROM persons;
				expect:
				- { count: 4 }
				~~~
				`),
			},
			wantErr:   "",
			wantTests: 1,
		},
		{
			name: "query test with global SQL fixture",
			args: args{
				src: testhelper.TrimIndent(t, `
				# Select Query

				~~~sql
				SELECT email, first_name, last_name FROM persons WHERE first_name=/*first_name*/'bob';
				~~~

				## Tests

				~~~sql
				INSERT INTO persons (employee_no, dept_no, email, first_name, last_name, created_at) VALUES (
					4, 13, 'dan@example.com', 'Dan', 'Conner', CURRENT_TIMESTAMP
				);
				~~~

				### Case: Query Evan

				~~~yaml
				params: { first_name: Dan }
				expect:
				- { email: dan@example.com, first_name: Dan, last_name: Conner }
				~~~
				`),
			},
			wantErr:   "",
			wantTests: 1,
		},
		{
			name: "query test with global yaml fixture",
			args: args{
				src: testhelper.TrimIndent(t, `
				# Select Query

				~~~sql
				SELECT email, first_name, last_name FROM persons WHERE first_name=/*first_name*/'bob';
				~~~

				## Tests

				~~~yaml
				fixtures:
				  persons:
				    - { employee_no: 4, dept_no: 13, email: 'dan@example.com', first_name: 'Dan', last_name: 'Conner', created_at: '2022-09-13 10:30:15' }
				~~~

				### Case: Query Dan

				~~~yaml
				params: { first_name: Dan }
				expect:
				- { email: dan@example.com, first_name: Dan, last_name: Conner }
				~~~
				`),
			},
			wantErr:   "",
			wantTests: 1,
		},
		{
			name: "query test with local yaml fixture (success)",
			args: args{
				src: testhelper.TrimIndent(t, `
				# Select Query

				~~~sql
				SELECT email, first_name, last_name FROM persons WHERE first_name=/*first_name*/'bob';
				~~~

				## Tests

				### Case: Query Dan

				~~~yaml
				fixtures:
				  persons:
				    - { employee_no: 4, dept_no: 13, email: 'dan@example.com', first_name: 'Dan', last_name: 'Conner', created_at: '2022-09-13 10:30:15' }
				params: { first_name: Dan }
				expect:
				- { email: dan@example.com, first_name: Dan, last_name: Conner }
				~~~
				`),
			},
			wantErr:   "",
			wantTests: 1,
		},
		{
			name: "query test with local yaml fixture (failure)",
			args: args{
				src: testhelper.TrimIndent(t, `
				# Select Query

				~~~sql
				SELECT email, first_name, last_name FROM persons WHERE first_name=/*first_name*/'bob';
				~~~

				## Tests

				### Case: Query Dan (fail)

				~~~yaml
				fixtures:
				  persons:
				    - { employee_no: 4, dept_no: 13, email: 'dan@example.com', first_name: 'Dan', last_name: 'Conner', created_at: '2022-09-13 10:30:15' }
				params: { first_name: Dan }
				expect:
				- { email: dan@example.com, first_name: Dan, last_name: Evan }
				~~~
				`),
			},
			wantErr:          "",
			wantErrorCount:   0,
			wantFailureCount: 1,
			wantTests:        1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := twowaysql.ParseMarkdownString(tt.args.src)
			assert.NoError(t, err)
			if err != nil {
				return
			}
			assert.Equal(t, tt.wantTests, len(doc.TestCases))
			failureCount, errCount, err := Run(context.Background(), db, doc, &dummyCallback{t: t})
			assert.Equal(t, tt.wantFailureCount, failureCount, "failure count")
			assert.Equal(t, tt.wantErrorCount, errCount, "err count")
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
			// rollback keeps row count
			newCount := rowCount(t, db)
			assert.Equal(t, count, newCount)
		})
	}
}

func TestRunInTest(t *testing.T) {
	driver := "pgx"
	srcStr := testhelper.SourceStr(t)
	db, err := sqlx.Open(driver, srcStr)
	if err != nil {
		panic(err)
	}

	src := testhelper.TrimIndent(t, `
	# Select Query

	~~~sql
	SELECT email, first_name, last_name FROM persons WHERE first_name=/*first_name*/'bob';
	~~~

	## Tests

	### Case: Query Evan

	~~~yaml
	params: { first_name: Evan }
	expect:
	- { email: evanmacmans@example.com, first_name: Evan, last_name: MacMans }
	~~~
	`)
	doc, err := twowaysql.ParseMarkdownString(src)
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	RunInTest(ctx, t, db, doc)
}
