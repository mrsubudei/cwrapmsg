package cwrapmsg

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

func getFileNames() ([]string, error) {
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

func getUnstagedFilesMap(enable bool) (map[string]struct{}, error) {
	if !enable {
		return nil, nil
	}

	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrap(err, "cmd.Output")
	}

	lines := strings.Split(string(output), "\n")

	filesMap := make(map[string]struct{}, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}

		status := line[:2]
		path := line[3:]

		if isUntrackedStatus(status) {
			if err := scanDir(path, filesMap); err != nil {
				return nil, errors.Wrap(err, "scanDir")
			}
		} else if isSutableStatus(status) && isNeeded(path) {
			filesMap[path] = struct{}{}
		}
	}

	return filesMap, nil
}

func scanDir(filePath string, filesMap map[string]struct{}) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return errors.Wrap(err, "os.Stat")
	}

	if info.IsDir() {
		err = filepath.Walk(filePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return errors.Wrap(err, "filepath.Walk")
			}
			if !info.IsDir() && isNeeded(info.Name()) {
				filesMap[path] = struct{}{}
			}

			return nil
		})
		if err != nil {
			return errors.Wrap(err, "Walk")
		}
	} else {
		if isNeeded(info.Name()) {
			filesMap[filePath] = struct{}{}
		}
	}

	return nil
}

func isSutableStatus(status string) bool {
	return status == " M" || // modified (unstaged)
		status == "M " || // modified (staged)
		status == "A " || // added (staged)
		status == "AM" || // added (staged) with modifications
		status == "R " || // renamed (staged)
		status == "RM" || // renamed (staged) with modifications
		status == "R?" || // renamed (unstaged) with modifications
		status == "C " || // copied (staged)
		status == "CM" || // copied (staged) with modifications
		status == "C?" // copied (unstaged)
}

func isUntrackedStatus(status string) bool {
	return status == "??" // untracked
}
