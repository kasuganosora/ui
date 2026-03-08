//go:build ignore

// icongen extracts SVG path data from Material Design Icons SVG files
// and generates a Go source file with embedded icon data.
//
// Usage:
//
//	go run cmd/icongen/main.go -src /path/to/material-design-icons/src -out icon/material/icons_generated.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	srcDir = flag.String("src", "", "path to material-design-icons/src directory")
	outFile = flag.String("out", "icon/material/icons_data.json", "output JSON file")
)

// pathDRegex extracts the d attribute from <path> elements, excluding fill="none" background rects.
var pathDRegex = regexp.MustCompile(`<path\s[^>]*?d="([^"]+)"[^>]*/?>`)
var fillNoneRegex = regexp.MustCompile(`fill="none"`)

func main() {
	flag.Parse()
	if *srcDir == "" {
		log.Fatal("must specify -src directory")
	}

	icons := map[string]string{}

	// Walk src/{category}/{name}/materialicons/24px.svg
	err := filepath.Walk(*srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		// Only process filled style 24px SVGs
		if !strings.HasSuffix(path, "/materialicons/24px.svg") &&
			!strings.HasSuffix(path, "\\materialicons\\24px.svg") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Extract icon name from path: src/{category}/{name}/materialicons/24px.svg
		rel, _ := filepath.Rel(*srcDir, path)
		parts := strings.Split(filepath.ToSlash(rel), "/")
		if len(parts) < 3 {
			return nil
		}
		name := parts[1] // {name}

		// Extract all <path d="..."> that don't have fill="none"
		content := string(data)
		matches := pathDRegex.FindAllStringSubmatch(content, -1)

		var paths []string
		for _, m := range matches {
			full := m[0]
			d := m[1]
			// Skip background rect (fill="none")
			if fillNoneRegex.MatchString(full) {
				continue
			}
			// Skip trivial "M0 0h24v24H0z" or "M0,0h24v24H0V0z" background paths
			if isBackgroundRect(d) {
				continue
			}
			paths = append(paths, d)
		}

		if len(paths) == 0 {
			return nil
		}

		// Combine multiple paths with space separator
		icons[name] = strings.Join(paths, " ")
		return nil
	})

	if err != nil {
		log.Fatalf("walk error: %v", err)
	}

	// Sort and output
	names := make([]string, 0, len(icons))
	for name := range icons {
		names = append(names, name)
	}
	sort.Strings(names)

	// Build ordered map for JSON
	type iconEntry struct {
		Name string `json:"n"`
		Path string `json:"p"`
	}
	entries := make([]iconEntry, len(names))
	for i, name := range names {
		entries[i] = iconEntry{Name: name, Path: icons[name]}
	}

	jsonData, err := json.Marshal(entries)
	if err != nil {
		log.Fatalf("json marshal: %v", err)
	}

	if err := os.WriteFile(*outFile, jsonData, 0644); err != nil {
		log.Fatalf("write file: %v", err)
	}

	fmt.Printf("Generated %d icons → %s (%d bytes)\n", len(names), *outFile, len(jsonData))
}

func isBackgroundRect(d string) bool {
	// Common background rect patterns
	d = strings.ReplaceAll(d, " ", "")
	d = strings.ReplaceAll(d, ",", "")
	return d == "M00h24v24H0z" || d == "M00h24v24H0V0z" || d == "M00h24v24H0Vz"
}
