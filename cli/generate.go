package cli

import (
	"bytes"
	"os"

	"github.com/alecthomas/chroma/quick"
	"github.com/future-architect/go-twowaysql"
)

func generateTemplate(outFile, language string) error {
	if outFile == "" || outFile == "--" {
		if isTerminal(os.Stdout) {
			b := &bytes.Buffer{}
			err := twowaysql.GenerateMarkdown(b, language)
			if err != nil {
				return err
			}
			quick.Highlight(os.Stdout, b.String(), "markdown", "terminal", "github")
			return nil
		} else {
			err := twowaysql.GenerateMarkdown(os.Stdout, language)
			if err != nil {
				return err
			}
			return nil
		}
	} else {
		f, err := os.Create(outFile)
		if err != nil {
			return err
		}
		defer f.Close()
		return twowaysql.GenerateMarkdown(f, language)
	}
}
