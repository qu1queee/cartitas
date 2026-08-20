package webexport

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildCatalog(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "languages.yaml"), "default: en\nlanguages:\n  en:\n    name: English\n")
	write(t, filepath.Join(root, "cards", "en", "Animals", "pets_early.md"), `# Pets

<!-- lang: en | age: early -->

Q: What do dogs wag?
A: Their tail.

---

C: A [cat] often purrs.
`)
	cat, err := BuildCatalog(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(cat.Cards) != 2 {
		t.Fatalf("cards=%d %+v", len(cat.Cards), cat.Cards)
	}
	if cat.Cards[0].Question != "What do dogs wag?" {
		t.Fatalf("qa question: %q", cat.Cards[0].Question)
	}
	if cat.Cards[1].Answer != "cat" || !strings.Contains(cat.Cards[1].Question, "_____") {
		t.Fatalf("cloze: %+v", cat.Cards[1])
	}
}

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
