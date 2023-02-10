package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var files = 0
var commentCount = 0

var lines = map[string]int{}

var verbose = false
var ignoreComments = true

var commentForKind = map[string]string{
	"ego":  "//",
	"go":   "//",
	"sh":   "#",
	"json": "//",
}

var ignoreKinds = map[string]bool{
	".DS_Store":                true,
	"exe":                      true,
	"db":                       true,
	"png":                      true,
	"jpeg":                     true,
	"jpg":                      true,
	"pdf":                      true,
	"log":                      true,
	"sum":                      true,
	"mod":                      true,
	"openapi-generator-ignore": true,
}

var ignorePaths = map[string]bool{
	".git": true,
}

// Count the lines in the file.
func count(file string) error {
	kind := strings.TrimPrefix(filepath.Ext(file), ".")
	base := filepath.Base(file)

	if kind == "" {
		return nil
	}

	if base == "" {
		kind = "text"
	}

	if ignoreKinds[kind] {
		return nil
	}

	if verbose {
		fmt.Printf("Scanning %s\n", file)
	}

	b, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	files = files + 1
	lineCount := 0

	text := strings.Split(string(b), "\n")
	for _, line := range text {
		shortLine := strings.TrimSpace(line)
		if len(shortLine) == 0 {
			continue
		}

		if ignoreComments {
			skip := false

			for fileKind, comment := range commentForKind {
				if fileKind == kind && strings.HasPrefix(shortLine, comment) {
					skip = true

					break
				}
			}

			if skip {
				commentCount++

				continue
			}
		}

		lineCount++
	}

	lines[kind] = lines[kind] + lineCount

	return nil
}

func scan(paths []string) error {
	for _, path := range paths {
		fullPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		entries, err := os.ReadDir(fullPath)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			if entry.IsDir() && ignorePaths[entry.Name()] {
				continue
			}

			file := filepath.Join(fullPath, entry.Name())

			if entry.IsDir() {
				err = scan([]string{file})
			} else {
				err = count(file)
			}

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func main() {
	pathList := make([]string, 0)

	for index := 1; index < len(os.Args); index++ {
		arg := os.Args[index]

		if arg == "-v" {
			verbose = true
		} else if arg == "-c" {
			ignoreComments = false
		} else {
			pathList = append(pathList, arg)
		}
	}

	if len(pathList) == 0 {
		pathList = []string{"."}
	}

	err := scan(pathList)
	if err != nil {
		fmt.Printf("Error: %v", err)
	} else {
		extensions := make([]string, 0)
		for k := range lines {
			extensions = append(extensions, k)
		}

		sort.Strings(extensions)

		if verbose {
			fmt.Println()
		}

		for _, k := range extensions {
			fmt.Printf("%-10s   %-7d\n", k, lines[k])
		}

		if verbose && ignoreComments && commentCount > 0 {
			fmt.Printf("\nIgnored %d comment lines\n", commentCount)
		}
	}
}
