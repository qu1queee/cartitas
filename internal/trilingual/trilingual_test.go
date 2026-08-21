package trilingual

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateStageMissingLang(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "draft", "en", "Geography")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "continents_early.md"), []byte("Q: x\nA: y\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	errs := ValidateStage(root, "draft", []string{"en", "es", "de"})
	if len(errs) != 1 {
		t.Fatalf("got %v", errs)
	}
	if !strings.Contains(errs[0], "missing [es, de]") {
		t.Fatalf("got %q", errs[0])
	}
}

func TestKeysFromPaths(t *testing.T) {
	keys := KeysFromPaths("queue", []string{
		"queue/en/Animals/pets_middle.md",
		"queue/es/Animals/pets_middle.md",
		"queue/de/Sports/olympics_middle.md",
		"cards/en/Animals/pets_middle.md",
		"queue/en/Animals",
	})
	if len(keys) != 2 {
		t.Fatalf("got %v", keys)
	}
	if keys[0] != "Animals/pets_middle.md" || keys[1] != "Sports/olympics_middle.md" {
		t.Fatalf("got %v", keys)
	}
}

func TestValidateStageFilterIgnoresUntouchedIncompleteSets(t *testing.T) {
	root := t.TempDir()
	for _, lang := range []string{"en", "es", "de"} {
		dir := filepath.Join(root, "queue", lang, "Animals")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "pets_middle.md"), []byte("ok"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	sports := filepath.Join(root, "queue", "de", "Sports")
	if err := os.MkdirAll(sports, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sports, "olympics_middle.md"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	if errs := ValidateStage(root, "queue", []string{"en", "es", "de"}); len(errs) != 1 {
		t.Fatalf("full scan should fail on olympics, got %v", errs)
	}
	only := KeysFromPaths("queue", []string{"queue/en/Animals/pets_middle.md"})
	if errs := ValidateStageFilter(root, "queue", []string{"en", "es", "de"}, only); len(errs) != 0 {
		t.Fatalf("PR-scoped scan should ignore olympics, got %v", errs)
	}
}

func TestValidateStageComplete(t *testing.T) {
	root := t.TempDir()
	for _, lang := range []string{"en", "es", "de"} {
		dir := filepath.Join(root, "draft", lang, "Geography")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "continents_early.md"), []byte("ok"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if errs := ValidateStage(root, "draft", []string{"en", "es", "de"}); len(errs) != 0 {
		t.Fatalf("unexpected %v", errs)
	}
}
