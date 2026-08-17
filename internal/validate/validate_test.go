package validate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileAllowsHeaderThenCards(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ok.md")
	content := `# Continents — early

Q: How many?
A: Seven.

---

C: Earth goes around the [Sun].
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if errs := File(dir, path); len(errs) != 0 {
		t.Fatalf("unexpected: %v", errs)
	}
}

func TestFileIncompleteQA(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.md")
	if err := os.WriteFile(path, []byte("Q: Only question?\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	errs := File(dir, path)
	if len(errs) != 1 || !strings.Contains(errs[0], "Q/A pair incomplete") {
		t.Fatalf("got %v", errs)
	}
}

func TestFileMixedCloze(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mix.md")
	if err := os.WriteFile(path, []byte("Q: Hi?\nA: There.\nC: [Nope]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	errs := File(dir, path)
	if len(errs) != 1 || !strings.Contains(errs[0], "cloze") {
		t.Fatalf("got %v", errs)
	}
}
