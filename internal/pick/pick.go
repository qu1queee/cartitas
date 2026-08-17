package pick

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/qu1queee/cartitas/internal/repo"
)

type Target struct {
	Topic    string
	Subtopic string
	AgeBand  string
	Count    int
	Coverage int
}

func LangCoverage(root, topic, subtopic, band string) int {
	topicName := repo.TopicDirName(topic)
	filename := fmt.Sprintf("%s_%s.md", subtopic, band)
	count := 0
	for _, lang := range []string{"en", "es", "de"} {
		for _, base := range []string{"cards", "queue", "draft"} {
			p := filepath.Join(root, base, lang, topicName, filename)
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				count++
				break
			}
		}
	}
	return count
}

func TargetFrom(root string, cfg *repo.PublishConfig) (Target, error) {
	best := Target{}
	found := false
	bestCov, bestIdx := 999, 999

	for topicIdx, topicKey := range repo.TopicOrder {
		topicCfg, ok := cfg.Topics[topicKey]
		if !ok || !topicCfg.IsEnabled() || !topicCfg.GenerateEnabled() {
			continue
		}
		cards := topicCfg.Generate.Cards
		if cards == 0 {
			cards = cfg.Generate.Defaults.Cards
		}
		if cards == 0 {
			cards = 5
		}
		bands := topicCfg.AgeBands
		if len(bands) == 0 {
			bands = []string{"early"}
		}
		for _, subtopic := range topicCfg.Subtopics {
			for _, band := range bands {
				coverage := LangCoverage(root, topicKey, subtopic, band)
				if coverage < bestCov || (coverage == bestCov && topicIdx < bestIdx) {
					bestCov, bestIdx = coverage, topicIdx
					best = Target{
						Topic:    topicKey,
						Subtopic: subtopic,
						AgeBand:  band,
						Count:    cards,
						Coverage: coverage,
					}
					found = true
				}
			}
		}
	}
	if !found {
		return Target{}, fmt.Errorf("no generate target found in publish.yaml")
	}
	return best, nil
}
