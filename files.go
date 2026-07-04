package cwrapmsg

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

func getFiles() ([]string, error) {
	var files []string

	err := filepath.Walk("./", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrap(err, "filepath.Walk")
		}

		if !info.IsDir() {
			if !isNeeded(info.Name()) {
				return nil
			}

			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, errors.Wrap(err, "Walk")
	}

	return files, nil
}

func isNeeded(name string) bool {
	if checkIfPb(name) || checkIfGen(name) {
		return false
	}

	sl := strings.Split(name, ".")

	if len(sl) == 0 {
		return false
	}

	if sl[len(sl)-1] != "go" {
		return false
	}

	return true
}

func checkIfPb(name string) bool {
	pattern := `\.pb\.`

	matched, err := regexp.MatchString(pattern, name)
	if err != nil {
		return false
	}

	return matched
}

func checkIfGen(name string) bool {
	pattern := `\_gen\.go`

	matched, err := regexp.MatchString(pattern, name)
	if err != nil {
		return false
	}

	return matched
}
