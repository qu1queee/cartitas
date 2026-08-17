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
