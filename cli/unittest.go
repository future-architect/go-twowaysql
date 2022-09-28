package cli

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/alecthomas/chroma/quick"
	"github.com/fatih/color"
	"github.com/future-architect/go-twowaysql"
	"github.com/future-architect/go-twowaysql/sqltest"
	"github.com/goccy/go-yaml"
	"github.com/hashicorp/go-multierror"
	"github.com/jmoiron/sqlx"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

type entry struct {
	path string
	doc  *twowaysql.Document
}

func newTestCallback(filePath string, verbose, quiet bool) *testCallback {
	return &testCallback{
		filePath: filePath,
		file:     color.New(color.FgHiBlue, color.Underline, color.Bold),
		testcase: color.New(color.FgHiCyan, color.Underline, color.Bold),
		name:     color.New(color.Bold),
		verbose:  verbose,
		quiet:    quiet,
	}
}

type testCallback struct {
	filePath string
	file     *color.Color
	testcase *color.Color
	name     *color.Color
	verbose  bool
	quiet    bool
}

func (c testCallback) StartTest(doc *twowaysql.Document, tc twowaysql.TestCase) {
	if c.verbose {
		if !c.quiet {
			c.testcase.Printf("## RUN  %s / %s\n", doc.Title, tc.Name)
		}
	}
}

func (c testCallback) ExecFixture(doc *twowaysql.Document, tc twowaysql.TestCase) {
	if c.verbose {
		fmt.Printf("  Running fixture SQL\n")
	}
}

func (c testCallback) InsertFixtureTable(doc *twowaysql.Document, tc twowaysql.TestCase, tb twowaysql.Table) {
	if c.verbose {
		fmt.Printf("  Inserting fixture table %s\n", c.name.Sprint(tb.Name))
	}
}

func (c testCallback) Exec(doc *twowaysql.Document, tc twowaysql.TestCase) {
	if c.verbose {
		sqlParamYaml, _ := yaml.Marshal(tc.Params)
		var buf bytes.Buffer
		quick.Highlight(&buf, string(sqlParamYaml), "yaml", "terminal", "monokai")
		fmt.Printf("  Exec SQL with: %s\n", strings.ReplaceAll(buf.String(), "\n", ""))
	}
}

func (c testCallback) ExecTestQuery(doc *twowaysql.Document, tc twowaysql.TestCase) {
	if c.verbose {
		fmt.Println("  Exec test query")
	}
}

func (c testCallback) EndTest(doc *twowaysql.Document, tc twowaysql.TestCase, failure error, err error) {
	if err != nil {
		if c.verbose {
			color.HiRed("  Test Error\n")
		} else {
			fmt.Printf("%s / %s: %s at %s\n\n", doc.Title, tc.Name, color.HiRedString("Test Error"), c.filePath)
		}
		color.HiRed(err.Error())
	} else if failure != nil {
		if c.verbose {
			color.Yellow("  Test Failure\n")
		} else {
			fmt.Printf("%s / %s: %s at %s\n\n", doc.Title, tc.Name, color.HiRedString("Test Failure"), c.filePath)
		}
		color.Yellow(failure.Error())
	} else if c.verbose {
		color.HiGreen("  Test OK\n")
	}
}

func unittest(driver, dbSrc string, filesOrDirs []string, verbose, quiet bool) (ok bool, err error) {
	if verbose {
		quiet = false
	}
	var entries []entry
	var errs *multierror.Error
	for _, f := range findFiles(filesOrDirs) {
		doc, err := twowaysql.ParseMarkdownFile(f)
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("%s: %w", f, err))
			continue
		}
		entries = append(entries, entry{path: f, doc: doc})
	}
	if errs != nil {
		return false, errs
	}

	// timeout
	ctx := context.Background()
	db, err := sqlx.Open(driver, dbSrc)
	if err != nil {
		return false, err
	}
	// quiet: show only error
	// verbose: show all
	// !quiet && !verbose: show test name and errors

	file := color.New(color.FgHiBlue, color.Underline, color.Bold)
	name := color.New(color.Bold)

	var totalFailureCount int
	var totalErrorCount int
	for _, e := range entries {
		if !quiet {
			fmt.Printf("%s at %s\n", file.Sprintf("# %s", e.doc.Title), name.Sprint(e.path))
		}
		if len(e.doc.TestCases) == 0 {
			if !quiet {
				color.Yellow("  No Test")
			}
			continue
		}
		if verbose {
			fmt.Print("\n")
			quick.Highlight(os.Stdout, e.doc.SQL, "sql", "terminal", "monokai")
			fmt.Print("\n\n")
		}
		failureCount, errCount, err := sqltest.Run(ctx, db, e.doc, newTestCallback(e.path, verbose, quiet))
		if err != nil {
			return false, err
		}
		if !quiet {
			if verbose {
				fmt.Print("\n")
			}
			fmt.Printf("%s %s\n", file.Sprintf("# %s: Result", e.doc.Title), formatResult(failureCount, errCount, "ok"))
			if verbose {
				fmt.Print("\n")
			}
		}
		totalFailureCount += failureCount
		totalErrorCount += errCount
	}
	fmt.Println(formatResult(totalFailureCount, totalErrorCount, "pass"))
	return (totalErrorCount + totalFailureCount) == 0, nil
}

func formatResult(failureCount, errorCount int, okMessage string) string {
	var result []string
	if failureCount == 1 {
		result = append(result, color.YellowString("1 failure"))
	} else if failureCount > 1 {
		result = append(result, color.YellowString("%d failures", failureCount))
	}
	if errorCount == 1 {
		result = append(result, color.HiRedString("1 error"))
	} else if errorCount > 1 {
		result = append(result, color.HiRedString("%d errors", errorCount))
	}
	if len(result) == 0 {
		result = append(result, color.HiGreenString(okMessage))
	}
	return strings.Join(result, "  ")
}

func findFiles(filesOrDirs []string) []string {
	result := []string{}
	found := make(map[string]bool)
	for _, f := range filesOrDirs {
		s, err := os.Stat(f)
		if err != nil {
			panic(err) // input should be existing files/dirs by using kingpin.
		}
		if s.IsDir() {
			filepath.Walk(f, func(path string, info fs.FileInfo, err error) error {
				if info.IsDir() {
					return nil
				}
				if strings.HasSuffix(info.Name(), ".sql.md") && !found[path] {
					result = append(result, path)
					found[path] = true
				}
				return nil
			})
		} else if !found[f] {
			result = append(result, f)
			found[f] = true
		}
	}
	sort.Strings(result)
	return result
}
