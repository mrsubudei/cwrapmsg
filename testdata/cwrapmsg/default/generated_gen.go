package defaults

import (
	"strconv"

	"github.com/pkg/errors"
)

// files with the _gen.go suffix are not checked
func generatedFileName() error {
	_, err := strconv.Atoi("1")
	if err != nil {
		return errors.Wrap(err, "wrongName")
	}

	return nil
}
