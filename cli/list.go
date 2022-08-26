package cli

import (
	"database/sql"
	"os"

	"github.com/alecthomas/chroma/quick"
	"github.com/goccy/go-yaml"
)

func listDriver() {
	drivers, _ := yaml.Marshal(sql.Drivers())
	quick.Highlight(os.Stdout, string(drivers), "yaml", "terminal", "monokai")
}
