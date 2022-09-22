package cli

import (
	"database/sql"
	"os"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app    = kingpin.New("twowaysql", "2-Way-SQL helper tool")
	driver = app.Flag("driver", `Database driver. TWOWAYSQL_DRIVER envvar is acceptable.`).Short('d').Envar("TWOWAYSQL_DRIVER").Enum(sql.Drivers()...)
	source = app.Flag("source", `Database source (e.g. postgres://user:pass@host/dbname?sslmode=disable). TWOWAYSQL_CONNECTION envvar is acceptable.`).Short('c').Envar("TWOWAYSQL_CONNECTION").String()

	runCommand      = app.Command("run", "Execute SQL file")
	runFile         = runCommand.Arg("file", "SQL/Markdown file").Required().NoEnvar().ExistingFile()
	runParam        = runCommand.Flag("param", "Parameter in single value or JSON (name=bob, or {\"name\": \"bob\"})").Short('p').NoEnvar().Strings()
	runExplain      = runCommand.Flag("explain", "Run with EXPLAIN to show execution plan").Short('e').NoEnvar().Bool()
	runRollback     = runCommand.Flag("rollback", "Run within transaction and then rollback").Short('r').NoEnvar().Bool()
	runOutputFormat = runCommand.Flag("output-format", "Result output format (default, md, json, yaml)").Short('o').Default("default").Enum("default", "md", "json", "yaml")

	testCommand = app.Command("test", "Run test")
	testFiles   = testCommand.Arg("file/dir", "Markdown file").Required().NoEnvar().ExistingFilesOrDirs()
	testVerbose = testCommand.Flag("verbose", "Show more information").Short('v').Bool()
	testQuiet   = testCommand.Flag("quiet", "Reduce information").Short('q').Bool()

	evalCommand = app.Command("eval", "Parse and evaluate SQL")
	evalFile    = evalCommand.Arg("file", "SQL/Markdown file").Required().NoEnvar().ExistingFile()
	evalParam   = evalCommand.Flag("param", "Parameter in single value or JSON (name=bob, or {\"name\": \"bob\"})").Short('p').NoEnvar().Strings()

	parseCommand    = app.Command("parse", "Parse SQL/Markdown source file")
	parseSrcFile    = parseCommand.Arg("file", "SQL file").Required().NoEnvar().ExistingFile()
	parseDumpFormat = parseCommand.Flag("dump-format", "Dump content in specified format (default, json, yaml)").Short('f').Default("default").Enum("default", "json", "yaml")

	generateCommand          = app.Command("generate", "Generate file")
	generateTemplateCommand  = generateCommand.Command("template", "Markdown template")
	generateTemplateOutput   = generateTemplateCommand.Arg("file", "Output file").String()
	generateTemplateLanguage = generateTemplateCommand.Flag("lang", "Language").Short('l').Enum("ja", "en")

	listCommand       = app.Command("list", "Inspection command")
	listDriverCommand = listCommand.Command("driver", "Show supported drivers")
)

func Main() {
	app.HelpFlag.Short('h')
	godotenv.Load(".env.local", ".env")

	var err error
	ok := true
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case listDriverCommand.FullCommand():
		listDriver()
	case evalCommand.FullCommand():
		err = eval(*evalFile, *evalParam)
	case runCommand.FullCommand():
		err = run(*driver, *source, *runFile, *runParam, *runExplain, *runRollback, *runOutputFormat, nil)
	case testCommand.FullCommand():
		ok, err = unittest(*driver, *source, *testFiles, *testVerbose, *testQuiet)
	case parseCommand.FullCommand():
		err = parseFile(*parseSrcFile, *parseDumpFormat)
	case generateTemplateCommand.FullCommand():
		err = generateTemplate(*generateTemplateOutput, *generateTemplateLanguage)
	}
	if err != nil {
		color.New(color.FgHiRed).Fprintln(os.Stderr, err.Error())
	}
	if err != nil || !ok {
		os.Exit(1)
	}
}
