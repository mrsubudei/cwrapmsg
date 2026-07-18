package cwrapmsg

import (
	"bufio"
	"log"
	"os"
	"strings"

	"github.com/pkg/errors"
)

type IgnoreData struct {
	funcNamesMap map[string]struct{}
	fullIgnore   bool
}

func GetIgnoreDataMap() map[string]IgnoreData {
	ignoreDataMap := make(map[string]IgnoreData)
	ignoreFiles := []string{"./.idea/cwrap.txt", "./.vscode/cwrap.txt", "./cwrap.txt"}

	for _, ignoreFile := range ignoreFiles {
		file, err := os.Open(ignoreFile)
		if err != nil {
			continue
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()

			if line == "" {
				break
			}

			ignoreData := IgnoreData{
				funcNamesMap: map[string]struct{}{},
			}

			sl := strings.Split(line, " ")
			dir := sl[0]
			if len(dir) > 0 && dir[len(dir)-1] == '/' {
				dir = dir[:len(dir)-1]
			}

			if len(sl) > 1 {
				funcNames := strings.Split(sl[1], ",")
				for _, funcName := range funcNames {
					ignoreData.funcNamesMap[funcName] = struct{}{}
				}
			} else {
				ignoreData.fullIgnore = true
			}

			ignoreDataMap[dir] = ignoreData
		}

		if err := scanner.Err(); err != nil {
			log.Fatal(errors.Wrap(err, "scanner.Err"))
		}
	}

	return ignoreDataMap
}

func getChunks(fileName string) []string {
	parts := strings.Split(fileName, "/")
	var chunks []string

	for idx := range parts {
		chunks = append(chunks, strings.Join(parts[:idx+1], "/"))
	}

	return chunks
}
