package pick

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/qu1queee/cartitas/internal/repo"
)

func TestTargetPicksLowestCoverageFirstTopic(t *testing.T) {
	root := t.TempDir()
	enabled := true
	cfg := &repo.PublishConfig{
		Topics: map[string]repo.TopicConfig{
			"geography": {
				Enabled:   &enabled,
				Subtopics: []string{"continents"},
				AgeBands:  []string{"early"},
			},
			"space": {
				Enabled:   &enabled,
				Subtopics: []string{"moon"},
				AgeBands:  []string{"early"},
			},
		},
	}
	cfg.Topics["geography"] = withGen(cfg.Topics["geography"], true, 5)
	cfg.Topics["space"] = withGen(cfg.Topics["space"], true, 5)

	// geography already exists in all langs; space missing.
	for _, lang := range []string{"en", "es", "de"} {
		dir := filepath.Join(root, "cards", lang, "Geography")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "continents_early.md"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	got, err := TargetFrom(root, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if got.Topic != "space" || got.Subtopic != "moon" || got.Coverage != 0 {
		t.Fatalf("got %+v", got)
	}
}

func withGen(t repo.TopicConfig, enabled bool, cards int) repo.TopicConfig {
	t.Generate.Enabled = &enabled
	t.Generate.Cards = cards
	return t
}
