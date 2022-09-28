package twowaysql

import (
	"log"
	"testing"

	"github.com/future-architect/go-twowaysql/private/testhelper"
	gocmp "github.com/google/go-cmp/cmp"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func TestParseMarkdown(t *testing.T) {
	type args struct {
		src string
	}
	tests := []struct {
		name    string
		args    args
		want    *Document
		wantErr string
	}{
		{
			name: "title and sql",
			args: args{
				src: testhelper.TrimIndent(t, `
				# Search User Query

				~~~sql
				SELECT email, name FROM persons WHERE first_name=/*first_name*/'bob';
				~~~
				`),
			},
			want: &Document{
				Title: "Search User Query",
				SQL:   `SELECT email, name FROM persons WHERE first_name=/*first_name*/'bob';`,
			},
		},
		{
			name: "with parameter",
			args: args{
				src: testhelper.TrimIndent(t, `
				# Search User Query

				~~~sql
				SELECT email, name FROM persons WHERE first_name=/*first_name*/'bob';
				~~~

				## Parameter

				| Name       | Type   | Description |
				|------------|--------|-------------|
				| first_name | string | search key  |
				`),
			},
			want: &Document{
				Title: "Search User Query",
				SQL:   `SELECT email, name FROM persons WHERE first_name=/*first_name*/'bob';`,
				Params: []Param{
					{
						Name:        "first_name",
						Type:        TextType,
						Description: "search key",
					},
				},
			},
		},
		{
			name: "with CRUD matrix",
			args: args{
				src: testhelper.TrimIndent(t, `
				# Search User Query

				~~~sql
				SELECT email, name FROM persons WHERE first_name=/*first_name*/'bob';
				~~~

				## CRUD Matrix

				| Table      | C | R | U | D | Description |
				|------------|---|---|---|---|-------------|
				| persons    | X |   |   |   |             |
				`),
			},
			want: &Document{
				Title: "Search User Query",
				SQL:   `SELECT email, name FROM persons WHERE first_name=/*first_name*/'bob';`,
				CRUDMatrix: []CRUDMatrix{
					{
						Table: "persons",
						C:     true,
						R:     false,
						U:     false,
						D:     false,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMarkdownString(tt.args.src)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
			} else {
				assert.NilError(t, err)
				assert.Check(t, cmp.DeepEqual(tt.want, got, gocmp.AllowUnexported(Document{})))
			}
		})
	}
}

func TestParseMarkdownTestCases(t *testing.T) {
	type args struct {
		src string
	}
	tests := []struct {
		name    string
		args    args
		want    *Document
		wantErr string
	}{
		{
			name: "with common test fixture in yaml (nested array)",
			args: args{
				src: testhelper.TrimIndent(t, `
				# Common Test Fixtures

				## Test

				~~~yaml
				fixtures:
				  persons:
				    - [employee_no, dept_no, first_name, last_name, email]
				    - [1, 10, Evan, MacMans, evanmacmans@example.com]
				    - [2, 11, Malvin, FitzSimons, malvinafitzsimons@example.com]
				    - [3, 12, Jimmie, Bruce, jimmiebruce@example.com]
				~~~
				`),
			},
			want: &Document{
				Title: "Common Test Fixtures",
				CommonTestFixture: Fixture{
					Lang: "yaml",
					Tables: []Table{
						{
							Name: "persons",
							Cells: [][]string{
								{"employee_no", "dept_no", "first_name", "last_name", "email"},
								{"1", "10", "Evan", "MacMans", "evanmacmans@example.com"},
								{"2", "11", "Malvin", "FitzSimons", "malvinafitzsimons@example.com"},
								{"3", "12", "Jimmie", "Bruce", "jimmiebruce@example.com"},
							},
						},
					},
				},
			},
		},
		{
			name: "with common test fixture in yaml (object list)",
			args: args{
				src: testhelper.TrimIndent(t, `
				# Common Test Fixtures

				## Test

				~~~yaml
				fixtures:
				  persons:
				    - { employee_no: 1, dept_no: 10, first_name: Evan, last_name: MacMans, email: evanmacmans@example.com }
				    - { employee_no: 2, dept_no: 11, first_name: Malvin, last_name: FitzSimons, email: malvinafitzsimons@example.com }
				    - { employee_no: 3, dept_no: 12, first_name: Jimmie, last_name: Bruce, email: jimmiebruce@example.com }
				~~~
				`),
			},
			want: &Document{
				Title: "Common Test Fixtures",
				CommonTestFixture: Fixture{
					Lang: "yaml",
					Tables: []Table{
						{
							Name: "persons",
							Cells: [][]string{
								{"dept_no", "email", "employee_no", "first_name", "last_name"},
								{"10", "evanmacmans@example.com", "1", "Evan", "MacMans"},
								{"11", "malvinafitzsimons@example.com", "2", "Malvin", "FitzSimons"},
								{"12", "jimmiebruce@example.com", "3", "Jimmie", "Bruce"},
							},
						},
					},
				},
			},
		},
		{
			name: "with common test fixture in sql",
			args: args{
				src: testhelper.TrimIndent(t, `
				# Common Test Fixtures

				~~~sql
				SELECT email FROM persons WHERE first_name=/*first_name*/'bob';
				~~~

				## Test

				~~~sql
				DELETE FROM persons;
				~~~
				`),
			},
			want: &Document{
				Title: "Common Test Fixtures",
				SQL:   "SELECT email FROM persons WHERE first_name=/*first_name*/'bob';",
				CommonTestFixture: Fixture{
					Lang: "sql",
					Code: "DELETE FROM persons;",
				},
			},
		},
		{
			name: "select test case (result is map list)",
			args: args{
				src: testhelper.TrimIndent(t, `
				# Test Cases

				~~~sql
				SELECT email FROM persons WHERE first_name=/*first_name*/'bob';
				~~~

				## Test

				### Case: select test

				~~~yaml
				fixtures:
				  persons:
				    - [employee_no, dept_no, first_name, last_name, email]
				    - [1, 10, Evan, MacMans, evanmacmans@example.com]
				params: { first_name: Evan }
				expect:
				  - { email: evanmacmans@example.com } 
				~~~
				`),
			},
			want: &Document{
				Title: "Test Cases",
				SQL:   "SELECT email FROM persons WHERE first_name=/*first_name*/'bob';",
				TestCases: []TestCase{
					{
						Name: "select test",
						Fixtures: []Table{
							{
								Name: "persons",
								Cells: [][]string{
									{"employee_no", "dept_no", "first_name", "last_name", "email"},
									{"1", "10", "Evan", "MacMans", "evanmacmans@example.com"},
								},
							},
						},
						Params: map[string]string{"first_name": "Evan"},
						Expect: [][]string{
							{"email"}, {"evanmacmans@example.com"},
						},
					},
				},
			},
		},
		{
			name: "select test case (result is nested list)",
			args: args{
				src: testhelper.TrimIndent(t, `
				# Test Cases

				~~~sql
				SELECT email FROM persons WHERE first_name=/*first_name*/'bob';
				~~~

				## Test

				### Case: select test

				~~~yaml
				fixtures:
				  persons:
				    - [employee_no, dept_no, first_name, last_name, email]
				    - [1, 10, Evan, MacMans, evanmacmans@example.com]
				params: { first_name: Evan }
				expect:
				  - [ email ]
				  - [ evanmacmans@example.com ]
				~~~
				`),
			},
			want: &Document{
				Title: "Test Cases",
				SQL:   "SELECT email FROM persons WHERE first_name=/*first_name*/'bob';",
				TestCases: []TestCase{
					{
						Name: "select test",
						Fixtures: []Table{
							{
								Name: "persons",
								Cells: [][]string{
									{"employee_no", "dept_no", "first_name", "last_name", "email"},
									{"1", "10", "Evan", "MacMans", "evanmacmans@example.com"},
								},
							},
						},
						Params: map[string]string{"first_name": "Evan"},
						Expect: [][]string{
							{"email"}, {"evanmacmans@example.com"},
						},
					},
				},
			},
		},
		{
			name: "delete test case",
			args: args{
				src: testhelper.TrimIndent(t, `
				# Test Cases

				~~~sql
				DELETE FROM persons;
				~~~

				## Test

				### Case: delete test

				~~~yaml
				testQuery: SELECT count(employee_no) FROM persons;
				expect:
				  - { count: 1 }
				~~~
				`),
			},
			want: &Document{
				Title: "Test Cases",
				SQL:   "DELETE FROM persons;",
				TestCases: []TestCase{
					{
						Name:      "delete test",
						TestQuery: `SELECT count(employee_no) FROM persons;`,
						Expect: [][]string{
							{"count"}, {"1"},
						},
					},
				},
			},
		},
		{
			name: "error: unknown field key in yaml",
			args: args{
				src: testhelper.TrimIndent(t, `
				# Test Cases

				~~~sql
				DELETE FROM persons;
				~~~

				## Test

				### Case: delete test

				testQueries should be testQuery.
				results should be result.

				~~~yaml
				testQueries: SELECT count(employee_no) FROM persons;
				results:
				  - { count: 1 }
				~~~
				`),
			},
			wantErr: "YAML keys results, testQueries is invalid in delete test of Test Cases (expect, fixtures, params, testQuery are acceptable)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMarkdownString(tt.args.src)
			if tt.wantErr != "" {
				assert.Error(t, err, tt.wantErr)
			} else {
				assert.NilError(t, err)
				assert.Check(t, cmp.DeepEqual(tt.want, got, gocmp.AllowUnexported(Document{})))
			}
		})
	}
}
