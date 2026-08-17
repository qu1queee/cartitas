package trilingual

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var Stages = []string{"draft", "queue"}

func FindLangGroups(base string) map[string]map[string]struct{} {
	groups := map[string]map[string]struct{}{}
	info, err := os.Stat(base)
	if err != nil || !info.IsDir() {
		return groups
	}
	entries, err := os.ReadDir(base)
	if err != nil {
		return groups
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		lang := e.Name()
		langDir := filepath.Join(base, lang)
		_ = filepath.WalkDir(langDir, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
				return nil
			}
			rel, err := filepath.Rel(base, path)
			if err != nil {
				return nil
			}
			parts := strings.Split(filepath.ToSlash(rel), "/")
			if len(parts) < 2 {
				return nil
			}
			key := strings.Join(parts[1:], "/")
			if groups[key] == nil {
				groups[key] = map[string]struct{}{}
			}
			groups[key][lang] = struct{}{}
			return nil
		})
	}
	return groups
}

func ValidateStage(root, stage string, required []string) []string {
	var errors []string
	groups := FindLangGroups(filepath.Join(root, stage))
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		present := groups[key]
		var missing []string
		for _, lang := range required {
			if _, ok := present[lang]; !ok {
				missing = append(missing, lang)
			}
		}
		if len(missing) == 0 {
			continue
		}
		have := sortedKeys(present)
		errors = append(errors, fmt.Sprintf(
			"%s/%s: has [%s], missing [%s] — needs all languages: %s",
			stage, key, strings.Join(have, ", "), strings.Join(missing, ", "), strings.Join(required, ", "),
		))
	}
	return errors
}

func sortedKeys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
