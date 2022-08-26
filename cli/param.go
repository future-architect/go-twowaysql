package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/hashicorp/go-multierror"
)

func parseParams(params []string, stdin io.Reader) (map[string]any, error) {
	err := &multierror.Error{}

	result := make(map[string]any)

	if stdin != nil {
		d := json.NewDecoder(stdin)
		e := d.Decode(&result)
		if e != nil {
			err = multierror.Append(err, fmt.Errorf("JSON parse error: %w", e))
		}
	}

	for _, s := range params {
		if strings.HasPrefix(s, "{") {
			d := json.NewDecoder(strings.NewReader(s))
			e := d.Decode(&result)
			if e != nil {
				err = multierror.Append(err, fmt.Errorf("JSON parse error: %w", e))
			}
		} else {
			key, raw, found := strings.Cut(s, "=")
			if !found {
				err = multierror.Append(err, fmt.Errorf("invalid format: '%s' key=value or JSON is supported", s))
			}
			if value, err := strconv.ParseFloat(raw, 64); err == nil {
				result[key] = value
			} else {
				result[key] = raw
			}
		}
	}

	if err.Len() > 0 {
		return nil, err
	}

	return result, nil
}
