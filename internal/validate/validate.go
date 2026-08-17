package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/qu1queee/cartitas/internal/cardsplit"
)

const maxAnswerLen = 500

func File(root, path string) []string {
	var errors []string
	data, err := os.ReadFile(path)
	if err != nil {
		return []string{fmt.Sprintf("%s: %v", rel(root, path), err)}
	}
	text := string(data)
	relPath := rel(root, path)

	if strings.TrimSpace(text) == "" {
		return []string{fmt.Sprintf("%s: file is empty", relPath)}
	}

	blocks := cardsplit.Split(text)
	for i, block := range blocks {
		n := i + 1
		hasQ, hasA, hasC := false, false, false
		for _, line := range strings.Split(block, "\n") {
			switch {
			case strings.HasPrefix(line, "Q:"):
				hasQ = true
			case strings.HasPrefix(line, "A:"):
				hasA = true
				if len(line) > maxAnswerLen {
					errors = append(errors, fmt.Sprintf("%s card %d: answer very long (%d chars)", relPath, n, len(line)))
				}
			case strings.HasPrefix(line, "C:"):
				hasC = true
			}
		}

		if hasC && (hasQ || hasA) {
			errors = append(errors, fmt.Sprintf("%s card %d: cloze (C:) mixed with Q/A in same block", relPath, n))
		} else if hasQ != hasA {
			errors = append(errors, fmt.Sprintf("%s card %d: Q/A pair incomplete", relPath, n))
		} else if !hasC && !hasQ {
			if n == 1 && strings.HasPrefix(block, "#") {
				continue
			}
			errors = append(errors, fmt.Sprintf("%s card %d: no Q/A or C: card found", relPath, n))
		}
	}
	return errors
}

func All(root string) []string {
	var all []string
	for _, dir := range []string{"cards", "queue", "draft"} {
		base := filepath.Join(root, dir)
		info, err := os.Stat(base)
		if err != nil || !info.IsDir() {
			continue
		}
		var files []string
		_ = filepath.WalkDir(base, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
				return nil
			}
			files = append(files, path)
			return nil
		})
		sort.Strings(files)
		for _, path := range files {
			all = append(all, File(root, path)...)
		}
	}
	return all
}

func rel(root, path string) string {
	r, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(r)
}
