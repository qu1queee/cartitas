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
	return ValidateStageFilter(root, stage, required, nil)
}

// KeysFromPaths maps repo paths like queue/en/Animals/pets_middle.md to
// relative keys (Animals/pets_middle.md) for the given stage.
func KeysFromPaths(stage string, paths []string) []string {
	prefix := stage + "/"
	seen := map[string]struct{}{}
	var keys []string
	for _, p := range paths {
		p = filepath.ToSlash(strings.TrimSpace(p))
		if p == "" || !strings.HasPrefix(p, prefix) {
			continue
		}
		parts := strings.Split(strings.TrimPrefix(p, prefix), "/")
		if len(parts) < 2 {
			continue
		}
		key := strings.Join(parts[1:], "/")
		if !strings.HasSuffix(key, ".md") {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// ValidateStageFilter checks trilingual completeness. If onlyKeys is nil, every
// set under the stage is checked. If onlyKeys is non-nil, only those keys are
// checked (empty means nothing to check). Keys with no remaining files are skipped.
func ValidateStageFilter(root, stage string, required []string, onlyKeys []string) []string {
	var errors []string
	groups := FindLangGroups(filepath.Join(root, stage))
	var keys []string
	if onlyKeys != nil {
		keys = append([]string{}, onlyKeys...)
		sort.Strings(keys)
	} else {
		keys = make([]string, 0, len(groups))
		for k := range groups {
			keys = append(keys, k)
		}
		sort.Strings(keys)
	}
	for _, key := range keys {
		present := groups[key]
		if onlyKeys != nil && len(present) == 0 {
			continue
		}
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
