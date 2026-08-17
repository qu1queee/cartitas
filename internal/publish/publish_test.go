package publish

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/qu1queee/cartitas/internal/repo"
)

func TestPublishMovesBatchAndLeavesRest(t *testing.T) {
	root := t.TempDir()
	qdir := filepath.Join(root, "queue", "en", "Geography")
	if err := os.MkdirAll(qdir, 0o755); err != nil {
		t.Fatal(err)
	}
	src := "Q: One?\nA: 1.\n\n---\n\nQ: Two?\nA: 2.\n\n---\n\nQ: Three?\nA: 3.\n"
	if err := os.WriteFile(filepath.Join(qdir, "continents_early.md"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	msgs := Topic(root, "geography", repo.TopicConfig{Rate: 2}, 2, "en", false)
	if len(msgs) != 1 || !strings.Contains(msgs[0], "published 2 card") {
		t.Fatalf("msgs=%v", msgs)
	}
	published, err := os.ReadFile(filepath.Join(root, "cards", "en", "Geography", "continents_early.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(published), "Q: One?") || !strings.Contains(string(published), "Q: Two?") || strings.Contains(string(published), "Q: Three?") {
		t.Fatalf("published=%s", published)
	}
	rest, err := os.ReadFile(filepath.Join(qdir, "continents_early.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(rest), "Q: Three?") || strings.Contains(string(rest), "Q: One?") {
		t.Fatalf("rest=%s", rest)
	}
}
